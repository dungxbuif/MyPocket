-- Expand the default Vietnamese finance catalog. Keys are stable so this is safe to rerun.
INSERT INTO categories (id, kind, name, system_key, is_system)
VALUES
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
ON CONFLICT (system_key) DO UPDATE SET name=EXCLUDED.name, kind=EXCLUDED.kind, is_system=true, updated_at=now();

UPDATE categories child
SET parent_id = parent.id, updated_at = now()
FROM categories parent
WHERE parent.system_key = 'income_root' AND child.system_key IN ('income_salary', 'income_bonus');
UPDATE categories child
SET parent_id = parent.id, name = 'Đi vay', updated_at = now()
FROM categories parent
WHERE parent.system_key = 'debt_root' AND child.system_key = 'debt_loan';

INSERT INTO categories (id, parent_id, kind, name, system_key, is_system)
SELECT v.id::uuid, p.id, v.kind, v.name, v.system_key, true
FROM (VALUES
 ('00000000-0000-4000-8000-000000000321','expense_food','expense','Ăn vặt','expense_food_snacks'),
 ('00000000-0000-4000-8000-000000000322','expense_food','expense','Cà phê','expense_food_coffee'),
 ('00000000-0000-4000-8000-000000000323','expense_food','expense','Cơm bữa','expense_food_meals'),
 ('00000000-0000-4000-8000-000000000324','expense_food','expense','Nhà hàng','expense_food_restaurant'),
 ('00000000-0000-4000-8000-000000000325','expense_bills','expense','Hoá đơn điện thoại','expense_bills_phone'),
 ('00000000-0000-4000-8000-000000000326','expense_bills','expense','Hoá đơn nước','expense_bills_water'),
 ('00000000-0000-4000-8000-000000000327','expense_bills','expense','Hoá đơn điện','expense_bills_electricity'),
 ('00000000-0000-4000-8000-000000000328','expense_bills','expense','Hoá đơn gas','expense_bills_gas'),
 ('00000000-0000-4000-8000-000000000329','expense_bills','expense','Hoá đơn TV','expense_bills_tv'),
 ('00000000-0000-4000-8000-000000000330','expense_bills','expense','Hoá đơn internet','expense_bills_internet'),
 ('00000000-0000-4000-8000-000000000331','expense_bills','expense','Thuê nhà','expense_bills_rent'),
 ('00000000-0000-4000-8000-000000000332','expense_bills','expense','Hoá đơn tiện ích khác','expense_bills_other'),
 ('00000000-0000-4000-8000-000000000333','expense_shopping','expense','Đồ dùng cá nhân','expense_shopping_personal'),
 ('00000000-0000-4000-8000-000000000334','expense_shopping','expense','Đồ gia dụng','expense_shopping_household'),
 ('00000000-0000-4000-8000-000000000335','expense_shopping','expense','Làm đẹp','expense_shopping_beauty'),
 ('00000000-0000-4000-8000-000000000336','expense_shopping','expense','Phụ kiện','expense_shopping_accessories'),
 ('00000000-0000-4000-8000-000000000337','expense_shopping','expense','Quần áo','expense_shopping_clothes'),
 ('00000000-0000-4000-8000-000000000338','expense_shopping','expense','Giày dép','expense_shopping_shoes'),
 ('00000000-0000-4000-8000-000000000339','expense_shopping','expense','Thiết bị điện tử','expense_shopping_electronics'),
 ('00000000-0000-4000-8000-000000000340','expense_family','expense','Sửa chữa nhà cửa','expense_family_home_repair'),
 ('00000000-0000-4000-8000-000000000341','expense_family','expense','Dịch vụ gia đình','expense_family_services'),
 ('00000000-0000-4000-8000-000000000342','expense_family','expense','Vật nuôi','expense_family_pets'),
 ('00000000-0000-4000-8000-000000000343','expense_family','expense','Con cái','expense_family_children'),
 ('00000000-0000-4000-8000-000000000344','expense_transport','expense','Bảo dưỡng xe','expense_transport_maintenance'),
 ('00000000-0000-4000-8000-000000000345','expense_transport','expense','Gửi xe','expense_transport_parking'),
 ('00000000-0000-4000-8000-000000000346','expense_transport','expense','Xăng dầu','expense_transport_fuel'),
 ('00000000-0000-4000-8000-000000000347','expense_transport','expense','Taxi','expense_transport_taxi'),
 ('00000000-0000-4000-8000-000000000348','expense_health','expense','Khám chữa bệnh','expense_health_medical'),
 ('00000000-0000-4000-8000-000000000349','expense_health','expense','Thể thao','expense_health_sport'),
 ('00000000-0000-4000-8000-000000000350','expense_health','expense','Chăm sóc cá nhân','expense_health_personal_care'),
 ('00000000-0000-4000-8000-000000000351','expense_health','expense','Thuốc','expense_health_medicine'),
 ('00000000-0000-4000-8000-000000000352','expense_education','expense','Sách','expense_education_books'),
 ('00000000-0000-4000-8000-000000000353','expense_entertainment','expense','Dịch vụ trực tuyến','expense_entertainment_online'),
 ('00000000-0000-4000-8000-000000000354','expense_entertainment','expense','Trò chơi','expense_entertainment_games'),
 ('00000000-0000-4000-8000-000000000355','expense_entertainment','expense','Phim ảnh','expense_entertainment_movies'),
 ('00000000-0000-4000-8000-000000000356','expense_gifts','expense','Cưới hỏi','expense_gifts_wedding'),
 ('00000000-0000-4000-8000-000000000357','expense_gifts','expense','Tang lễ','expense_gifts_funeral'),
 ('00000000-0000-4000-8000-000000000358','expense_gifts','expense','Từ thiện','expense_gifts_charity'),
 ('00000000-0000-4000-8000-000000000359','expense_gifts','expense','Bạn bè & Người yêu','expense_gifts_relationships'),
 ('00000000-0000-4000-8000-000000000360','expense_general','expense','Các chi phí khác','expense_general_other'),
 ('00000000-0000-4000-8000-000000000361','expense_general','expense','Tiền chuyển đi','expense_general_transfer'),
 ('00000000-0000-4000-8000-000000000362','expense_general','expense','Trả lãi','expense_general_interest'),
 ('00000000-0000-4000-8000-000000000363','expense_general','expense','Khoản chi chưa phân loại','expense_general_uncategorized'),
 ('00000000-0000-4000-8000-000000000364','expense_general','expense','Rút tiền','expense_general_withdrawal'),
 ('00000000-0000-4000-8000-000000000365','expense_general','expense','Chi phí','expense_general_cost'),
 ('00000000-0000-4000-8000-000000000366','income_root','income','Lương','income_salary'),
 ('00000000-0000-4000-8000-000000000367','income_root','income','Thu nhập khác','income_other'),
 ('00000000-0000-4000-8000-000000000368','income_root','income','Tiền chuyển đến','income_transfer_in'),
 ('00000000-0000-4000-8000-000000000369','income_root','income','Thu lãi','income_interest'),
 ('00000000-0000-4000-8000-000000000370','income_root','income','Khoản thu chưa phân loại','income_uncategorized'),
 ('00000000-0000-4000-8000-000000000371','income_root','income','Được tặng','income_gift'),
 ('00000000-0000-4000-8000-000000000372','income_root','income','Thưởng','income_bonus'),
 ('00000000-0000-4000-8000-000000000373','income_root','income','Bán đồ','income_sale'),
 ('00000000-0000-4000-8000-000000000374','debt_root','debt','Cho vay','debt_lend'),
 ('00000000-0000-4000-8000-000000000375','debt_root','debt','Trả nợ','debt_repay'),
 ('00000000-0000-4000-8000-000000000376','debt_root','debt','Đi vay','debt_loan'),
 ('00000000-0000-4000-8000-000000000377','debt_root','debt','Thu nợ','debt_collect')
) AS v(id, parent_key, kind, name, system_key)
JOIN categories p ON p.system_key = v.parent_key
ON CONFLICT (system_key) DO UPDATE
SET parent_id = EXCLUDED.parent_id, name = EXCLUDED.name, kind = EXCLUDED.kind, is_system = true, updated_at = now();
