package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateRider(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		var input struct {
			FullName     string  `json:"full_name" binding:"required"`
			PhoneNumber  string  `json:"phone_number" binding:"required"`
			LicenseNo    string  `json:"license_no" binding:"required"`
			DailyTarget  float64 `json:"daily_target"`
			MotorcycleID string  `json:"motorcycle_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		db := cfg.MongoClient.Database(cfg.DBName)
		ridersCol := db.Collection("riders")
		motoCol := db.Collection("motorcycles")

		motorcycleID, err := primitive.ObjectIDFromHex(input.MotorcycleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid motorcycle id"})
			return
		}

		count, err := motoCol.CountDocuments(ctx, bson.M{
			"_id":     motorcycleID,
			"user_id": userID,
		})
		if err != nil || count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "motorcycle does not exist"})
			return
		}

		phoneCount, _ := ridersCol.CountDocuments(ctx, bson.M{
			"phone_number": input.PhoneNumber,
		})
		if phoneCount > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "phone number already exists"})
			return
		}

		licenseCount, _ := ridersCol.CountDocuments(ctx, bson.M{
			"license_no": input.LicenseNo,
		})
		if licenseCount > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "license number already exists"})
			return
		}

		bikeAssignedCount, _ := ridersCol.CountDocuments(ctx, bson.M{
			"motorcycle_id": motorcycleID,
		})
		if bikeAssignedCount > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "motorcycle already assigned"})
			return
		}

		rider := models.Rider{
			ID:              primitive.NewObjectID(),
			UserID:          userID,
			FullName:        input.FullName,
			PhoneNumber:     input.PhoneNumber,
			LicenseNo:       input.LicenseNo,
			DailyTarget:     input.DailyTarget,
			MotorcycleID:    motorcycleID,
			OutstandingDebt: 0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		_, err = ridersCol.InsertOne(ctx, rider)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed creating rider"})
			return
		}

		c.JSON(http.StatusCreated, rider)
	}
}

func ListRiders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("riders")

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"user_id": userID}}},
			{{
				Key: "$lookup",
				Value: bson.M{
					"from":         "motorcycles",
					"localField":   "motorcycle_id",
					"foreignField": "_id",
					"as":           "motorcycle",
				},
			}},
			{{
				Key: "$unwind",
				Value: bson.M{
					"path":                       "$motorcycle",
					"preserveNullAndEmptyArrays": true,
				},
			}},
		}

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed fetching riders"})
			return
		}

		var riders []bson.M
		cursor.All(ctx, &riders)

		c.JSON(http.StatusOK, riders)
	}
}

func GetRider(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		riderID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rider id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("riders")

		pipeline := mongo.Pipeline{
			{{
				Key: "$match",
				Value: bson.M{
					"_id":     riderID,
					"user_id": userID,
				},
			}},
			{{
				Key: "$lookup",
				Value: bson.M{
					"from":         "motorcycles",
					"localField":   "motorcycle_id",
					"foreignField": "_id",
					"as":           "motorcycle",
				},
			}},
			{{
				Key: "$unwind",
				Value: bson.M{
					"path":                       "$motorcycle",
					"preserveNullAndEmptyArrays": true,
				},
			}},
			{{
				Key: "$lookup",
				Value: bson.M{
					"from":         "payments",
					"localField":   "_id",
					"foreignField": "rider_id",
					"as":           "payments",
				},
			}},
		}

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed fetching rider"})
			return
		}

		var riders []bson.M
		cursor.All(ctx, &riders)

		if len(riders) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "rider not found"})
			return
		}

		c.JSON(http.StatusOK, riders[0])
	}
}

func UpdateRider(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		riderID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rider id"})
			return
		}

		var input struct {
			FullName     string  `json:"full_name"`
			PhoneNumber  string  `json:"phone_number"`
			DailyTarget  float64 `json:"daily_target"`
			MotorcycleID string  `json:"motorcycle_id"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("riders")

		update := bson.M{
			"updated_at": time.Now(),
		}

		if input.FullName != "" {
			update["full_name"] = input.FullName
		}

		if input.PhoneNumber != "" {
			update["phone_number"] = input.PhoneNumber
		}

		if input.DailyTarget > 0 {
			update["daily_target"] = input.DailyTarget
		}

		if input.MotorcycleID != "" {
			motorcycleID, err := primitive.ObjectIDFromHex(input.MotorcycleID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid motorcycle id"})
				return
			}

			count, _ := col.CountDocuments(ctx, bson.M{
				"motorcycle_id": motorcycleID,
				"_id":           bson.M{"$ne": riderID},
			})

			if count > 0 {
				c.JSON(http.StatusConflict, gin.H{
					"error": "motorcycle already assigned",
				})
				return
			}

			update["motorcycle_id"] = motorcycleID
		}

		res, err := col.UpdateOne(
			ctx,
			bson.M{
				"_id":     riderID,
				"user_id": userID,
			},
			bson.M{"$set": update},
		)

		if err != nil || res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "rider not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "rider updated"})
	}
}

func DeleteRider(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rider id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("riders")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		res, err := col.DeleteOne(ctx, bson.M{
			"_id":     id,
			"user_id": userID,
		})

		if err != nil || res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "rider not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "rider deleted"})
	}
}