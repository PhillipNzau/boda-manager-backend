package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/models"
	"github.com/PhillipNzau/boda-manager-backend/utils"
)

// =============================
// Register (send OTP only)
// =============================
func Register(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Check existing email
		count, err := users.CountDocuments(ctx, bson.M{"email": input.Email})
		if err != nil {
			log.Printf("count error: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}

		user := models.User{
			ID:        primitive.NewObjectID(),
			Name:      input.Name,
			Email:     input.Email,
			Password:  hashedPassword,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = users.InsertOne(ctx, user)
		if err != nil {
			log.Printf("insert error: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		// if _, err := users.InsertOne(ctx, user); err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		// 	return
		// }

		c.JSON(http.StatusCreated, gin.H{
			"status":  201,
			"message": "user created successfully",
		})
	}
}

// =============================
// Login (send OTP only)
// =============================
func Login(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var user models.User
		if err := users.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// Compare password
		if err := utils.CheckPassword(user.Password, input.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// Generate tokens
		accessToken, refreshToken, err := createTokensForUser(user.ID, cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}

		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
			"$set": bson.M{"refresh_token": refreshToken},
		})

		c.JSON(http.StatusOK, gin.H{
			"status":        200,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user": gin.H{
				"id":    user.ID.Hex(),
				"name":  user.Name,
				"email": user.Email,
			},
		})
	}
}


// =============================
// Refresh Token (unchanged)
// =============================
func RefreshToken(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh_token"})
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(input.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
			return cfg.JWTSecret, nil
		})
		if err != nil || !token.Valid || claims["type"] != "refresh" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		uid, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
			return
		}

		users := cfg.MongoClient.Database(cfg.DBName).Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var user models.User
		objID, _ := primitive.ObjectIDFromHex(uid)
		if err := users.FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		if user.RefreshToken != input.RefreshToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token mismatch"})
			return
		}

		// Create new tokens
		accessToken, refreshToken, _ := createTokensForUser(user.ID, cfg)

		// Rotate refresh token
		users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"refresh_token": refreshToken}})

		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	}
}

// =============================
// Helpers
// =============================
func createTokensForUser(uid primitive.ObjectID, cfg *config.Config) (accessToken string, refreshToken string, err error) {
	// Access Token (short-lived)
	accessClaims := jwt.MapClaims{
		"user_id": uid.Hex(),
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = access.SignedString(cfg.JWTSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (long-lived)
	refreshClaims := jwt.MapClaims{
		"user_id": uid.Hex(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refresh.SignedString(cfg.JWTSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
