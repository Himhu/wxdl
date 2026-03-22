package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService           *service.UserService
	legacySiteService     *service.LegacySiteService
	legacyTransferService *service.LegacyTransferService
	wechatClient          *utils.WeChatClient
	jwtConfig             config.JWTConfig
}

func NewUserHandler(userService *service.UserService, legacySiteService *service.LegacySiteService, legacyTransferService *service.LegacyTransferService, wechatClient *utils.WeChatClient, jwtConfig config.JWTConfig) *UserHandler {
	return &UserHandler{
		userService:           userService,
		legacySiteService:     legacySiteService,
		legacyTransferService: legacyTransferService,
		wechatClient:          wechatClient,
		jwtConfig:             jwtConfig,
	}
}

// WechatLogin 微信登录（自动注册）
// POST /api/user/auth/wechat/login
func (h *UserHandler) WechatLogin(c *gin.Context) {
	var req struct {
		Code       string `json:"code" binding:"required"`
		Nickname   string `json:"nickname"`
		Avatar     string `json:"avatar"`
		InviteCode string `json:"inviteCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("参数错误: code 不能为空"))
		return
	}

	// 用 code 换取 openId
	session, err := h.wechatClient.GetSession(c.Request.Context(), req.Code)
	if err != nil {
		utils.Error(c, utils.InternalError(fmt.Errorf("微信登录失败: %w", err)))
		return
	}

	// 自动注册或登录
	user, err := h.userService.WechatLoginOrRegister(
		c.Request.Context(),
		session.OpenID,
		session.UnionID,
		req.Nickname,
		req.Avatar,
	)
	if err != nil {
		utils.Error(c, utils.InternalError(fmt.Errorf("登录处理失败: %w", err)))
		return
	}

	// 生成 JWT（role = "user"，复用 AgentID 字段存 userID）
	token, err := utils.GenerateTokenWithRole(h.jwtConfig, uint(user.ID), user.Nickname, utils.RoleUser)
	if err != nil {
		utils.Error(c, utils.InternalError(fmt.Errorf("生成令牌失败: %w", err)))
		return
	}

	utils.Success(c, "登录成功", gin.H{
		"token":    token,
		"userInfo": user.ToResponse(),
	})
}

// Profile 获取当前用户信息
// GET /api/user/auth/profile
func (h *UserHandler) Profile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), uint64(userID))
	if err != nil {
		utils.Error(c, utils.NotFoundError("用户不存在"))
		return
	}

	invite, _ := h.userService.GetOrCreateInvite(c.Request.Context(), uint64(userID))
	pendingApplication, _ := h.userService.GetPendingApplication(c.Request.Context(), uint64(userID))

	userInfo := user.ToResponse()
	if agentID, ok := middleware.GetAgentID(c); ok {
		agent, err := h.userService.GetAgentByID(c.Request.Context(), uint64(agentID))
		if err == nil && agent != nil {
			userInfo.Role = model.UserRoleAgent
			userInfo.AgentBalance = agent.Balance.StringFixed(2)
		}
	}

	utils.Success(c, "ok", gin.H{
		"userInfo":           userInfo,
		"inviteCode":         func() string { if invite != nil { return invite.Code }; return "" }(),
		"pendingApplication": pendingApplication,
	})
}

// AdminListUsers 管理员查询用户列表
// GET /api/admin/users
func (h *UserHandler) AdminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	role := c.Query("role")
	keyword := c.Query("keyword")

	users, total, err := h.userService.ListUsers(c.Request.Context(), page, pageSize, role, keyword)
	if err != nil {
		utils.Error(c, utils.InternalError(fmt.Errorf("查询失败: %w", err)))
		return
	}

	list := make([]interface{}, 0, len(users))
	for _, u := range users {
		list = append(list, u.ToResponse())
	}

	utils.Success(c, "ok", gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// AdminGetUser 管理员查看用户详情
// GET /api/admin/users/:id
func (h *UserHandler) AdminGetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.BadRequestError("无效的用户ID"))
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, utils.NotFoundError("用户不存在"))
		return
	}

	utils.Success(c, "ok", gin.H{
		"userInfo": user.ToResponse(),
	})
}

// AdminUpdateUserRole 管理员修改用户角色
// PUT /api/admin/users/:id/role
func (h *UserHandler) AdminUpdateUserRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.BadRequestError("无效的用户ID"))
		return
	}

	var req struct {
		Role   string `json:"role" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("参数错误"))
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("管理员未登录"))
		return
	}

	user, err := h.userService.UpdateUserRole(
		c.Request.Context(),
		id,
		req.Role,
		uint64(adminID),
		req.Remark,
	)
	if err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	utils.Success(c, "角色更新成功", gin.H{
		"userInfo": user.ToResponse(),
	})
}

func (h *UserHandler) UserApplyAgent(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	var req struct {
		InviteCode string `json:"inviteCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("参数错误"))
		return
	}

	application, err := h.userService.ApplyAgent(c.Request.Context(), uint64(userID), req.InviteCode)
	if err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	utils.Success(c, "申请已提交", gin.H{
		"application": application,
	})
}

func buildInviteSharePath(inviteCode string) string {
	if inviteCode == "" {
		return "/pages/common/login/index"
	}
	return "/pages/common/login/index?inviteCode=" + inviteCode
}

func (h *UserHandler) UserCurrentInvite(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	invite, err := h.userService.GetOrCreateInvite(c.Request.Context(), uint64(userID))
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}

	records, err := h.userService.ListMyInvitedApplications(c.Request.Context(), uint64(userID))
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}

	sharePath := ""
	miniProgramCodeURL := ""
	if invite != nil {
		sharePath = buildInviteSharePath(invite.Code)
		miniProgramCodeURL = "/api/v1/user/auth/invite/mini-code"
	}

	utils.Success(c, "ok", gin.H{
		"invite":             invite,
		"records":            records,
		"sharePath":          sharePath,
		"miniProgramCodeUrl": miniProgramCodeURL,
	})
}

func (h *UserHandler) UserInviteMiniCode(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	invite, err := h.userService.GetOrCreateInvite(c.Request.Context(), uint64(userID))
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}
	if invite == nil || invite.Code == "" || invite.Status != model.InviteStatusActive {
		utils.Error(c, utils.BadRequestError("邀请码不可用"))
		return
	}

	codeResp, err := h.wechatClient.GetUnlimitedMiniProgramCode(c.Request.Context(), "inviteCode="+invite.Code)
	if err != nil {
		utils.Error(c, utils.InternalError(fmt.Errorf("获取小程序码失败: %w", err)))
		return
	}

	c.Header("Content-Type", codeResp.ContentType)
	c.Header("Cache-Control", "private, max-age=300")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"invite-%s.png\"", invite.Code))
	c.Data(http.StatusOK, codeResp.ContentType, codeResp.Data)
}

func (h *UserHandler) AdminListApplications(c *gin.Context) {
	status := c.Query("status")
	applications, err := h.userService.ListApplications(c.Request.Context(), status)
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}
	utils.Success(c, "ok", gin.H{"list": applications})
}

func (h *UserHandler) AdminReviewApplication(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.BadRequestError("无效的申请ID"))
		return
	}

	var req struct {
		Approved     bool   `json:"approved"`
		RejectReason string `json:"rejectReason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("参数错误"))
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("管理员未登录"))
		return
	}

	application, err := h.userService.ReviewApplication(c.Request.Context(), id, req.Approved, req.RejectReason, uint64(adminID))
	if err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	utils.Success(c, "操作成功", gin.H{"application": application})
}

// LegacyBalance 查询旧站余额
// POST /api/v1/user/data-transfer/legacy/balance
func (h *UserHandler) LegacyBalance(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("请输入旧站账号和密码"))
		return
	}

	result, err := h.legacySiteService.FetchBalance(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	utils.Success(c, "查询成功", gin.H{
		"userId":            userID,
		"legacyUserId":      result.LegacyUserID,
		"legacyUsername":    result.LegacyUsername,
		"balance":           result.Balance,
		"role":              result.Role,
		"agentType":         result.AgentType,
		"svipRemainingDays": result.SVIPRemainingDays,
	})
}

// ConfirmLegacyTransfer 确认将旧站余额转移到新站
// POST /api/v1/user/data-transfer/legacy/confirm
func (h *UserHandler) ConfirmLegacyTransfer(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("请输入旧站账号和密码"))
		return
	}

	result, err := h.legacyTransferService.Confirm(c.Request.Context(), service.LegacyTransferConfirmInput{
		UserID:         uint64(userID),
		LegacyUsername: req.Username,
		LegacyPassword: req.Password,
	})
	if err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	token, err := utils.GenerateTokenWithRole(h.jwtConfig, result.Agent.ID, result.User.Nickname, utils.RoleAgent)
	if err != nil {
		utils.Error(c, utils.InternalError(fmt.Errorf("生成新令牌失败: %w", err)))
		return
	}

	userInfo := result.User.ToResponse()
	userInfo.Role = model.UserRoleAgent
	userInfo.AgentBalance = result.NewBalance

	utils.Success(c, "转移成功", gin.H{
		"token":             token,
		"userInfo":          userInfo,
		"legacyUserId":      result.LegacyUserID,
		"legacyUsername":    result.LegacyUsername,
		"transferredAmount": result.TransferredAmount,
		"legacyDisabled":    result.LegacyDisabled,
		"newBalance":        result.NewBalance,
	})
}

// UploadImage 上传图片到对象存储
// POST /api/v1/user/files/upload-image
func (h *UserHandler) UploadImage(c *gin.Context) {
	_, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.Error(c, utils.BadRequestError("请上传图片文件"))
		return
	}
	defer file.Close()

	scene := c.DefaultPostForm("scene", "general")

	ossCfg, err := h.userService.GetObjectStorageConfig(c.Request.Context())
	if err != nil || !ossCfg.Enabled {
		utils.Error(c, utils.BadRequestError("对象存储未开启"))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	result, err := utils.UploadToS3Compatible(c.Request.Context(), *ossCfg, file, contentType, scene)
	if err != nil {
		utils.Error(c, utils.BadRequestError(fmt.Sprintf("上传失败: %v", err)))
		return
	}

	utils.Success(c, "上传成功", gin.H{
		"url": result.URL,
		"key": result.Key,
	})
}

// UpdateAvatar 更新用户头像
// PUT /api/v1/user/auth/avatar
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	var req struct {
		Avatar string `json:"avatar" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("请提供头像 URL"))
		return
	}

	if err := h.userService.UpdateAvatar(c.Request.Context(), uint64(userID), req.Avatar); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	utils.Success(c, "头像更新成功", gin.H{"avatar": req.Avatar})
}

// PublicUploadAvatar 公开头像上传（不需要 token，但有严格限制）
// POST /api/miniapp/upload-avatar
func (h *UserHandler) PublicUploadAvatar(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.Error(c, utils.BadRequestError("请上传图片文件"))
		return
	}
	defer file.Close()

	if header.Size > 2*1024*1024 {
		utils.Error(c, utils.BadRequestError("头像文件不能超过 2MB"))
		return
	}

	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
	}
	if !allowedTypes[contentType] {
		utils.Error(c, utils.BadRequestError("只支持 jpg/png/gif/webp 格式"))
		return
	}

	ossCfg, err := h.userService.GetObjectStorageConfig(c.Request.Context())
	if err != nil || !ossCfg.Enabled {
		utils.Error(c, utils.BadRequestError("对象存储未开启"))
		return
	}

	result, err := utils.UploadToS3Compatible(c.Request.Context(), *ossCfg, file, contentType, "avatar")
	if err != nil {
		utils.Error(c, utils.BadRequestError(fmt.Sprintf("上传失败: %v", err)))
		return
	}

	utils.Success(c, "上传成功", gin.H{
		"url": result.URL,
		"key": result.Key,
	})
}
