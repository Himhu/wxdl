-- Baseline rollback: drop all tables in reverse dependency order

DROP TABLE IF EXISTS admin_role_permissions;
DROP TABLE IF EXISTS admin_users;
DROP TABLE IF EXISTS admin_roles;
DROP TABLE IF EXISTS system_setting_revisions;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS mini_program_config_items;
DROP TABLE IF EXISTS login_logs;
DROP TABLE IF EXISTS recharge_requests;
DROP TABLE IF EXISTS balance_logs;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS stations;
