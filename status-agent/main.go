package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	mongo "status-agent/database"
	setup "status-agent/database/setup"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Status        string    `json:"status"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	ID            string    `json:"id"`
}

func serializeInstance(instance bson.D) []byte {
	docMap := make(map[string]interface{})

	for _, elem := range instance {
		if oid, ok := elem.Value.(primitive.ObjectID); ok {
			docMap[elem.Key] = oid.Hex()
		} else {
			docMap[elem.Key] = elem.Value
		}
	}
	payload, err := json.Marshal(docMap)

	if err != nil {
		log.Printf("Error marshaling instance: %v", err)
		return nil
	}
	return payload
}

func publishNATSRequest(payload []byte, natsConn *nats.Conn) {
	err := natsConn.PublishRequest("activity-status.request", "status-queue-group", payload)
	if err != nil {
		log.Printf("Error publishing message %v to NATS: %v", string(payload), err)
	} else {
		log.Printf("Instance published to activity-status.request: %v", string(payload))
	}

}

func publishStatusRequests(instances []interface{}, natsConn *nats.Conn) {
	for _, instance := range instances {
		payload := serializeInstance(instance.(bson.D))
		publishNATSRequest(payload, natsConn)
	}
}

func retrieveInstances(db *mongo.MongoDB) []interface{} {
	log.Println("Getting all instances...")

	collection := db.Client.Database("instances").Collection("status")
	instances, err := db.FindAll(context.Background(), collection)
	if err != nil {
		log.Fatalf("Error finding instances: %v", err)
	}
	log.Println("Successfully retrieved all instances. Total number of instances:", len(instances))
	return instances
}

func task(db *mongo.MongoDB, natsConn *nats.Conn) {
	fmt.Println("Task started running at:", time.Now())
	instances := retrieveInstances(db)
	publishStatusRequests(instances, natsConn)
	log.Println("Task completed.")
}

func handleInstanceActivityResponse(msg *nats.Msg, db *mongo.MongoDB) {
	var message Message

	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	log.Printf("Received message: %+v", message)

	collection := db.Client.Database("instances").Collection("status")

	id, err := primitive.ObjectIDFromHex(message.ID)
	if err != nil {
		panic(err)
	}

	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: message.Status},
			{Key: "last_updated", Value: time.Now()},
		}},
	}

	if message.Status == "ACTIVE" {
		update = append(update, bson.E{Key: "$set", Value: bson.D{
			{Key: "last_active", Value: time.Now()},
		}})
	}

	log.Printf("Update instance payload: %+v", update)

	db.UpdateDocument(context.Background(), collection, filter, update)
}

func main() {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	db := mongo.NewMongoDB("mongodb://mongodb:27017", "root", "changeme")

	cancel, err := db.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer cancel()

	log.Println("Setting up database.")
	setup.Setup(db)

	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	log.Println("Connected to NATS.")
	defer nc.Close()

	sub, err := nc.QueueSubscribe("activity-status.response", "status-queue-group", func(msg *nats.Msg) {
		handleInstanceActivityResponse(msg, db)
	})

	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	go func() {
		for {
			task(db, nc)
			time.Sleep(5 * time.Minute)
		}
	}()

	sig := <-sigChannel
	fmt.Println("\nReceived signal:", sig)
	fmt.Println("Shutting down gracefully...")

	fmt.Println("Exiting...")
}
