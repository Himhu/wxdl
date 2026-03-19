package handler

import (
	"time"

	"backend/internal/config"
	"backend/pkg/version"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	config *config.Config
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{config: cfg}
}

func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"service": h.config.App.Name,
			"env":     h.config.App.Env,
			"version": version.Version,
			"time":    time.Now().Format(time.RFC3339),
		},
	})
}
