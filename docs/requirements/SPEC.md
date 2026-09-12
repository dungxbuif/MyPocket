---
artifact_type: requirement_spec
id: REQ-MASTER
status: in_review
owner: human
human_fields: [product_summary, goals, non_goals, users_and_stakeholders, acceptance_criteria]
ai_fields: [functional_requirements, non_functional_requirements, constraints, linked_decisions]
shared_fields: [status, trace]
updated: 2026-09-13
trace:
  context: ../CONTEXT.md
  backlog: ../work/BACKLOG.md
  requirements: REQUIREMENTS.md
  business_rules: BUSINESS_RULES.md
  reports: REPORTS.md
  user_stories: USER_STORIES.md
  validation: ../work/VALIDATION_MATRIX.md
  release_notes: ../releases/CHANGELOG.md
---

# MyPocket — Đặc tả sản phẩm

## Mục tiêu và trạng thái

MyPocket là ứng dụng tài chính cá nhân ưu tiên tiếng Việt, dùng trên web/PWA. Người dùng ghi chép giao dịch, quản lý ví, budget, hũ, tiết kiệm, tín dụng, khoản nợ và danh mục tài sản; xem báo cáo và dùng AI hỗ trợ nhập liệu, phân tích.

Tài liệu mô tả toàn bộ chức năng để chủ sản phẩm review. Đây là yêu cầu thiết kế, không phải tuyên bố đã triển khai. Các tình huống còn cần thống nhất nằm trong [BUSINESS_RULES.md](BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).

Quyết định trực tiếp của chủ sản phẩm có ưu tiên cao nhất. Tài liệu nghiên cứu và ứng dụng tham chiếu cung cấp bối cảnh; nội dung cũ không tự ghi đè đặc tả này. Chủ sản phẩm sở hữu ý định và chấp nhận nghiệp vụ; AI duy trì mô tả, truy vết và bằng chứng.

## Account, tiền tệ và thời gian

- Account quản lý dữ liệu riêng, cấu hình hiển thị và múi giờ riêng.
- Seed cứng một tiền tệ `VND`; số tiền giao dịch là số nguyên đồng. Nhập liệu tự dùng VND.
- **Ghi chú tiền tệ:** đa tiền tệ chưa được thiết kế. Không suy diễn trường currency thành hợp đồng tỷ giá, quy đổi hoặc báo cáo nhiều tiền tệ. Dữ liệu hiện tại đều có ý nghĩa VND.
- Thời điểm thực tế lưu UTC. Ngày thuần có thể dùng `YYYY-MM-DD`; tháng cấu hình và ghi chú dùng `YYYY-MM`.
- Query ngày/tháng lấy múi giờ trong account để xác định ranh giới UTC; không tự dùng múi giờ máy chủ hoặc trình duyệt.

## Ví và danh mục

| Thành phần | Chức năng |
| --- | --- |
| Ví thường (`basic`) | Tên, số dư đầu, giao dịch, số dư hiện tại, lựa chọn tính vào tổng |
| Ví tiết kiệm (`goal`) | Mục tiêu, hạn tùy chọn, nạp/rút, tiến độ từ số dư, đạt mục tiêu và theo dõi lại khi số dư giảm |
| Ví tín dụng (`credit`) | Hạn mức, sao kê, hạn thanh toán, mua hàng, hoàn tiền, phí/lãi, trả một phần/toàn bộ, quá hạn |
| Tổng ví | View tổng hợp các ví được chọn; không phải đối tượng nhận giao dịch |

User thêm, sửa, xóa ví. Xóa cần nêu rõ dữ liệu liên quan và ảnh hưởng báo cáo; quy tắc xử lý giao dịch nối hai ví cần thống nhất trước triển khai.

Danh mục phân loại thu, chi và các nhóm nghiệp vụ liên quan. Giữ bộ nhóm mặc định trong tài liệu sản phẩm cũ. User quản lý danh mục cá nhân, cha/con tối đa hai cấp, thứ tự và phạm vi sử dụng theo ví.

## Giao dịch

User tạo, sửa, xóa, nhân bản, tìm kiếm, lọc và thao tác hàng loạt. Giao dịch có số tiền, loại, ví, thời điểm/ngày nghiệp vụ, danh mục phù hợp, ghi chú, chứng từ và liên kết ngữ cảnh khi có.

Nghiệp vụ gồm thu, chi thường, chuyển ví, phí chuyển, điều chỉnh số dư, giải ngân/thu hồi/trả nợ và hoạt động ví tiết kiệm/tín dụng. Hiệu ứng số dư và phân loại báo cáo dùng chung quy tắc.

Chi thường có thể gắn tối đa một hũ. Hũ, danh mục và sự kiện/chuyến đi là các chiều theo dõi độc lập. Xuất hiện ở nhiều báo cáo không nhân đôi hiệu ứng tiền.

## Budget

Budget dùng hành vi Money Lover làm cơ sở: số tiền dự kiến chi theo ví, danh mục và kỳ. Tiến độ tự tính từ giao dịch khớp phạm vi, gồm danh mục con theo quy tắc tránh đếm trùng.

Chức năng gồm tạo/sửa/xóa; kỳ tuần, tháng, quý, năm hoặc ngày tùy chỉnh; lặp kỳ cố định; kỳ đang chạy/đã kết thúc; đã chi/còn lại, cảnh báo, giao dịch chi tiết, chi theo ngày và dự báo cuối kỳ. Kỳ tùy chỉnh không tự lặp. Budget kết thúc có thể xem/xóa; sửa giao dịch nguồn vẫn cập nhật số liệu.

Budget và hũ là hai chức năng riêng. User không gắn giao dịch trực tiếp vào budget.

## Hũ

Hũ là nhóm theo dõi chi tiêu, lấy cảm hứng từ quy tắc sáu hũ. User thêm/sửa/xóa hũ; số lượng và tên không bị cố định theo một bộ sáu hũ.

- Cấu hình theo tháng; tháng mới lấy cấu hình tháng gần nhất đã có.
- Mức phân bổ tùy chọn: % tổng thu thực tế trong tháng hoặc số tiền nhập tay.
- Mức phân bổ chỉ tham khảo. Vượt mức hoặc phân bổ vượt thu tạo cảnh báo, không chặn lưu giao dịch/cấu hình hợp lệ.
- Một hũ có nhiều chi thường. Chưa gắn hũ vẫn lưu giao dịch và xuất hiện trong báo cáo chung.
- Hũ có giao dịch nhưng chưa có mức phân bổ vẫn báo cáo chi thực tế.
- Không tự chuyển dư sang tháng sau. Sao chép cấu hình không sao chép số đã chi hoặc dư.
- Trang Hũ có view tháng và cộng dồn: tổng chi, diễn biến theo tháng; thêm tổng phân bổ/chênh lệch tham khảo khi có dữ liệu.
- Chênh lệch phân bổ trừ chi không phải tiền thật trong ví. Nếu chỉ một số tháng có phân bổ, view nêu rõ phạm vi có cấu hình.

## Recurring, sự kiện và Travel Mode

Lịch do user tạo/bật tự sinh giao dịch bình thường khi đến hạn. Giao dịch sửa/xóa độc lập; sửa lịch tác động các lần tiếp theo. Ghi chú mặc định: `Giao dịch định kỳ — {nội dung}`, lấy ghi chú lịch hoặc tên lịch. Một lần đến hạn chỉ sinh một giao dịch kể cả khi chạy lại.

Sự kiện/chuyến đi nhóm giao dịch để xem tổng và bối cảnh. Account có tối đa một Travel Mode đang bật. Giao dịch mới đủ điều kiện tự gắn chuyến đang hoạt động; tắt/đổi chế độ không viết lại giao dịch cũ. User được sửa liên kết sau đó. Giao dịch từ recurring không được tự gắn Travel Mode.

Chuyến đi được tự dùng làm bối cảnh báo cáo tháng, giảm nhu cầu gõ lại mô tả vào từng note. Nó không ghi đè nội dung user đã viết. Nhập bù/offline là tình huống cần thống nhất riêng.

## Tiết kiệm, tín dụng và khoản nợ

Ví tiết kiệm sở hữu mục tiêu và tiến độ từ hoạt động thực tế. Chuyển giữa các ví của cùng account chỉ đổi nơi giữ tiền, không tự tạo thêm thu/chi thường.

Ví tín dụng thể hiện dư nợ, hạn mức còn lại, số sao kê, số còn phải trả và hạn thanh toán. Mua hàng ghi nhận chi; trả nợ không tính cùng chi phí lần thứ hai. Hoàn tiền, phí/lãi, trả dư và phân bổ thanh toán cần ví dụ đầy đủ trước triển khai.

Khoản vay/cho vay gồm đối tác, chiều vay/cho vay, gốc, ngày/hạn, giải ngân và các lần thanh toán liên kết. Phải đối chiếu được số còn phải thu/trả; phân biệt tiền gốc với thu/chi thường.

## Reports và Money Insider

Reports tập trung các báo cáo tổng quan, dòng tiền, danh mục, xu hướng, budget, hũ tháng/cộng dồn, tiết kiệm, tín dụng, nợ, sự kiện/du lịch, portfolio và tài sản ròng. Money Insider tổng hợp so sánh, dự báo, điểm nổi bật và dẫn về báo cáo chi tiết.

Trong tháng xem realtime. Cuối tháng tổng kết/chốt kỳ để nhìn lại; số liệu vẫn tính lại nếu dữ liệu nguồn đổi. Chốt kỳ không khóa sửa giao dịch hoặc đóng băng số lịch sử.

Mỗi account có một note cho mỗi tháng; user được sửa/xóa chủ động. Tính lại báo cáo không xóa/viết lại note. Báo cáo tự đính kèm bối cảnh có dữ liệu như du lịch, hũ, tiết kiệm và tín dụng.

AI tạo phần tổng kết dễ đọc từ số liệu và bối cảnh thực tế. Phần AI phân biệt với note user, có thời điểm/phạm vi dữ liệu, có thể tạo lại khi dữ liệu đổi. Hệ thống tính công thức; AI không tự tạo số liệu. Chi tiết tại [REPORTS.md](REPORTS.md).

## Portfolio

Theo dõi tài sản, đơn vị/số lượng, mua/bán, giá, phí, giá trị thị trường và lãi/lỗ. Giá nhập tay hoặc từ nguồn được cấu hình; phân biệt giá hiện tại, cũ và thiếu.

Giá vốn dùng **bình quân gia quyền di động**: mỗi lần mua tính lại bình quân của lượng đang giữ; khi bán dùng giá bình quân ngay trước lúc bán. Ví dụ mua một đơn vị 100.000 và một đơn vị 140.000 thì bình quân 120.000; bán một đơn vị 150.000 lãi 30.000 khi chưa tính phí.

Biến động giá chỉ đổi định giá/lãi lỗ, không tự sửa số dư ví hoặc thu/chi. Quan hệ mua/bán tài sản với thanh toán từ ví còn cần thống nhất.

## API key và API

API key có tên, user tạo, secret hiển thị một lần, có thể thu hồi. Key thực hiện chức năng chủ account thao tác được trên UI trong cùng phạm vi sở hữu và quy tắc nghiệp vụ. Key không tạo/quản lý key khác; được xem thông tin an toàn/thu hồi chính key đang dùng.

API có hợp đồng, lỗi rõ ràng, phân trang, chống tạo trùng khi retry và phát hiện phiên bản cũ. Quyền API không bỏ qua điều kiện nghiệp vụ trên UI.

## AI nhập liệu, tư vấn và OCR

- AI nhập liệu nhận nội dung/chứng từ, hỏi dữ liệu thiếu, tạo đề xuất sửa/xác nhận được; chỉ xác nhận mới tác động số dư.
- OCR đọc chứng từ thuộc account để review; lỗi không tự tạo giao dịch.
- AI tư vấn dùng công cụ chỉ đọc lấy dữ liệu mới, nêu phạm vi/nguồn và phân biệt dữ kiện với suy luận.
- AI tổng kết tháng chỉ cập nhật phần tự sinh, không sửa giao dịch hoặc note thủ công.
- Quyền của từng loại AI vẫn áp dụng khi user gọi qua API key.

## Offline, thông báo và vòng đời dữ liệu

PWA lưu dữ liệu riêng theo user, ghi nhận offline ở chức năng được hỗ trợ, đồng bộ có chống trùng và xung đột rõ ràng. Báo cáo cache thể hiện dữ liệu chưa đồng bộ. Đăng xuất dọn dữ liệu riêng khỏi trình duyệt.

Thông báo trong app lưu các sự kiện đến hạn, cảnh báo và xung đột; push là kênh bổ sung. Export lấy đúng dữ liệu account. Xóa/reset nêu rõ ảnh hưởng, xử lý liên kết, tệp riêng, key và cache nhất quán.

## Điều kiện review

- Chức năng có mã yêu cầu và tiêu chí quan sát được trong [REQUIREMENTS.md](REQUIREMENTS.md), [USER_STORIES.md](USER_STORIES.md).
- Giao dịch, budget, hũ và reports dùng thống nhất tập dữ liệu/thời gian.
- Cảnh báo hũ không chặn lưu giao dịch.
- Báo cáo lịch sử cập nhật; note user độc lập với số liệu và AI.
- Tình huống chưa chốt không được xem là đã duyệt.
- [Validation matrix](../work/VALIDATION_MATRIX.md) tách review tài liệu khỏi bằng chứng triển khai.
