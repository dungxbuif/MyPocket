---
artifact_type: business_rules
id: BUSINESS-RULES
status: in_review
owner: shared
updated: 2026-09-13
human_fields: [business_intent, acceptance, open_questions]
ai_fields: [rule_rows, examples, trace_links]
shared_fields: [status, trace]
trace:
  spec: SPEC.md
  reports: REPORTS.md
  requirements: REQUIREMENTS.md
  stories: USER_STORIES.md
  context: ../CONTEXT.md
  validation: ../work/VALIDATION_MATRIX.md
  release_notes: ../releases/CHANGELOG.md
---

# MyPocket — Quy tắc nghiệp vụ

Đây là hợp đồng nghiệp vụ đang review. Mục cuối tách riêng tình huống chưa thống nhất; các đề xuất từ cuộc rà gap không tự trở thành quyết định.

## ACC — Account và tiền tệ

- ACC-01: Mỗi đối tượng riêng tư thuộc một account; truy vấn, chỉnh sửa, export, sync và AI đều giới hạn theo account đã xác thực.
- ACC-02: Tiền tệ được seed cố định `VND`, tự dùng khi nhập; tiền giao dịch là số nguyên đồng. Đa tiền tệ chưa được thiết kế; không có giả định về tỷ giá hoặc chuyển đổi.
- ACC-03: Cấu hình múi giờ/hiển thị lưu theo user/account. Thay đổi cấu hình không sửa giá trị tiền hoặc thời điểm UTC nguồn.

## TIME — Thời gian và phân kỳ

- TIME-01: Thời điểm xảy ra thực tế lưu UTC. Các API nhận thời điểm phải xác định offset hoặc chuyển từ giờ địa phương bằng timezone account; không đoán timezone máy chủ.
- TIME-02: Ngày nghiệp vụ thuần lưu kiểu date, biểu diễn `YYYY-MM-DD`; tháng cấu hình/note dùng `YYYY-MM`. Không biến một ngày thuần thành UTC nửa đêm rồi dịch sang ngày khác.
- TIME-03: Lọc theo ngày/tháng chuyển đầu kỳ và đầu kỳ kế tiếp trong timezone account thành khoảng UTC nửa mở `[start, next_start)`. Cùng một giao dịch chỉ thuộc một kỳ trong cùng cách query.
- TIME-04: Giao dịch có thời điểm phân kỳ theo thời điểm đó ở timezone account; dữ liệu ngày thuần phân kỳ theo ngày đã chọn. Kiểu dữ liệu phải thể hiện rõ ý nghĩa, không dùng lẫn hai cách cho cùng trường.
- TIME-05: Khi đổi timezone, các report dùng timestamp được tính lại theo cấu hình mới. Note vẫn gắn với nhãn tháng user đã ghi; không tự di chuyển hoặc viết lại note. Report/AI ghi nhận timezone đã dùng để nhận biết kết quả cũ.
- TIME-06: Recurring dùng lịch địa phương của account; kho dữ liệu ghi nhận thời điểm chạy UTC. Quy tắc đổi timezone khi có kỳ chờ chạy cần thống nhất cùng xử lý chạy bù.

## WAL, CAT — Ví và danh mục

- WAL-01: Loại ví là `basic`, `goal`, `credit`. Tổng ví là view các ví được chọn.
- WAL-02: Số dư đến từ số dư đầu và giao dịch đã ghi sổ; draft, OCR, AI và biến động giá tài sản không làm đổi số dư.
- WAL-03: User tạo/sửa/xóa ví. Xóa là thao tác loại bỏ dữ liệu được xác nhận; màn hình phải nêu các dữ liệu/giao dịch liên quan sẽ bị xóa và ảnh hưởng báo cáo. Xử lý các liên kết sang ví còn tồn tại phải được chốt trước triển khai.
- WAL-04: Lựa chọn ví tính vào tổng áp dụng nhất quán ở dashboard, report và thành phần tài sản ròng. Bộ lọc ví cụ thể phải hiển thị rõ phạm vi đang dùng.
- WAL-05: Cho phép các ví trùng tên trong cùng account; định danh và liên kết dùng ID, không dùng tên.
- CAT-01: Giữ nhóm danh mục mặc định đã có; danh mục tương thích với loại giao dịch và có tối đa hai cấp, không có vòng lặp.
- CAT-02: User quản lý danh mục cá nhân và thứ tự/phạm vi sử dụng; cách xóa danh mục đang có giao dịch cần bảo toàn tính giải thích được của lịch sử.
- CAT-03: Quản lý nhóm nằm trong tab Tài khoản. Chọn nhóm có hành động Sửa, cho chọn nhóm cha và ví áp dụng; yêu cầu trường “category” được ghi trong [design](../design/system/DESIGN.md#account--quản-lý-nhóm) và đang chờ làm rõ ý nghĩa.

## TX — Giao dịch và số dư

- TX-01: Số tiền giao dịch thường dương, chính xác tới đồng. Chiều tác động do loại nghiệp vụ quyết định.
- TX-02: Thu tăng tiền, chi giảm tiền ở ví thường/tiết kiệm. Chuyển ví có hai vế và được ghi nhận nguyên tử; không tạo thu/chi thường.
- TX-03: Phí chuyển là hiệu ứng chi riêng; không giấu phí trong số tiền chuyển. Điều chỉnh số dư là chênh lệch tới số đích, không tự phân loại thành thu/chi thường.
- TX-04: Sửa/xóa phải đảo hiệu ứng cũ rồi áp dụng hiệu ứng mới trong cùng thao tác nhất quán. Cập nhật mọi report liên quan và vô hiệu kết quả AI cũ nếu cần.
- TX-05: Nhân bản là giao dịch mới, khác với retry cùng yêu cầu. Thao tác hàng loạt phải cho biết kết quả từng bản ghi.
- TX-06: Danh mục, hũ và chuyến đi là các chiều liên kết; không tự tạo thêm bút toán. Ảnh chứng từ là bằng chứng riêng tư; gỡ ảnh không sửa số tiền.
- TX-07: Giao dịch chưa gắn hũ vẫn hợp lệ. Chỉ chi thường được gắn hũ và tối đa một hũ tại một thời điểm.

## BUD — Budget

Budget theo mô hình Money Lover được chuẩn hoá trực tiếp trong tài liệu nghiệp vụ hiện hành; các dòng M20–M24 dưới đây là contract của MyPocket, không phải bằng chứng triển khai từ ứng dụng khác.

- BUD-01: Budget gồm số tiền, phạm vi ví/danh mục và kỳ. Giao dịch khớp phạm vi tự vào tiến độ; không có yêu cầu gắn trực tiếp một budget vào giao dịch.
- BUD-02: Kỳ gồm tuần, tháng, quý, năm, ngày tùy chỉnh. Kỳ cố định có lựa chọn lặp; kỳ tùy chỉnh không lặp.
- BUD-03: Không tạo budget trùng phạm vi cùng ví/danh mục và thời gian chồng lấn. Khi danh mục cha/con hoặc toàn bộ danh mục cùng hiện diện, tổng overview phải khử đếm trùng.
- BUD-04: Budget đang chạy được quản lý; budget kết thúc xem/xóa được. Sửa giao dịch lịch sử cập nhật số tiến độ lịch sử, không tự đổi cấu hình budget.
- BUD-05: Hiển thị đã chi, còn lại, tỷ lệ, cảnh báo, giao dịch, chi theo ngày, gợi ý chi/ngày và dự báo. Cảnh báo 75% theo nguồn tham chiếu; vượt 100% hiển thị vượt mức. Cảnh báo được chống gửi trùng theo budget/kỳ/ngưỡng.
- BUD-06: Ngân sách tham khảo không phải số dư ví. Việc hũ gắn hay không gắn không làm thay đổi tập giao dịch budget.

## JAR — Hũ theo tháng và cộng dồn

- JAR-01: Hũ là nhóm theo dõi chi tiêu tương tự một chiều phân loại. Cảm hứng sáu hũ không áp đặt tên, số lượng hoặc tỷ lệ bắt buộc.
- JAR-02: User thêm/sửa/xóa hũ theo tháng. Tháng mới lấy cấu hình tháng gần nhất đã có của account; tháng đã có cấu hình không bị tự ghi đè khi tháng cũ được sửa.
- JAR-03: Một hũ có định danh ổn định qua các tháng để query cộng dồn; cấu hình tên/mức phân bổ của từng tháng giữ riêng. Hũ khác nhau nhưng trùng tên không tự được nhập làm một.
- JAR-04: Hũ có thể chỉ có giao dịch để theo dõi. Mức phân bổ tùy chọn theo % tổng thu thực tế trong tháng hoặc số tiền nhập tay; không dùng thu nhập dự kiến để tính phần trăm.
- JAR-05: Tổng % vượt 100%, tổng phân bổ vượt thu, hoặc chi vượt phân bổ đều chỉ cảnh báo. Không chặn lưu vì vượt kế hoạch; vẫn kiểm tra định dạng số, sở hữu và loại giao dịch hợp lệ.
- JAR-06: Tiến độ tính từ chi thường đang gắn hũ, theo ngày/tháng và phạm vi report. Sửa tiền/ngày/liên kết hoặc xóa giao dịch tính lại các kỳ bị ảnh hưởng.
- JAR-07: Không tự chuyển phần dư/âm sang tháng sau. Sao chép cấu hình chỉ sao chép hũ và cách đặt phân bổ, không sao chép chi thực tế hoặc chênh lệch.
- JAR-08: Không có mức phân bổ thì hiển thị tổng chi và cơ cấu/thời gian chi; không giả định phân bổ bằng 0 và không tạo cảnh báo vượt mức giả.
- JAR-09: Thiếu dữ liệu thì hiển thị trạng thái rỗng hoặc bỏ phần phân tích không có dữ liệu. Giao dịch chưa phân hũ vẫn có thể xem ở bộ lọc riêng và báo cáo chung.
- JAR-10: Cộng dồn theo hũ từ tháng đầu có dữ liệu đến tháng đang xem hoặc khoảng tháng user chọn. Luôn có tổng chi và diễn biến từng tháng; tổng phân bổ/chênh lệch chỉ tính trong các tháng có cấu hình tương ứng và phải nêu rõ độ bao phủ.
- JAR-11: Chênh lệch cộng dồn là phép tổng hợp để nhìn lại kế hoạch, không phải tiền trong một ví và không cấp thêm phân bổ cho tháng mới.

### Công thức và ví dụ hũ

```text
Phân bổ % tháng = tổng thu thực tế tháng × tỷ lệ / 100
Chi tháng       = tổng chi thường hợp lệ gắn hũ trong tháng
Chênh lệch      = phân bổ tháng − chi tháng (khi có phân bổ)
Chi cộng dồn    = tổng chi tháng trong phạm vi đã chọn
```

Ví dụ tháng 1 phân bổ 1.000.000, chi 800.000; tháng 2 phân bổ 1.000.000, chi 1.100.000. View cộng dồn cho thấy phân bổ 2.000.000, chi 1.900.000, chênh lệch +100.000. Tháng 2 vẫn có mức riêng 1.000.000 và cảnh báo vượt 100.000; không được cộng dư 200.000 của tháng 1 vào mức tháng 2.

Nếu tháng 3 có chi 300.000 nhưng không đặt phân bổ, tổng chi ba tháng là 2.200.000. Chênh lệch +100.000 chỉ có ý nghĩa trên hai tháng đã đặt phân bổ; không trình bày thành số dư cho cả ba tháng.

## REC — Giao dịch định kỳ

- REC-01: User tạo lịch với tần suất, nội dung, dữ liệu giao dịch, thời điểm tiếp theo, ngày kết thúc tùy chọn; bật/tạm dừng/xóa lịch.
- REC-02: Lịch bật tạo giao dịch bình thường khi đến hạn. Mỗi lần đến hạn có danh tính riêng để retry không tạo trùng.
- REC-03: Giao dịch đã tạo được sửa/xóa độc lập. Sửa/tạm dừng/xóa lịch không viết lại các giao dịch trước đó. Xóa giao dịch đã sinh không làm retry tự tái sinh chính lần đó.
- REC-04: Note mặc định `Giao dịch định kỳ — {nội dung}`. Lấy note lịch nếu có, nếu không lấy tên lịch; cả hai trống dùng `Giao dịch định kỳ`. User có thể sửa note sau.
- REC-05: Giao dịch recurring không tự gắn Travel Mode, kể cả worker chạy trong chuyến đi. Đổi note không làm mất nguồn recurring.
- REC-06: Hũ không tồn tại/không được chọn không phải lý do chặn giao dịch thường hợp lệ. Quy tắc ánh xạ hũ của lịch qua tháng và lỗi dữ liệu bắt buộc cần thống nhất tiếp.

## TRV — Sự kiện và Travel Mode

- TRV-01: Một account có tối đa một Travel Mode đang bật. Chế độ thuộc một sự kiện/chuyến đi có tên và ngữ cảnh để báo cáo.
- TRV-02: Giao dịch mới đủ điều kiện được tự gắn chuyến đang bật. Bật chế độ không tự sinh giao dịch; tắt/đổi chuyến không viết lại liên kết cũ.
- TRV-03: Giao dịch recurring được loại khỏi tự gắn chuyến. Với AI/OCR, liên kết chỉ có hiệu lực kế toán khi giao dịch được xác nhận.
- TRV-04: User có thể sửa/gỡ liên kết chuyến. Liên kết chỉ tác động nhóm/báo cáo, không đổi số dư hay tự sửa note thủ công.
- TRV-05: Báo cáo tháng tự đính kèm chuyến liên quan và giao dịch có thật; AI có thể dùng ngữ cảnh này để giải thích diễn biến chi tiêu, không suy đoán hoạt động không có dữ liệu.

## SAV, CRD, DEBT — Tiết kiệm, tín dụng và nợ

- SAV-01: Ví tiết kiệm sở hữu mục tiêu, hạn tùy chọn và số hiện có. Tiến độ tính từ số dư thực tế; nạp/rút/sửa/xóa cập nhật tiến độ.
- SAV-02: Đạt mục tiêu khi số hiện có đạt mức mục tiêu. Khi rút làm thấp hơn mục tiêu, trạng thái phản ánh lại tiến độ hiện tại. Chuyển tiền vào tiết kiệm không nhân đôi chi/thu.
- CRD-01: Ví tín dụng quản lý hạn mức, dư nợ, ngày sao kê, hạn trả, các khoản mua/hoàn tiền/phí/lãi và thanh toán liên kết.
- CRD-02: Mua làm tăng dư nợ; trả nợ/hoàn tiền giảm dư nợ theo hợp đồng thanh toán. Hạn mức còn lại = hạn mức − dư nợ; đóng góp tài sản ròng là giá trị nghĩa vụ mang dấu âm, với trường hợp trả dư cần quy tắc riêng.
- CRD-03: Sao kê thể hiện khoản trong kỳ, số cần trả, đã trả, còn lại, ngày đến hạn và trạng thái chưa trả/trả một phần/đã trả/quá hạn. Tính trạng thái theo số liệu nguồn, không chỉ theo nút người dùng bấm.
- CRD-04: Chi mua hàng ghi một lần; thanh toán từ ví tiền sang thẻ xử lý hai vế, không thành chi lần hai. Lãi/phí là khoản riêng có nguồn rõ ràng.
- DEBT-01: Khoản vay/cho vay có đối tác, gốc, chiều, ngày/hạn và các lần giải ngân/thu hồi/trả liên kết. Số gốc còn lại đối chiếu được với các lần phát sinh.
- DEBT-02: Phải tách chuyển động tiền gốc, lãi và thu/chi thường trong reports. Quy tắc hoàn tiền và phân loại vào tổng thu của hũ còn cần thống nhất.

## AST — Portfolio

- AST-01: Vị thế có tài sản, đơn vị/số lượng, lịch sử mua/bán, giá/phí VND; số lượng dùng số thập phân chính xác.
- AST-02: Giá vốn dùng bình quân gia quyền di động. Mua cộng giá trị mua và phí mua vào tổng giá vốn rồi chia cho lượng mới. Bán loại giá vốn theo giá bình quân ngay trước bán; phí bán giảm tiền thu.
- AST-03: Bán không tự đổi giá bình quân phần còn lại. Khi hết vị thế, giá vốn còn lại về 0; lần mua mới bắt đầu giá vốn mới.
- AST-04: Xử lý lịch sử theo thứ tự xác định, gồm ngày/thời điểm và tiêu chí phụ ổn định. Sửa lịch sử cần tính lại giá vốn/lãi lỗ phía sau. Hành vi xóa khiến lượng bán sau đó không còn được bảo đảm cần thống nhất.
- AST-05: Giá có thời điểm và nguồn; thiếu giá là chưa biết, không phải giá 0. Giá cũ phải có nhãn thời điểm; lỗi nguồn giữ giá cuối và cho nhập tay.
- AST-06: Biến động giá không tạo thu/chi hoặc hiệu ứng tiền ví. Tổng ví, giá trị đầu tư và tài sản ròng phải trình bày thành các thành phần đối chiếu được.

```text
Giá vốn bình quân mới = (giá vốn còn lại + tiền mua + phí mua)
                       / (lượng còn lại + lượng mua)
Giá vốn lượng bán    = lượng bán × giá vốn bình quân trước bán
Lãi/lỗ đã thực hiện  = tiền bán − phí bán − giá vốn lượng bán
```

Ví dụ không phí: mua 1 đơn vị giá 100.000 và 1 đơn vị giá 140.000, bình quân 120.000. Bán 1 đơn vị giá 150.000: lãi 30.000, còn 1 đơn vị có giá vốn 120.000. Phép tính trung gian giữ độ chính xác; hiển thị VND làm tròn nhất quán, không làm trôi tổng giá vốn.

## REP, MONTH — Báo cáo, tổng kết tháng và ghi chú

- REP-01: Reports dùng dữ liệu đã ghi nhận trong phạm vi account/ví/ngày/timezone và lựa chọn tính vào báo cáo. Draft/proposal không có hiệu ứng tiền.
- REP-02: Cùng chỉ số và phạm vi phải cho cùng số trên dashboard, Reports và Insider. Drilldown giải thích được tổng; chuyển nội bộ/điều chỉnh không tự thành thu/chi thường.
- REP-03: Sửa/xóa dữ liệu nguồn cập nhật cả báo cáo hiện tại và lịch sử liên quan. Cache phải được làm mới hoặc ghi rõ thời điểm; có SQL không tự bảo đảm cache/AI đã cập nhật.
- MONTH-01: Trong tháng có preview realtime. Khi account-local calendar sang tháng mới, report của tháng trước tự mang trạng thái hoàn tất trên lần đọc kế tiếp; hệ thống không cần cron close, không đóng băng số liệu và không khóa giao dịch. Report luôn tính lại từ ledger hiện tại.
- MONTH-02: Một ghi chú user theo `(account, YYYY-MM)`. User sửa/xóa chủ động; query, thay đổi giao dịch, tổng kết tự động và tạo lại AI không ghi đè note.
- MONTH-03: Bối cảnh tự đính kèm từ dữ liệu thực: chuyến đi/sự kiện, hũ, tiết kiệm, tín dụng và các biến động liên quan. Nội dung tự sinh phân biệt với note user.
- MONTH-04: AI nhận dữ liệu báo cáo đã tính và bối cảnh đúng account/kỳ. Không tự tạo số, suy diễn lý do như một sự thật, hoặc sửa tài chính.
- MONTH-05: Nội dung AI gắn kỳ, timezone, phạm vi và phiên bản/thời điểm dữ liệu. Khi nguồn đổi, không được hiển thị kết luận cũ như kết quả mới; đánh dấu cần cập nhật và tạo lại phần AI. Lỗi AI không cản report, note hoặc trạng thái hoàn tất kỳ.
- MONTH-06: Note thuộc tháng của account, không nhân bản theo bộ lọc ví. Nội dung AI/bối cảnh theo bộ lọc phải có nhãn phạm vi; không gán kết luận một ví thành kết luận toàn account.

## KEY, API, AI — Tích hợp và trợ lý

- KEY-01: User tạo key có tên; plaintext hiển thị một lần, lưu đại diện một chiều, metadata an toàn và trạng thái thu hồi.
- KEY-02: Key có quyền chức năng tương đương chủ account, vẫn kiểm tra sở hữu và quy tắc. Key không tạo/liệt kê/quản lý key khác; xem metadata/thu hồi chính key đang dùng được phép.
- KEY-03: Thu hồi có hiệu lực với mọi đường xác thực/cache. Không ghi secret vào log; kiểm soát tần suất và truy vết theo key ID an toàn.
- API-01: API có hợp đồng công khai, lỗi ổn định, phân trang; retry cùng thao tác không thêm hiệu ứng. Sửa phiên bản cũ phải báo xung đột thay vì âm thầm ghi đè.
- AI-01: Nhập liệu tạo bản nháp từ từng giao dịch nhận diện được để user sửa/duyệt, không từ chối cả lô vì thiếu trường hoặc ảnh trùng. AI gộp bản sao rõ ràng trong cùng lần gửi; giữ các giao dịch khác ngày/mã tham chiếu và bản nghi trùng kèm câu hỏi. Tự chọn ví phù hợp trong danh sách, không rõ thì chọn ví đầu tiên làm mặc định; chưa có ví thì để trống để user bổ sung. Không bịa số tiền/ngày, không biến số dư thành giao dịch, không tự xác nhận. Tool chỉ gọi dịch vụ nghiệp vụ với quyền do server xác định.
- AI-02: Tư vấn chỉ đọc dữ liệu hiện tại; số có nguồn và phạm vi. Trí nhớ hội thoại không thay cho dữ liệu báo cáo.
- AI-03: OCR là tiền xử lý chứng từ riêng tư; output là dữ liệu chưa tin cậy, cần review. Lỗi/ảnh không thuộc account không được tạo giao dịch.
- AI-04: AI tổng kết tháng có thể tự sinh phần riêng; không cần biến thao tác tạo văn bản thành xác nhận giao dịch. Có nhãn AI, dữ liệu nguồn và trạng thái cũ/lỗi.

## SYNC, LIFE — Đồng bộ và dữ liệu

- SYNC-01: SQL là dữ liệu có thẩm quyền; cache/outbox theo account và có trạng thái chờ gửi/xung đột. Replay không tạo trùng hiệu ứng.
- SYNC-02: Đồng bộ kiểm tra phiên bản/sở hữu. Dữ liệu đã xóa không được âm thầm sống lại do thiết bị cũ gửi lại; chính sách xử lý xóa phải bao phủ outbox/cache.
- SYNC-03: Đăng xuất dọn dữ liệu riêng của account khỏi trình duyệt, gồm dữ liệu tài chính, ảnh chờ gửi và AI/cache.
- LIFE-01: Thông báo trong app bền vững; push là bổ sung. Cảnh báo hũ không khóa giao dịch.
- LIFE-02: Export phản ánh dữ liệu/phạm vi tại thời điểm xuất; không trở thành nguồn tự đồng bộ ngược.
- LIFE-03: Xóa/reset nêu rõ dữ liệu tác động và xử lý nhất quán database, liên kết, file, key/cache thuộc phạm vi. Retry không gây thêm lần xóa ngoài ý định.

## Các tình huống cần thống nhất tiếp

Đây là khoảng trống thiết kế để tiếp tục thảo luận, không phải quyết định mặc định đã được duyệt:

1. Xóa hũ/danh mục/ví đang có giao dịch: giữ giao dịch và bỏ liên kết, chuyển nhóm, hay xóa giao dịch phụ thuộc; riêng chuyển ví/trả thẻ cần xử lý vế ở ví còn lại. Đã chốt xóa thực, nhưng chưa có ma trận ảnh hưởng từng loại liên kết.
2. Travel Mode khi nhập giao dịch quá khứ, đồng bộ offline hoặc xác nhận draft sau khi đổi chuyến: chọn theo lúc ghi nhận, lúc giao dịch xảy ra hay lúc xác nhận.
3. Recurring ngày 29–31, chạy bù, đổi timezone và ví/hũ trong lịch bị xóa: cách chọn kỳ, báo lỗi và ánh xạ cấu hình. Hũ thiếu không chặn giao dịch hợp lệ đã được xác định ở REC-06.
4. Tổng thu thực tế của hũ và tiền hoàn mua hàng: cách xử lý thu hồi gốc, tiền vay, bán tài sản, hoàn tiền qua tháng và giao dịch loại khỏi báo cáo.
5. Credit: trả dư, hoàn tiền sau sao kê, thứ tự phân bổ khoản trả, phí/lãi và tác động của sửa giao dịch lên sao kê đã lập.
6. Portfolio: mua/bán có liên kết ví thanh toán hay chỉ ghi sổ tài sản; xóa/sửa lịch sử làm thiếu số lượng cho lần bán sau.

Các lựa chọn này phải được bổ sung trước khi triển khai phần phụ thuộc. Chúng không ngăn review những quy tắc báo cáo tháng, hũ mềm và thời gian đã được ghi ở trên.
