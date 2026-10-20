# Quản lý đầu tư trên iPhone

Mở **Tài khoản → Danh mục đầu tư** để xem tài sản, giá vốn, số lượng và lịch sử mua/bán.
Dữ liệu đầu tư độc lập với ví và báo cáo thu/chi.

Các chỉ số vị thế dùng cùng kiểu hiển thị gọn như tổng quan; toàn bộ giá vốn,
bình quân, lãi/lỗ và lịch sử vẫn có thể xem bằng cách cuộn. Bản native dùng
chung card, menu thao tác và editor của hệ thống UI; việc cài đặt/Face ID và
touch UAT trên thiết bị thật vẫn là bước nghiệm thu riêng.

- **Thêm tài sản** mở editor riêng. Nhập mã/tên rồi lưu. Mã trùng được báo lỗi;
  bản nháp giữ nguyên để bạn sửa. Hủy quay lại mà không tạo tài sản.
- **Ghi giao dịch**, hoặc **Giao dịch** trên tài sản, mở editor mua/bán. Chọn tài
  sản, khối lượng, giá, phí và ngày rồi xác nhận lưu.
- Có thể sửa **mã** và/hoặc **tên** tài sản từ màn hình chi tiết. Việc này chỉ đổi
  metadata của tài sản; toàn bộ giao dịch mua/bán vẫn giữ nguyên ID, lịch sử và
  giá vốn. Mã trùng bị từ chối; giá trị chỉ gồm khoảng trắng hoặc vượt giới hạn
  cũng bị báo lỗi.
- Khối lượng chấp nhận dấu phẩy hoặc dấu chấm thập phân, tối đa8 chữ số: `1,25`
  và `1.25` đều gửi thành `1.25`. Không nhập dấu phân nhóm hoặc trộn cả hai dấu.
  Giá và phí vẫn là số tiền nguyên VND.
- **Cập nhật giá** mở hộp nhập riêng; để trống để bỏ giá hiện tại. Khi chưa có giá,
  app hiển thị **Chưa có giá**, không giả định bằng0.
- Xóa tài sản chỉ áp dụng khi chưa có giao dịch. Nếu một giao dịch vừa được lưu,
  máy chủ giữ nguyên tài sản và lịch sử, đồng thời từ chối yêu cầu xóa.
- Ở từng dòng lịch sử, mở menu ba chấm để **Sửa giao dịch** hoặc **Hủy giao dịch**.
  Sửa mở editor với dữ liệu gốc đã điền sẵn; hủy chỉ cần nhập lý do. Nút **Xem
  thay đổi** gọi preview read-only và hiển thị diff khối lượng, giá vốn, lãi/lỗ
  và phiên bản danh mục. Chỉ **Xác nhận sửa** mới gửi commit với idempotency key;
  bản nháp và lỗi vẫn giữ nguyên trong editor.
- Trong lúc lưu, app khóa thao tác liên quan để tránh gửi trùng. Lỗi giữ editor
  và bản nháp; lưu thành công mới đóng rồi tải lại vị thế/lịch sử.

Trade history is append-only. There is intentionally no direct row edit/delete
operation. A reviewed correction uses an immutable replace/void revision; the
original trade remains in audit history and the complete ordered history is
replayed for cost basis, realized P&L and inventory validation.

## Correction contract (current)

The owner-authenticated correction flow is a two-step preview/commit pair:

1. `POST /api/v1/portfolio/assets/{asset_id}/trade-corrections/preview` with
   `source_trade_id`, `expected_asset_version`, `operation` (`replace` or
   `void`), a non-empty `reason`, and replacement fields when `operation` is
   `replace`.
2. After the user reviews the before/after positions, repeat the same body at
   `POST /api/v1/portfolio/assets/{asset_id}/trade-corrections` with a stable
   `Idempotency-Key`.

The server rejects stale asset versions, duplicate/conflicting keys, invalid
replacement fields and any projected history that oversells. Backdated
replacement timestamps are accepted only when the replayed chronological
history remains valid; later trades are reported in `affected_trade_ids` so a
client can make the impact visible. Fees, realized P&L and cost basis are not
inferred: replacement fields must be explicit and all values are recomputed by
the weighted-average ledger. No wallet cash entry is created.

The native client exposes the same preview/commit API adapter. `PortfolioView`
keeps preview as a draft, renders before/after position deltas and requires
explicit confirmation; it does not delete rows or encode an implicit reversal
in free-form notes. The original trade remains in audit history after commit.

## AI-agent reference

Routes relative to `/api/v1`: `/portfolio/assets`, `/portfolio/assets/:id`,
`/portfolio/assets/:id/trades`, `/portfolio/assets/:id/price`,
`/portfolio/assets/:id/trade-corrections/preview`,
`/portfolio/assets/:id/trade-corrections`, `/portfolio/summary`.
Keep authenticated owner, idempotency and authorization requirements.
`PATCH /portfolio/assets/:id` accepts a partial JSON body with `symbol` and/or
`name`; values are trimmed, validated and owner-scoped. A duplicate symbol
returns conflict. Trade `quantity` remains an ASCII
decimal string; native UI alone normalizes comma before sending. No API type or
schema changes. Trade history is append-only; corrections use immutable
replace/void revisions, never direct DB editing. [API reference](../architecture/API.md).

Price PATCH requires an explicit `latest_price` field: `{"latest_price":250}`
sets an integer VND unit price, `{"latest_price":0}` sets a known zero price,
and `{"latest_price":null}` clears it. An omitted field or top-level `null`
returns 400 without changing the saved valuation. `DELETE /portfolio/assets/:id`
only deletes unused assets; a concurrent committed trade makes deletion return
409 and preserves history.
