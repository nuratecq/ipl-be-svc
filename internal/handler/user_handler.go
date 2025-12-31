package handler

import (
	"strconv"

	"ipl-be-svc/internal/models/response"
	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"
	"ipl-be-svc/pkg/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService service.UserService
	logger      *logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// UserDetailResponse represents the user detail response structure
type UserDetailResponse struct {
	ID           uint   `json:"id" example:"1"`
	NamaPenghuni string `json:"nama_penghuni" example:"John Doe"`
	NamaPemilik  string `json:"nama_pemilik" example:"Jane Doe"`
	Blok         string `json:"blok" example:"A1"`
	Rt           int    `json:"rt" example:"5"`
	NoHP         string `json:"no_hp" example:"+6281234567890"`
	NoTelp       string `json:"no_telp" example:"021-12345678"`
	DocumentID   string `json:"document_id" example:"abc123def456"`
	Email        string `json:"email" example:"john.doe@example.com"`
	UserID       uint   `json:"user_id" example:"123"`
	RoleName     string `json:"role_name" example:"Administrator"`
	RoleID       uint   `json:"role_id" example:"1"`
	RoleType     string `json:"role_type" example:"admin"`
}

// GetUserDetailByProfileID handles GET /api/v1/users/profile/:user_id
// @Summary Get user detail by profile ID
// @Description Get detailed user information by profile ID
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} utils.APIResponse{data=UserDetailResponse} "User detail retrieved successfully"
// @Failure 400 {object} utils.APIResponse "Invalid user ID"
// @Failure 404 {object} utils.APIResponse "User not found"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/users/profile/{user_id} [get]
func (h *UserHandler) GetUserDetailByProfileID(c *gin.Context) {
	// Get user ID from path parameter
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("user_id_param", userIDParam).Error("Invalid user ID parameter")
		utils.BadRequestResponse(c, "Invalid user ID", err)
		return
	}

	// Get user detail
	userDetail, err := h.userService.GetUserDetailByProfileID(uint(userID))
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user detail")

		// Check if it's a not found error
		if err.Error() == "record not found" || err.Error() == "sql: no rows in result set" {
			utils.NotFoundResponse(c, "User not found")
			return
		}

		utils.InternalServerErrorResponse(c, "Failed to get user detail", err)
		return
	}

	// Convert to response format
	response := UserDetailResponse{
		ID:           userDetail.ID,
		NamaPenghuni: userDetail.NamaPenghuni,
		NamaPemilik:  userDetail.NamaPemilik,
		Blok:         userDetail.Blok,
		Rt:           userDetail.Rt,
		NoHP:         userDetail.NoHP,
		NoTelp:       userDetail.NoTelp,
		DocumentID:   userDetail.DocumentID,
		Email:        userDetail.Email,
		UserID:       userDetail.UserID,
		RoleName:     userDetail.RoleName,
		RoleID:       userDetail.RoleID,
		RoleType:     userDetail.RoleType,
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"profile_id": userDetail.ID,
		"email":      userDetail.Email,
	}).Info("User detail retrieved successfully")

	utils.SuccessResponse(c, "User detail retrieved successfully", response)
}

// GetPenghuniUsers handles GET /api/v1/users/penghuni
// @Summary Get all penghuni users
// @Description Get list of all users with role type "penghuni"
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]response.PenghuniUserResponse} "Penghuni users retrieved successfully"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/users/penghuni [get]
func (h *UserHandler) GetPenghuniUsers(c *gin.Context) {
	// Get penghuni users
	users, err := h.userService.GetPenghuniUsers()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get penghuni users")
		utils.InternalServerErrorResponse(c, "Failed to get penghuni users", err)
		return
	}

	// Convert to response format
	var responses []response.PenghuniUserResponse
	for _, user := range users {
		responses = append(responses, response.PenghuniUserResponse{
			ID:           user.ID,
			Username:     user.Username,
			Email:        user.Email,
			NamaPenghuni: user.NamaPenghuni,
			NamaPemilik:  user.NamaPemilik,
			Blok:         user.Blok,
			Rt:           user.Rt,
			NoHP:         user.NoHP,
			NoTelp:       user.NoTelp,
			DocumentID:   user.DocumentID,
			RoleName:     user.RoleName,
			RoleID:       user.RoleID,
			RoleType:     user.RoleType,
		})
	}

	h.logger.WithField("count", len(responses)).Info("Penghuni users retrieved successfully")

	utils.SuccessResponse(c, "Penghuni users retrieved successfully", responses)
}
