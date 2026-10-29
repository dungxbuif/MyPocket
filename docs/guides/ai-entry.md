# Nhập giao dịch bằng AI: kiểm tra chứng từ trước khi duyệt

AI có thể đọc nội dung bạn nhập và tệp JPEG, PNG, PDF hoặc XLSX để tạo các **đề xuất giao dịch**. Mỗi tệp tối đa 5 MiB; một lần gửi nhận tối đa 20 tệp. Kết quả chỉ là bản nháp để xem và sửa. Chưa có giao dịch nào được ghi vào sổ cho đến khi bạn duyệt từng đề xuất.

Với XLSX ngân hàng được hỗ trợ, ứng dụng đọc trực tiếp từng dòng: ngày, chiều thu/chi, số tiền VND, mã tham chiếu và định danh dòng không do AI đoán lại. AI chỉ đề xuất phần ngữ nghĩa như nhóm và ghi chú. Không gửi XLSX chung với ảnh/PDF trong một lượt; hãy tách thành hai yêu cầu để mỗi nguồn được xử lý an toàn.

## Kiểm tra từng dòng

1. Mở nhập liệu AI, thêm chứng từ và gửi một lần. Khi có kết quả, đối chiếu từng đề xuất với dòng tương ứng trên chứng từ: loại thu/chi, ví, nhóm, ngày, số tiền VND và ghi chú. Đừng dựa vào lời tóm tắt của AI để duyệt cả lô.
2. Nếu ngày để trống hoặc có câu hỏi về ngày, hãy chọn **ngày giao dịch trên chứng từ**. Ứng dụng không dùng ngày hôm nay để thay cho ngày không rõ.
3. Nếu số tiền là **0** hoặc có câu hỏi về ngoại tệ/đơn vị tiền, hãy kiểm tra nguồn và nhập số **VND đã xác minh**. AI không tự đổi ngoại tệ sang VND và không lấy phần nguyên của số tiền ngoại tệ. Đừng duyệt mức 0 hoặc một số tiền chỉ được đoán từ chứng từ.
4. So sánh các dòng cùng số tiền. Nhiều dòng giống nhau trên cùng bảng sao kê có thể là nhiều giao dịch thật; giữ và duyệt riêng khi chứng từ xác nhận. Nếu cùng một giao dịch xuất hiện trong các ảnh chồng lấn, kiểm tra mã tham chiếu và chỉ giữ một đề xuất. AI không đối chiếu với các giao dịch đã có trong sổ, nên kiểm tra sổ trước khi duyệt để tránh nhập lại.
5. Sửa từng bản nháp cần chỉnh. Chỉ nhấn **Lưu giao dịch** sau khi ngày, số tiền VND, ví, nhóm và việc có tính vào báo cáo đã đúng. **Bỏ qua** đề xuất không thể xác minh; bỏ qua không ghi giao dịch.

Nếu một trường vẫn không rõ, để đề xuất chờ xem xét hoặc từ chối. Chỉ một đề xuất đã được duyệt mới tạo giao dịch và thay đổi số dư/báo cáo theo các trường đã lưu.

**Trạng thái phát hành:** quy tắc giữ ngày trống, số tiền ngoại tệ/chưa rõ ở mức 0, giữ các dòng trùng hình thức và nhập XLSX deterministic đã được kiểm thử trong backend local ngày 27/10/2026; chưa triển khai lên production. Hãy kiểm tra bản ứng dụng/backend đang dùng trước khi dựa vào quy tắc này cho chứng từ thật.

## Dành cho AI agent

Tài liệu Markdown này là bề mặt đọc ổn định cho quy trình người dùng. Hợp đồng route, request/response, giới hạn tệp, khôi phục yêu cầu và điều kiện duyệt nằm tại [AI entry review trong API reference](../architecture/API.md). [Work note an toàn PDF](../work/bugs/AI-ENTRY-PDF-REVIEW-SAFETY-20261026.md) ghi phạm vi kiểm thử và trạng thái phát hành.

| Bước | Hợp đồng |
| --- | --- |
| Gửi chứng từ | `POST /api/v1/ai/entry/process`; một yêu cầu multipart có `request_id`, `text`, `timezone` và tối đa 20 `files` JPEG/PNG/PDF/XLSX, mỗi tệp tối đa 5 MiB. Không trộn XLSX với file OCR trong cùng yêu cầu. |
| Đọc lại kết quả | `GET /api/v1/ai/entry/requests/{request_id}`; sau kết quả gửi không chắc chắn, đọc lại cùng ID trước khi người dùng quyết định thử gửi lại. |
| Đề xuất | `proposals[]` là bản nháp. `draft.occurred_at=""` nghĩa là chưa xác minh ngày; `draft.amount=0` nghĩa là chưa có số tiền VND duyệt được. `questions[]` nêu trường cần kiểm tra. Trường nguồn tiền tệ của model chỉ dùng nội bộ, không có trong proposal JSON. |
| Dòng trùng | Giữ nhiều dòng nguồn riêng biệt dù cùng ngày/số tiền/mô tả; chỉ gộp bản sao nguồn rõ ràng từ ảnh chồng lấn. Không có đối chiếu tự động với sổ hiện có. |
| XLSX có cấu trúc | Mọi dòng phải xuất hiện đúng một lần theo `source_id` nội bộ. `text` chỉ hướng dẫn cách diễn giải/phân loại, không lọc bỏ dòng hay thay đổi ngày, số tiền, chiều giao dịch hoặc tiền tệ. |
| Sửa và duyệt | `PATCH /api/v1/ai/entry/proposals/{id}` lưu trường đã sửa; `POST /api/v1/ai/entry/proposals/{id}/approve` mới ghi một giao dịch thu/chi sau kiểm tra backend. `POST /api/v1/ai/entry/proposals/{id}/reject` không ghi sổ. Ngày rỗng và số tiền 0 không đủ điều kiện duyệt. |

Không diễn giải câu hỏi trong `questions[]` là bằng chứng trường đã được xác minh. Trạng thái production của thay đổi an toàn này vẫn cần triển khai và kiểm tra lại ở chế độ chỉ xem.
