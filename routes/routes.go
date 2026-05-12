package routes

import (
	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/controllers"
	"github.com/PhillipNzau/boda-manager-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// public auth
	r.POST("/auth/register", controllers.Register(cfg))
	r.POST("/auth/login", controllers.Login(cfg))
	r.POST("/auth/refresh", controllers.RefreshToken(cfg))

	auth := middleware.AuthMiddleware(cfg)

	// =====================================================
	// BODA MANAGER ROUTES
	// =====================================================

	// Riders
	riders := r.Group("/riders")
	riders.Use(auth)
	{
		riders.POST("", controllers.CreateRider(cfg))
		riders.GET("", controllers.ListRiders(cfg))
		riders.DELETE(":id", controllers.DeleteRider(cfg))
	}

	// Rides
	rides := r.Group("/rides")
	rides.Use(auth)
	{
		rides.POST("", controllers.CreateRide(cfg))
		rides.GET("", controllers.ListRides(cfg))
	}

	// Payments
	payments := r.Group("/payments")
	payments.Use(auth)
	{
		payments.POST("", controllers.CreatePayment(cfg))
		payments.GET("", controllers.ListPayments(cfg))
		payments.GET(":id", controllers.GetPayment(cfg))
		payments.DELETE(":id", controllers.DeletePayment(cfg))
	}

	// Expenses
	expenses := r.Group("/expenses")
	expenses.Use(auth)
	{
		expenses.POST("", controllers.CreateExpense(cfg))
		expenses.GET("", controllers.ListExpenses(cfg))
	}

	// Dashboard
	dashboard := r.Group("/dashboard")
	dashboard.Use(auth)
	{
		dashboard.GET("", controllers.DashboardSummary(cfg))
	}

	// Alerts
	alerts := r.Group("/alerts")
	alerts.Use(auth)
	{
		alerts.GET("/missed-payments", controllers.MissedPayments(cfg))
	}

	// Analytics
	analytics := r.Group("/analytics")
	analytics.Use(auth)
	{
		analytics.GET("/monthly", controllers.MonthlyAnalytics(cfg))
		analytics.GET("/motorcycles/:id", controllers.MotorcycleProfitability(cfg))
	}

	motorcycles := r.Group("/motorcycles")
	motorcycles.Use(auth)
	{
		motorcycles.POST("", controllers.CreateMotorcycle(cfg))
		motorcycles.GET("", controllers.ListMotorcycles(cfg))
	}
}