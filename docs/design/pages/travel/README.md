# Travel Mode — sự kiện gắn giao dịch

Status: implemented web slice; owner browser/UAT vẫn cần xác nhận.
Route: `/account/travel`.

Travel Mode là một chiều ngữ cảnh của giao dịch, không phải loại ví. Màn hình
dùng `SurfaceCard`, `BaseButton`, `BaseTextInput`, `FormField`,
`BaseBottomSheet`, `StatusMessage` và `Text` hiện có.

## Hành vi

- Tạo, sửa, xóa chuyến; ngày bắt đầu/kết thúc là ngày thuần `YYYY-MM-DD`.
- Tại một thời điểm chỉ có một chuyến bật. Tắt/đổi chuyến không sửa các liên
  kết lịch sử.
- Giao dịch thu/chi thường tạo mới khi không truyền `travel_event_id` sẽ nhận
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

Báo cáo tháng chưa dựng narrative riêng theo chuyến; đây là phần của Money
Insider. Offline/backfill và lựa chọn chuyến cho AI draft vẫn giữ nguyên link
thủ công theo thời điểm duyệt.
