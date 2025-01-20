package setup

import (
	"context"
	"log"
	mongo "status-agent/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Document struct {
	Port        int        `bson:"port"`
	Host        string     `bson:"host"`
	Status      *string    `bson:"status"`
	LastChecked *time.Time `bson:"last_checked"`
}

func Setup(db *mongo.MongoDB) {
	dbExists, err := db.CheckDatabaseExists(context.Background(), "instances")
	if err != nil {
		log.Fatalf("Error checking if database exists: %v", err)
	}

	if !dbExists {
		log.Println("Database 'instances' does not exist. Will create the database...")

	} else {
		log.Println("Database 'instances' already exists.")
	}

	collectionExists, err := db.CheckCollectionExists(context.Background(), "instances", "status")
	if err != nil {
		log.Fatalf("Error checking if collection exists: %v", err)
	}

	if !collectionExists {
		log.Println("Collection 'status' does not exist. Creating the collection...")
		if err := db.CreateCollection(context.Background(), "instances", "status"); err != nil {
			log.Fatalf("Failed to create collection: %v", err)
		}
	}

	collection := db.Client.Database("instances").Collection("status")
	var documents []interface{}
	for port := 4001; port <= 4009; port++ {
		filter := bson.M{"port": port}
		var existingDoc Document
		err := collection.FindOne(context.Background(), filter).Decode(&existingDoc)
		if err == nil {
			log.Printf("Document with port %d already exists, skipping...\n", port)
			continue
		}

		var status *string
		var lastChecked *time.Time

		doc := Document{
			Port:        port,
			Host:        "localhost",
			Status:      status,
			LastChecked: lastChecked,
		}

		documents = append(documents, doc)
	}

	if len(documents) > 0 {
		for _, doc := range documents {
			if err := db.InsertDocument(context.Background(), collection, doc); err != nil {
				log.Fatalf("Failed to insert document: %v", err)
			}
		}
	} else {
		log.Println("No new documents to insert.")
	}
}
