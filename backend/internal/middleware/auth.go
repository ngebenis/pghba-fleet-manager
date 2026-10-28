package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type Claims struct {
	UserID             uuid.UUID       `json:"uid"`
	Username           string          `json:"sub"`
	Role               models.UserRole `json:"role"`
	MustChangePassword bool            `json:"must_change_password,omitempty"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(user *models.User, cfg *config.Config) (string, error) {
	claims := Claims{
		UserID:             user.ID,
		Username:           user.Username,
		Role:               user.Role,
		MustChangePassword: user.MustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTAccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pghba-manager",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func GenerateRefreshToken() string {
	raw := uuid.New().String() + uuid.New().String()
	return raw
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func ParseToken(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// AuthRequired validates JWT and attaches claims to context
func AuthRequired(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := ParseToken(tokenStr, cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Re-check status/role against the DB on every request instead of
		// trusting what the token carries for its full lifetime (up to
		// JWTAccessExpiry — several hours). Without this, deactivating or
		// demoting a user has no effect until their token happens to
		// expire or they explicitly log out.
		var current struct {
			Status             models.UserStatus
			Role               models.UserRole
			MustChangePassword bool
		}
		if err := db.Table("users").Select("status, role, must_change_password").Where("id = ?", claims.UserID).Scan(&current).Error; err != nil || current.Role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account no longer exists"})
			return
		}
		if current.Status != models.StatusActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account is inactive"})
			return
		}
		claims.Role = current.Role
		claims.MustChangePassword = current.MustChangePassword

		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// requirePasswordChangeAllowlist are the only endpoints a user whose
// account is flagged must_change_password may still reach — just enough
// to see who they are, change their password, and log out.
var requirePasswordChangeAllowlist = map[string]bool{
	"/api/auth/me":              true,
	"/api/auth/logout":          true,
	"/api/auth/change-password": true,
}

// RequirePasswordChanged blocks every other protected endpoint for an
// account flagged must_change_password (see migration 006 — this is how
// the published default Super Admin's well-known password gets forced to
// change instead of staying usable forever on a deployment that never
// rotated it). Must run after AuthRequired.
func RequirePasswordChanged() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims != nil && claims.MustChangePassword && !requirePasswordChangeAllowlist[c.Request.URL.Path] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "password change required before continuing",
				"code":  "MUST_CHANGE_PASSWORD",
			})
			return
		}
		c.Next()
	}
}

// RequireRole enforces minimum role hierarchy
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[models.UserRole]bool)
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if !allowed[role.(models.UserRole)] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

// GetClaims retrieves JWT claims from context (safe)
func GetClaims(c *gin.Context) *Claims {
	v, _ := c.Get("claims")
	claims, _ := v.(*Claims)
	return claims
}
