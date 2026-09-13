INSERT INTO categories (id, kind, name, system_key, is_system) VALUES
('00000000-0000-4000-8000-000000000301','expense','Ăn uống','expense_food',true),
('00000000-0000-4000-8000-000000000302','expense','Hoá đơn & Tiện ích','expense_bills',true),
('00000000-0000-4000-8000-000000000303','expense','Mua sắm','expense_shopping',true),
('00000000-0000-4000-8000-000000000304','expense','Gia đình','expense_family',true),
('00000000-0000-4000-8000-000000000305','expense','Di chuyển','expense_transport',true),
('00000000-0000-4000-8000-000000000306','expense','Sức khỏe','expense_health',true),
('00000000-0000-4000-8000-000000000307','expense','Giáo dục','expense_education',true),
('00000000-0000-4000-8000-000000000308','expense','Giải trí','expense_entertainment',true),
('00000000-0000-4000-8000-000000000309','expense','Quà tặng & Quyên góp','expense_gifts',true),
('00000000-0000-4000-8000-000000000310','expense','Bảo hiểm','expense_insurance',true),
('00000000-0000-4000-8000-000000000311','expense','Đầu tư','expense_investment',true),
('00000000-0000-4000-8000-000000000312','expense','Du lịch','expense_travel',true),
('00000000-0000-4000-8000-000000000313','expense','Tiết kiệm','expense_savings',true),
('00000000-0000-4000-8000-000000000314','expense','Kinh doanh','expense_business',true),
('00000000-0000-4000-8000-000000000315','expense','Các chi phí chung','expense_general',true),
('00000000-0000-4000-8000-000000000316','income','Thu nhập','income_root',true),
('00000000-0000-4000-8000-000000000317','debt','Vay/Nợ','debt_root',true)
ON CONFLICT (system_key) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind, is_system = true;

INSERT INTO categories (id, parent_id, kind, name, system_key, is_system)
SELECT child.id, parent.id, child.kind, child.name, child.system_key, true
FROM (VALUES
('00000000-0000-4000-8000-000000000321','expense_food','expense','Ăn vặt','expense_food_snacks'),
('00000000-0000-4000-8000-000000000322','expense_food','expense','Cà phê','expense_food_coffee'),
('00000000-0000-4000-8000-000000000323','expense_food','expense','Cơm bữa','expense_food_meals'),
('00000000-0000-4000-8000-000000000324','expense_food','expense','Nhà hàng','expense_food_restaurant'),
('00000000-0000-4000-8000-000000000366','income_root','income','Lương','income_salary'),
('00000000-0000-4000-8000-000000000367','income_root','income','Thu nhập khác','income_other'),
('00000000-0000-4000-8000-000000000372','income_root','income','Thưởng','income_bonus'),
('00000000-0000-4000-8000-000000000374','debt_root','debt','Cho vay','debt_lend'),
('00000000-0000-4000-8000-000000000375','debt_root','debt','Trả nợ','debt_repay'),
('00000000-0000-4000-8000-000000000376','debt_root','debt','Đi vay','debt_loan'),
('00000000-0000-4000-8000-000000000377','debt_root','debt','Thu nợ','debt_collect')
) AS child(id, parent_key, kind, name, system_key)
JOIN categories parent ON parent.system_key = child.parent_key
ON CONFLICT (system_key) DO UPDATE SET parent_id = EXCLUDED.parent_id, name = EXCLUDED.name, kind = EXCLUDED.kind, is_system = true;
