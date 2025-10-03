// internal/auth/handlers.go
package auth

import (
	"net/http"

	"quizapi/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	svc *Service
	val *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, val: validator.New()}
}

type AuthReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

func (h *Handler) Register(c *gin.Context) {
	var req AuthReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.val.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	// For this example, first user is admin, others are public.
	// In a real app, this logic would be more complex.
	var userCount int64
	h.svc.db.Model(&models.User{}).Count(&userCount)
	role := models.RoleUser
	if userCount == 0 {
		role = models.RoleAdmin
	}

	user, err := h.svc.RegisterUser(req.Username, req.Password, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username, "role": user.Role})
}

func (h *Handler) Login(c *gin.Context) {
	var req AuthReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.svc.LoginUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
