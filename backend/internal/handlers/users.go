package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// describeUserWriteError turns a raw DB error from creating/updating a user
// into an accurate HTTP status + user-facing message, instead of always
// blaming "username or email already exists" regardless of what actually
// went wrong. Postgres error codes:
// https://www.postgresql.org/docs/current/errcodes-appendix.html
func describeUserWriteError(err error) (int, string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return http.StatusConflict, "username or email already exists"
		case "22P02": // invalid_text_representation — e.g. a role value the
			// user_role enum doesn't have yet (pending migration)
			return http.StatusBadRequest, "invalid role — the database doesn't recognize this role yet; make sure its migration has been applied"
		}
	}
	return http.StatusInternalServerError, "failed to save user: " + err.Error()
}

type UserHandler struct {
	db    *gorm.DB
	audit *services.AuditService
}

func NewUserHandler(db *gorm.DB, audit *services.AuditService) *UserHandler {
	return &UserHandler{db: db, audit: audit}
}

func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	q := h.db
	if role := c.Query("role"); role != "" {
		q = q.Where("role = ?", role)
	}
	q.Order("created_at").Find(&users)
	c.JSON(http.StatusOK, users)
}

type createUserReq struct {
	Username string           `json:"username"  binding:"required,min=3,max=64"`
	FullName string           `json:"full_name" binding:"required"`
	Email    string           `json:"email"     binding:"required,email"`
	Password string           `json:"password"  binding:"required,min=8"`
	Role     models.UserRole  `json:"role"      binding:"required"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	user := &models.User{
		Username: req.Username,
		FullName: req.FullName,
		Email:    req.Email,
		Password: string(hashed),
		Role:     req.Role,
		Status:   models.StatusActive,
	}
	if err := h.db.Create(user).Error; err != nil {
		status, msg := describeUserWriteError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventAddUser, user.Username, nil, "", c.ClientIP(), nil)
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var user models.User
	if err := h.db.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	type updateReq struct {
		FullName string           `json:"full_name"`
		Email    string           `json:"email"`
		Role     models.UserRole  `json:"role"`
		Status   models.UserStatus `json:"status"`
		Theme    string           `json:"theme"`
	}
	var req updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.FullName != "" { updates["full_name"] = req.FullName }
	if req.Email != ""    { updates["email"] = req.Email }
	if req.Role != ""     { updates["role"] = req.Role }
	if req.Status != ""   { updates["status"] = req.Status }
	if req.Theme != ""    { updates["theme"] = req.Theme }
	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		status, msg := describeUserWriteError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEditUser, user.Username, nil, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	type req struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(r.NewPassword), 12)
	h.db.Model(&models.User{}).Where("id = ?", id).Update("password", string(hashed))
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEditUser, id.String(), nil, "", c.ClientIP(), map[string]string{"action": "reset_password"})
	c.JSON(http.StatusOK, gin.H{"message": "password reset"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	claims := middleware.GetClaims(c)
	if id == claims.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}
	var user models.User
	if err := h.db.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	h.db.Delete(&user)
	h.audit.Log(&claims.UserID, claims.Username, models.EventDeleteUser, user.Username, nil, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
