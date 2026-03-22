ALTER TABLE users
    DROP COLUMN agent_level,
    DROP COLUMN parent_user_id,
    ADD COLUMN inviter_user_id BIGINT UNSIGNED DEFAULT NULL COMMENT '邀请人用户ID' AFTER role,
    ADD INDEX idx_inviter_user_id (inviter_user_id);

ALTER TABLE user_role_change_logs
    DROP COLUMN old_agent_level,
    DROP COLUMN new_agent_level;

CREATE TABLE IF NOT EXISTS user_invites (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(32) NOT NULL UNIQUE COMMENT '邀请码',
    inviter_user_id BIGINT UNSIGNED NOT NULL COMMENT '邀请人ID',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active/inactive',
    expires_at DATETIME DEFAULT NULL COMMENT '过期时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_inviter_user_id (inviter_user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户邀请表';

CREATE TABLE IF NOT EXISTS agent_applications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '申请用户ID',
    inviter_user_id BIGINT UNSIGNED DEFAULT NULL COMMENT '邀请人ID',
    invite_code VARCHAR(32) DEFAULT '' COMMENT '邀请码',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态: pending/approved/rejected',
    reject_reason VARCHAR(500) DEFAULT '' COMMENT '驳回原因',
    reviewed_by_admin_id BIGINT UNSIGNED DEFAULT NULL COMMENT '审核管理员ID',
    reviewed_at DATETIME DEFAULT NULL COMMENT '审核时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_inviter_user_id (inviter_user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理申请表';
