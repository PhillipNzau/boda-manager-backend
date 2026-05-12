package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// USERS
func EnsureUserIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("users")

	emailIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "email", Value: 1}},
		Options: options.Index().
			SetUnique(true),
	}

	_, err := col.Indexes().CreateOne(ctx, emailIdx)
	if err != nil {
		log.Printf("Could not create user indexes: %v", err)
		return
	}

	log.Println("User indexes ensured")
}

// RIDERS
func EnsureRiderIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("riders")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "phone_number", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "license_no", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Could not create rider indexes: %v", err)
		return
	}

	log.Println("Rider indexes ensured")
}

// MOTORCYCLES
func EnsureMotorcycleIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("motorcycles")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "plate_number", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	}

	_, err := col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Could not create motorcycle indexes: %v", err)
		return
	}

	log.Println("Motorcycle indexes ensured")
}

// RIDES
func EnsureRideIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("rides")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "motorcycle_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	
	}

	_, err := col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Could not create ride indexes: %v", err)
		return
	}

	log.Println("Ride indexes ensured")
}

// EXPENSES
func EnsureExpenseIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("expenses")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "motorcycle_id", Value: 1}},
		},
		
	}

	_, err := col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Could not create expense indexes: %v", err)
		return
	}

	log.Println("Expense indexes ensured")
}

// PAYOUTS
func EnsurePayoutIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("payouts")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		
	}

	_, err := col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Could not create payout indexes: %v", err)
		return
	}

	log.Println("Payout indexes ensured")
}

// ALL INDEXES
func EnsureAllIndexes(client *mongo.Client, dbName string) {
	EnsureUserIndexes(client, dbName)
	EnsureRiderIndexes(client, dbName)
	EnsureMotorcycleIndexes(client, dbName)
	EnsureRideIndexes(client, dbName)
	EnsureExpenseIndexes(client, dbName)
	EnsurePayoutIndexes(client, dbName)
}