package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureHubIndexes creates indexes for the hubs collection
func EnsureHubIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("hubs")

	userIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	typeIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "type", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{userIdx, typeIdx})
	if err != nil {
		log.Printf("⚠️ Could not create hub indexes: %v", err)
	} else {
		log.Println("✅ Hub indexes ensured")
	}
}

// EnsureCategoryIndexes creates indexes for the categories collection
func EnsureCategoryIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("categories")

	userIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	nameIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{userIdx, nameIdx})
	if err != nil {
		log.Printf("⚠️ Could not create category indexes: %v", err)
	} else {
		log.Println("✅ Category indexes ensured")
	}
}

// EnsureCredentialIndexes creates indexes for the credentials collection
func EnsureCredentialIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("credentials")

	userIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	siteIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "site", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	uniqueUserSiteIdx := mongo.IndexModel{
    Keys: bson.D{
        {Key: "user_id", Value: 1},
        {Key: "site_name", Value: 1},
    },
    Options: options.Index().SetUnique(true).SetBackground(true),
}

	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{userIdx, siteIdx, uniqueUserSiteIdx})
	if err != nil {
		log.Printf("⚠️ Could not create credential indexes: %v", err)
	} else {
		log.Println("✅ Credential indexes ensured")
	}
}

// EnsureSubscriptionIndexes creates indexes for the subscriptions collection
func EnsureSubscriptionIndexes(client *mongo.Client, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := client.Database(dbName).Collection("subscriptions")

	userIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	serviceIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "service_name", Value: 1}},
		Options: options.Index().SetBackground(true),
	}

	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{userIdx, serviceIdx})
	if err != nil {
		log.Printf("⚠️ Could not create subscription indexes: %v", err)
	} else {
		log.Println("✅ Subscription indexes ensured")
	}
}

// EnsureAllIndexes creates indexes for all collections
func EnsureAllIndexes(client *mongo.Client, dbName string) {
	EnsureHubIndexes(client, dbName)
	EnsureCategoryIndexes(client, dbName)
	EnsureCredentialIndexes(client, dbName)
	EnsureSubscriptionIndexes(client, dbName)
}
