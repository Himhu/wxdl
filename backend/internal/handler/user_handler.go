package handler

import (
	"fmt"
	"strconv"

	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService  *service.UserService
	wechatClient *utils.WeChatClient
	jwtConfig    config.JWTConfig
}

func NewUserHandler(userService *service.UserService, wechatClient *utils.WeChatClient, jwtConfig config.JWTConfig) *UserHandler {
	return &UserHandler{
		userService:  userService,
		wechatClient: wechatClient,
		jwtConfig:    jwtConfig,
	}
}

// WechatLogin 微信登录（自动注册）
// POST /api/user/auth/wechat/login
func (h *UserHandler) WechatLogin(c *gin.Context) {
	var req struct {
		Code     string `json:"code" binding:"required"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
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

	utils.Success(c, "ok", gin.H{
		"userInfo": user.ToResponse(),
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
		Role         string  `json:"role" binding:"required"`
		AgentLevel   *int    `json:"agentLevel"`
		ParentUserID *uint64 `json:"parentUserId"`
		Remark       string  `json:"remark"`
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
		req.AgentLevel,
		req.ParentUserID,
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
