---
artifact_type: reporting_spec
id: REPORTS
status: in_review
owner: shared
updated: 2026-09-13
human_fields: [report_intent, acceptance]
ai_fields: [report_inventory, formulas, scenarios]
shared_fields: [status, trace]
trace:
  spec: SPEC.md
  business_rules: BUSINESS_RULES.md
  requirements: REQUIREMENTS.md
  stories: USER_STORIES.md
  validation: ../work/VALIDATION_MATRIX.md
---

# Reports, Money Insider và tổng kết tháng

## Nguyên tắc chung

Reports là trung tâm xem số liệu; Money Insider là phần phân tích tổng hợp dẫn về các report tương ứng. Mọi số liệu dùng cùng công thức và bộ lọc account/ví/ngày/timezone. Số tiền dùng VND. Hành vi nghiệp vụ nền ở các nhóm REP, MONTH, TIME và JAR của [BUSINESS_RULES.md](BUSINESS_RULES.md).

Report luôn thể hiện kỳ, bộ lọc, timezone và thời điểm dữ liệu. Drilldown mở đúng tập giao dịch tạo nên chỉ số. Khi lấy dữ liệu offline/cache phải có trạng thái chưa cập nhật. Dữ liệu lịch sử được query lại khi giao dịch thay đổi.

## Danh sách report và trang xem

| Report / view | Nội dung | Drilldown |
| --- | --- | --- |
| Tổng quan | Số dư đầu/cuối, thu, chi, thu trừ chi; nghĩa vụ và hiệu ứng khác trình bày riêng | Ví, loại nghiệp vụ, giao dịch |
| Dòng tiền | Thu/chi theo ngày/kỳ; chuyển nội bộ, điều chỉnh và tiền gốc có nhóm riêng | Giao dịch theo ngày/loại |
| Thu và chi theo danh mục | Số tiền, %, cha/con, dạng nhóm/phẳng, biểu đồ tròn/cột | Danh mục con và giao dịch |
| Chi tiết danh mục | Xu hướng, tổng, trung bình, so với kỳ trước và lịch sử ba tháng khi đủ dữ liệu | Giao dịch và nhóm con |
| Xu hướng / so sánh | Ngày, tuần, tháng, quý, năm, toàn bộ hoặc khoảng tùy chọn; kỳ hiện tại và lịch sử | Kỳ thành phần |
| Budget | Phạm vi ví/danh mục, mức, đã chi/còn lại, cảnh báo, khuyến nghị/ngày, dự báo | Budget và giao dịch khớp phạm vi |
| Hũ tháng | Các hũ có dữ liệu, chi thực tế, mức phân bổ nếu có, chênh lệch và cảnh báo | Chi thường thuộc hũ; chưa phân hũ |
| Hũ cộng dồn | Tổng chi từng hũ, diễn biến tháng, phân bổ/chênh lệch trong tháng có cấu hình | Hũ → tháng → giao dịch |
| Tiết kiệm | Mục tiêu, số hiện có, còn cần, nạp/rút và tiến độ | Ví tiết kiệm và hoạt động liên quan |
| Tín dụng | Dư nợ, hạn mức còn, sao kê, cần trả/đã trả/còn lại, hạn, quá hạn, phí/lãi | Sao kê và mua/hoàn/trả liên kết |
| Vay / cho vay | Gốc giải ngân, đã trả/thu hồi, còn phải thu/trả, đối tác, ngày đến hạn | Khoản nợ và từng lần thanh toán |
| Sự kiện / du lịch | Thời gian, giao dịch liên kết, thu/chi phù hợp, cơ cấu và xu hướng | Chuyến đi và giao dịch |
| Portfolio | Lượng, giá vốn bình quân, giá hiện tại/cũ/thiếu, thị giá, lãi/lỗ đã/chưa thực hiện | Tài sản, mua/bán và giá |
| Tài sản ròng | Các thành phần ví, nghĩa vụ và đầu tư; tổng và biến động có thể đối chiếu | Thành phần tương ứng |
| Tổng kết tháng | Số liệu cập nhật, note user, bối cảnh tự đính kèm, phần tổng kết AI | Các report ở trên cùng tháng |

Không có dữ liệu thì không tạo phân tích giả. Report cần cho biết rỗng, chưa cấu hình hoặc thiếu thông tin tùy trường hợp; không coi mọi giá trị thiếu là 0.

## Money Insider

### Thẻ tổng quan

Thẻ trang chủ nêu kỳ đang xem, tổng chi, danh mục phát sinh chi thường xuyên nhất, chi trung bình/ngày và so sánh kỳ trước. Tần suất theo số giao dịch khác với top danh mục theo số tiền; nhãn phải phân biệt hai chỉ số.

Thẻ báo cáo tháng có thể so với trung bình ba tháng trước khi đủ dữ liệu. Nếu lịch sử ít hơn, ghi rõ số kỳ thực dùng hoặc báo chưa đủ dữ liệu.

### Trang chi tiết

- Chọn ví, danh mục, tuần/tháng; biểu đồ sáu kỳ để thấy xu hướng.
- Thu/chi, tỷ lệ chi trên thu, trung bình ngày và so sánh cùng tiến độ kỳ trước.
- Đường tham chiếu budget chỉ xuất hiện khi phạm vi tương thích và đã khử đếm trùng.
- Dự báo chi cuối kỳ, top ba danh mục theo tiền, top năm giao dịch chi.
- Dẫn đến hũ tháng/cộng dồn, tiết kiệm, tín dụng/nợ, du lịch và tài sản khi có dữ liệu liên quan.
- Kết luận mô tả tăng/giảm và điểm nổi bật bằng số thật. Giải thích nguyên nhân chỉ dùng bối cảnh xác nhận được hoặc thể hiện rõ là nhận định.

Insider dùng số tính toán của report, không tạo một hệ thống tổng thu/chi riêng. Giao dịch recurring bị loại khỏi tự gắn du lịch vẫn là chi thực tế trong các report thường phù hợp.

## Công thức và trường hợp thiếu dữ liệu

```text
Thu trừ chi              = thu thường − chi thường
Chi trung bình/ngày      = chi trong kỳ tính tới hiện tại / số ngày đã qua
Tỷ lệ chi trên thu       = chi / thu × 100, khi thu > 0
Dự báo chi cuối kỳ       = chi / số ngày đã qua × số ngày toàn kỳ
Thay đổi so với kỳ trước = (hiện tại − trước) / trước × 100, khi trước > 0
```

- Ngày đã qua tính theo ngày địa phương account, tối thiểu 1 cho kỳ hiện tại có ngày hôm nay. Kỳ kết thúc dùng toàn bộ ngày kỳ; kỳ tương lai không giả tạo tốc độ chi.
- So sánh kỳ chưa kết thúc dùng đoạn tương ứng của kỳ trước, cắt tại cuối kỳ trước nếu ngắn hơn; nhãn ghi rõ phạm vi. Không so một phần tháng với cả tháng mà không thông báo.
- Thu bằng 0: tỷ lệ chi/thu là không áp dụng, vẫn hiển thị số chi. Kỳ trước bằng 0: hiển thị chênh lệch tiền và dữ liệu mới, không chia cho 0.
- Dự báo là ước tính từ tốc độ hiện tại, không phải giao dịch thật hoặc cam kết tương lai. Kỳ đã kết thúc hiển thị số thực tế.
- Số dư đầu/cuối phải tính cả các hiệu ứng phù hợp khác; không ép `cuối = đầu + thu − chi` nếu còn chuyển ví, điều chỉnh, gốc hoặc hoạt động tín dụng trong phạm vi đã chọn.
- Giá tài sản thiếu là chưa biết, không là 0. Thành phần tài sản ròng thiếu dữ liệu phải được nêu rõ thay vì trình bày tổng chưa đủ như tổng hoàn chỉnh.

## Tổng kết tháng: số liệu cập nhật, note giữ riêng

Trong tháng, user xem báo cáo realtime và ghi note bất kỳ lúc nào. Khi account-local calendar sang tháng mới, report tháng trước tự hiển thị trạng thái hoàn tất trên lần đọc kế tiếp; không cần job đóng tháng hay thao tác chốt tay. Số liệu luôn được tính lại từ ledger hiện tại, kể cả khi user sửa/nhập bù/xóa giao dịch cũ.

Màn hình có bốn phần độc lập:

| Phần | Nguồn | Khi dữ liệu tháng đổi |
| --- | --- | --- |
| Số liệu, biểu đồ | Query/report service | Cập nhật theo dữ liệu hiện tại; không phải snapshot bất biến |
| Ghi chú của tôi | User nhập, một note/account/tháng | Giữ nguyên trừ khi user sửa/xóa |
| Bối cảnh tháng | Liên kết sự kiện, chuyến đi, hũ và các hoạt động có thật | Tự đính kèm/cập nhật; không ghi đè note |
| Tổng kết AI | Dữ liệu report đã tính và bối cảnh | Nhận biết phiên bản cũ, tạo lại phần AI riêng |

Ví dụ tháng 8 có chuyến Đà Nẵng và user ghi “Tháng này chuyển việc”. Report tự đính kèm chuyến cùng tổng chi có nguồn. Khi user bổ sung hóa đơn tháng 8, số chi đổi và phần AI được làm mới; câu ghi chú của user vẫn giữ nguyên. Không cần tự viết lại chuyến đi vào note để report biết bối cảnh.

Note gắn nhãn `YYYY-MM` của account, không nhân bản theo ví hoặc theo mỗi lần query. Đổi timezone có thể làm giao dịch gần ranh giới vào tháng khác; note vẫn thuộc nhãn tháng cũ mà user đã chọn. Ngày thuần không bị dịch ngày.

## AI tổng kết và ngữ cảnh

AI được dùng cho preview và tổng kết tháng, tạo nhận xét từ dữ liệu đã có. Nội dung có nhãn AI, kỳ, timezone, bộ lọc và thời điểm/phiên bản nguồn. Số liệu quan trọng dẫn về report/giao dịch liên quan.

Thay đổi giao dịch, hũ, liên kết du lịch hoặc phạm vi query có thể làm nội dung AI cũ không còn phù hợp. Hệ thống đánh dấu chưa cập nhật và chỉ thay bằng kết quả tạo từ phiên bản phù hợp; một lần chạy AI cũ hoàn tất muộn không được ghi đè kết quả mới.

AI không tự suy ra user đi đâu, mua gì hay lý do chi tiêu khi không có dữ liệu. Có thể diễn giải dữ liệu như “chi du lịch tăng”, nhưng không coi tương quan là nguyên nhân đã xác nhận. Ghi chú user và OCR là nội dung đầu vào, không phải lệnh cho AI sửa tiền hoặc quyền truy cập.

Nếu AI lỗi/chưa sẵn sàng, số liệu, bối cảnh xác định được và note vẫn dùng bình thường. Tổng kết kỳ không phụ thuộc việc gọi AI thành công. Tạo lại văn bản AI không tạo hoặc sửa giao dịch.

## View hũ cộng dồn

User chọn hũ và khoảng tháng hoặc toàn bộ lịch sử. Cộng dồn dựa trên định danh hũ qua các cấu hình tháng, không dựa đơn thuần vào tên. Hiển thị tổng chi, số giao dịch, diễn biến tháng và danh mục chi.

Nếu có mức phân bổ, thêm tổng phân bổ và chênh lệch trên đúng các tháng có mức. Nếu có tháng chỉ có chi, tổng chi toàn kỳ vẫn hiển thị đầy đủ, nhưng không gọi chênh lệch phạm vi nhỏ hơn là “tiền còn trong hũ” cho toàn kỳ. Tháng mới không được nhận thêm phần dư chỉ vì view này đang cộng dồn.

## Tiêu chí kiểm tra

1. Cùng scope cho cùng tổng ở Reports, Insider và drilldown.
2. Sửa giao dịch tháng đã tổng kết cập nhật số/bối cảnh, giữ note user.
3. Recurring không xuất hiện trong tổng du lịch vì chế độ tự gắn; vẫn nằm trong chi thường phù hợp.
4. Hũ không có mức phân bổ vẫn có thống kê chi; vượt mức chỉ cảnh báo.
5. View cộng dồn không đếm trùng và không tự chuyển dư; tháng thiếu phân bổ được phân biệt.
6. Query dùng account timezone; date thuần và UTC không bị trộn nghĩa.
7. Thu 0, thiếu lịch sử, thiếu giá, AI lỗi hoặc AI cũ không tạo số liệu giả.
8. Note user không bị AI hoặc tác vụ nền ghi đè; lọc ví không tạo note tháng mới. Trạng thái hoàn tất kỳ được suy ra từ timezone account, không phụ thuộc scheduler.
