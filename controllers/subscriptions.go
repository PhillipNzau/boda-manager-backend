package controllers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/models"
	"github.com/PhillipNzau/boda-manager-backend/utils"
)

// CreateSubscription - Only the authenticated user can create their own subscription
// parseDateFlexible attempts to parse a date string in RFC3339 or YYYY-MM-DD formats
func parseDateFlexible(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Try full RFC3339 first (e.g. "1984-07-17T00:00:00Z")
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return &t, nil
	}

	// Fallback to date-only format (e.g. "1984-07-17")
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return &t, nil
	}

	// If both fail, return an error
	return nil, fmt.Errorf("invalid date format: %s", dateStr)
}

func CreateSubscription(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		var input struct {
			ServiceName string  `json:"service_name" binding:"required"`
			PlanName    string  `json:"plan_name"`
			StartDate   string  `json:"start_date"`
			RenewalDate string  `json:"renewal_date"`
			Price       float64 `json:"price" binding:"required"`
			Currency    string  `json:"currency" binding:"required"`
			Status      string  `json:"status" binding:"required"`
			Notes       string  `json:"notes"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sd, _ := parseDateFlexible(input.StartDate)
		rd, _ := parseDateFlexible(input.RenewalDate)

		sub := models.Subscription{
			ID:          primitive.NewObjectID(),
			UserID:      userID,
			ServiceName: input.ServiceName,
			PlanName:    input.PlanName,
			StartDate:   sd,
			RenewalDate: rd,
			Price:       input.Price,
			Currency:    input.Currency,
			Status:      input.Status,
			Notes:       input.Notes,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("subscriptions")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if _, err := col.InsertOne(ctx, sub); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": sub.ID.Hex(), "message": "created"})
	}
}

// ListSubscriptions - Only the authenticated user can list their own subscriptions
func ListSubscriptions(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)

		col := cfg.MongoClient.Database(cfg.DBName).Collection("subscriptions")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{"user_id": userID}
		if status := c.Query("status"); status != "" {
			filter["status"] = status
		}

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed"})
			return
		}

		var subs []models.Subscription
		if err := cursor.All(ctx, &subs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decode failed"})
			return
		}

		// --- Build ETag for collection ---
		combined := ""
		var lastModified time.Time
		for _, s := range subs {
			combined += fmt.Sprintf("%s-%d", s.ID.Hex(), s.UpdatedAt.UnixNano())
			if s.UpdatedAt.After(lastModified) {
				lastModified = s.UpdatedAt
			}
		}
		hash := md5.Sum([]byte(combined))
		collectionETag := `"` + hex.EncodeToString(hash[:]) + `"`

		// Handle If-None-Match
		if match := c.GetHeader("If-None-Match"); match != "" && match == collectionETag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", collectionETag)

		// --- Add Last-Modified (latest subscription) ---
		if !lastModified.IsZero() {
			c.Header("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
		}

		c.JSON(http.StatusOK, subs)
	}
}

// GetSubscription - Only the owner can fetch their subscription
func GetSubscription(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, _ := primitive.ObjectIDFromHex(uid)
		oid, _ := primitive.ObjectIDFromHex(c.Param("id"))

		col := cfg.MongoClient.Database(cfg.DBName).Collection("subscriptions")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var sub models.Subscription
		if err := col.FindOne(ctx, bson.M{"_id": oid, "user_id": userID}).Decode(&sub); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// --- Generate ETag ---
		etag := utils.GenerateETag(sub.ID, sub.UpdatedAt)
		if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)

		// --- Add Last-Modified ---
		c.Header("Last-Modified", sub.UpdatedAt.UTC().Format(http.TimeFormat))

		c.JSON(http.StatusOK, sub)
	}
}

// UpdateSubscription - Only the owner can update their subscription
func UpdateSubscription(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ Validate user ID
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		// ✅ Validate object ID
		id := c.Param("id")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
			return
		}

		// ✅ Bind request JSON
		var input struct {
			ServiceName string   `json:"service_name"`
			PlanName    string   `json:"plan_name"`
			StartDate   string   `json:"start_date"`
			RenewalDate string   `json:"renewal_date"`
			Price       *float64 `json:"price"`
			Currency    string   `json:"currency"`
			Status      string   `json:"status"`
			Notes       string   `json:"notes"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ✅ Prepare update document
		update := bson.M{"updated_at": time.Now()}

		if input.ServiceName != "" {
			update["service_name"] = input.ServiceName
		}
		if input.PlanName != "" {
			update["plan_name"] = input.PlanName
		}
		if t, err := parseDateFlexible(input.StartDate); err == nil && t != nil {
			update["start_date"] = *t
		}
		if t, err := parseDateFlexible(input.RenewalDate); err == nil && t != nil {
			update["renewal_date"] = *t
		}
		if input.Price != nil {
			update["price"] = *input.Price
		}
		if input.Currency != "" {
			update["currency"] = input.Currency
		}
		if input.Status != "" {
			update["status"] = input.Status
		}
		if input.Notes != "" {
			update["notes"] = input.Notes
		}

		// ❗ Make sure at least one field (other than `updated_at`) is being updated
		if len(update) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		// ✅ Update with ownership enforcement
		col := cfg.MongoClient.Database(cfg.DBName).Collection("subscriptions")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		res, err := col.UpdateOne(ctx, bson.M{"_id": oid, "user_id": userID}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
		if res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "subscription updated"})
	}
}


// DeleteSubscription - Only the owner can delete their subscription
func DeleteSubscription(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ Extract and validate user ID
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		// ✅ Extract and validate subscription ID
		idParam := c.Param("id")
		oid, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("subscriptions")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// ✅ Delete only if user owns the subscription
		res, err := col.DeleteOne(ctx, bson.M{"_id": oid, "user_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
			return
		}
		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "subscription deleted", "id": oid.Hex()})
	}
}
