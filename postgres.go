package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type UserInsertRequest struct {
	Count int `json:"count"`
}

func registerPostgresHandlers(router *http.ServeMux, pool *pgxpool.Pool) {
	router.HandleFunc("/test/postgres/users/insert", func(w http.ResponseWriter, r *http.Request) {
		insertTestUserPostgresHandler(w, r, pool)
	})
	router.HandleFunc("/test/postgres/users/deleteAll", func(w http.ResponseWriter, r *http.Request) {
		deleteAllUsersDataPostgresHandler(w, r, pool)
	})
	router.HandleFunc("/test/postgres/users/update", func(w http.ResponseWriter, r *http.Request) {
		updateAllUsersPasswordsPostgresHandler(w, r, pool)
	})
	router.HandleFunc("/test/postgres/users/addNotification", func(w http.ResponseWriter, r *http.Request) {
		addNotificationForUsersPostgresHandler(w, r, pool)
	})
}

func addNotificationForUsersPostgresHandler(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	start := time.Now()

	ctx := context.Background()
	commandTag, err := pool.Exec(ctx, `
		WITH LatestReadingHistory AS (
		    SELECT user_id, book_id, device_id, last_page_read,
		           ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) as rn
		    FROM user_book_reading_histories
		)
		INSERT INTO notifications (user_id, message, is_read, created_at)
		SELECT u.id AS user_id,
		       'Hey there! You last stopped at page ' || lrh.last_page_read || 
		       ' in ''' || b.title || ''' by ' || a.name || 
		       ', a ' || string_agg(g.name, ', ') || ' genre book, on your ' || d.device_type || ' device. Can''t wait to see you back!' AS message,
		       FALSE,
		       NOW()
		FROM users u
		JOIN LatestReadingHistory lrh ON u.id = lrh.user_id AND lrh.rn = 1
		JOIN books b ON lrh.book_id = b.id
		JOIN authors a ON b.author_id = a.id
		JOIN book_genres bg ON b.id = bg.book_id
		JOIN genres g ON bg.genre_id = g.id
		JOIN user_devices d ON lrh.device_id = d.id
		GROUP BY u.id, lrh.last_page_read, b.title, a.name, d.device_type;
	`)
	if err != nil {
		http.Error(w, "Error generating notifications", http.StatusInternalServerError)
		log.Printf("Error generating notifications: %v\n", err)
		return
	}

	log.Printf("Notifications generated, rows affected: %v\n", commandTag.RowsAffected())

	duration := time.Since(start).Seconds()
	sendResponse(w, "PostgreSQL", duration, "ADD_NOTIFICATIONS")
}

func updateAllUsersPasswordsPostgresHandler(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	start := time.Now()

	commandTag, err := pool.Exec(context.Background(), "UPDATE users SET password_hash = password_hash || '1'")
	if err != nil {
		http.Error(w, "Error updating user passwords", http.StatusInternalServerError)
		log.Printf("Error updating user passwords: %v\n", err)
		return
	}

	log.Printf("Password hashes updated, rows affected: %v\n", commandTag.RowsAffected())
	duration := time.Since(start).Seconds()
	sendResponse(w, "PostgreSQL", duration, "UPDATE")
}

func deleteAllUsersDataPostgresHandler(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	start := time.Now()
	tx, err := pool.Begin(context.Background())
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	tables := []string{"user_book_reading_histories", "user_devices", "notifications"}
	for _, table := range tables {
		if _, err := tx.Exec(context.Background(), "DELETE FROM "+table); err != nil {
			http.Error(w, "Error deleting data from table "+table, http.StatusInternalServerError)
			tx.Rollback(context.Background())
			return
		}
	}

	if _, err := tx.Exec(context.Background(), "DELETE FROM users"); err != nil {
		http.Error(w, "Error deleting users", http.StatusInternalServerError)
		tx.Rollback(context.Background())
		return
	}

	if err := tx.Commit(context.Background()); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	log.Println("All user data deleted successfully")
	duration := time.Since(start).Seconds()
	sendResponse(w, "PostgreSQL", duration, "DELETE")
}

func insertTestUserPostgresHandler(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
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

	start := time.Now()
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	for i := 0; i < req.Count; i++ {
		userID, err := insertUser(ctx, tx)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Error inserting user", http.StatusInternalServerError)
			return
		}

		deviceIDs, err := insertRandomDevices(ctx, tx, userID, 3)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Error inserting devices", http.StatusInternalServerError)
			return
		}

		err = insertUserBookReadingHistories(ctx, tx, userID, 3, deviceIDs)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Error inserting user book reading histories", http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	duration := time.Since(start).Seconds()
	sendResponse(w, "PostgreSQL", duration, "CREATE")
}

func insertUser(ctx context.Context, tx pgx.Tx) (int, error) {
	var userID int
	err := tx.QueryRow(ctx, `INSERT INTO users (name, email, password_hash, created_at, updated_at) VALUES ('Test User', 'test@example.com', 'hash', NOW(), NOW()) RETURNING id`).Scan(&userID)
	return userID, err
}

func insertRandomDevices(ctx context.Context, tx pgx.Tx, userID int, count int) ([]int, error) {
	var deviceIDs []int
	for i := 0; i < count; i++ {
		deviceType := fmt.Sprintf("DeviceType%d", i)
		var deviceID int
		err := tx.QueryRow(ctx, `INSERT INTO user_devices (user_id, device_type, device_token, registered_at) VALUES ($1, $2, 'token', NOW()) RETURNING id`, userID, deviceType).Scan(&deviceID)
		if err != nil {
			return nil, err
		}
		deviceIDs = append(deviceIDs, deviceID)
	}
	return deviceIDs, nil
}

func insertUserBookReadingHistories(ctx context.Context, tx pgx.Tx, userID int, count int, deviceIds []int) error {
	rows, err := tx.Query(ctx, `SELECT id FROM books ORDER BY random() LIMIT $1`, count)
	if err != nil {
		return err
	}

	var bookIDs []int
	for rows.Next() {
		var bookID int
		if err := rows.Scan(&bookID); err != nil {
			rows.Close()
			return err
		}
		bookIDs = append(bookIDs, bookID)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}

	rows.Close()

	for _, bookID := range bookIDs {
		for _, deviceId := range deviceIds {
			_, err := tx.Exec(ctx, `INSERT INTO user_book_reading_histories (user_id, book_id, start_timestamp, end_timestamp, last_page_read, device_id, created_at) VALUES ($1, $2, NOW(), NOW(), $3, $4, NOW())`, userID, bookID, rand.Intn(500), deviceId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
