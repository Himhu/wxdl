package handler

import (
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/model"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, authHandler *AuthHandler, cardHandler *CardHandler, agentHandler *AgentHandler, pointsHandler *PointsHandler, auditHandler *AuditHandler, healthHandler *HealthHandler, adminAuthHandler *AdminAuthHandler, adminAgentHandler *AdminAgentHandler, adminMiniProgramConfigHandler *AdminMiniProgramConfigHandler, miniProgramConfigHandler *MiniProgramConfigHandler, adminSystemSettingHandler *AdminSystemSettingHandler, userHandler *UserHandler, jwtConfig config.JWTConfig, adminRepo middleware.AdminPermissionReader) {
	register := func(group *gin.RouterGroup) {
		group.GET("/health", healthHandler.Ping)

		authGroup := group.Group("/auth")
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/wechat/login", authHandler.WechatLogin)

		protectedAuth := authGroup.Group("")
		protectedAuth.Use(middleware.JWTAgentMiddleware(jwtConfig))
		protectedAuth.POST("/wechat/bind", authHandler.BindWechat)
		protectedAuth.GET("/me", authHandler.Me)
		protectedAuth.GET("/profile", authHandler.Me)

		cardGroup := group.Group("/cards")
		cardGroup.Use(middleware.JWTAgentMiddleware(jwtConfig))
		cardGroup.GET("", cardHandler.List)
		cardGroup.GET("/stats", cardHandler.Stats)
		cardGroup.GET("/:id", cardHandler.Detail)
		cardGroup.DELETE("/:id", cardHandler.Destroy)

		agentGroup := group.Group("/agents")
		agentGroup.Use(middleware.JWTAgentMiddleware(jwtConfig))
		agentGroup.POST("", agentHandler.Create)
		agentGroup.GET("", agentHandler.List)
		agentGroup.GET("/:id", agentHandler.Detail)
		agentGroup.PUT("/:id", agentHandler.Update)
		agentGroup.PUT("/:id/status", agentHandler.UpdateStatus)

		pointsGroup := group.Group("/points")
		pointsGroup.Use(middleware.JWTAgentMiddleware(jwtConfig))
		pointsGroup.GET("/balance", pointsHandler.Balance)
		pointsGroup.GET("/records", pointsHandler.Records)
		pointsGroup.GET("/stats", pointsHandler.Stats)
		pointsGroup.POST("/recharge/apply", pointsHandler.ApplyRecharge)
		pointsGroup.GET("/recharge/pending", pointsHandler.PendingRechargeRequests)
		pointsGroup.POST("/recharge/approve/:id", pointsHandler.ApproveRecharge)
		pointsGroup.POST("/recharge/reject/:id", pointsHandler.RejectRecharge)
		pointsGroup.GET("/recharge/history", pointsHandler.RechargeHistory)

		auditGroup := group.Group("/audit")
		auditGroup.Use(middleware.JWTUserMiddleware(jwtConfig))
		auditGroup.GET("/logs", auditHandler.ListLogs)

		adminGroup := group.Group("/admin")
		adminAuthGroup := adminGroup.Group("/auth")
		adminAuthGroup.POST("/login", adminAuthHandler.Login)

		protectedAdmin := adminAuthGroup.Group("")
		protectedAdmin.Use(middleware.JWTAdminMiddleware(jwtConfig))
		protectedAdmin.GET("/me", adminAuthHandler.Me)

		// 小程序配置管理
		miniProgramGroup := adminGroup.Group("/mini-program")
		miniProgramGroup.Use(middleware.JWTAdminMiddleware(jwtConfig))
		miniProgramGroup.GET("/configs",
			middleware.RequireAdminPermission(adminRepo, model.PermissionMiniProgramConfigRead, model.PermissionMiniProgramAll, model.PermissionAll),
			adminMiniProgramConfigHandler.ListConfigs)
		miniProgramGroup.GET("/configs/:id",
			middleware.RequireAdminPermission(adminRepo, model.PermissionMiniProgramConfigRead, model.PermissionMiniProgramAll, model.PermissionAll),
			adminMiniProgramConfigHandler.GetConfig)
		miniProgramGroup.PUT("/configs/:id",
			middleware.RequireAdminPermission(adminRepo, model.PermissionMiniProgramConfigWrite, model.PermissionMiniProgramAll, model.PermissionAll),
			adminMiniProgramConfigHandler.UpdateConfig)

		// 系统设置管理
		systemSettingGroup := adminGroup.Group("/system-settings")
		systemSettingGroup.Use(middleware.JWTAdminMiddleware(jwtConfig))
		systemSettingGroup.GET("/wechat",
			middleware.RequireAdminPermission(adminRepo, model.PermissionSystemWechatRead, model.PermissionSystemAll, model.PermissionAll),
			adminSystemSettingHandler.GetWeChatSettings)
		systemSettingGroup.PUT("/wechat",
			middleware.RequireAdminPermission(adminRepo, model.PermissionSystemWechatWrite, model.PermissionSystemAll, model.PermissionAll),
			adminSystemSettingHandler.UpdateWeChatSettings)
		systemSettingGroup.GET("/object-storage",
			middleware.RequireAdminPermission(adminRepo, model.PermissionSystemAll, model.PermissionAll),
			adminSystemSettingHandler.GetObjectStorageSettings)
		systemSettingGroup.PUT("/object-storage",
			middleware.RequireAdminPermission(adminRepo, model.PermissionSystemAll, model.PermissionAll),
			adminSystemSettingHandler.UpdateObjectStorageSettings)

		// 小程序公开API
		miniappGroup := group.Group("/miniapp")
		miniappGroup.GET("/config/bootstrap", miniProgramConfigHandler.GetBootstrapConfig)
		miniappGroup.POST("/upload-avatar", userHandler.PublicUploadAvatar)

		// 小程序用户认证（新用户体系）
		userAuthGroup := group.Group("/user/auth")
		userAuthGroup.POST("/wechat/login", userHandler.WechatLogin)

		protectedUser := userAuthGroup.Group("")
		protectedUser.Use(middleware.JWTUserMiddleware(jwtConfig))
		protectedUser.GET("/profile", userHandler.Profile)
		protectedUser.GET("/invite", userHandler.UserCurrentInvite)
		protectedUser.GET("/invite/mini-code", userHandler.UserInviteMiniCode)
		protectedUser.POST("/apply-agent", userHandler.UserApplyAgent)
		protectedUser.PUT("/avatar", userHandler.UpdateAvatar)

		// 数据转移
		dataTransferGroup := group.Group("/user/data-transfer")
		dataTransferGroup.Use(middleware.JWTUserMiddleware(jwtConfig))
		dataTransferGroup.POST("/legacy/balance", userHandler.LegacyBalance)
		dataTransferGroup.POST("/legacy/confirm", userHandler.ConfirmLegacyTransfer)

		// 文件上传
		fileGroup := group.Group("/user/files")
		fileGroup.Use(middleware.JWTUserMiddleware(jwtConfig))
		fileGroup.POST("/upload-image", userHandler.UploadImage)

		// 管理员用户管理
		adminUserGroup := adminGroup.Group("/users")
		adminUserGroup.Use(middleware.JWTAdminMiddleware(jwtConfig))
		adminUserGroup.GET("", userHandler.AdminListUsers)
		adminUserGroup.GET("/:id", userHandler.AdminGetUser)
		adminUserGroup.PUT("/:id/role", userHandler.AdminUpdateUserRole)
		adminUserGroup.GET("/applications", userHandler.AdminListApplications)
		adminUserGroup.POST("/applications/:id/review", userHandler.AdminReviewApplication)

		adminAgentGroup := adminGroup.Group("/agents")
		adminAgentGroup.Use(middleware.JWTAdminMiddleware(jwtConfig))
		adminAgentGroup.GET("", adminAgentHandler.List)
		adminAgentGroup.PUT("/:id/status", adminAgentHandler.UpdateStatus)
	}

	register(router.Group("/api"))
	register(router.Group("/api/v1"))
}
