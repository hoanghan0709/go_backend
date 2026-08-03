package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	common "github.com/han/go-ecommerce/internal/common/model"
	"github.com/han/go-ecommerce/internal/token"
)

const UserIDContextKey = "userID"

func RequireAuth(tokenService *token.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")

		parts := strings.Fields(authorization)
		if len(parts) != 2 ||

			!strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				common.Response{
					StatusCode: http.StatusUnauthorized,
					Message:    "Missing or invalid authorization header",
				},
			)

			return
		}

		claims, err := tokenService.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				common.Response{
					StatusCode: http.StatusUnauthorized,
					Message:    "Invalid or expired token",
				},
			)
			return
		}

		c.Set(UserIDContextKey, claims.UserID)
		c.Next()
	}
}
