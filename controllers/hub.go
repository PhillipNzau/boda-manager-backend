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
)

// CreateHub - only authenticated user can create
func CreateHub(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		fmt.Println("user_id in context:", uid)

		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		var input struct {
			Title string `json:"title" binding:"required"`
			Type  string `json:"type" binding:"required"` // e.g., bookmark, ip, note
			Value string `json:"value" binding:"required"`
			Notes string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hub := models.Hub{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			Title:     input.Title,
			Type:      input.Type,
			Value:     input.Value,
			Notes:     input.Notes,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("hubs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = col.InsertOne(ctx, hub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save hub entry"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": hub.ID.Hex(), "message": "hub entry created"})
	}
}

// ListHubs - Show all hubs for current user
func ListHubs(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("hubs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"user_id": userID}
		cursor, err := col.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch hub entries"})
			return
		}

		var hubs []models.Hub
		if err := cursor.All(ctx, &hubs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode hub entries"})
			return
		}

		// --- Generate ETag and Last-Modified ---
		combined := ""
		var lastModified time.Time
		for _, h := range hubs {
			combined += fmt.Sprintf("%s-%d", h.ID.Hex(), h.UpdatedAt.UnixNano())
			if h.UpdatedAt.After(lastModified) {
				lastModified = h.UpdatedAt
			}
		}
		hash := md5.Sum([]byte(combined))
		collectionETag := `"` + hex.EncodeToString(hash[:]) + `"`

		// Check If-None-Match header
		if match := c.GetHeader("If-None-Match"); match != "" && match == collectionETag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", collectionETag)

		// Last-Modified header
		if !lastModified.IsZero() {
			c.Header("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
		}

		c.JSON(http.StatusOK, hubs)
	}
}


// GetHub - view single hub item
func GetHub(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hub id"})
			return
		}

		var hub models.Hub
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = cfg.MongoClient.Database(cfg.DBName).
			Collection("hubs").
			FindOne(ctx, bson.M{"_id": oid}).
			Decode(&hub)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "hub entry not found"})
			return
		}

		c.JSON(http.StatusOK, hub)
	}
}

// UpdateHub - Edit hub
func UpdateHub(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ Get and validate user ID
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}

		// ✅ Get and validate Hub ID
		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hub ID"})
			return
		}

		// ✅ Parse input
		var input struct {
			Title string `json:"title"`
			Type  string `json:"type"`
			Value string `json:"value"`
			Notes string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ✅ Prepare update fields
		update := bson.M{"updated_at": time.Now()}
		if input.Title != "" {
			update["title"] = input.Title
		}
		if input.Type != "" {
			update["type"] = input.Type
		}
		if input.Value != "" {
			update["value"] = input.Value
		}
		if input.Notes != "" {
			update["notes"] = input.Notes
		}

		// ✅ Perform update with ownership check
		col := cfg.MongoClient.Database(cfg.DBName).Collection("hubs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := col.UpdateOne(ctx, bson.M{"_id": oid, "user_id": userID}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update hub"})
			return
		}
		if res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "hub entry not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "hub entry updated", "id": oid.Hex()})
	}
}

// DeleteHub - Remove hub entry
func DeleteHub(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ Get and validate user ID
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}

		// ✅ Get and validate Hub ID
		hubID := c.Param("id")
		oid, err := primitive.ObjectIDFromHex(hubID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hub ID"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("hubs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// ✅ Ensure user ownership before deletion
		res, err := col.DeleteOne(ctx, bson.M{"_id": oid, "user_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete hub entry"})
			return
		}
		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "hub entry not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "hub entry deleted", "id": oid.Hex()})
	}
}

