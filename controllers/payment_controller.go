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

func CreatePayment(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		var input struct {
			RiderID string  `json:"rider_id" binding:"required"`
			Amount  float64 `json:"amount" binding:"required"`
			Method  string  `json:"method"`
			Notes   string  `json:"notes"`
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ridersCol := cfg.MongoClient.Database(cfg.DBName).Collection("riders")
		paymentsCol := cfg.MongoClient.Database(cfg.DBName).Collection("payments")

		// Get rider
		var rider models.Rider
		err = ridersCol.FindOne(ctx, bson.M{
			"_id": riderID,
			"user_id": userID,
		}).Decode(&rider)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "rider not found"})
			return
		}

		expected := rider.DailyTarget
		balance := expected - input.Amount

		status := "paid"
		if input.Amount < expected {
			status = "partial"
		}
		if input.Amount > expected {
			status = "overpaid"
		}

		now := time.Now()

		payment := models.Payment{
			ID:             primitive.NewObjectID(),
			UserID:         userID,
			RiderID:        riderID,
			Amount:         input.Amount,
			Method:         input.Method,
			Date:           now,
			ExpectedAmount: expected,
			Balance:        balance,
			Status:         status,
			Notes:          input.Notes,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		_, err = paymentsCol.InsertOne(ctx, payment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save payment"})
			return
		}

		// Update rider debt
		newDebt := rider.OutstandingDebt + balance

		_, err = ridersCol.UpdateOne(
			ctx,
			bson.M{"_id": riderID},
			bson.M{
				"$set": bson.M{
					"outstanding_debt": newDebt,
					"last_payment_date": now,
					"updated_at": now,
				},
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed updating rider debt"})
			return
		}

		c.JSON(http.StatusCreated, payment)
	}
}

// ListPayments - fetch all payments for logged-in user
func ListPayments(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("payments")

		filter := bson.M{"user_id": userID}

		// optional filters
		if riderID := c.Query("rider_id"); riderID != "" {
			oid, err := primitive.ObjectIDFromHex(riderID)
			if err == nil {
				filter["rider_id"] = oid
			}
		}

		if status := c.Query("status"); status != "" {
			filter["status"] = status
		}

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch payments"})
			return
		}

		var payments []models.Payment
		if err := cursor.All(ctx, &payments); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode payments"})
			return
		}

		c.JSON(http.StatusOK, payments)
	}
}

// GetPayment - fetch single payment
func GetPayment(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		paymentID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("payments")

		var payment models.Payment
		err = col.FindOne(ctx, bson.M{
			"_id": paymentID,
			"user_id": userID,
		}).Decode(&payment)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		c.JSON(http.StatusOK, payment)
	}
}

// DeletePayment - delete payment
func DeletePayment(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		paymentID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		col := cfg.MongoClient.Database(cfg.DBName).Collection("payments")

		res, err := col.DeleteOne(ctx, bson.M{
			"_id": paymentID,
			"user_id": userID,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete payment"})
			return
		}

		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "payment deleted",
			"id":      paymentID.Hex(),
		})
	}
}