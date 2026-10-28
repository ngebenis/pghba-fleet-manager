package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db    *gorm.DB
	cfg   *config.Config
	audit *services.AuditService
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config, audit *services.AuditService) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg, audit: audit}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("username = ? AND status = ?", req.Username, models.StatusActive).First(&user).Error; err != nil {
		h.audit.Log(nil, req.Username, models.EventLoginFailed, req.Username, nil, "", c.ClientIP(), nil)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.audit.Log(nil, req.Username, models.EventLoginFailed, req.Username, nil, "", c.ClientIP(), nil)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, err := middleware.GenerateAccessToken(&user, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	rawRefresh := middleware.GenerateRefreshToken()
	refreshHash := middleware.HashToken(rawRefresh)
	rt := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(h.cfg.JWTRefreshExpiry),
	}
	h.db.Create(rt)

	// Update last_login
	now := time.Now()
	h.db.Model(&user).Update("last_login", now)

	h.audit.Log(&user.ID, user.Username, models.EventLogin, user.Username, nil, "", c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": rawRefresh,
		"expires_in":    int(h.cfg.JWTAccessExpiry.Seconds()),
		"user": gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"full_name":            user.FullName,
			"email":                user.Email,
			"role":                 user.Role,
			"theme":                user.Theme,
			"must_change_password": user.MustChangePassword,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	type logoutReq struct {
		RefreshToken string `json:"refresh_token"`
	}
	var req logoutReq
	_ = c.ShouldBindJSON(&req)
	if req.RefreshToken != "" {
		h.db.Where("token_hash = ?", middleware.HashToken(req.RefreshToken)).Delete(&models.RefreshToken{})
	}
	claims := middleware.GetClaims(c)
	if claims != nil {
		h.audit.Log(&claims.UserID, claims.Username, models.EventLogout, "", nil, "", c.ClientIP(), nil)
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	type req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash := middleware.HashToken(r.RefreshToken)
	var rt models.RefreshToken
	if err := h.db.Where("token_hash = ? AND revoked = false AND expires_at > NOW()", hash).First(&rt).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	var user models.User
	h.db.First(&user, "id = ?", rt.UserID)
	token, _ := middleware.GenerateAccessToken(&user, h.cfg)
	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

// ChangePassword lets a logged-in user set their own new password. The
// only way to clear must_change_password (see migration 006), so this is
// reachable even while that flag blocks every other protected endpoint.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claims := middleware.GetClaims(c)
	type req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"      binding:"required,min=8"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}
	if r.NewPassword == r.CurrentPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different from the current password"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(r.NewPassword), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"password":             string(hashed),
		"must_change_password": false,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Issue a fresh token — the old one still carries must_change_password
	// in its claims and would keep getting blocked until it expires.
	user.MustChangePassword = false
	newToken, err := middleware.GenerateAccessToken(&user, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password changed but failed to issue new token — please log in again"})
		return
	}

	h.audit.Log(&user.ID, user.Username, models.EventEditUser, user.Username, nil, "", c.ClientIP(),
		map[string]string{"action": "change_own_password"})

	c.JSON(http.StatusOK, gin.H{
		"access_token": newToken,
		"expires_in":   int(h.cfg.JWTAccessExpiry.Seconds()),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var user models.User
	h.db.First(&user, "id = ?", claims.UserID)
	c.JSON(http.StatusOK, user)
}
