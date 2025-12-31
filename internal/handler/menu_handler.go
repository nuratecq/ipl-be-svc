package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ipl-be-svc/internal/models/response"
	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"
	"ipl-be-svc/pkg/utils"
)

// MenuHandler handles menu-related HTTP requests
type MenuHandler struct {
	menuService service.MenuService
	logger      *logger.Logger
}

// NewMenuHandler creates a new menu handler
func NewMenuHandler(menuService service.MenuService, logger *logger.Logger) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
		logger:      logger,
	}
}

// GetMenusByUserID handles GET /api/v1/menus/user/:user_id
// @Summary Get menus by user ID
// @Description Get list of menus accessible by a specific user ID
// @Tags menus
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} utils.APIResponse{data=[]response.MenuResponse} "Menus retrieved successfully"
// @Failure 400 {object} utils.APIResponse "Invalid user ID"
// @Failure 404 {object} utils.APIResponse "No menus found"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/menus/user/{user_id} [get]
func (h *MenuHandler) GetMenusByUserID(c *gin.Context) {
	userID, err := utils.GetIDParam(c)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID parameter")
		utils.BadRequestResponse(c, "Invalid user ID", err)
		return
	}

	menus, err := h.menuService.GetMenusByUserID(userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get menus")
		if strings.Contains(err.Error(), "invalid user ID") {
			utils.BadRequestResponse(c, "Invalid user ID", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get menus", err)
		return
	}

	if len(menus) == 0 {
		h.logger.WithField("user_id", userID).Info("No menus found for user")
		utils.SuccessResponse(c, "No menus found for this user", []interface{}{})
		return
	}

	// Convert to response format
	var menuResponses []response.MenuResponse
	for _, menu := range menus {
		var publishedAt *string
		if menu.PublishedAt != nil {
			pubAt := menu.PublishedAt.Format("2006-01-02T15:04:05.000Z")
			publishedAt = &pubAt
		}

		menuResponses = append(menuResponses, response.MenuResponse{
			ID:          menu.ID,
			DocumentID:  menu.DocumentID,
			NamaMenu:    menu.NamaMenu,
			KodeMenu:    menu.KodeMenu,
			UrutanMenu:  menu.UrutanMenu,
			IsActive:    menu.IsActive,
			PublishedAt: publishedAt,
		})
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"menu_count": len(menuResponses),
	}).Info("Menus retrieved successfully")

	utils.SuccessResponse(c, "Menus retrieved successfully", menuResponses)
}
