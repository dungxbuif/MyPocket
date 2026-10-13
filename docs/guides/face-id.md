# Đăng nhập MyPocket bằng Face ID trên iPhone

Đăng nhập bằng email và mật khẩu một lần, mở **Tài khoản → Bảo mật →
Bật đăng nhập Face ID**, rồi nhập lại mật khẩu để xác nhận. iPhone cần có
mã khóa và Face ID đã được thiết lập trong Cài đặt iOS.

Sau khi đăng xuất, màn đăng nhập vẫn nhớ email và có nút **Đăng nhập bằng
Face ID**. Chạm nút và xác thực trên iPhone để bắt đầu một phiên mới. Face ID
cũng dùng được khi phiên cũ đã hết hạn. Đăng nhập cần kết nối mạng.

**Đăng xuất** kết thúc phiên hiện tại nhưng giữ đăng nhập Face ID trên iPhone
này. **Tắt đăng nhập Face ID** thu hồi quyền đăng nhập đã nhớ nhưng giữ phiên
đang dùng. **Gỡ tài khoản khỏi thiết bị** vừa thu hồi quyền đã nhớ vừa đăng xuất.
Nếu không kết nối được máy chủ, thao tác đăng xuất/thu hồi báo lỗi để bạn thử
lại; ứng dụng không báo đã thu hồi khi máy chủ chưa xác nhận.

MyPocket không lưu mật khẩu và không nhận dữ liệu khuôn mặt. iOS bảo vệ một
credential riêng trong Keychain của thiết bị; credential không đồng bộ qua
iCloud. Đăng ký có hiệu lực tối đa 180 ngày. Đổi mật khẩu tài khoản, thiết lập
lại Face ID, thu hồi hoặc hết hạn đăng ký có thể yêu cầu đăng nhập bằng mật
khẩu và bật lại Face ID. Hủy xác thực hoặc lỗi mạng không xóa đăng ký còn hiệu
lực.

Khi nâng cấp từ bản cũ, lần đăng xuất đầu tiên có thể yêu cầu đăng nhập lại
trên những thiết bị khác vẫn dùng phiên trước nâng cấp. Các phiên đăng nhập
mới bằng mật khẩu hoặc Face ID không bị ảnh hưởng bởi việc chuyển đổi này.

## Ghi nhớ ví đang xem

Màn Giao dịch ghi nhớ ID ví được chọn trên từng thiết bị, riêng cho mỗi tài
khoản. Chọn **Tổng cộng** sẽ bỏ lựa chọn ví đã nhớ. Nếu ví bị xóa hoặc không
còn truy cập được, ứng dụng trở về Tổng cộng. Đây là lựa chọn giao diện trong
`UserDefaults`; số dư, ví và giao dịch vẫn do backend quản lý.

API cho tích hợp/AI agents: [Authentication contract](../architecture/AUTH_API.md)
và OpenAPI của backend. Face ID/device enrollment là chức năng tài khoản cần
session; API key không có quyền đăng ký hay thu hồi thiết bị.
