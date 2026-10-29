---
artifact_type: user_stories
id: USER_STORIES
status: in_review
owner: shared
human_fields: [role, need, benefit, acceptance_criteria]
ai_fields: [story_rows, trace_links, status_updates]
shared_fields: [stories]
updated: 2026-09-13
trace:
  requirements: REQUIREMENTS.md
  business_rules: BUSINESS_RULES.md
  reports: REPORTS.md
  validation: ../work/VALIDATION_MATRIX.md
---

# Hành trình review MyPocket

## Field Ownership

- Human owns user intent and acceptance criteria.
- AI maintains story rows, trace links, and status updates.

Các tiêu chí dưới đây phục vụ review và kiểm thử khi triển khai, chưa phải UAT đã chạy.

| ID | User muốn | Kết quả cần quan sát | Yêu cầu |
| --- | --- | --- | --- |
| US-01 | Chọn timezone account | Giao dịch gần nửa đêm vào đúng ngày/tháng; mọi số tiền VND | REQ-01 |
| US-02 | Quản lý nhiều ví | Tạo/sửa/xóa; tổng đúng lựa chọn; xóa nêu rõ ảnh hưởng | REQ-02 |
| US-03 | Ghi nhanh và sửa sai | Sửa/xóa cập nhật số dư/report một lần; chuyển ví không thành thu/chi kép | REQ-03, REQ-04 |
| US-04 | Theo dõi budget ăn uống | Đúng ví/danh mục/kỳ tự vào tiến độ; cha/con không đếm trùng | REQ-05 |
| US-05 | Theo dõi chi bằng hũ | Tối đa một hũ/chi thường; chưa gắn hũ vẫn lưu; vượt chỉ cảnh báo | REQ-06 |
| US-06 | Xem hũ tháng mới và cộng dồn | Copy cấu hình gần nhất, không chuyển dư; tổng chi đúng; thiếu phân bổ không coi là 0 | REQ-06 |
| US-07 | Ghi tiền thuê định kỳ | Tự tạo giao dịch/note, retry không trùng, sửa giao dịch không sửa lịch | REQ-07 |
| US-08 | Bật du lịch | Giao dịch đủ điều kiện tự gắn chuyến; recurring không tự gắn; tắt không đổi lịch sử | REQ-08 |
| US-09 | Tiết kiệm cho mục tiêu | Nạp/rút cập nhật tiến độ; chuyển giữa ví không nhân đôi thu/chi | REQ-09 |
| US-10 | Theo dõi thẻ và nợ | Thấy dư nợ/sao kê/hạn/từng lần trả; trả thẻ không tính mua hàng lần hai | REQ-10, REQ-11 |
| US-11 | Hiểu chi tiêu | Reports/Insider cùng phạm vi cùng tổng, mở được giao dịch nguồn | REQ-12 |
| US-12 | Nhớ lại một tháng | Note, chuyến đi, AI; sửa giao dịch tháng cũ đổi số nhưng giữ note user | REQ-13 |
| US-13 | Theo dõi lãi/lỗ | Mua đổi giá vốn bình quân, bán dùng giá hiện hành; đổi giá không đổi ví | REQ-14 |
| US-14 | Kết nối công cụ riêng | Key thao tác như owner, không quản lý key khác, thu hồi có hiệu lực | REQ-15, REQ-16 |
| US-15 | Nhờ AI đọc và tư vấn | Nhập liệu xác nhận; tư vấn chỉ đọc; số có nguồn; OCR lỗi không tạo giao dịch | REQ-17 |
| US-16 | Ghi offline | Thấy chờ gửi/xung đột, retry không trùng, logout dọn cache riêng | REQ-18 |

## Ca kiểm tra cho quyết định mới

1. Sửa giao dịch tháng đã tổng kết: số cập nhật, note giữ nguyên, AI cũ không được trình bày như kết luận mới.
2. Hũ có chi nhưng không đặt mức: thấy chi thực tế, không cảnh báo vượt mức 0 giả.
3. Hũ phân bổ 1.000.000, chi 1.200.000: lưu được, cảnh báo vượt 200.000, tháng sau không tự mang phần âm sang.
4. Cộng dồn qua tháng có/không có phân bổ: tổng chi toàn kỳ tách khỏi chênh lệch chỉ trên tháng có cấu hình.
5. UTC query theo timezone account đúng phân kỳ; date thuần không bị dịch ngày.
