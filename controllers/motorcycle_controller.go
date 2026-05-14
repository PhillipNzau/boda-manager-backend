package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateMotorcycle(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		var input struct {
			Name         string `json:"name" binding:"required"`
			PlateNumber  string `json:"plate_number" binding:"required"`
			Model        string `json:"model"`
			EngineNumber string `json:"engine_number"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("motorcycles")

		count, _ := col.CountDocuments(ctx, bson.M{
			"user_id":      userID,
			"plate_number": input.PlateNumber,
		})

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "motorcycle already exists"})
			return
		}

		m := models.Motorcycle{
			ID:           primitive.NewObjectID(),
			UserID:       userID,
			Name:         input.Name,
			PlateNumber:  input.PlateNumber,
			Model:        input.Model,
			EngineNumber: input.EngineNumber,
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		_, err := col.InsertOne(ctx, m)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create motorcycle"})
			return
		}

		c.JSON(http.StatusCreated, m)
	}
}

func ListMotorcycles(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		col := cfg.MongoClient.Database(cfg.DBName).Collection("motorcycles")

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		cursor, _ := col.Find(ctx, bson.M{"user_id": userID})

		var motos []models.Motorcycle
		cursor.All(ctx, &motos)

		c.JSON(http.StatusOK, motos)
	}
}

func UpdateMotorcycle(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		motorcycleID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid motorcycle id"})
			return
		}

		var input struct {
			Name         string `json:"name"`
			PlateNumber  string `json:"plate_number"`
			Model        string `json:"model"`
			EngineNumber string `json:"engine_number"`
			Status       string `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("motorcycles")

		update := bson.M{
			"updated_at": time.Now(),
		}

		if input.Name != "" {
			update["name"] = input.Name
		}

		if input.Model != "" {
			update["model"] = input.Model
		}

		if input.EngineNumber != "" {
			update["engine_number"] = input.EngineNumber
		}

		if input.Status != "" {
			update["status"] = input.Status
		}

		if input.PlateNumber != "" {
			count, err := col.CountDocuments(ctx, bson.M{
				"user_id":      userID,
				"plate_number": input.PlateNumber,
				"_id":          bson.M{"$ne": motorcycleID},
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed checking plate number"})
				return
			}

			if count > 0 {
				c.JSON(http.StatusConflict, gin.H{
					"error": "plate number already exists",
				})
				return
			}

			update["plate_number"] = input.PlateNumber
		}

		res, err := col.UpdateOne(
			ctx,
			bson.M{
				"_id":     motorcycleID,
				"user_id": userID,
			},
			bson.M{
				"$set": update,
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed updating motorcycle"})
			return
		}

		if res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "motorcycle not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "motorcycle updated"})
	}
}

func DeleteMotorcycle(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		motorcycleID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid motorcycle id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		db := cfg.MongoClient.Database(cfg.DBName)

		// prevent deleting assigned motorcycle
		ridersCol := db.Collection("riders")
		assignedCount, err := ridersCol.CountDocuments(ctx, bson.M{
			"user_id":       userID,
			"motorcycle_id": motorcycleID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed checking motorcycle assignments",
			})
			return
		}

		if assignedCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "cannot delete motorcycle assigned to rider",
			})
			return
		}

		col := db.Collection("motorcycles")

		res, err := col.DeleteOne(ctx, bson.M{
			"_id":     motorcycleID,
			"user_id": userID,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed deleting motorcycle"})
			return
		}

		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "motorcycle not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "motorcycle deleted",
			"id":      motorcycleID.Hex(),
		})
	}
}