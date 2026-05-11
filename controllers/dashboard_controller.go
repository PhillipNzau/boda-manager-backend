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

func DashboardSummary(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db := cfg.MongoClient.Database(cfg.DBName)

		ridersCol := db.Collection("riders")
		paymentsCol := db.Collection("payments")

		// total riders
		totalRiders, _ := ridersCol.CountDocuments(ctx, bson.M{"user_id": userID})

		// today's payments
		start := time.Now().Truncate(24 * time.Hour)
		end := start.Add(24 * time.Hour)

		cursor, _ := paymentsCol.Find(ctx, bson.M{
			"user_id": userID,
			"date": bson.M{
				"$gte": start,
				"$lt":  end,
			},
		})

		type Payment struct {
			Amount float64 `bson:"amount"`
		}

		var payments []Payment
		cursor.All(ctx, &payments)

		var todayTotal float64
		for _, p := range payments {
			todayTotal += p.Amount
		}

		// outstanding debt
		riderCursor, _ := ridersCol.Find(ctx, bson.M{"user_id": userID})

		type Rider struct {
			OutstandingDebt float64 `bson:"outstanding_debt"`
		}

		var riders []Rider
		riderCursor.All(ctx, &riders)

		var totalDebt float64
		for _, r := range riders {
			totalDebt += r.OutstandingDebt
		}

		c.JSON(http.StatusOK, gin.H{
			"total_riders": totalRiders,
			"today_collected": todayTotal,
			"outstanding_debt": totalDebt,
		})
	}
}