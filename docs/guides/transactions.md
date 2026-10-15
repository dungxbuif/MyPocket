# Xem giao dịch theo ví và theo tuần

## Ghi chú trên iPhone

Trong màn thêm/sửa giao dịch, chạm dòng **Ghi chú** để mở ô ghi chú riêng, có thể
nhập nhiều dòng. **Lưu ghi chú** đưa nội dung về bản nháp giao dịch; bạn vẫn phải
nhấn **Lưu giao dịch/Lưu thay đổi** để lưu lên hệ thống. Đóng màn ghi chú bỏ phần
vừa sửa. **Thêm chi tiết** dùng cho tùy chọn **Tính vào báo cáo**.

AI-agent contract: note-sheet save is local draft-only; transaction submit alone
persists optional `note` through the existing transaction API. No endpoint/schema change.

Mở **Giao dịch** ở thanh điều hướng. Chọn **Ví → Tổng cộng** để xem giao dịch của mọi ví, hoặc chọn một ví thường để chỉ xem giao dịch của ví đó. Hai phạm vi dùng cùng giao diện: chọn kỳ, tổng **Tiền vào / Tiền ra** và danh sách giao dịch theo ngày.

**Tổng cộng**, ví thường và màn xem ví tín dụng hiện tại hiển thị **tuần này** (Thứ hai–Chủ nhật theo múi giờ tài khoản) nếu tuần có giao dịch. Nếu không, ứng dụng mở tuần gần nhất có giao dịch trong phạm vi ví đang chọn; nếu chưa có giao dịch nào thì giữ tuần này. Dùng mũi tên xem tuần cũ hoặc chọn khoảng ngày tùy chỉnh. Ví thường liệt kê mọi giao dịch của ví **trong kỳ đã chọn**, không lẫn giao dịch ví khác. Không có tab thời gian “Tất cả”.

Trên iPhone, ứng dụng ghi nhớ ví đang chọn trên thiết bị cho từng tài khoản đã đăng nhập. Dữ liệu tài chính không được lưu thêm vào tùy chọn này; nếu ví đã bị xóa hoặc không còn truy cập được, ứng dụng tự quay về **Tổng cộng**.

Khi chọn **ví tiết kiệm**, màn hình hiển thị số dư, phần còn thiếu để đạt mục tiêu, hạn/tiến độ và lịch sử giao dịch của ví theo tháng. Dòng chữ **Tất cả các giao dịch** là tiêu đề lịch sử của ví tiết kiệm, không phải bộ lọc thời gian chung.

Khi xem **Tổng cộng**, ví thường hoặc ví tín dụng, có hai cách chọn thời gian:

- **Theo tuần**: tuần hiện tại, từ Thứ hai đến Chủ nhật theo múi giờ tài khoản. Dùng mũi tên trái/phải để xem các tuần cũ và quay lại. Không thể đi quá tuần hiện tại.
- **Tùy chọn**: chọn khoảng **Từ ngày / Đến ngày**. Nhấn **Xong** để áp dụng; đóng hộp thoại thì bộ lọc cũ vẫn giữ nguyên. Chọn lại **Theo tuần** để quay về tuần hiện tại.

Khoảng ngày đang xem lọc cả danh sách lẫn tổng thu/chi. Nếu khoảng đó không có giao dịch, màn hình nói rõ không có kết quả; dữ liệu giao dịch không bị xóa. Chuyển tab hoặc đổi ví chỉ thay cách xem, không tạo/sửa giao dịch.

Giao dịch thu/chi có tùy chọn **Tính vào báo cáo**. Khi tắt, lựa chọn này
được lưu cùng giao dịch và khoản đó không đóng góp vào tổng thu/chi, xu hướng
hoặc các tổng hợp báo cáo; giao dịch vẫn còn trong sổ và số dư ví vẫn phản ánh
giao dịch. Chuyển ví và điều chỉnh luôn nằm ngoài các tổng thu/chi báo cáo.

Mỗi nhóm ngày trong danh sách dùng cùng cấu trúc trực quan: số ngày lớn,
thứ trong tuần và tháng/năm ở múi giờ tài khoản, kèm tổng thu/chi có dấu.
Khi bật **Xem theo nhóm**, tiêu đề chuyển sang tên danh mục và số giao dịch;
không đổi nguồn dữ liệu hay phạm vi lọc.
Mỗi dòng cũng dùng biểu tượng và tông màu của danh mục; tiền vào màu hành
động, khoản chi màu cảnh báo, còn chuyển ví/điều chỉnh dùng tông trung tính.

## Dành cho AI agent

| Trường | Giá trị |
| --- | --- |
| Route | `/transactions` |
| Kỳ tuần | Thứ hai 00:00 đến hết Chủ nhật theo `account.timezone`, lọc theo ngày lịch cục bộ (inclusive) |
| Phạm vi ví | `Tổng cộng` = giao dịch của mọi ví owner trong kỳ đã chọn; ví thường = chỉ giao dịch của ví trong cùng kỳ; ví tín dụng = phạm vi tuần/ngày đang chọn |
| Biến thể goal | Không có week strip; hiển thị tiến độ mục tiêu và toàn bộ lịch sử thực của ví theo tháng |
| Biến thể basic | Cùng layout và kỳ với `Tổng cộng`; bộ lọc ví áp dụng đồng thời cho tổng tiền và từng dòng |
| Nguồn dữ liệu | `GET /api/v1/transactions`, `GET /api/v1/wallets`; lọc hiển thị ở frontend |
| Tác động dữ liệu | Không ghi dữ liệu, không thay đổi report hoặc số dư |
| Tính vào báo cáo | `included_in_reports=false` được lưu rõ ràng; giao dịch vẫn ở sổ và số dư, nhưng không được cộng vào số liệu Reports |
| Hạn chế | Không có endpoint kỳ tuần riêng; Travel Mode/chuyến đi không phải loại ví và chưa có màn riêng |

Đặc tả kỹ thuật: [trang giao dịch](../design/pages/transactions/README.md). Trạng thái Travel Mode: [ticket](../work/tickets/TICKET-05-03-su-kien-travel-mode.md).
