package config

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoClient *mongo.Client
	DBName      string
	JWTSecret   []byte
	AESKey      []byte
}

func LoadConfig() (*Config, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "boda_manager"
	}

	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		return nil, errors.New("JWT_SECRET required")
	}

	aes := os.Getenv("AES_KEY")
	if len(aes) != 32 {
		return nil, errors.New("AES_KEY must be exactly 32 bytes")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	cfg := &Config{
		MongoClient: client,
		DBName:      dbName,
		JWTSecret:   []byte(jwt),
		AESKey:      []byte(aes),
	}

	if err := ensureIndexes(cfg); err != nil {
		log.Printf("index creation error: %v", err)
	}

	return cfg, nil
}

func ensureIndexes(cfg *Config) error {
	db := cfg.MongoClient.Database(cfg.DBName)
	ctx := context.Background()

	// USERS
	users := db.Collection("users")
	_, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "email", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// RIDERS
	riders := db.Collection("riders")
	_, err = riders.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "national_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return err
	}

	// MOTORCYCLES
	motorcycles := db.Collection("motorcycles")
	_, err = motorcycles.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "plate_number", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "assigned_rider_id", Value: 1}},
		},
	})
	if err != nil {
		return err
	}

	// RIDES
	rides := db.Collection("rides")
	_, err = rides.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "motorcycle_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "rider_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "ride_date", Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	// EXPENSES
	expenses := db.Collection("expenses")
	_, err = expenses.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "motorcycle_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "expense_date", Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	// PAYOUTS
	payouts := db.Collection("payouts")
	_, err = payouts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "rider_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "payout_date", Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	return nil
}