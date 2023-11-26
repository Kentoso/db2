package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func registerMongoHandlers(router *http.ServeMux, client *mongo.Client) {
	db := client.Database("summary")

	router.HandleFunc("/test/mongo/users/insert", func(w http.ResponseWriter, r *http.Request) {
		insertTestUserMongoHandler(w, r, db)
	})
	router.HandleFunc("/test/mongo/users/deleteAll", func(w http.ResponseWriter, r *http.Request) {
		deleteAllUsersDataMongoHandler(w, r, db)
	})
	router.HandleFunc("/test/mongo/users/update", func(w http.ResponseWriter, r *http.Request) {
		updateAllUsersPasswordsMongoHandler(w, r, db)
	})
	router.HandleFunc("/test/mongo/users/addNotification", func(w http.ResponseWriter, r *http.Request) {
		addNotificationForUsersMongoHandler(w, r, db)
	})
}

func addNotificationForUsersMongoHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	usersCollection := db.Collection("users")
	booksCollection := db.Collection("books")

	ctx := context.Background()
	start := time.Now()

	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		log.Printf("Error fetching users: %v\n", err)
		return
	}

	var users []bson.M
	if err = cursor.All(ctx, &users); err != nil {
		http.Error(w, "Error decoding user data", http.StatusInternalServerError)
		log.Printf("Error decoding user data: %v\n", err)
		return
	}

	for _, user := range users {
		readingHistories := user["user_book_reading_histories"].(bson.A)
		lastReading := readingHistories[len(readingHistories)-1].(bson.M)
		bookID := lastReading["book_id"].(primitive.ObjectID)

		var book bson.M
		err := booksCollection.FindOne(ctx, bson.M{"_id": bookID}).Decode(&book)
		if err != nil {
			http.Error(w, "Error fetching book details", http.StatusInternalServerError)
			log.Printf("Error fetching book details for book ID %v: %v\n", bookID, err)
			return
		}
		bookTitle := book["title"].(string)
		authorName := book["author"].(bson.M)["name"].(string)

		lastPageRead := lastReading["last_page_read"].(int32)
		message := fmt.Sprintf("Hey there! You last stopped at page %d in '%s' by %s. Can't wait to see you back!", lastPageRead, bookTitle, authorName)

		update := bson.M{
			"$push": bson.M{
				"notifications": bson.D{
					{"message", message},
					{"is_read", false},
					{"created_at", time.Now()},
				},
			},
		}
		_, err = usersCollection.UpdateOne(ctx, bson.M{"_id": user["_id"]}, update)
		if err != nil {
			http.Error(w, "Error adding notification to user", http.StatusInternalServerError)
			log.Printf("Error adding notification to user %v: %v\n", user["_id"], err)
			return
		}
	}

	duration := time.Since(start).Seconds()

	sendResponse(w, "MongoDB", duration, "ADD_NOTIFICATIONS")
}

func updateAllUsersPasswordsMongoHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()

	usersCollection := db.Collection("users")
	updatePipeline := mongo.Pipeline{
		{{"$set", bson.D{{"password_hash", bson.M{"$concat": []interface{}{"$password_hash", "1"}}}}}},
	}
	result, err := usersCollection.UpdateMany(context.Background(), bson.M{}, updatePipeline)
	if err != nil {
		http.Error(w, "Error updating user passwords", http.StatusInternalServerError)
		log.Printf("Error updating user passwords: %v\n", err)
		return
	}

	log.Printf("Password hashes updated, number of users affected: %v\n", result.ModifiedCount)

	duration := time.Since(start).Seconds()

	sendResponse(w, "MongoDB", duration, "UPDATE")
}

func deleteAllUsersDataMongoHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()

	usersCollection := db.Collection("users")

	if _, err := usersCollection.DeleteMany(context.Background(), bson.D{}); err != nil {
		http.Error(w, "Error deleting users", http.StatusInternalServerError)
		log.Printf("Error deleting users: %v\n", err)
		return
	}

	log.Println("All user data deleted successfully")
	duration := time.Since(start).Seconds()

	sendResponse(w, "MongoDB", duration, "DELETE")
}

func insertTestUserMongoHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UserInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Count <= 0 {
		http.Error(w, "Count must be a positive integer", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	usersCollection := db.Collection("users")
	booksCollection := db.Collection("books")

	start := time.Now()

	cursor, err := booksCollection.Find(ctx, bson.D{})
	if err != nil {
		log.Printf("Error fetching book IDs: %v\n", err)
		http.Error(w, "Error fetching book IDs", http.StatusInternalServerError)
		return
	}

	var books []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err = cursor.All(ctx, &books); err != nil {
		log.Printf("Error decoding book IDs: %v\n", err)
		http.Error(w, "Error decoding book IDs", http.StatusInternalServerError)
		return
	}

	bookIDs := make([]primitive.ObjectID, len(books))
	for i, book := range books {
		bookIDs[i] = book.ID
	}

	getThreeRandomIndexes := func(length int) []int {
		indexes := rand.Perm(length)
		return indexes[:3]
	}

	var userDocuments []interface{}
	for i := 0; i < req.Count; i++ {
		randomIndexes := getThreeRandomIndexes(len(bookIDs))
		userDocument := bson.D{
			{"name", fmt.Sprintf("Test User %d", i)},
			{"email", fmt.Sprintf("test%d@example.com", i)},
			{"password_hash", "hash"},
			{"created_at", time.Now()},
			{"updated_at", time.Now()},
			{"user_devices", bson.A{
				bson.D{{"device_type", "DeviceType1"}, {"device_token", "token1"}, {"registered_at", time.Now()}},
				bson.D{{"device_type", "DeviceType2"}, {"device_token", "token2"}, {"registered_at", time.Now()}},
				bson.D{{"device_type", "DeviceType3"}, {"device_token", "token3"}, {"registered_at", time.Now()}},
			}},
			{"user_book_reading_histories", bson.A{
				bson.D{{"book_id", bookIDs[randomIndexes[0]]}, {"device", 0}, {"start_timestamp", time.Now()}, {"end_timestamp", time.Now()}, {"last_page_read", rand.Intn(500)}},
				bson.D{{"book_id", bookIDs[randomIndexes[1]]}, {"device", 1}, {"start_timestamp", time.Now()}, {"end_timestamp", time.Now()}, {"last_page_read", rand.Intn(500)}},
				bson.D{{"book_id", bookIDs[randomIndexes[2]]}, {"device", 2}, {"start_timestamp", time.Now()}, {"end_timestamp", time.Now()}, {"last_page_read", rand.Intn(500)}},
			}},
			{"notifications", bson.A{}},
		}
		userDocuments = append(userDocuments, userDocument)
	}

	_, err = usersCollection.InsertMany(ctx, userDocuments)
	if err != nil {
		log.Printf("Error inserting users: %v\n", err)
		http.Error(w, "Error inserting users", http.StatusInternalServerError)
		return
	}

	duration := time.Since(start).Seconds()
	sendResponse(w, "MongoDB", duration, "CREATE")
}
