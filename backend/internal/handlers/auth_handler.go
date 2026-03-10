package handlers

import (
	"log"
	"net/http"
	"strconv"

	"smsystem-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	AuthService *services.AuthService
	LogService  *services.LogService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *services.AuthService, logService *services.LogService) *AuthHandler {
	return &AuthHandler{AuthService: authService, LogService: logService}
}

// Register handles user registration.
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	response, err := h.AuthService.Register(input)
	if err != nil {
		log.Printf("[Registration Error] %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "email already registered" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   response.Token,
		"user":    response.User,
	})
}

// Login handles user authentication.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	response, err := h.AuthService.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Log the login action
	h.LogService.Record(response.User.ID, "LOGIN", "System", strconv.Itoa(int(response.User.ID)), "User logged in", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   response.Token,
		"user":    response.User,
	})
}

// GetMe returns the current authenticated user.
// GET /api/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, err := h.AuthService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
