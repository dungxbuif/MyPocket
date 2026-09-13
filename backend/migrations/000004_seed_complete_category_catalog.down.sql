DELETE FROM categories WHERE is_system = true AND system_key LIKE 'expense_bills_%';
DELETE FROM categories WHERE is_system = true AND system_key LIKE 'expense_shopping_%';
DELETE FROM categories WHERE is_system = true AND system_key LIKE 'expense_family_%';
DELETE FROM categories WHERE is_system = true AND system_key LIKE 'expense_transport_%';
DELETE FROM categories WHERE is_system = true AND system_key LIKE 'expense_health_%';
DELETE FROM categories WHERE is_system = true AND system_key LIKE 'expense_entertainment_%';
