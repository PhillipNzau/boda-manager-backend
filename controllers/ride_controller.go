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

func CreateRide(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		// Removed MotorcycleID from input as it is now system-resolved
		var input struct {
			RiderID  string  `json:"rider_id" binding:"required"`
			Date     string  `json:"date"`
			Trips    int     `json:"trips"`
			Income   float64 `json:"income"`
			FuelCost float64 `json:"fuel_cost"`
			Expenses float64 `json:"expenses"`
			Notes    string  `json:"notes"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		riderID, err := primitive.ObjectIDFromHex(input.RiderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rider id"})
			return
		}

		// 1. Fetch the Rider from the database to get their assigned MotorcycleID
		riderCol := cfg.MongoClient.Database(cfg.DBName).Collection("riders")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var rider models.Rider // Ensure this model has a MotorcycleID field
		err = riderCol.FindOne(ctx, bson.M{"_id": riderID}).Decode(&rider)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "rider not found"})
			return
		}

		// Optional: Check if the rider actually has a motorcycle assigned
		if rider.MotorcycleID.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this rider has no motorcycle assigned"})
			return
		}

		// 2. Create the Ride using the MotorcycleID retrieved from the Rider document
		ride := models.Ride{
			ID:           primitive.NewObjectID(),
			UserID:       userID,
			RiderID:      riderID,
			MotorcycleID: rider.MotorcycleID, // Automatically assigned
			Date:         input.Date,
			Trips:        input.Trips,
			Income:       input.Income,
			FuelCost:     input.FuelCost,
			Expenses:     input.Expenses,
			Notes:        input.Notes,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// 3. Save the ride
		ridesCol := cfg.MongoClient.Database(cfg.DBName).Collection("rides")
		_, err = ridesCol.InsertOne(ctx, ride)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save ride"})
			return
		}

		c.JSON(http.StatusCreated, ride)
	}
}

func ListRides(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		col := cfg.MongoClient.Database(cfg.DBName).Collection("rides")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, err := col.Find(ctx, bson.M{"user_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rides"})
			return
		}

		var rides []models.Ride
		cursor.All(ctx, &rides)

		c.JSON(http.StatusOK, rides)
	}
}

func MotorcycleProfitability(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		bikeID := c.Param("id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db := cfg.MongoClient.Database(cfg.DBName)

		// expenses for bike
		expCursor, _ := db.Collection("expenses").Find(
			ctx,
			bson.M{
				"user_id": userID,
				"motorcycle_id": bikeID,
			},
		)

		type Expense struct {
			Amount float64 `bson:"amount"`
		}

		var expenses []Expense
		expCursor.All(ctx, &expenses)

		var totalExpenses float64
		for _, e := range expenses {
			totalExpenses += e.Amount
		}

		c.JSON(http.StatusOK, gin.H{
			"motorcycle_id": bikeID,
			"expenses": totalExpenses,
		})
	}
}