package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MissedPayments(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now().Truncate(24 * time.Hour)

		cursor, err := cfg.MongoClient.Database(cfg.DBName).
			Collection("riders").
			Find(ctx, bson.M{
				"user_id": userID,
				"$or": bson.A{
					bson.M{"last_payment_date": bson.M{"$lt": start}},
					bson.M{"last_payment_date": bson.M{"$exists": false}},
				},
			})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed fetching alerts"})
			return
		}

		var riders []bson.M
		cursor.All(ctx, &riders)

		c.JSON(http.StatusOK, riders)
	}
}