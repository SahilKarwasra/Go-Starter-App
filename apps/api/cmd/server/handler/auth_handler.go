package handler

import (
	"api/cmd/server/services"
	"api/cmd/server/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var req services.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, utils.FormatValidationError(err))
		return
	}

	res, err := h.authService.SignUp(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, utils.ErrEmailAlreadyExists) {
			utils.Conflict(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "User registered successfully", res)
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var req services.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, utils.FormatValidationError(err))
		return
	}

	res, err := h.authService.SignIn(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, utils.ErrInvalidCredentials) {
			utils.Unauthorized(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Signed in successfully", res)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req services.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, utils.FormatValidationError(err))
		return
	}

	res, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, utils.ErrInvalidRefreshToken) {
			utils.Unauthorized(c, err.Error())
			return
		}
		if errors.Is(err, utils.ErrUserNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Tokens refreshed successfully", res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "unauthorized")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		utils.InternalServerError(c, "invalid user context")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Logged out successfully", nil)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "unauthorized")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		utils.InternalServerError(c, "invalid user context")
		return
	}

	res, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, utils.ErrUserNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Profile retrieved successfully", res)
}
