package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"web-crawler-go/internal/core/domain"
	"web-crawler-go/internal/core/ports"
)

// MongoDBRepository implements the ProductRepository interface
type MongoDBRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
	logger     ports.Logger
}

// NewMongoDBRepository creates a new MongoDB repository
func NewMongoDBRepository(ctx context.Context, connectionURI, dbName, collectionName string, logger ports.Logger) (*MongoDBRepository, error) {
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if connectionURI == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectionURI).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the primary to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get a handle to the specified database and collection
	collection := client.Database(dbName).Collection(collectionName)

	logger.Info("connected to MongoDB", "database", dbName, "collection", collectionName)

	return &MongoDBRepository{
		client:     client,
		collection: collection,
		logger:     logger,
	}, nil
}

// UpsertProduct saves a product to MongoDB
func (m *MongoDBRepository) UpsertProduct(ctx context.Context, product *domain.Product) error {
	m.logger.Info("upserting product to MongoDB", "name", product.Name)

	type Document struct {
		Domain string         `bson:"domain"`
		Data   domain.Product `bson:"data"`
	}

	document := Document{
		Domain: "attfrench.cross-right.tw",
		Data:   *product,
	}

	// Define the filter to find existing document
	filter := map[string]interface{}{
		"domain":    document.Domain,
		"data.name": product.Name,
	}

	// Set upsert option to true
	opts := options.Replace().SetUpsert(true)

	// Upsert the product into the collection
	result, err := m.collection.ReplaceOne(ctx, filter, document, opts)
	if err != nil {
		m.logger.Error("failed to upsert product to MongoDB", "error", err)
		return fmt.Errorf("failed to upsert product to MongoDB: %w", err)
	}

	if result.UpsertedCount > 0 {
		m.logger.Info("product inserted to MongoDB", "name", product.Name, "upsertedID", result.UpsertedID)
	} else {
		m.logger.Info("product updated in MongoDB", "name", product.Name, "modifiedCount", result.ModifiedCount)
	}

	return nil
}

// Close closes the MongoDB connection
func (m *MongoDBRepository) Close(ctx context.Context) error {
	m.logger.Info("closing MongoDB connection")
	if err := m.client.Disconnect(ctx); err != nil {
		m.logger.Error("failed to disconnect from MongoDB", "error", err)
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	return nil
}
