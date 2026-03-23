package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := utils.NewLogger(cfg.Log, cfg.App.Env)
	if err != nil {
		panic(err)
	}
	defer utils.SyncLogger(logger)

	db, err := config.InitDatabase(cfg.Database, logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	if err := database.RunMigrations(db, logger); err != nil {
		logger.Fatal("failed to run database migrations", zap.Error(err))
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	agentRepository := repository.NewAgentRepository(db)
	cardRepository := repository.NewCardRepository(db)
	pointsRepository := repository.NewPointsRepository(db)
	adminRepository := repository.NewAdminRepository(db)
	miniProgramConfigRepository := repository.NewMiniProgramConfigRepository(db)
	systemSettingRepository := repository.NewSystemSettingRepository(db)
	userRepository := repository.NewUserRepository(db)
	txManager := repository.NewTxManager(db)

	// 初始化配置加密工具
	masterKey := strings.TrimSpace(os.Getenv("APP_CONFIG_MASTER_KEY"))
	var cipher utils.SecretCipher

	if masterKey == "" {
		if cfg.App.Env != "development" {
			logger.Fatal("APP_CONFIG_MASTER_KEY is required outside development")
		}

		logger.Warn("APP_CONFIG_MASTER_KEY is not set; system setting encryption is disabled in development")
		cipher = utils.NewDisabledSecretCipher("APP_CONFIG_MASTER_KEY is not configured")
	} else {
		cipher, err = utils.NewSecretCipher(masterKey, "v1")
		if err != nil {
			logger.Fatal("failed to initialize secret cipher", zap.Error(err))
		}
	}

	// 初始化运行时配置提供者
	runtimeProvider := service.NewRuntimeSettingsProvider(systemSettingRepository, cipher, cfg.WeChat)

	// 初始化微信配置适配器和客户端
	wechatConfigAdapter := service.NewWeChatConfigAdapter(runtimeProvider)
	wechatClient := utils.NewWeChatClient(wechatConfigAdapter, cfg.WeChat.TimeoutSeconds)

	authService := service.NewAuthService(agentRepository, cfg.JWT, wechatClient)
	agentService := service.NewAgentService(agentRepository)
	cardService := service.NewCardService(agentService, cardRepository, txManager)
	pointsService := service.NewPointsService(agentService, pointsRepository, txManager)
	adminAuthService := service.NewAdminAuthService(adminRepository, cfg.JWT)
	miniProgramConfigService := service.NewMiniProgramConfigService(miniProgramConfigRepository)
	systemSettingService := service.NewSystemSettingService(systemSettingRepository, txManager, cipher, runtimeProvider)
	userService := service.NewUserService(userRepository, agentRepository, systemSettingRepository, cipher)
	legacySiteService := service.NewLegacySiteService("http://dl.jiexi6.cn", 10)
	legacyTransferService := service.NewLegacyTransferService(legacySiteService, userRepository, agentRepository, txManager, "admin", "wugui1996")

	authHandler := handler.NewAuthHandler(authService)
	cardHandler := handler.NewCardHandler(cardService, cardRepository, systemSettingService)
	agentHandler := handler.NewAgentHandler(agentService)
	pointsHandler := handler.NewPointsHandler(pointsService)
	adminAuthHandler := handler.NewAdminAuthHandler(adminAuthService)
	adminMiniProgramConfigHandler := handler.NewAdminMiniProgramConfigHandler(miniProgramConfigService)
	adminSystemSettingHandler := handler.NewAdminSystemSettingHandler(systemSettingService)
	miniProgramConfigHandler := handler.NewMiniProgramConfigHandler(miniProgramConfigService)
	adminAgentHandler := handler.NewAdminAgentHandler(agentRepository)
	adminCardHandler := handler.NewAdminCardHandler(systemSettingService, cardRepository)
	agentCardHandler := handler.NewAgentCardHandler(systemSettingService, agentRepository, cardRepository, txManager)
	auditHandler := handler.NewAuditHandler(agentRepository, userRepository, db)
	healthHandler := handler.NewHealthHandler(cfg)
	userHandler := handler.NewUserHandler(userService, legacySiteService, legacyTransferService, wechatClient, cfg.JWT)
	dashboardHandler := handler.NewDashboardHandler(db)

	router := gin.New()
	router.Use(
		cors.New(cors.Config{
			AllowOrigins:     cfg.CORS.AllowOrigins,
			AllowMethods:     cfg.CORS.AllowMethods,
			AllowHeaders:     cfg.CORS.AllowHeaders,
			ExposeHeaders:    cfg.CORS.ExposeHeaders,
			AllowCredentials: cfg.CORS.AllowCredentials,
			MaxAge:           time.Duration(cfg.CORS.MaxAge) * time.Second,
		}),
		middleware.RequestLogger(logger),
		middleware.Recovery(logger),
	)
	handler.RegisterRoutes(router, authHandler, cardHandler, agentHandler, pointsHandler, auditHandler, healthHandler, adminAuthHandler, adminAgentHandler, adminCardHandler, agentCardHandler, adminMiniProgramConfigHandler, miniProgramConfigHandler, adminSystemSettingHandler, userHandler, dashboardHandler, cfg.JWT, adminRepository)

	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	go func() {
		logger.Info("server started",
			zap.String("addr", cfg.Server.Address()),
			zap.String("env", cfg.App.Env),
		)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server exited unexpectedly", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to gracefully shutdown server", zap.Error(err))
		return
	}

	logger.Info("server stopped")
}
