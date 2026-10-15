# Travel Mode — sự kiện gắn giao dịch

Status: native implemented; owner/device UAT vẫn cần xác nhận.
Entry: Tài khoản → Travel Mode. Web route cũ chỉ là lịch sử, không còn maintain.

Travel Mode là một chiều ngữ cảnh của giao dịch, không phải loại ví. Màn hình
dùng `MPSurfaceCard`, `MPButton`, `MPTextField`, `MPTextEditor`,
`MPOptionalDateField`, `MPEditorScaffold` và `MPStatusMessage` hiện có.
Thêm/sửa mở editor riêng thay danh sách, với Hủy trên đầu và Lưu ghim dưới.
Đang lưu khóa thao tác lặp; lỗi giữ editor/nội dung. Ngày là tùy chọn, bật để
chọn bằng lịch theo múi giờ tài khoản; tắt để xóa. Không nhập chuỗi ngày thủ công.
Hủy không lưu; mở lại cùng chuyến trong màn hiện tại giữ nháp input, nhưng cập
nhật trạng thái bật/tắt từ dữ liệu mới. Chọn chuyến khác/tạo mới tách nháp. Lưu
thành công xóa nháp; chưa cam kết giữ nháp khi thoát màn hoặc khởi động lại app.

## Hành vi

- Tạo, sửa, xóa chuyến; ngày bắt đầu/kết thúc là ngày thuần `YYYY-MM-DD`.
- Tại một thời điểm chỉ có một chuyến bật. Tắt/đổi chuyến không sửa các liên
  kết lịch sử.
- Giao dịch thu/chi thường tạo mới khi không truyền `travel_event_id` và không
  tắt `auto_link_travel` sẽ nhận
  chuyến đang bật. Recurring, chuyển ví, điều chỉnh và AI approval không tự
  gắn chuyến.
- API riêng `PATCH /api/v1/transactions/{id}/travel` cho phép gắn hoặc gỡ
  giao dịch thường. Credit, transfer và adjustment bị từ chối.
- Xóa chuyến chỉ gỡ `travel_event_id`; không xóa giao dịch, không đổi số dư,
  và không đổi cờ tính vào báo cáo.

## API

`GET/POST/PATCH/DELETE /api/v1/travel/events`,
`POST /api/v1/travel/events/{id}/activate`,
`POST /api/v1/travel/events/{id}/deactivate`, và
`PATCH /api/v1/transactions/{id}/travel`.

## Gaps

AI-agent PATCH reference: bỏ qua hoặc gửi null cho `context`, `starts_on`,
`ends_on` sẽ giữ giá trị cũ. Gửi chuỗi rỗng để xóa. Native full-form update
thực hiện chuyển đổi này; không đổi hành vi partial PATCH của backend.
Actual local API UI proof: `/tmp/mypocket-travel-selection-20261014.xcresult`
1/1, tạo và xóa mô tả/ngày với server readback. Device UAT vẫn mở.
Gate mới `/tmp/mypocket-draft-retention-final-20261014.xcresult`188/188 giữ
regression ngày/mô tả và kiểm chứng hủy không ghi, mở lại cùng chuyến giữ nháp.

Báo cáo tháng chưa dựng narrative riêng theo chuyến; đây là phần của Money
Insider. Offline/backfill và lựa chọn chuyến cho AI draft vẫn giữ nguyên link
thủ công theo thời điểm duyệt.

## Delayed transaction context — 2026-10-14

New local create contract accepts auto_link_travel=false to retain explicit
selected trip or no trip rather than attaching the trip active at delayed send.
Ordinary online input keeps its current omitted-field auto-link behavior. Explicit
trip is owner-validated; deletion rejects new writes, while completed receipt
replay returns its original result without resurrecting links. Native outbox
freezes false and supplied trip/time before persistence, but current screen/cache/
sync integration is not enabled. [Contract/proof](../../../work/tickets/OFFLINE-FROZEN-TRAVEL-20261014.md).
