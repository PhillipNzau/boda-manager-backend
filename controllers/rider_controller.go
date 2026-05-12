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

		ctx, cancel := context.WithTimeout(context.Background(), 38*time.Second)
		defer cancel()

		db := cfg.MongoClient.Database(cfg.DBName)
		ridersCol := db.Collection("riders")
		// ridesCol := db.Collection("rides")

		// // 1. Ensure motorcycle exists in rides
		// rideCount, err := ridesCol.CountDocuments(ctx, bson.M{
		// 	"user_id":       userID,
		// 	"motorcycle_id": input.MotorcycleID,
		// })
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed checking motorcycle"})
		// 	return
		// }

		// if rideCount == 0 {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error": "motorcycle does not exist in rides history",
		// 	})
		// 	return
		// }

		// 2. Unique phone number
		phoneCount, err := ridersCol.CountDocuments(ctx, bson.M{
			"phone_number": input.PhoneNumber,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed checking phone number"})
			return
		}

		if phoneCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "phone number already exists",
			})
			return
		}

		// 3. Unique license number
		licenseCount, err := ridersCol.CountDocuments(ctx, bson.M{
			"license_no": input.LicenseNo,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed checking license"})
			return
		}

		if licenseCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "license number already exists",
			})
			return
		}

		// 4. Optional: one rider per motorcycle
		bikeAssignedCount, err := ridersCol.CountDocuments(ctx, bson.M{
			"motorcycle_id": input.MotorcycleID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed checking motorcycle assignment"})
			return
		}

		if bikeAssignedCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "motorcycle already assigned to another rider",
			})
			return
		}

		motorcycleID, err := primitive.ObjectIDFromHex(input.MotorcycleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid motorcycle id"})
			return
		}

		motoCol := cfg.MongoClient.Database(cfg.DBName).Collection("motorcycles")

		count, err := motoCol.CountDocuments(ctx, bson.M{
			"_id":     motorcycleID,
			"user_id": userID,
		})
		if err != nil {
			log.Println("motorcycle count error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed checking motorcycle",
				"details": err.Error(),
			})
			return
		}

		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "motorcycle does not exist"})
			return
		}

		rider := models.Rider{
			ID:              primitive.NewObjectID(),
			UserID:          userID,
			FullName:        input.FullName,
			PhoneNumber:     input.PhoneNumber,
			LicenseNo:       input.LicenseNo,
			DailyTarget:     input.DailyTarget,
			MotorcycleID: 	 motorcycleID,
			OutstandingDebt: 0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		_, err = ridersCol.InsertOne(ctx, rider)
		if err != nil {
			log.Println("insert error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed creating rider",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, rider)
	}
}

func ListRiders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		col := cfg.MongoClient.Database(cfg.DBName).Collection("riders")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		cursor, err := col.Find(ctx, bson.M{"user_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch riders"})
			return
		}

		var riders []models.Rider
		cursor.All(ctx, &riders)

		c.JSON(http.StatusOK, riders)
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
			"_id": id,
			"user_id": userID,
		})

		if err != nil || res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "rider not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "rider deleted"})
	}
}