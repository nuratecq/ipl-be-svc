package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"
)

// Routes sets up all API routes
func SetupRoutes(
	router *gin.Engine,
	menuService service.MenuService,
	paymentService service.PaymentService,
	userService service.UserService,
	billingService service.BillingService,
	masterMenuService service.MasterMenuService,
	roleMenuService service.RoleMenuService,
	dashboardService service.DashboardService,
	logger *logger.Logger,
) {
	// Initialize handlers
	menuHandler := NewMenuHandler(menuService, logger)
	paymentHandler := NewPaymentHandler(paymentService, logger)
	userHandler := NewUserHandler(userService, logger)
	bulkBillingHandler := NewBulkBillingHandler(billingService, logger)
	masterMenuHandler := NewMasterMenuHandler(masterMenuService, logger)
	roleMenuHandler := NewRoleMenuHandler(roleMenuService, logger)
	dashboardHandler := NewDashboardHandler(dashboardService, logger)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", HealthCheck)

		// Menu routes
		menus := v1.Group("/menus")
		{
			menus.GET("/user/:id", menuHandler.GetMenusByUserID)
		}

		// Payment routes
		payments := v1.Group("/payments")
		{
			payments.POST("/billing/:id/link", paymentHandler.CreatePaymentLink)
			payments.POST("/billing/link", paymentHandler.CreatePaymentLinkMultiple)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.GET("/profile/:user_id", userHandler.GetUserDetailByProfileID)
			users.GET("/penghuni", userHandler.GetPenghuniUsers)
		}

		// Billing routes
		billings := v1.Group("/billings")
		{
			billings.POST("/bulk-monthly", bulkBillingHandler.CreateBulkMonthlyBillings)
			billings.POST("/bulk-custom", bulkBillingHandler.CreateBulkCustomBillings)
			// Payment confirmation webhook endpoint
			billings.POST("/confirm-payment", bulkBillingHandler.ConfirmPaymentWebhook)
			// Confirm single billing via JSON body {billing_id}
			billings.POST("/confirm-single", bulkBillingHandler.ConfirmPaymentSingle)
			// Admin endpoint to confirm payments by billing IDs
			billings.GET("/penghuni", bulkBillingHandler.GetBillingPenghuni)
			// Paginated/search endpoint
			billings.GET("/penghuni/search", bulkBillingHandler.GetBillingPenghuniSearch)
			// Get profile billing with optional filters
			billings.GET("/profile", bulkBillingHandler.GetProfileBillingWithFilters)
			// Get billing by profile ID with optional filters
			billings.GET("/by-profile", bulkBillingHandler.GetBillingByProfileID)
			// Get billing statistics with optional filters
			billings.GET("/statistics", bulkBillingHandler.GetBillingStatistics)
			// Export billing to Excel
			billings.GET("/export", bulkBillingHandler.ExportBillingToExcel)
			// Billing attachments
			billings.POST("/:id/attachments", bulkBillingHandler.UploadBillingAttachment)
			billings.GET("/:id/attachments", bulkBillingHandler.ListBillingAttachments)
			billings.GET("/:id/attachments/:attachment_id", bulkBillingHandler.DownloadBillingAttachment)
		}

		// Master Menu routes
		masterMenus := v1.Group("/master-menus")
		{
			masterMenus.POST("", masterMenuHandler.CreateMasterMenu)
			masterMenus.GET("", masterMenuHandler.GetAllMasterMenus)
			masterMenus.GET("/:id", masterMenuHandler.GetMasterMenu)
			masterMenus.PUT("/:id", masterMenuHandler.UpdateMasterMenu)
			masterMenus.DELETE("/:id", masterMenuHandler.DeleteMasterMenu)
		}

		// Role Menu routes
		roleMenus := v1.Group("/role-menus")
		{
			roleMenus.POST("", roleMenuHandler.CreateRoleMenu)
			roleMenus.GET("", roleMenuHandler.GetAllRoleMenus)
			roleMenus.GET("/:id", roleMenuHandler.GetRoleMenu)
			roleMenus.PUT("/:id", roleMenuHandler.UpdateRoleMenu)
			roleMenus.DELETE("/:id", roleMenuHandler.DeleteRoleMenu)

			// Master menu attachments
			roleMenus.POST("/:id/master-menus", roleMenuHandler.AttachMasterMenu)
			roleMenus.DELETE("/:id/master-menus/:master_menu_id", roleMenuHandler.DetachMasterMenu)

			// Role attachments
			roleMenus.POST("/:id/roles", roleMenuHandler.AttachRole)
			roleMenus.DELETE("/:id/roles/:role_id", roleMenuHandler.DetachRole)
		}

		// Role-specific role menu routes
		roles := v1.Group("/roles")
		{
			roles.GET("/:role_id/role-menus", roleMenuHandler.GetRoleMenusByRoleID)
		}

		// Dashboard routes
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/statistics", dashboardHandler.GetDashboardStatistics)
			dashboard.GET("/billings", dashboardHandler.GetBillingList)
		}
	}
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Server is running",
		"service": "IPL Backend Service",
	})
}
