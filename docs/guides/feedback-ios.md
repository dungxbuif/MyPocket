# Gửi phản hồi trong ứng dụng iPhone

## Theo dõi và xác nhận vấn đề

Hai tab **Chưa giải quyết** và **Đã giải quyết** có số lượng. Tick cạnh tiêu đề
khi bạn xác nhận vấn đề đã hết; sau khi lưu thành công, item chuyển sang tab
Đã giải quyết. Nếu lỗi xuất hiện lại, vào tab đó bỏ tick để đưa về Chưa giải
quyết. Lưu theo tài khoản trên backend; lỗi mạng giữ nguyên item và báo thử lại.

Xác nhận của bạn độc lập với trạng thái dev/changelog; mở lại không xóa mô tả,
ngày dev sửa hay release liên quan. Feedback cũ mặc định chưa được xác nhận.
Chạm **+** để gửi mới; **Xem changelog** mở lịch sử phát hành.

AI-agent reference: `PATCH /api/v1/feedback/{id}/resolution`, JSON
`{"resolved":true}` hoặc `{"resolved":false}`; JWT hoặc user API key có
`feedback:write`, chỉ tác động feedback của tài khoản được xác thực. Response
có `resolved`, `resolved_at` khi đang resolved. Gửi cùng giá trị không đổi ngày
xác nhận. Không truyền user_id; agent service token không được dùng route này.

Chạm bong bóng phản hồi ở góc dưới bên phải, hoặc mở **Tài khoản →
Phản hồi & changelog**. Bong bóng nổi trên nội dung; không chiếm một hàng
riêng và không rút ngắn vùng cuộn phía trên thanh điều hướng.

Chọn loại phản hồi, nhập tiêu đề và mô tả. Khi mở từ bong bóng, ảnh màn hình
hiện tại được điền sẵn nếu chụp được. Xem thumbnail trước khi gửi; chạm **Bỏ
tệp** để bỏ ảnh hoặc **Thay ảnh** để chọn ảnh khác. Màn phản hồi riêng có nút
**Đính kèm ảnh**. Kiểm tra ảnh không chứa mật khẩu/API key hay thông tin không
muốn chia sẻ với người xử lý phản hồi.

Mỗi phản hồi có một ảnh tùy chọn theo API hiện tại. Ảnh được chuyển thành PNG
và thu nhỏ tới tối đa 1 MiB trước khi gửi; ảnh nguồn tối đa 25 MiB. Nếu không
đọc được ảnh, ứng dụng giữ nội dung bạn đã nhập và cho chọn lại. Khi kho ảnh
chưa sẵn sàng, hệ thống có thể nhận phản hồi chữ mà không lưu ảnh.

Các trường nhập có thể cuộn; nút gửi ở ngoài vùng cuộn của hộp thoại. Phản hồi
và ảnh không được đưa vào hội thoại AI. Trạng thái **Đã xử lý** và phiên bản
chỉ xuất hiện sau khi bản sửa đã được xác nhận/phát hành.

Tài liệu cho tích hợp và AI agents: [Feedback API](../architecture/FEEDBACK_API.md).
Ảnh được lưu riêng tư; người dùng chỉ truy cập phản hồi/ảnh của mình. Service
token dành cho quy trình xử lý feedback không được nhúng trong app.
