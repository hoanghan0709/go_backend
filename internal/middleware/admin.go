package middleware

import (
	"github.com/gin-gonic/gin"
	auth "github.com/han/go-ecommerce/internal/auth/model"
	"gorm.io/gorm"
	"net/http"
)

// Đọc quyền từ DB thay vì tin role do client gửi hoặc role cũ trong JWT.
// Vì vậy thu hồi quyền có hiệu lực ngay với access token còn hạn.
func RequireAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user auth.User
		id, exists := c.Get(UserIDContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "authentication required"})
			return
		}
		if err := db.First(&user, id).Error; err != nil || user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "admin role required"})
			return
		}
		c.Next()
	}
}
