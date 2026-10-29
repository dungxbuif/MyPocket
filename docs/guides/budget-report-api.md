# Xem chi tiêu và dự báo ngân sách

Các API này dùng `Authorization: Bearer <JWT hoặc API key của bạn>` như các
API tài chính khác. Dữ liệu luôn thuộc tài khoản của credential; không truyền
`owner_id` để đổi tài khoản. Bản triển khai hiện tại là local candidate, chưa
được rollout lên production.

## Báo cáo chi tiết

`GET /api/v1/budgets/{id}/report`

```json
{
  "data": {
    "budget_id": "owned-budget-uuid",
    "timezone": "Asia/Ho_Chi_Minh",
    "as_of": "2026-09-03T05:00:00Z",
    "start_date": "2026-09-01",
    "end_date": "2026-09-30",
    "base_limit_amount": 4000000,
    "carryover_amount": 1000000,
    "effective_limit_amount": 5000000,
    "spent": 300000,
    "remaining": 4700000,
    "forecast": {
      "status": "estimated",
      "method": "linear_daily_v1",
      "elapsed_days": 3,
      "total_days": 30,
      "projected_spent": 3000000,
      "projected_remaining": 2000000,
      "over_budget": false
    },
    "daily": [
      {"date": "2026-09-01", "spent": 0},
      {"date": "2026-09-02", "spent": 300000},
      {"date": "2026-09-03", "spent": 0}
    ]
  },
  "meta": null
}
```

Số tiền là số nguyên VND. Dự báo tính bằng code:
`ceil(chi thực tế đến hiện tại × tổng ngày kỳ / ngày đã qua)`, tính cả ngày
không chi theo múi giờ tài khoản. Đây là ước tính theo tốc độ hiện tại, không
phải giao dịch đã lên lịch. Các giao dịch tương lai chưa vào dự báo hiện tại.

- `estimated`: kỳ đang diễn ra, projection là ước tính.
- `actual`: kỳ đã kết thúc, projection bằng số chi thực tế.
- `not_started`: kỳ chưa bắt đầu, các projection và `over_budget` là null.

`remaining` có thể âm khi vượt ngân sách. Hạn mức hiệu lực là hạn mức gốc cộng
phần dư dương từ kỳ trước nếu chuỗi đã bật chuyển dư. Sửa giao dịch lịch sử sẽ
tính lại số chi, phần dư và dự báo.

## Mở giao dịch nguồn

`GET /api/v1/budgets/{id}/transactions?limit=50&offset=0`

Trả `data: {budget_id, timezone, as_of, items, total, limit, offset}`; `items`
chứa giao dịch theo thứ tự thời gian giảm dần, ID giảm dần khi cùng thời điểm.
`limit` từ 1 đến 100; `offset` không âm. `total` là tổng toàn phạm vi trước phân
trang. Các trang dùng cùng bộ lọc chi, ví, danh mục/con, ngày, report inclusion
và cutoff thời gian như báo cáo. Tổng từ mọi trang đối soát được `spent` nếu
ledger không thay đổi giữa các lần đọc.

401 khi thiếu/sai credential; 404 khi ID không thuộc tài khoản; 400 khi ID hay
phân trang sai. Không coi API lỗi hoặc projection null là số tiền bằng 0.

## Cho AI agents và tích hợp

Tài liệu Markdown này là contract dễ fetch cho agent. Dùng kết quả server làm
nguồn số tiền; giữ `timezone`, `as_of`, `forecast.status` và phạm vi khi diễn
giải. Không tự cộng `base_limit_amount` với `effective_limit_amount`. Không dùng
LLM để tính lại phép cộng/trừ hoặc tạo số forecast khác với API.

Xem thêm [ngân sách định kỳ](budget-recurrence-api.md) và
[báo cáo tiết kiệm](savings-report-api.md).
