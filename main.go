package main

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

type Response struct {
	Operation string  `json:"operation"`
	Database  string  `json:"database"`
	Duration  float64 `json:"duration"`
}

func main() {
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:notacopyofheadway@localhost:5432/summary")
	if err != nil {
		log.Fatalf("Unable to connect to postgres database: %v\n", err)
	}
	defer pool.Close()
	log.Print("Connected to Postgres")

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Unable to connect to MongoDB: %v\n", err)
	}
	defer mongoClient.Disconnect(context.Background())
	log.Print("Connected to Mongo")

	router := http.NewServeMux()

	registerPostgresHandlers(router, pool)
	registerMongoHandlers(router, mongoClient)

	log.Print("Starting server on :4000")
	log.Fatal(http.ListenAndServe("127.0.0.1:4000", router))
}

func sendResponse(w http.ResponseWriter, database string, duration float64, operation string) {
	response := Response{
		Operation: operation,
		Database:  database,
		Duration:  duration,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
