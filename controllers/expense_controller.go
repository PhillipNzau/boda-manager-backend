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

func CreateExpense(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		var input struct {
			MotorcycleID string  `json:"motorcycle_id" binding:"required"`
			Category     string  `json:"category" binding:"required"`
			Description  string  `json:"description"`
			Amount       float64 `json:"amount" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}


		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		// ✅ FIX: convert motorcycle ID properly
		motorcycleID, err := primitive.ObjectIDFromHex(input.MotorcycleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid motorcycle id"})
			return
		}
		col := cfg.MongoClient.Database(cfg.DBName).Collection("motorcycles")

		count, _ := col.CountDocuments(ctx, bson.M{
			"_id":    motorcycleID,
			"user_id": userID,
		})

		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "motorcycle not found"})
			return
		}


		now := time.Now()

		expense := models.Expense{
			ID:           primitive.NewObjectID(),
			UserID:       userID,
			MotorcycleID: motorcycleID, // ✅ FIXED HERE
			Category:     input.Category,
			Description:  input.Description,
			Amount:       input.Amount,
			Date:         now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}


		_, err = cfg.MongoClient.Database(cfg.DBName).
			Collection("expenses").
			InsertOne(ctx, expense)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create expense"})
			return
		}

		c.JSON(http.StatusCreated, expense)
	}
}

func ListExpenses(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, err := cfg.MongoClient.Database(cfg.DBName).
			Collection("expenses").
			Find(ctx, bson.M{"user_id": userID})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch expenses"})
			return
		}

		var expenses []models.Expense
		cursor.All(ctx, &expenses)

		c.JSON(http.StatusOK, expenses)
	}
}