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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, _ := col.Find(ctx, bson.M{"user_id": userID})

		var motos []models.Motorcycle
		cursor.All(ctx, &motos)

		c.JSON(http.StatusOK, motos)
	}
}