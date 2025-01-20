package mongo

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database interface {
	Connect(ctx context.Context) (*mongo.Client, context.CancelFunc, error)
	CheckDatabaseExists(ctx context.Context, dbName string) (bool, error)
	CheckCollectionExists(ctx context.Context, dbName, collectionName string) (bool, error)
	CreateDatabase(ctx context.Context, dbName string) error
	CreateCollection(ctx context.Context, dbName, collectionName string) error
	InsertDocument(ctx context.Context, collection *mongo.Collection, doc interface{}) error
	FindAllDocuments(ctx context.Context, collection *mongo.Collection) ([]interface{}, error)
	FindDocument(ctx context.Context, collection *mongo.Collection, filter bson.D) (interface{}, error)
	UpdateDocument(ctx context.Context, collection *mongo.Collection, filter, update bson.D) (*mongo.UpdateResult, error)
}

type MongoDB struct {
	URI        string
	Username   string
	Password   string
	Client     *mongo.Client
	Context    context.Context
	CancelFunc context.CancelFunc
}

func NewMongoDB(uri, username, password string) *MongoDB {
	return &MongoDB{
		URI:      uri,
		Username: username,
		Password: password,
	}
}

func (db *MongoDB) Connect(ctx context.Context) (context.CancelFunc, error) {
	log.Println("Connecting to MongoDB...")
	client, cancel := context.WithTimeout(ctx, 10*time.Second)
	credentials := options.Credential{
		Username: db.Username,
		Password: db.Password,
	}
	clientInstance, err := mongo.Connect(client, options.Client().ApplyURI(db.URI).SetAuth(credentials))
	if err != nil {
		return cancel, err
	}
	db.Client = clientInstance

	// Ping MongoDB to verify connection
	err = clientInstance.Ping(client, nil)
	if err != nil {
		cancel()
		return cancel, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB.")
	return cancel, nil
}

func (db *MongoDB) CheckDatabaseExists(ctx context.Context, dbName string) (bool, error) {
	log.Printf("Checking if database '%s' exists...", dbName)
	databases, err := db.Client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return false, err
	}
	for _, database := range databases {
		if database == dbName {
			return true, nil
		}
	}
	return false, nil
}

func (db *MongoDB) CheckCollectionExists(ctx context.Context, dbName, collectionName string) (bool, error) {
	collections, err := db.Client.Database(dbName).ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return false, err
	}
	for _, col := range collections {
		if col == collectionName {
			return true, nil
		}
	}
	return false, nil
}

func (db *MongoDB) CreateCollection(ctx context.Context, dbName, collectionName string) error {
	collection := db.Client.Database(dbName).Collection(collectionName)
	if err := collection.Database().CreateCollection(ctx, collectionName); err != nil {
		return err
	}
	log.Printf("Collection '%s' created successfully.", collectionName)
	return nil
}

func (db *MongoDB) InsertDocument(ctx context.Context, collection *mongo.Collection, doc interface{}) error {
	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	log.Printf("Document inserted successfully.")
	return nil
}

func (db *MongoDB) FindAll(ctx context.Context, collection *mongo.Collection) ([]interface{}, error) {
	var results []interface{}
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var document interface{}
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		results = append(results, document)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *MongoDB) FindDocument(ctx context.Context, collection *mongo.Collection, filter bson.D) (interface{}, error) {
	var result interface{}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no document found matching filter: %w", err)
		}
		return nil, err
	}
	log.Printf("Document found: %+v", result)
	return result, nil
}

func (db *MongoDB) UpdateDocument(ctx context.Context, collection *mongo.Collection, filter, update bson.D) (*mongo.UpdateResult, error) {
	log.Printf("Updating document with filter: %+v", filter)
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("no document matched the filter for update")
	}
	log.Printf("Document updated successfully. Matched %d, Modified %d", result.MatchedCount, result.ModifiedCount)
	return result, nil
}
