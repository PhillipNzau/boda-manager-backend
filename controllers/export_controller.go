package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ExportReportExcel(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		db := cfg.MongoClient.Database(cfg.DBName)

		file := excelize.NewFile()

		if err := buildPaymentsSheet(ctx, db.Collection("payments"), file, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed building payments sheet",
				"details": err.Error(),
			})
			return
		}

		if err := buildRidersSheet(ctx, db.Collection("riders"), file, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed building riders sheet",
				"details": err.Error(),
			})
			return
		}

		if err := buildRidesSheet(ctx, db.Collection("rides"), file, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed building rides sheet",
				"details": err.Error(),
			})
			return
		}

		file.DeleteSheet("Sheet1")

		filename := fmt.Sprintf("boda_report_%d.xlsx", time.Now().Unix())

		c.Header(
			"Content-Type",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		)
		c.Header("Content-Disposition", "attachment; filename="+filename)

		if err := file.Write(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed generating excel",
			})
			return
		}
	}
}

func buildPaymentsSheet(
	ctx context.Context,
	col *mongo.Collection,
	file *excelize.File,
	userID primitive.ObjectID,
) error {
	sheet := "Payments"
	file.NewSheet(sheet)

	headers := []string{
		"Date",
		"Rider Name",
		"Phone Number",
		"Motorcycle Plate",
		"Amount Paid",
		"Expected Amount",
		"Balance",
		"Status",
		"Method",
		"Notes",
	}

	for i, h := range headers {
		file.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "riders",
				"localField":   "rider_id",
				"foreignField": "_id",
				"as":           "rider",
			},
		}},
		{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$rider",
				"preserveNullAndEmptyArrays": true,
			},
		}},
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "motorcycles",
				"localField":   "motorcycle_id",
				"foreignField": "_id",
				"as":           "motorcycle",
			},
		}},
		{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$motorcycle",
				"preserveNullAndEmptyArrays": true,
			},
		}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	var rows []bson.M
	if err := cursor.All(ctx, &rows); err != nil {
		return err
	}

	for i, row := range rows {
		r := i + 2

		rider, _ := row["rider"].(bson.M)
		motorcycle, _ := row["motorcycle"].(bson.M)

		file.SetCellValue(sheet, fmt.Sprintf("A%d", r), row["date"])
		file.SetCellValue(sheet, fmt.Sprintf("B%d", r), getMapValue(rider, "full_name"))
		file.SetCellValue(sheet, fmt.Sprintf("C%d", r), getMapValue(rider, "phone_number"))
		file.SetCellValue(sheet, fmt.Sprintf("D%d", r), getMapValue(motorcycle, "plate_number"))
		file.SetCellValue(sheet, fmt.Sprintf("E%d", r), row["amount"])
		file.SetCellValue(sheet, fmt.Sprintf("F%d", r), row["expected_amount"])
		file.SetCellValue(sheet, fmt.Sprintf("G%d", r), row["balance"])
		file.SetCellValue(sheet, fmt.Sprintf("H%d", r), row["status"])
		file.SetCellValue(sheet, fmt.Sprintf("I%d", r), row["method"])
		file.SetCellValue(sheet, fmt.Sprintf("J%d", r), row["notes"])
	}

	return nil
}

func buildRidersSheet(
	ctx context.Context,
	col *mongo.Collection,
	file *excelize.File,
	userID primitive.ObjectID,
) error {
	sheet := "Riders"
	file.NewSheet(sheet)

	headers := []string{
		"Full Name",
		"Phone Number",
		"License Number",
		"Daily Target",
		"Outstanding Debt",
		"Last Payment Date",
		"Motorcycle Plate",
	}

	for i, h := range headers {
		file.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "motorcycles",
				"localField":   "motorcycle_id",
				"foreignField": "_id",
				"as":           "motorcycle",
			},
		}},
		{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$motorcycle",
				"preserveNullAndEmptyArrays": true,
			},
		}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	var rows []bson.M
	if err := cursor.All(ctx, &rows); err != nil {
		return err
	}

	for i, row := range rows {
		r := i + 2
		motorcycle, _ := row["motorcycle"].(bson.M)

		file.SetCellValue(sheet, fmt.Sprintf("A%d", r), row["full_name"])
		file.SetCellValue(sheet, fmt.Sprintf("B%d", r), row["phone_number"])
		file.SetCellValue(sheet, fmt.Sprintf("C%d", r), row["license_no"])
		file.SetCellValue(sheet, fmt.Sprintf("D%d", r), row["daily_target"])
		file.SetCellValue(sheet, fmt.Sprintf("E%d", r), row["outstanding_debt"])
		file.SetCellValue(sheet, fmt.Sprintf("F%d", r), row["last_payment_date"])
		file.SetCellValue(sheet, fmt.Sprintf("G%d", r), getMapValue(motorcycle, "plate_number"))
	}

	return nil
}

func buildRidesSheet(
	ctx context.Context,
	col *mongo.Collection,
	file *excelize.File,
	userID primitive.ObjectID,
) error {
	sheet := "Rides"
	file.NewSheet(sheet)

	headers := []string{
		"Date",
		"Rider",
		"Motorcycle",
		"Trips",
		"Income",
		"Fuel Cost",
		"Expenses",
		"Net Profit",
		"Notes",
	}

	for i, h := range headers {
		file.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "riders",
				"localField":   "rider_id",
				"foreignField": "_id",
				"as":           "rider",
			},
		}},
		{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$rider",
				"preserveNullAndEmptyArrays": true,
			},
		}},
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "motorcycles",
				"localField":   "motorcycle_id",
				"foreignField": "_id",
				"as":           "motorcycle",
			},
		}},
		{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$motorcycle",
				"preserveNullAndEmptyArrays": true,
			},
		}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	var rows []bson.M
	if err := cursor.All(ctx, &rows); err != nil {
		return err
	}

	for i, row := range rows {
		r := i + 2

		rider, _ := row["rider"].(bson.M)
		motorcycle, _ := row["motorcycle"].(bson.M)

		income := toFloat(row["income"])
		fuel := toFloat(row["fuel_cost"])
		expenses := toFloat(row["expenses"])
		netProfit := income - fuel - expenses

		file.SetCellValue(sheet, fmt.Sprintf("A%d", r), row["date"])
		file.SetCellValue(sheet, fmt.Sprintf("B%d", r), getMapValue(rider, "full_name"))
		file.SetCellValue(sheet, fmt.Sprintf("C%d", r), getMapValue(motorcycle, "plate_number"))
		file.SetCellValue(sheet, fmt.Sprintf("D%d", r), row["trips"])
		file.SetCellValue(sheet, fmt.Sprintf("E%d", r), income)
		file.SetCellValue(sheet, fmt.Sprintf("F%d", r), fuel)
		file.SetCellValue(sheet, fmt.Sprintf("G%d", r), expenses)
		file.SetCellValue(sheet, fmt.Sprintf("H%d", r), netProfit)
		file.SetCellValue(sheet, fmt.Sprintf("I%d", r), row["notes"])
	}

	return nil
}

func getMapValue(m bson.M, key string) string {
	if m == nil {
		return ""
	}

	if value, ok := m[key]; ok {
		return fmt.Sprintf("%v", value)
	}

	return ""
}

func toFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}