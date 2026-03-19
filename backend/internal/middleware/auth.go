package middleware

import (
	"strings"

	"backend/internal/config"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextAgentIDKey  = "agentID"
	ContextAdminIDKey  = "adminID"
	ContextUserIDKey   = "userID"
	ContextUsernameKey = "username"
	ContextRoleKey     = "role"
)

func jwtRoleMiddleware(cfg config.JWTConfig, allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			utils.Error(c, utils.UnauthorizedError("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.Error(c, utils.UnauthorizedError("invalid authorization header"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(cfg.Secret, parts[1])
		if err != nil {
			utils.Error(c, utils.UnauthorizedError("invalid or expired token"))
			c.Abort()
			return
		}

		if claims.Role == "" {
			utils.Error(c, utils.UnauthorizedError("token role is missing"))
			c.Abort()
			return
		}

		if _, ok := allowed[claims.Role]; !ok {
			utils.Error(c, utils.ForbiddenError("insufficient permissions"))
			c.Abort()
			return
		}

		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextRoleKey, claims.Role)

		switch claims.Role {
		case utils.RoleAgent:
			c.Set(ContextAgentIDKey, claims.AgentID)
		case utils.RoleAdmin:
			c.Set(ContextAdminIDKey, claims.AgentID)
		case utils.RoleUser:
			c.Set(ContextUserIDKey, claims.AgentID)
		}

		c.Next()
	}
}

func JWTAgentMiddleware(cfg config.JWTConfig) gin.HandlerFunc {
	return jwtRoleMiddleware(cfg, utils.RoleAgent)
}

func JWTAdminMiddleware(cfg config.JWTConfig) gin.HandlerFunc {
	return jwtRoleMiddleware(cfg, utils.RoleAdmin)
}

// JWTUserMiddleware 小程序用户认证（允许 user 和 agent 角色）
func JWTUserMiddleware(cfg config.JWTConfig) gin.HandlerFunc {
	return jwtRoleMiddleware(cfg, utils.RoleUser, utils.RoleAgent)
}

func GetAgentID(c *gin.Context) (uint, bool) {
	role, exists := c.Get(ContextRoleKey)
	if !exists || role != utils.RoleAgent {
		return 0, false
	}

	value, exists := c.Get(ContextAgentIDKey)
	if !exists {
		return 0, false
	}

	agentID, ok := value.(uint)
	return agentID, ok
}

// GetAdminID returns the admin ID from context, only if the role is "admin".
func GetAdminID(c *gin.Context) (uint, bool) {
	role, exists := c.Get(ContextRoleKey)
	if !exists || role != utils.RoleAdmin {
		return 0, false
	}

	value, exists := c.Get(ContextAdminIDKey)
	if !exists {
		return 0, false
	}

	adminID, ok := value.(uint)
	return adminID, ok
}

// GetUserID returns the user ID from context (works for both "user" and "agent" roles).
func GetUserID(c *gin.Context) (uint, bool) {
	role, exists := c.Get(ContextRoleKey)
	if !exists {
		return 0, false
	}

	r, _ := role.(string)
	switch r {
	case utils.RoleUser:
		value, exists := c.Get(ContextUserIDKey)
		if !exists {
			return 0, false
		}
		id, ok := value.(uint)
		return id, ok
	case utils.RoleAgent:
		value, exists := c.Get(ContextAgentIDKey)
		if !exists {
			return 0, false
		}
		id, ok := value.(uint)
		return id, ok
	default:
		return 0, false
	}
}
