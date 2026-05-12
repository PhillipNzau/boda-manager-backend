package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/models"
)

// CreateCategory - only authenticated user can create
func CreateCategory(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		var input struct {
			Name  string `json:"name" binding:"required"`
			Notes string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		category := models.Category{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			Name:      input.Name,
			Notes:     input.Notes,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("categories")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		_, err = col.InsertOne(ctx, category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save category"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": category.ID.Hex(), "message": "category created"})
	}
}

// ListCategories - Show all categories for all users
func ListCategories(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		
		col := cfg.MongoClient.Database(cfg.DBName).Collection("categories")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		cursor, err := col.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch categories"})
			return
		}

		var categories []models.Category
		if err := cursor.All(ctx, &categories); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode categories"})
			return
		}

		// Return the categories as a response
		c.JSON(http.StatusOK, categories)
	}
}



// GetCategory - anyone can view
func GetCategory(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {

        catID, err := primitive.ObjectIDFromHex(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
            return
        }

        var category models.Category
        ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
        defer cancel()

        err = cfg.MongoClient.Database(cfg.DBName).
            Collection("categories").
            FindOne(ctx, bson.M{"_id": catID}).
            Decode(&category)

        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "category not found or not owned"})
            return
        }

        c.JSON(http.StatusOK, category)
    }
}

// UpdateCategory - Edit category
func UpdateCategory(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
			return
		}

		var input struct {
			Name  string `json:"name"`
			Notes string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{
			"updated_at": time.Now(),
		}

		if input.Name != "" {
			update["name"] = input.Name
		}
		if input.Notes != "" {
			update["notes"] = input.Notes
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("categories")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		res, err := col.UpdateOne(ctx, bson.M{"_id": oid, "user_id": userID}, bson.M{"$set": update})
		if err != nil || res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "category updated"})
	}
}

// DeleteCategory - Remove category
func DeleteCategory(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("categories")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		res, err := col.DeleteOne(ctx, bson.M{"_id": oid, "user_id": userID})
		if err != nil || res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
	}
}
