package handlers

import (
    "net/http"
    "logi/internal/models"
    "logi/internal/services"
    "logi/pkg/auth"

    "github.com/gin-gonic/gin"
)

type UserHandler struct {
    Service     *services.UserService
    AuthService *auth.AuthService
}

func NewUserHandler(service *services.UserService, authService *auth.AuthService) *UserHandler {
    return &UserHandler{
        Service:     service,
        AuthService: authService,
    }
}

func (h *UserHandler) Register(c *gin.Context) {
    var payload struct {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := c.BindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    user := &models.User{
        Name:  payload.Name,
        Email: payload.Email,
    }

    err := h.Service.Register(user, payload.Password)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
    var payload struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := c.BindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    user, err := h.Service.Login(payload.Email, payload.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    token, err := h.AuthService.GenerateJWT(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}
