INSERT INTO mini_program_config_items (namespace, config_key, scope_type, scope_code, published_value, visibility, description, status)
VALUES
  ('general',  'app_name',         'global', 'default', '"代理系统"', 'public', '应用名称', 1),
  ('general',  'support_wechat',   'global', 'default', '""',        'public', '客服微信号', 1),
  ('feature',  'recharge_enabled', 'global', 'default', 'true',      'public', '充值功能开关', 1),
  ('recharge', 'min_amount',       'global', 'default', '100',       'public', '最小充值金额', 1),
  ('recharge', 'max_amount',       'global', 'default', '5000',      'public', '最大充值金额', 1)
ON DUPLICATE KEY UPDATE description = VALUES(description);
