-- Owner-approved default category catalog. System rows are read-only in v1.
WITH roots(system_key, kind, name) AS (
  VALUES
    ('expense_food','expense','Ăn uống'),
    ('expense_bills','expense','Hoá đơn & Tiện ích'),
    ('expense_shopping','expense','Mua sắm'),
    ('expense_family','expense','Gia đình'),
    ('expense_transport','expense','Di chuyển'),
    ('expense_health','expense','Sức khoẻ'),
    ('expense_entertainment','expense','Giải trí'),
    ('expense_education','expense','Giáo dục'),
    ('expense_gifts','expense','Quà tặng & Quyên góp'),
    ('expense_insurance','expense','Bảo hiểm'),
    ('expense_investment','expense','Đầu tư'),
    ('expense_general','expense','Các chi phí khác'),
    ('expense_transfer_out','expense','Tiền chuyển đi'),
    ('expense_interest_paid','expense','Trả lãi'),
    ('expense_uncategorized','expense','Khoản chi chưa phân loại'),
    ('expense_withdrawal','expense','Rút tiền'),
    ('expense_accessories','expense','Phụ kiện'), ('expense_books','expense','Sách'),
    ('expense_business','expense','Kinh doanh'), ('expense_charity','expense','Từ thiện'),
    ('expense_funeral','expense','Tang lễ'), ('expense_games','expense','Trò chơi'),
    ('expense_home_repair','expense','Sửa chữa nhà cửa'), ('expense_wedding','expense','Cưới hỏi'),
    ('expense_movies','expense','Phim ảnh'), ('expense_parking','expense','Gửi xe'),
    ('expense_personal_care','expense','Chăm sóc cá nhân'), ('expense_fuel','expense','Xăng dầu'),
    ('expense_medicine','expense','Thuốc'), ('expense_restaurant','expense','Nhà hàng'),
    ('expense_sports','expense','Thể thao'), ('expense_taxi','expense','Taxi'),
    ('expense_travel','expense','Du lịch'), ('expense_savings','expense','Tiết Kiệm'),
    ('expense_children','expense','Con cái'), ('expense_clothes','expense','Quần áo'),
    ('expense_medical_treatment','expense','Khám chữa bệnh'), ('expense_electronics','expense','Thiết bị điện tử'),
    ('expense_cost','expense','Chi phí'), ('expense_shoes','expense','Giày dép'),
    ('expense_friends','expense','Bạn bè & Người yêu'),
    ('income_salary','income','Lương'), ('income_other','income','Thu nhập khác'),
    ('income_transfer_in','income','Tiền chuyển đến'), ('income_interest','income','Thu lãi'),
    ('income_uncategorized','income','Khoản thu chưa phân loại'), ('income_gift','income','Được tặng'),
    ('income_bonus','income','Thưởng'), ('income_sell_items','income','Bán đồ'),
    ('debt_lend','debt','Cho vay'), ('debt_repay','debt','Trả nợ'),
    ('debt_loan','debt','Đi vay'), ('debt_collect','debt','Thu nợ')
)
INSERT INTO categories (id, kind, name, system_key, is_system)
SELECT '00000000-0000-4000-8000-' || lpad((500 + row_number() OVER ())::text, 12, '0'), kind, name, system_key, true FROM roots
ON CONFLICT (system_key) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind, parent_id = NULL, is_system = true;

WITH children(system_key, parent_key, kind, name) AS (
  VALUES
    ('expense_food_snacks','expense_food','expense','Ăn vặt'),
    ('expense_food_coffee','expense_food','expense','Cà phê'),
    ('expense_food_meals','expense_food','expense','Cơm Bữa'),
    ('expense_bills_phone','expense_bills','expense','Hoá đơn điện thoại'),
    ('expense_bills_water','expense_bills','expense','Hoá đơn nước'),
    ('expense_bills_electricity','expense_bills','expense','Hoá đơn điện'),
    ('expense_bills_gas','expense_bills','expense','Hoá đơn gas'),
    ('expense_bills_tv','expense_bills','expense','Hoá đơn TV'),
    ('expense_bills_internet','expense_bills','expense','Hoá đơn internet'),
    ('expense_bills_rent','expense_bills','expense','Thuê nhà'),
    ('expense_bills_other','expense_bills','expense','Hoá đơn tiện ích khác'),
    ('expense_shopping_personal','expense_shopping','expense','Đồ dùng cá nhân'),
    ('expense_shopping_household','expense_shopping','expense','Đồ gia dụng'),
    ('expense_shopping_beauty','expense_shopping','expense','Làm đẹp'),
    ('expense_family_home_decor','expense_family','expense','Sửa & trang trí nhà'),
    ('expense_family_services','expense_family','expense','Dịch vụ gia đình'),
    ('expense_family_pets','expense_family','expense','Vật nuôi'),
    ('expense_transport_maintenance','expense_transport','expense','Bảo dưỡng xe'),
    ('expense_health_checkup','expense_health','expense','Khám sức khoẻ'),
    ('expense_health_fitness','expense_health','expense','Thể dục thể thao'),
    ('expense_entertainment_online','expense_entertainment','expense','Dịch vụ trực tuyến'),
    ('expense_entertainment_fun','expense_entertainment','expense','Vui - chơi')
)
INSERT INTO categories (id, parent_id, kind, name, system_key, is_system)
SELECT '00000000-0000-4000-8000-' || lpad((600 + row_number() OVER ())::text, 12, '0'), parent.id, child.kind, child.name, child.system_key, true
FROM children child JOIN categories parent ON parent.system_key = child.parent_key
ON CONFLICT (system_key) DO UPDATE SET parent_id = EXCLUDED.parent_id, name = EXCLUDED.name, kind = EXCLUDED.kind, is_system = true;

-- Remove retired system rows after reparenting any personal child to root.
UPDATE categories SET parent_id = NULL WHERE is_system = false AND parent_id IN (SELECT id FROM categories WHERE is_system = true AND system_key IN ('expense_food_restaurant','income_root','debt_root'));
DELETE FROM categories WHERE is_system = true AND system_key IN ('expense_food_restaurant','income_root','debt_root');
