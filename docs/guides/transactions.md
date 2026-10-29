# Xem giao dịch theo ví và theo tuần

Trên iPhone, nút **+** cũng có ở trang chính **Tài khoản / Mở rộng**. Nhấn để
thêm giao dịch; **Hủy** hoặc lưu xong sẽ đưa bạn về trang Tài khoản. Trong các
màn con, dùng nút thêm riêng của chức năng đó. Bạn có thể vuốt từ cạnh trái để
quay về từ màn con; khi đang sửa bản nháp, hãy dùng **Hủy/Quay lại** trong màn
để tránh mất nội dung ngoài ý muốn. Các dòng giao dịch có khoảng cách dọc rộng
hơn, vẫn giữ nguyên số tiền và cách nhóm dữ liệu.

## Ghi chú trên iPhone

Trong màn thêm/sửa giao dịch, chạm dòng **Ghi chú** để mở ô ghi chú riêng, có thể
nhập nhiều dòng. **Lưu ghi chú** đưa nội dung về bản nháp giao dịch; bạn vẫn phải
nhấn **Lưu giao dịch/Lưu thay đổi** để lưu lên hệ thống. Đóng màn ghi chú bỏ phần
vừa sửa. **Thêm chi tiết** dùng cho tùy chọn **Tính vào báo cáo**.
Ô chọn này mặc định được bật khi thêm giao dịch. Bỏ chọn trước khi lưu để giao
dịch vẫn nằm trong sổ và số dư ví nhưng không tính vào báo cáo; mở lại màn sửa
sẽ thấy lựa chọn đã lưu.

AI-agent contract: note-sheet save is local draft-only; transaction submit alone
persists optional `note` through the existing transaction API. No endpoint/schema change.

Màn thêm/sửa dùng một vùng cuộn cho các trường và giữ nút lưu cố định ở dưới,
nên bàn phím không che thao tác lưu. Các trường Ví, Số tiền, Danh mục, Ghi chú,
Ngày và Hũ (khi phù hợp) nằm trong cùng nhóm; **Thêm chi tiết** mở tùy chọn
**Tính vào báo cáo**. Màn chi tiết dùng cùng cấu trúc nhóm để hiển thị số tiền,
danh mục, ví, ghi chú, ngày và trạng thái báo cáo trước các hành động hợp lệ.
Chuyển ví, điều chỉnh số dư và bút toán tín dụng vẫn dùng quy tắc sửa/xóa riêng.

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
Khi xem theo ngày, tên nhóm hoặc loại khoản là dòng chính của giao dịch; khi
**Xem theo nhóm**, dòng chính chuyển thành ngày giao dịch. Dòng phụ hiện ghi chú
nếu có; khi không có ghi chú, dòng này hiện ngày ngắn theo múi giờ tài khoản.
Tên ví không lặp lại trên từng giao dịch: chọn ví ở bộ lọc phía trên để xem
phạm vi ví đang dùng.

## Tìm giao dịch trên iPhone

Trong sổ **Giao dịch** của **Tổng cộng** hoặc một ví thường, nhấn biểu tượng
kính lúp. Ứng dụng cho biết ví và kỳ đang được tìm. Nhập một phần tên danh mục
hoặc ghi chú; chữ hoa/chữ thường không ảnh hưởng kết quả. Tìm kiếm chỉ dùng các
giao dịch đã tải trong đúng ví và kỳ đang chọn, nên không phải là tìm toàn bộ
lịch sử trên máy chủ.

Nút xóa trong ô tìm kiếm xóa từ khóa nhưng vẫn giữ chế độ tìm. Nút đóng rời chế
độ tìm và đưa danh sách đầy đủ của kỳ trở lại. Nếu không có kết quả, ứng dụng
hiển thị trạng thái trống; giao dịch không bị xóa. Số lượng kết quả và tổng ở
từng nhóm chỉ mô tả các dòng khớp, còn **Tiền vào / Tiền ra / Chênh lệch** vẫn
mô tả toàn bộ ví và kỳ đang chọn. Các bản nháp đang chờ gửi vẫn nằm ở khu vực
hàng đợi riêng và không tham gia tìm kiếm.

## Dành cho AI agent

| Trường | Giá trị |
| --- | --- |
| Route | `/transactions` |
| Native Account Add | Account root → ordinary transaction editor → save/cancel restores Account; child destinations and Assistant hide global Add |
| Native Back | Scoped native interactive pop for More secondary routes; editors, dirty month notes and Account drafts block gesture; explicit controls remain |
| Native row geometry | Shared transaction row: 44pt minimum content plus 6pt top/bottom padding; Dynamic Type can expand |
| Native row content | Category/type title by day, transaction date title when grouped by category; note when present, otherwise account-local short date; signed amount; no wallet name in row. Scope selector owns wallet context. |
| Native editor shell | One keyboard-safe scroll region plus shared pinned footer; explicit Cancel and Save; grouped wallet/amount/category/note/date/jar/report rows |
| Native detail content | Maskable amount plus read-only category, wallet, optional note, account-local date and report-inclusion status; action availability follows ordinary/transfer/adjustment/credit policy |
| Kỳ tuần | Thứ hai 00:00 đến hết Chủ nhật theo `account.timezone`, lọc theo ngày lịch cục bộ (inclusive) |
| Phạm vi ví | `Tổng cộng` = giao dịch của mọi ví owner trong kỳ đã chọn; ví thường = chỉ giao dịch của ví trong cùng kỳ; ví tín dụng = phạm vi tuần/ngày đang chọn |
| Biến thể goal | Không có week strip; hiển thị tiến độ mục tiêu và toàn bộ lịch sử thực của ví theo tháng |
| Biến thể basic | Cùng layout và kỳ với `Tổng cộng`; bộ lọc ví áp dụng đồng thời cho tổng tiền và từng dòng |
| Nguồn dữ liệu | `GET /api/v1/transactions`, `GET /api/v1/wallets`; lọc hiển thị ở frontend |
| Tìm kiếm native | Chỉ aggregate/basic ledger; local substring match không phân biệt hoa thường trên category display name hoặc optional note, sau wallet+period filter; query rỗng là prompt state; không tìm pending queue/server history |
| Tổng khi tìm kiếm | Income/expense/net dùng toàn bộ confirmed rows trong wallet+period; result count, groups và group totals dùng matched rows |
| Tác động dữ liệu | Không ghi dữ liệu, không thay đổi report hoặc số dư |
| Tính vào báo cáo | `included_in_reports=false` được lưu rõ ràng; giao dịch vẫn ở sổ và số dư, nhưng không được cộng vào số liệu Reports |
| Mặc định biểu mẫu | Giao dịch thu/chi mới: `included_in_reports=true`; màn sửa khôi phục giá trị đã lưu; ô chọn nằm trong **Thêm chi tiết** |
| Hạn chế | Không có endpoint kỳ tuần riêng; Travel Mode/chuyến đi không phải loại ví và chưa có màn riêng |

Nếu màn chi tiết được mở từ một luồng cũ chưa cung cấp đủ tên danh mục hoặc tên
ví cho phần trình bày, ứng dụng bỏ các dòng ngữ cảnh đó thay vì đoán rằng ví đã
bị xóa. Số tiền, ghi chú, ngày, trạng thái báo cáo và các hành động hợp lệ vẫn
giữ nguyên; không có dữ liệu giao dịch nào bị thay đổi bởi cách hiển thị này.

Đặc tả kỹ thuật: [trang giao dịch](../design/pages/transactions/README.md). Trạng thái Travel Mode: [ticket](../work/tickets/TICKET-05-03-su-kien-travel-mode.md).
