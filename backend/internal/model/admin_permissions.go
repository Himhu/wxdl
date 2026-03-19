package model

// Admin permission constants
const (
	// System settings permissions
	PermissionSystemWechatRead  = "system:wechat:read"
	PermissionSystemWechatWrite = "system:wechat:write"

	// Mini program config permissions
	PermissionMiniProgramConfigRead  = "mini_program:config:read"
	PermissionMiniProgramConfigWrite = "mini_program:config:write"

	// Wildcard permissions
	PermissionSystemAll      = "system:*"
	PermissionMiniProgramAll = "mini_program:*"
	PermissionAll            = "*"
)
