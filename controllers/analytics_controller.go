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

func MonthlyAnalytics(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db := cfg.MongoClient.Database(cfg.DBName)

		start := time.Date(
			time.Now().Year(),
			time.Now().Month(),
			1,
			0, 0, 0, 0,
			time.UTC,
		)

		end := start.AddDate(0, 1, 0)

		// PAYMENTS
		paymentCursor, _ := db.Collection("payments").Find(
			ctx,
			bson.M{
				"user_id": userID,
				"date": bson.M{
					"$gte": start,
					"$lt":  end,
				},
			},
		)

		type Payment struct {
			Amount float64 `bson:"amount"`
		}

		var payments []Payment
		paymentCursor.All(ctx, &payments)

		var totalRevenue float64
		for _, p := range payments {
			totalRevenue += p.Amount
		}

		// EXPENSES
		expenseCursor, _ := db.Collection("expenses").Find(
			ctx,
			bson.M{
				"user_id": userID,
				"date": bson.M{
					"$gte": start,
					"$lt":  end,
				},
			},
		)

		type Expense struct {
			Amount float64 `bson:"amount"`
		}

		var expenses []Expense
		expenseCursor.All(ctx, &expenses)

		var totalExpenses float64
		for _, e := range expenses {
			totalExpenses += e.Amount
		}

		profit := totalRevenue - totalExpenses

		c.JSON(http.StatusOK, gin.H{
			"month":            time.Now().Month().String(),
			"revenue":          totalRevenue,
			"expenses":         totalExpenses,
			"profit":           profit,
		})
	}
}

