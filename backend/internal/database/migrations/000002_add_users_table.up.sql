-- 统一用户表（微信登录自动注册）
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    open_id VARCHAR(128) NOT NULL UNIQUE COMMENT '微信openId',
    union_id VARCHAR(128) DEFAULT '' COMMENT '微信unionId',
    nickname VARCHAR(100) DEFAULT '微信用户' COMMENT '微信昵称',
    avatar VARCHAR(500) DEFAULT '' COMMENT '头像URL',
    mobile VARCHAR(20) DEFAULT '' COMMENT '手机号',
    role VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '角色: user/agent',
    agent_level TINYINT DEFAULT NULL COMMENT '代理等级: 1=一级 2=二级 3=三级',
    parent_user_id BIGINT UNSIGNED DEFAULT NULL COMMENT '上级用户ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 0=禁用',
    last_login_at DATETIME DEFAULT NULL COMMENT '最后登录时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_role (role),
    INDEX idx_parent_user_id (parent_user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='统一用户表';

-- 角色变更日志表
CREATE TABLE IF NOT EXISTS user_role_change_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    old_role VARCHAR(20) NOT NULL COMMENT '原角色',
    new_role VARCHAR(20) NOT NULL COMMENT '新角色',
    old_agent_level TINYINT DEFAULT NULL COMMENT '原代理等级',
    new_agent_level TINYINT DEFAULT NULL COMMENT '新代理等级',
    changed_by_admin_id BIGINT UNSIGNED DEFAULT NULL COMMENT '操作管理员ID',
    remark VARCHAR(500) DEFAULT '' COMMENT '备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色变更日志';
