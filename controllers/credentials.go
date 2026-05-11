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

// CreateCredential - Add new password/credential
func CreateCredential(cfg *config.Config) gin.HandlerFunc {
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

		// ✅ Parse and validate input
		var input struct {
			SiteName string `json:"site_name" binding:"required"`
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
			LoginURL string `json:"login_url"`
			Notes    string `json:"notes"`
			Category string `json:"category"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ✅ Encrypt password (already returns base64 string)
		enc, err := utils.Encrypt(cfg.AESKey, input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
			return
		}

		// ✅ Build credential object
		cred := models.Credential{
			ID:                primitive.NewObjectID(),
			UserID:            userID,
			SiteName:          input.SiteName,
			Username:          input.Username,
			PasswordEncrypted: enc, // already safe string
			LoginURL:          input.LoginURL,
			Notes:             input.Notes,
			Category:          input.Category,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// ✅ Insert into MongoDB
		col := cfg.MongoClient.Database(cfg.DBName).Collection("credentials")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := col.InsertOne(ctx, cred); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "could not save credential",
				"details": err.Error(),
			})
			return
		}

		// ✅ Success response
		c.JSON(http.StatusCreated, gin.H{
			"id":      cred.ID.Hex(),
			"message": "credential created",
		})
	}
}



// ListCredentials - Show all credentials for logged-in user
func ListCredentials(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}


		col := cfg.MongoClient.Database(cfg.DBName).Collection("credentials")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{"user_id": userID}
		if q := c.Query("q"); q != "" {
			filter["$or"] = bson.A{
				bson.M{"site_name": bson.M{"$regex": q, "$options": "i"}},
				bson.M{"username": bson.M{"$regex": q, "$options": "i"}},
			}
		}

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch credentials"})
			return
		}

		var creds []models.Credential
		if err := cursor.All(ctx, &creds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode creds"})
			return
		}

		// --- Build ETag for collection ---
		combined := ""
		var lastModified time.Time
		for _, cr := range creds {
			combined += fmt.Sprintf("%s-%d", cr.ID.Hex(), cr.UpdatedAt.UnixNano())
			if cr.UpdatedAt.After(lastModified) {
				lastModified = cr.UpdatedAt
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

		// --- Add Last-Modified (latest credential) ---
		if !lastModified.IsZero() {
			c.Header("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
		}

		// Decrypt passwords for output
		out := make([]gin.H, 0, len(creds))
		for _, cr := range creds {
			pass, err := utils.Decrypt(cfg.AESKey, cr.PasswordEncrypted)
			if err != nil {
				pass = "" // or skip the record, or log it
			}

			out = append(out, gin.H{
				"id":         cr.ID.Hex(),
				"site_name":  cr.SiteName,
				"username":   cr.Username,
				"password":   pass,
				"login_url":  cr.LoginURL,
				"notes":      cr.Notes,
				"category":   cr.Category,
				"created_at": cr.CreatedAt,
				"updated_at": cr.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, out)
	}
}

// GetCredential - Fetch single credential
func GetCredential(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}


		credID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
			return
		}

		var credential models.Credential
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = cfg.MongoClient.Database(cfg.DBName).
			Collection("credentials").
			FindOne(ctx, bson.M{"_id": credID, "user_id": userID}).
			Decode(&credential)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found or not owned"})
			return
		}

		// --- Generate ETag via utils ---
		etag := utils.GenerateETag(credential.ID, credential.UpdatedAt)
		if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)

		// --- Add Last-Modified ---
		c.Header("Last-Modified", credential.UpdatedAt.UTC().Format(http.TimeFormat))

		// Decrypt before sending
		pass, _ := utils.Decrypt(cfg.AESKey, credential.PasswordEncrypted)
		credential.PasswordEncrypted = pass

		c.JSON(http.StatusOK, credential)
	}
}

// UpdateCredential - Edit credential
func UpdateCredential(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ Get and validate user ID
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

		// ✅ Get and validate credential ID
		oid, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
			return
		}

		// ✅ Bind input
		var input struct {
			SiteName string `json:"site_name"`
			Username string `json:"username"`
			Password string `json:"password"`
			LoginURL string `json:"login_url"`
			Notes    string `json:"notes"`
			Category string `json:"category"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ✅ Find the credential and ensure ownership
		col := cfg.MongoClient.Database(cfg.DBName).Collection("credentials")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var existing models.Credential
		err = col.FindOne(ctx, bson.M{"_id": oid, "user_id": userID}).Decode(&existing)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found or not owned"})
			return
		}

		// ✅ Build update document
		update := bson.M{"updated_at": time.Now()}

		if input.SiteName != "" {
			update["site_name"] = input.SiteName
		}
		if input.Username != "" {
			update["username"] = input.Username
		}
		if input.Password != "" {
			enc, err := utils.Encrypt(cfg.AESKey, input.Password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt password"})
				return
			}
			update["password_encrypted"] = enc
		}
		if input.LoginURL != "" {
			update["login_url"] = input.LoginURL
		}
		if input.Notes != "" {
			update["notes"] = input.Notes
		}
		if input.Category != "" {
			update["category"] = input.Category
		}

		// ❗ Ensure at least one field is being updated (besides updated_at)
		if len(update) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		// ✅ Perform update
		res, err := col.UpdateOne(ctx, bson.M{"_id": oid, "user_id": userID}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update credential"})
			return
		}
		if res.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "credential updated", "id": oid.Hex()})
	}
}

// DeleteCredential - Remove credential
func DeleteCredential(cfg *config.Config) gin.HandlerFunc {
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

		// ✅ Extract and validate credential ID
		idParam := c.Param("id")
		oid, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
			return
		}

		col := cfg.MongoClient.Database(cfg.DBName).Collection("credentials")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// ✅ Delete only if the user owns the credential
		res, err := col.DeleteOne(ctx, bson.M{"_id": oid, "user_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete credential"})
			return
		}
		if res.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found or not owned"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "credential deleted", "id": oid.Hex()})
	}
}


