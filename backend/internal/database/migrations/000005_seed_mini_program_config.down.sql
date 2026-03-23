DELETE FROM mini_program_config_items
WHERE scope_type = 'global' AND scope_code = 'default'
  AND (
    (namespace = 'general'  AND config_key IN ('app_name', 'support_wechat'))
    OR (namespace = 'feature'  AND config_key = 'recharge_enabled')
    OR (namespace = 'recharge' AND config_key IN ('min_amount', 'max_amount'))
  );
