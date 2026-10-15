# Quản lý đầu tư trên iPhone

Mở **Tài khoản → Danh mục đầu tư** để xem tài sản, giá vốn, số lượng và lịch sử mua/bán.
Dữ liệu đầu tư độc lập với ví và báo cáo thu/chi.

Các chỉ số vị thế dùng cùng kiểu hiển thị gọn như tổng quan; toàn bộ giá vốn,
bình quân, lãi/lỗ và lịch sử vẫn có thể xem bằng cách cuộn. Bản native đang phát
triển còn lỗi bong bóng phản hồi che một phần tên tài sản; chưa coi đây là bản
UI hoàn tất hoặc đã cài lên iPhone.

- **Thêm tài sản** mở editor riêng. Nhập mã/tên rồi lưu. Mã trùng được báo lỗi;
  bản nháp giữ nguyên để bạn sửa. Hủy quay lại mà không tạo tài sản.
- **Ghi giao dịch**, hoặc **Giao dịch** trên tài sản, mở editor mua/bán. Chọn tài
  sản, khối lượng, giá, phí và ngày rồi xác nhận lưu.
- Khối lượng chấp nhận dấu phẩy hoặc dấu chấm thập phân, tối đa8 chữ số: `1,25`
  và `1.25` đều gửi thành `1.25`. Không nhập dấu phân nhóm hoặc trộn cả hai dấu.
  Giá và phí vẫn là số tiền nguyên VND.
- **Cập nhật giá** mở hộp nhập riêng; để trống để bỏ giá hiện tại. Khi chưa có giá,
  app hiển thị **Chưa có giá**, không giả định bằng0.
- Trong lúc lưu, app khóa thao tác liên quan để tránh gửi trùng. Lỗi giữ editor
  và bản nháp; lưu thành công mới đóng rồi tải lại vị thế/lịch sử.

## AI-agent reference

Routes relative to `/api/v1`: `/portfolio/assets`, `/portfolio/assets/:id/trades`,
`/portfolio/assets/:id/price`, `/portfolio/summary`. Keep authenticated owner,
idempotency and authorization requirements. Trade `quantity` remains an ASCII
decimal string; native UI alone normalizes comma before sending. No API type or
schema changes. Trade history is append-only; corrections use compensating
trades, never direct DB editing. [API reference](../architecture/API.md).
