package middleware

import (
	"context"
	"strings"

	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AdminPermissionReader defines the interface for reading admin permissions
type AdminPermissionReader interface {
	GetPermissions(ctx context.Context, adminID uint) ([]string, error)
}

// RequireAdminPermission creates a middleware that checks if the current admin has any of the required permissions
func RequireAdminPermission(repo AdminPermissionReader, required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, ok := GetAdminID(c)
		if !ok {
			utils.Error(c, utils.UnauthorizedError("failed to resolve current admin"))
			c.Abort()
			return
		}

		permissions, err := repo.GetPermissions(c.Request.Context(), adminID)
		if err != nil {
			utils.Error(c, utils.InternalError(err))
			c.Abort()
			return
		}

		if !hasAnyPermission(permissions, required...) {
			utils.Error(c, utils.ForbiddenError("insufficient permissions"))
			c.Abort()
			return
		}

		c.Set("adminPermissions", permissions)
		c.Next()
	}
}

// hasAnyPermission checks if the admin has any of the required permissions
func hasAnyPermission(granted []string, required ...string) bool {
	for _, need := range required {
		if hasPermission(granted, need) {
			return true
		}
	}
	return false
}

// hasPermission checks if a specific permission is granted, supporting wildcard matching
func hasPermission(granted []string, required string) bool {
	for _, perm := range granted {
		// Exact match
		if perm == required {
			return true
		}
		// Full wildcard
		if perm == "*" {
			return true
		}
		// Prefix wildcard (only allow colon-based wildcards like "system:*")
		if strings.HasSuffix(perm, ":*") {
			prefix := strings.TrimSuffix(perm, "*")
			if strings.HasPrefix(required, prefix) {
				return true
			}
		}
	}
	return false
}
