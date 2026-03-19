package middleware

import (
	"fmt"

	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered",
			zap.String("error", fmt.Sprintf("%v", recovered)),
			zap.String("path", c.Request.URL.Path),
		)

		utils.Error(c, utils.InternalError(nil))
	})
}
