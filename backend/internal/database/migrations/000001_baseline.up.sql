-- Baseline migration: consolidated schema for agent_system
-- Merges init.sql + admin_rbac.sql + AutoMigrate-only tables
-- Adds missing stations table and foreign key constraints
-- Note: database is already selected via connection string

-- ============================================================
-- 1. stations (referenced by agents, cards)
-- ============================================================
CREATE TABLE IF NOT EXISTS stations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '站点名称',
    code VARCHAR(50) NOT NULL COMMENT '站点编码',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1-启用 2-禁用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='站点表';

-- ============================================================
-- 2. agents
-- ============================================================
CREATE TABLE IF NOT EXISTS agents (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL COMMENT '账号',
    password VARCHAR(255) NOT NULL COMMENT '密码(bcrypt)',
    real_name VARCHAR(50) DEFAULT NULL COMMENT '真实姓名',
    phone VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    level TINYINT NOT NULL DEFAULT 1 COMMENT '1-一级 2-二级 3-三级',
    parent_id BIGINT UNSIGNED DEFAULT NULL COMMENT '上级代理ID',
    balance DECIMAL(20,2) NOT NULL DEFAULT 0.00 COMMENT '积分余额',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 2-禁用',
    station_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '站点ID',
    wechat_open_id VARCHAR(100) DEFAULT NULL COMMENT '微信OpenID',
    wechat_union_id VARCHAR(100) DEFAULT NULL COMMENT '微信UnionID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_username (username),
    UNIQUE KEY uk_wechat_open_id (wechat_open_id),
    KEY idx_parent_id (parent_id),
    KEY idx_level (level),
    KEY idx_status (status),
    KEY idx_station_id (station_id),
    KEY idx_wechat_union_id (wechat_union_id),
    CONSTRAINT fk_agents_parent FOREIGN KEY (parent_id) REFERENCES agents(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理商表';

-- ============================================================
-- 3. cards
-- ============================================================
CREATE TABLE IF NOT EXISTS cards (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    card_key VARCHAR(32) NOT NULL COMMENT '卡密',
    agent_id BIGINT UNSIGNED NOT NULL COMMENT '所属代理商ID',
    station_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '站点ID',
    cost DECIMAL(20,2) NOT NULL DEFAULT 1.00 COMMENT '卡密成本积分',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1-未使用 2-已使用 3-已销毁',
    used_at DATETIME DEFAULT NULL COMMENT '使用时间',
    used_by VARCHAR(100) DEFAULT NULL COMMENT '使用者标识',
    destroyed_at DATETIME DEFAULT NULL COMMENT '销毁时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_card_key (card_key),
    KEY idx_agent_id (agent_id),
    KEY idx_station_id (station_id),
    KEY idx_status (status),
    KEY idx_created_at (created_at),
    CONSTRAINT fk_cards_agent FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='卡密表';

-- ============================================================
-- 4. balance_logs (Go model: PointsRecord)
-- ============================================================
CREATE TABLE IF NOT EXISTS balance_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    agent_id BIGINT UNSIGNED NOT NULL COMMENT '代理商ID',
    type TINYINT NOT NULL COMMENT '1-充值 2-消费 3-退款',
    amount DECIMAL(20,2) NOT NULL COMMENT '金额',
    balance_before DECIMAL(20,2) NOT NULL COMMENT '变动前余额',
    balance_after DECIMAL(20,2) NOT NULL COMMENT '变动后余额',
    remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
    related_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_agent_id (agent_id),
    KEY idx_type (type),
    KEY idx_related_id (related_id),
    KEY idx_created_at (created_at),
    CONSTRAINT fk_balance_logs_agent FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分流水表';
-- PLACEHOLDER_CONTINUE_2

-- ============================================================
-- 5. recharge_requests
-- ============================================================
CREATE TABLE IF NOT EXISTS recharge_requests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    agent_id BIGINT UNSIGNED NOT NULL COMMENT '申请代理商ID',
    amount DECIMAL(20,2) NOT NULL COMMENT '申请充值金额',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0-待审核 1-已通过 2-已拒绝',
    payment_method VARCHAR(50) DEFAULT NULL COMMENT '支付方式',
    payment_proof VARCHAR(500) DEFAULT NULL COMMENT '支付凭证URL',
    remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
    reviewed_by BIGINT UNSIGNED DEFAULT NULL COMMENT '审核人ID',
    reviewed_at DATETIME DEFAULT NULL COMMENT '审核时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_agent_id (agent_id),
    KEY idx_status (status),
    KEY idx_reviewed_by (reviewed_by),
    KEY idx_created_at (created_at),
    CONSTRAINT fk_recharge_requests_agent FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值申请表';

-- ============================================================
-- 6. login_logs
-- ============================================================
CREATE TABLE IF NOT EXISTS login_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    agent_id BIGINT UNSIGNED DEFAULT NULL COMMENT '代理商ID',
    username VARCHAR(50) NOT NULL COMMENT '登录账号',
    login_type TINYINT NOT NULL COMMENT '1-账号密码 2-微信授权',
    status TINYINT NOT NULL COMMENT '0-失败 1-成功',
    ip VARCHAR(50) DEFAULT NULL COMMENT 'IP地址',
    user_agent VARCHAR(255) DEFAULT NULL COMMENT 'User Agent',
    fail_reason VARCHAR(255) DEFAULT NULL COMMENT '失败原因',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_agent_id (agent_id),
    KEY idx_username (username),
    KEY idx_status (status),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';

-- ============================================================
-- 7. mini_program_config_items (previously AutoMigrate-only)
-- ============================================================
CREATE TABLE IF NOT EXISTS mini_program_config_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    namespace VARCHAR(50) NOT NULL COMMENT '命名空间',
    config_key VARCHAR(100) NOT NULL COMMENT '配置键',
    scope_type VARCHAR(20) NOT NULL DEFAULT 'global' COMMENT '作用域类型',
    scope_code VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '作用域编码',
    published_value TEXT DEFAULT NULL COMMENT '已发布值',
    visibility VARCHAR(20) NOT NULL DEFAULT 'public' COMMENT '可见性',
    description VARCHAR(255) DEFAULT NULL COMMENT '描述',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1-启用 2-禁用',
    updated_by BIGINT UNSIGNED DEFAULT NULL COMMENT '更新人ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='小程序配置项表';
-- PLACEHOLDER_CONTINUE_3

-- ============================================================
-- 8. system_settings (previously AutoMigrate-only)
-- ============================================================
CREATE TABLE IF NOT EXISTS system_settings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    category VARCHAR(50) NOT NULL COMMENT '分类',
    setting_key VARCHAR(100) NOT NULL COMMENT '设置键',
    display_name VARCHAR(100) NOT NULL COMMENT '显示名称',
    value_type VARCHAR(20) NOT NULL COMMENT '值类型',
    is_secret TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否敏感',
    value_plain TEXT DEFAULT NULL COMMENT '明文值',
    value_ciphertext TEXT DEFAULT NULL COMMENT '密文值',
    value_masked VARCHAR(255) DEFAULT NULL COMMENT '脱敏值',
    checksum VARCHAR(64) DEFAULT NULL COMMENT '校验和',
    key_version VARCHAR(32) DEFAULT NULL COMMENT '密钥版本',
    source VARCHAR(20) NOT NULL DEFAULT 'database' COMMENT '来源',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1-启用 2-禁用',
    description VARCHAR(255) DEFAULT NULL COMMENT '描述',
    version INT NOT NULL DEFAULT 1 COMMENT '版本号',
    updated_by BIGINT UNSIGNED DEFAULT NULL COMMENT '更新人ID',
    published_at DATETIME DEFAULT NULL COMMENT '发布时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_setting (category, setting_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

-- ============================================================
-- 9. system_setting_revisions (previously AutoMigrate-only)
-- ============================================================
CREATE TABLE IF NOT EXISTS system_setting_revisions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    setting_id BIGINT UNSIGNED NOT NULL COMMENT '设置ID',
    category VARCHAR(50) NOT NULL COMMENT '分类',
    setting_key VARCHAR(100) NOT NULL COMMENT '设置键',
    version INT NOT NULL COMMENT '版本号',
    operation VARCHAR(20) NOT NULL COMMENT '操作类型',
    old_value_masked VARCHAR(255) DEFAULT NULL COMMENT '旧值(脱敏)',
    new_value_masked VARCHAR(255) DEFAULT NULL COMMENT '新值(脱敏)',
    old_checksum VARCHAR(64) DEFAULT NULL COMMENT '旧校验和',
    new_checksum VARCHAR(64) DEFAULT NULL COMMENT '新校验和',
    change_note VARCHAR(255) DEFAULT NULL COMMENT '变更说明',
    changed_by BIGINT UNSIGNED DEFAULT NULL COMMENT '变更人ID',
    changed_ip VARCHAR(64) DEFAULT NULL COMMENT '变更IP',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_setting_id (setting_id),
    CONSTRAINT fk_revisions_setting FOREIGN KEY (setting_id) REFERENCES system_settings(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置变更记录表';

-- ============================================================
-- 10. admin_roles
-- ============================================================
CREATE TABLE IF NOT EXISTS admin_roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL COMMENT '角色编码',
    name VARCHAR(100) NOT NULL COMMENT '角色名称',
    description VARCHAR(255) DEFAULT NULL COMMENT '角色描述',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1-启用 2-禁用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_code (code),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员角色表';
-- PLACEHOLDER_CONTINUE_4

-- ============================================================
-- 11. admin_users
-- ============================================================
CREATE TABLE IF NOT EXISTS admin_users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL COMMENT '管理员账号',
    password VARCHAR(255) NOT NULL COMMENT '密码(bcrypt)',
    real_name VARCHAR(50) DEFAULT NULL COMMENT '真实姓名',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1-启用 2-禁用',
    role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    last_login_at DATETIME DEFAULT NULL COMMENT '最后登录时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_username (username),
    KEY idx_role_id (role_id),
    KEY idx_status (status),
    CONSTRAINT fk_admin_users_role FOREIGN KEY (role_id) REFERENCES admin_roles(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员账号表';

-- ============================================================
-- 12. admin_role_permissions
-- ============================================================
CREATE TABLE IF NOT EXISTS admin_role_permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    permission_code VARCHAR(100) NOT NULL COMMENT '权限编码',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_permission (role_id, permission_code),
    KEY idx_permission_code (permission_code),
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES admin_roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限表';

-- ============================================================
-- Seed Data
-- ============================================================

-- Default station
INSERT INTO stations (name, code, status) VALUES ('默认站点', 'default', 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Super admin role
INSERT INTO admin_roles (code, name, description, status)
SELECT 'super_admin', '超级管理员', '平台最高权限角色', 1
WHERE NOT EXISTS (SELECT 1 FROM admin_roles WHERE code = 'super_admin');

-- Super admin permissions
INSERT INTO admin_role_permissions (role_id, permission_code)
SELECT r.id, p.code
FROM admin_roles r
JOIN (
    SELECT '*' AS code
    UNION ALL SELECT 'system:*'
    UNION ALL SELECT 'system:wechat:read'
    UNION ALL SELECT 'system:wechat:write'
    UNION ALL SELECT 'mini_program:*'
    UNION ALL SELECT 'mini_program:config:read'
    UNION ALL SELECT 'mini_program:config:write'
) p
WHERE r.code = 'super_admin'
  AND NOT EXISTS (
      SELECT 1 FROM admin_role_permissions arp
      WHERE arp.role_id = r.id AND arp.permission_code = p.code
  );

-- Default admin account (password: admin123)
INSERT INTO admin_users (username, password, real_name, status, role_id)
SELECT 'admin', '$2y$10$x.OjbycVEjdwGpx9ml65MutwwfIyo2jXci9b2DjgnWWM2O1yRBNSi', '平台管理员', 1, r.id
FROM admin_roles r
WHERE r.code = 'super_admin'
  AND NOT EXISTS (SELECT 1 FROM admin_users WHERE username = 'admin');
