---
artifact_type: detail_design
status: draft
owner: shared
approval: conversational_direction_accepted_planning_requested
---

# Tài sản tích lũy: vàng tự cập nhật và tài sản nhập tay

Owner chọn tên chung “Tài sản tích lũy”, tái sử dụng portfolio hiện có. Vàng là loại đầu tiên có giá hệ thống; loại khác/custom nhập giá thủ công. Chọn sản phẩm hệ thống liên kết nguồn giá dùng chung; backend kiểm tra giá mỗi 10 phút. Giá ước tính bán lại dùng chiều cửa hàng mua vào, kèm nguồn và thời điểm giá. BTMC và BTMH đã được đối chiếu với pipeline n8n production, có adapter đọc official HTML và catalog mapping thật.

Đã cân nhắc tách Đầu tư/Tích lũy thành hai module hoặc viết portfolio mới; chọn nâng cấp portfolio vì cùng một holding không được đếm hai lần. Bản đầu không có automatic stocks, biểu đồ lịch sử giá, cảnh báo giá, AI định giá hoặc tự mua bán.

Nhập mua/bán: số lượng, đơn giá VND/đơn vị, phí, ngày; tổng trước phí = số lượng × đơn giá. Mua trả tổng + phí; bán nhận tổng − phí. Dùng fixed-point/rounding canonical hiện có; client chỉ preview, server là authority. Giá thị trường không sửa lịch sử mua, giá vốn hoặc giao dịch tiền mặt. Không tự tạo cash transaction/đổi số dư ví.

Mỗi tài sản có đơn vị. Ở phiên bản đầu, backend chỉ công bố loại Vàng và ba unit cho người dùng chọn: chỉ, lượng, phân; không có ô nhập unit tự do. Sản phẩm hệ thống có unit cố định theo provider mapping đã kiểm chứng. Tài sản cũ giữ quantity/unit_price/fee/correction history và latest price, mặc định manual, unit chưa xác nhận. Không đoán chỉ/lượng/gram, không quy đổi quantity hoặc price. Cho người dùng xác nhận nhãn đơn vị họ đã dùng, preview và expected_version; không cho đổi đơn vị đã xác nhận khi có trades. Chưa làm conversion workflow. Liên kết vàng cũ chỉ sau khi xác nhận đúng sản phẩm/đơn vị; mismatch bị chặn. Các API legacy vẫn hoạt động với tài sản manual chưa xác nhận; không cho legacy caller ghi giá lên asset automatic. Khi có provider product đúng unit, UI mới cho chọn giá tự động; lượng/phân hiện chỉ nhập giá thủ công.

Product catalog có provider_key, external_product_id, tên/loại/unit/currency và valuation_side=buyback; holdings owner-scoped liên kết catalog, không copy một holding thành tài sản khác. Mỗi product chỉ dùng một nguồn xác định, không average/fallback sang nguồn khác. Hai nguồn có thể chứa sản phẩm khác nhau; mapping do hệ thống cấu hình, không dùng tên làm identity. Custom không liên kết catalog thì manual. Để quay từ automatic về manual cần explicit update/version và preview giữ giá gần nhất với provenance, sau đó user tự cập nhật; không đổi product/unit cho holding đã có trades trong bản đầu.

Quote có product/source/provider_observed_at/fetched_at/price/unit/currency/status. Chỉ nhận giá dương đúng product/unit/currency, timestamp hợp lệ và không cũ hơn quote hiện có. Không dùng fetched_at làm thời gian giá. Đề xuất freshness TTL 60 phút, clock skew tolerance5phút, HTTP timeout10giây và max body1MiB; cấu hình nguồn được override theo hợp đồng thực. Hết TTL hoặc lỗi nguồn giữ giá gần nhất kèm nhãn “Giá cũ”; chưa có giá hiển thị chưa định giá, không 0. Tổng định giá có unknown/stale counts; nếu có asset chưa định giá thì total hoàn chỉnh unavailable, phần đã biết được gắn nhãn, không ngụ ý tổng đầy đủ.

Worker polling shared catalog active products, một lần/source/batch cho mọi owner, DB claim theo provider+10-minute slot chống overlap. Kết quả cập nhật quote atomically, summary resolve latest quote bằng product binding; không ghi hàng loạt owner prices hay bump asset version khi quote đổi. Failure không xóa quote. Restart có thể retry read-only provider GET với cùng slot; upsert chống lùi giá/duplicate. Active products tự động chỉ được công bố khi adapter configured; custom vẫn dùng bình thường. Secrets/endpoint cấu hình backend, không nhận arbitrary URL từ user.

UI light mode/shared bases: danh sách holdings, chi tiết, tạo custom/chọn hệ thống, form số lượng/đơn giá/tổng/phí, nguồn/thời gian giá, sửa metadata khác sửa giao dịch khác cập nhật giá. History headline là tổng tiền, đơn giá ở subtitle có unit. Nút “Sửa giao dịch” rõ trong chi tiết, tái sử dụng correction preview/confirm hiện có, expected asset version/idempotency. Không xóa correction trail. Mask tiền và VoiceOver. Price mode/manual-unit badges dùng shared status bases.

Impact: additive portfolio schema/DTO/API, catalog/quote tables và worker config; giữ `/portfolio` routes. Test migration preservation, owner/version checks, fractional math/overflow, stale/unknown, out-of-order quotes, worker overlap, source failure, correction/replay và native form masking. Source contract evidence lives in [GOLD-SOURCE-CONTRACT](../../work/tickets/GOLD-SOURCE-CONTRACT.md); production deployment still requires migration rollout and a bounded live smoke check.

Trace: [Backlog](../../work/BACKLOG.md), [native phase](../../work/phases/PHASE-002-native-ios-swiftui.md), [API](../../architecture/API.md), [validation](../../work/VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [Reports design](2026-10-31-reports-insider-design.md). Implementation reconcile API/ERD/ADR/base/screen docs/changelog. Đây là design companion của plan, không runtime evidence.
