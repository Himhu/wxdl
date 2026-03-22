DROP TABLE IF EXISTS agent_applications;
DROP TABLE IF EXISTS user_invites;

ALTER TABLE user_role_change_logs
    ADD COLUMN old_agent_level TINYINT DEFAULT NULL COMMENT '原代理等级',
    ADD COLUMN new_agent_level TINYINT DEFAULT NULL COMMENT '新代理等级';

ALTER TABLE users
    DROP COLUMN inviter_user_id,
    ADD COLUMN agent_level TINYINT DEFAULT NULL COMMENT '代理等级: 1=一级 2=二级 3=三级' AFTER role,
    ADD COLUMN parent_user_id BIGINT UNSIGNED DEFAULT NULL COMMENT '上级用户ID' AFTER agent_level,
    ADD INDEX idx_parent_user_id (parent_user_id);
