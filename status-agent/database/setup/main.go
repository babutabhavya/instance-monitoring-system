package setup

import (
	"context"
	"fmt"
	"log"
	"os"
	mongo "status-agent/database"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Document struct {
	Port        int        `bson:"port"`
	Host        string     `bson:"host"`
	Status      *string    `bson:"status"`
	LastUpdated *time.Time `bson:"last_updated"`
	Name        string     `bson:"name"`
	Region      string     `bson:"region"`
	LastActive  *time.Time `bson:"last_active"`
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
	asiaCount, _ := os.LookupEnv("ASIA_COUNT")
	europeCount, _ := os.LookupEnv("EUROPE_COUNT")
	usaCount, _ := os.LookupEnv("USA_COUNT")

	asiaCountInt, err := strconv.Atoi(asiaCount)
	if err != nil {
		log.Printf("Error parsing ASIA_COUNT, using default (2): %v", err)
		asiaCountInt = 2
	}

	europeCountInt, err := strconv.Atoi(europeCount)
	if err != nil {
		log.Printf("Error parsing EUROPE_COUNT, using default (2): %v", err)
		europeCountInt = 2
	}

	usaCountInt, err := strconv.Atoi(usaCount)
	if err != nil {
		log.Printf("Error parsing USA_COUNT, using default (2): %v", err)
		usaCountInt = 2
	}

	regionCounts := map[string]int{
		"asia":   asiaCountInt,
		"europe": europeCountInt,
		"usa":    usaCountInt,
	}

	var documents []interface{}

	for region, count := range regionCounts {
		for i := 1; i <= count; i++ {
			name := fmt.Sprintf("instance-to-be-monitored-%s-%d", region, i)
			host := fmt.Sprintf("instance-to-be-monitored-%s-%d-service.%s.svc.cluster.local", region, i, region)

			// Define the document structure
			doc := Document{
				Port:        8080,
				Host:        host,
				Status:      nil,
				LastUpdated: nil,
				Name:        name,
				Region:      region,
				LastActive:  nil,
			}

			filter := bson.D{{Key: "host", Value: host}}
			existingDoc, err := db.FindDocument(context.Background(), collection, filter)
			if existingDoc != nil || err == nil {
				log.Printf("Document with port %d already exists, skipping...\n", 8080)
				continue
			} else {
				documents = append(documents, doc)

			}
		}
	}
	if len(documents) > 0 {
		db.BulkInsert(context.Background(), collection, documents)

	}
}
