package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/models"
	"github.com/PhillipNzau/boda-manager-backend/utils"
)

//
// ================== EXPORT ==================
//

// Export Credentials to Excel
func ExportCredentialsExcel(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{"user_id": userID}
		cursor, err := config.GetCollection("credentials").Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var credentials []*models.Credential
		if err := cursor.All(ctx, &credentials); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		f := excelize.NewFile()
		sheet := "Credentials"
		f.NewSheet(sheet)

		headers := []string{"SiteName", "Username", "Password", "LoginURL", "Notes", "Category"}
		for i, h := range headers {
			cell := string(rune('A'+i)) + "1"
			f.SetCellValue(sheet, cell, h)
		}

		for i, cred := range credentials {
			row := strconv.Itoa(i + 2)
			password, err := utils.Decrypt(cfg.AESKey, cred.PasswordEncrypted)
			if err != nil {
				password = "[decryption error]"
			}

			f.SetCellValue(sheet, "A"+row, cred.SiteName)
			f.SetCellValue(sheet, "B"+row, cred.Username)
			f.SetCellValue(sheet, "C"+row, password)
			f.SetCellValue(sheet, "D"+row, cred.LoginURL)
			f.SetCellValue(sheet, "E"+row, cred.Notes)
			f.SetCellValue(sheet, "F"+row, cred.Category)
		}

		c.Header("Content-Disposition", "attachment; filename=credentials.xlsx")
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

		if err := f.Write(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}



// Export Subscriptions to Excel
func ExportSubscriptionsExcel(c *gin.Context) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := config.GetCollection("subscriptions").Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var subscriptions []models.Subscription
	if err := cursor.All(ctx, &subscriptions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "Subscriptions"
	f.NewSheet(sheet)

	headers := []string{
		"ServiceName", "PlanName", "StartDate", "RenewalDate", "Price",
		"Currency", "Status", "Notes", "CreatedAt", "UpdatedAt",
	}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		f.SetCellValue(sheet, cell, h)
	}

	for i, s := range subscriptions {
		row := strconv.Itoa(i + 2)
		f.SetCellValue(sheet, "A"+row, s.ServiceName)
		f.SetCellValue(sheet, "B"+row, s.PlanName)
		if s.StartDate != nil {
			f.SetCellValue(sheet, "C"+row, s.StartDate.Format("2006-01-02"))
		}
		if s.RenewalDate != nil {
			f.SetCellValue(sheet, "D"+row, s.RenewalDate.Format("2006-01-02"))
		}
		f.SetCellValue(sheet, "E"+row, s.Price)
		f.SetCellValue(sheet, "F"+row, s.Currency)
		f.SetCellValue(sheet, "G"+row, s.Status)
		f.SetCellValue(sheet, "H"+row, s.Notes)
		f.SetCellValue(sheet, "I"+row, s.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "J"+row, s.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	c.Header("Content-Disposition", "attachment; filename=subscriptions.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}



// Export Resources to Excel
func ExportResourcesExcel(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	cursor, err := config.GetCollection("hubs").Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var hubs []*models.Hub
	if err := cursor.All(ctx, &hubs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "Hubs"
	f.NewSheet(sheet)

	headers := []string{"Title", "Type", "Value", "Notes", "CreatedAt"}
	for i, h := range headers {
		cell := string(rune('A' + i)) + "1"
		f.SetCellValue(sheet, cell, h)
	}

	for i, s := range hubs {
		row := strconv.Itoa(i + 2)
		f.SetCellValue(sheet, "A"+row, s.Title)
		f.SetCellValue(sheet, "B"+row, s.Type)
		f.SetCellValue(sheet, "C"+row, s.Value)
		f.SetCellValue(sheet, "D"+row, s.Notes)
		f.SetCellValue(sheet, "E"+row, s.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	c.Header("Content-Disposition", "attachment; filename=resources.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}


//
// ================== IMPORT ==================
//

// Import Credentials from Excel
func ImportCredentialsExcel(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get logged-in user ID from context
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer src.Close()

		f, err := excelize.OpenReader(src)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Excel file"})
			return
		}

		rows, err := f.GetRows("Credentials")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sheet Credentials not found"})
			return
		}

		var credentials []*models.Credential
		for i, row := range rows {
			if i == 0 {
				continue // skip header
			}
			if len(row) < 6 {
				continue // skip incomplete rows
			}

			enc, err := utils.Encrypt(cfg.AESKey, row[2]) // password is column C (index 2)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt password"})
				return
			}

			cred := &models.Credential{
				ID:                primitive.NewObjectID(),
				UserID:            userID,
				SiteName:          row[0], // SiteName = A
				Username:          row[1], // Username = B
				PasswordEncrypted: enc,    // Password = C (encrypted)
				LoginURL:          row[3], // LoginURL = D
				Notes:             row[4], // Notes = E
				Category:          row[5], // Category = F
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}
			credentials = append(credentials, cred)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var docs []any
		for _, cred := range credentials {
			docs = append(docs, cred)
		}

		if _, err := config.GetCollection("credentials").InsertMany(ctx, docs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"imported": len(credentials)})
	}
}


// Import Subscriptions from Excel
func ImportSubscriptionsExcel(c *gin.Context) {
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

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Excel file"})
		return
	}

	rows, err := f.GetRows("Subscriptions")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sheet 'Subscriptions' not found"})
		return
	}

	var subs []*models.Subscription
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}

		if len(row) < 7 {
			continue // skip incomplete rows
		}

		price, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			price = 0
		}

		// Optional date parsing
		var startDate *time.Time
		if t, err := time.Parse("2006-01-02", row[2]); err == nil {
			startDate = &t
		}

		var renewalDate *time.Time
		if t, err := time.Parse("2006-01-02", row[3]); err == nil {
			renewalDate = &t
		}

		sub := &models.Subscription{
			ID:          primitive.NewObjectID(),
			UserID:      userID,
			ServiceName: row[0],
			PlanName:    row[1],
			StartDate:   startDate,
			RenewalDate: renewalDate,
			Price:       price,
			Currency:    row[5],
			Status:      row[6],
			Notes:       "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if len(row) >= 8 {
			sub.Notes = row[7]
		}

		subs = append(subs, sub)
	}

	if len(subs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid data to import"})
		return
	}

	// Insert into DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	docs := make([]interface{}, len(subs))
	for i, s := range subs {
		docs[i] = s
	}

	if _, err := config.GetCollection("subscriptions").InsertMany(ctx, docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imported": len(subs)})
}


// Import Resources from Excel
func ImportResourcesExcel(c *gin.Context) {
	// 1. Get logged-in user ID from context
    uid := c.GetString("user_id")
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID, err := primitive.ObjectIDFromHex(uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
        return
    }

	// 2. Handle uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Excel file"})
		return
	}

	rows, err := f.GetRows("Hubs")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sheet Hubs not found"})
		return
	}

	// 3. Build Hubs list
	var hubs []*models.Hub
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 8 {
			continue
		}

		hubs = append(hubs, &models.Hub{
			ID:         primitive.NewObjectID(),
			UserID:		userID,         // 🔑 attach logged-in user
			Title:		row[1],
			Type: 		row[2],
			Value:  	row[5],
			Notes:		row[6],
			CreatedAt:	time.Now(),
			UpdatedAt:	time.Now(),
		})
	}

	// 4. Insert into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var docs []interface{}
	for _, s := range hubs {
		docs = append(docs, s)
	}

	if _, err := config.GetCollection("hubs").InsertMany(ctx, docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imported": len(hubs)})
}
