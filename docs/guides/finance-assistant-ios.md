# Hỏi trợ lý tài chính trên iPhone

Hướng dẫn này mô tả bản native hiện tại. Async chat và dọn lịch sử sau thời gian
không hoạt động đã được ghi nhận triển khai production ngày 16/10/2026;
[bản ghi triển khai](../operations/DEPLOYMENT.md) phân biệt các bản local,
production và iPhone. Kiểm thử tự động không thay thế nghiệm thu trên máy thật.

Mở tab **Trợ lý** (tab chính của shell) để hỏi về chi tiêu hoặc tìm giao dịch. Trợ lý
không nằm trong **Tài khoản/More**. Chọn câu gợi ý để điền
vào ô chat, sửa lại nếu cần rồi nhấn mũi tên **Gửi**.

- Tổng quan thu nhập/chi tiêu nằm trên cùng. Chạm để xem số liệu và nguồn.
- Khi mở bàn phím, thanh tab tạm ẩn để dành chỗ cho hội thoại. Nút biểu đồ
  trên header mở cùng Tổng quan; nút bàn phím đóng bàn phím và khôi phục tabs.
- Ở cỡ chữ trợ năng, dùng menu chọn màn hình có nhãn đầy đủ thay thanh bốn tab.
  Đóng bàn phím trước khi đổi màn; câu hỏi chưa gửi vẫn giữ trong phiên hiện tại.
- Ô nhập mở rộng tối đa bốn dòng; nội dung dài hơn cuộn bên trong ô.
- App gửi chat qua API bất đồng bộ và nhận mã lượt chạy trước khi model hoàn
  tất. **Dừng xử lý** dùng mã đó để yêu cầu hủy; backend chặn ghi câu trả lời
  muộn của lượt đã hủy. Ngắt ngay model ở process khác vẫn cần cơ chế truyền
  tín hiệu hủy riêng, không được suy ra từ kiểm thử cùng process.
- **Chat mới** hủy best-effort các provider run đang chạy trong process khi
  reset được xác nhận; reset replay không hủy run mới. Provider work ở process
  khác vẫn dựa vào generation/lease guard để từ chối ghi muộn.
- Backend lưu lịch sử. Cuộn lên đầu hội thoại và chọn **Tải tin cũ** khi có.
  Tải tin cũ giữ vùng đang đọc, không tự đưa bạn về tin mới nhất. Mở lại app sẽ
  tải trang mới nhất; bạn vẫn có thể tải các trang cũ.
- Dịch vụ không khả dụng: vẫn soạn nháp được nhưng chưa gửi được.
  Lỗi không tự động gửi lại; nội dung chưa được tiếp nhận vẫn giữ trong ô.
- Chế độ che số dư áp dụng cả thẻ lạ/sai định dạng, thay dữ liệu thô bằng
  thông báo nội dung ẩn.
- Số tiền thiếu hoặc không hợp lệ hiện **Chưa có dữ liệu**, không tự đổi thành0
  hoặc cắt phần lẻ. Số0 hợp lệ và số dư âm vẫn được giữ; nhánh tiến độ chỉ vẽ
  khi có số tiền/mục tiêu hợp lệ.

Trợ lý hiện đọc/phân tích dữ liệu. Đề xuất tạo ví chỉ thực hiện sau khi bạn
xác nhận trên thẻ; với thẻ có `action_id` và context máy chủ, ứng dụng gọi
dispatcher `POST /api/v1/ai/advisor/actions/{action_id}/confirm` và chỉ gửi
conversation/generation binding. Backend lấy lại payload đã lưu theo owner,
không nhận số tiền/tên ví từ client, rồi dùng ledger idempotency. Thẻ lịch sử
cũ thiếu context vẫn dùng fallback wallet endpoint. Chat không tự ý sửa giao
dịch hay kết nối ngân hàng.
Đề xuất tạo **hũ chi tiêu** dùng cùng ranh giới: tool `propose_create_jar` chỉ
trả tên/tháng/quy tắc phân bổ để xem trước; nút xác nhận gọi cùng dispatcher
với `action_id`, và backend ghi qua ledger operation `jar.create`. Retry cùng
action không tạo hũ thứ hai. Web cũ không phải bề mặt được duy trì cho slice
native này.
Nếu bạn yêu cầu trợ lý tạo hoặc cập nhật hũ, trợ lý sẽ chuẩn bị thẻ đề xuất để
bạn kiểm tra; tool `propose_update_jar` yêu cầu `jar_id` và tháng đích, sau đó
thẻ gọi cùng dispatcher để chuyển tới PUT owner-scoped. Cập nhật là thay thế
cấu hình tháng hiện có, không tạo hũ mới. Trợ lý không được tự ghi hũ,
sửa/xóa hũ khác, hay thực hiện thanh toán khi chưa xác nhận.
Đề xuất tạo ngân sách cũng chỉ là bản nháp: `propose_create_budget` hiển thị
tên, hạn mức, khoảng ngày và phạm vi ví/danh mục để bạn duyệt. Nút xác nhận
dùng cùng dispatcher và `budget.create` idempotency key; gửi lại cùng action
không tạo ngân sách thứ hai. Các ID ví/danh mục vẫn được backend kiểm tra theo
owner, không tin dữ liệu do model tự gắn.
Đề xuất cập nhật ngân sách dùng `propose_update_budget`, yêu cầu `budget_id`
và thay thế cấu hình qua PATCH sau khi bạn xác nhận. Thẻ dùng cùng action
binding/generation; gửi lại action không tạo ngân sách mới.
Thẻ tạo ví thiếu/sai loại ví, số tiền, mục tiêu hoặc hạn mức sẽ hiện thông báo
chưa đủ dữ liệu thay vì tự đổi thành ví thường hay số dư0. Backend dùng
conversation/generation binding và mutation ledger để chống tạo lại cùng
đề xuất khi xác nhận lại. Thao tác thuộc chat đã hết hạn/reset bị từ chối.
Bong bóng phản hồi ở góc phải dưới dùng chung với các màn khác.

### Hướng dẫn trong ứng dụng

Khi câu hỏi liên quan tới cách dùng My Pocket, trợ lý có thể trả về thẻ
**Hướng dẫn trong ứng dụng** bằng tool đọc `get_product_help`. Thẻ chỉ hiển thị
những đoạn trong handbook nhúng, có phiên bản và mã nguồn ổn định để đối chiếu;
không chứa dữ liệu tài chính của chủ tài khoản, không truy cập URL bên ngoài và
không được dùng để tự suy đoán chính sách. Nếu không có đoạn phù hợp, thẻ trả
trạng thái “Chưa tìm thấy hướng dẫn phù hợp” thay vì đổ toàn bộ handbook vào
cuộc trò chuyện. Đây là bộ tra cứu xác định, chưa phải RAG/live-docs hay bằng
chứng provider production. Catalog local v1.2 bổ sung
hướng dẫn hũ và các đề xuất ví/hũ/ngân sách đã hỗ trợ; handbook hệ thống vẫn
có phiên bản độc lập v1.3.

App native khóa ở chế độ dọc để giữ layout tài chính ổn định trên iPhone. Bản
local đã vượt kiểm thử bàn phím, nháp và menu ở cỡ chữ lớn; yêu cầu xoay ngang
được giữ ở viewport dọc và có UI regression test riêng. Đã có bản cài thử
iPhone15/10, chưa phát hành/nghiệm thu hay chứng nhận trợ năng toàn app.
[Phạm vi/bằng chứng](../work/tickets/NATIVE-CHAT-ADAPTIVE-20261014.md).

## Cải tiến đang chờ phát hành

Bản local ngày 20/10 kiểm tra lại ví/danh mục và phiên bản giao dịch trước khi
hiện đề xuất. Khi bật chế độ trả lời có đối chiếu, số liệu và nhãn được lấy từ
kết quả hệ thống; hướng dẫn lấy từ handbook. Trợ lý chỉ giữ một phần hội thoại
gần nhất và các tham chiếu đã kiểm tra, luôn đọc lại số dư khi cần. Nếu kết quả
quá dài, app vẫn có thể hiện các thẻ dữ liệu đã lấy được. Những thay đổi này
chưa xác nhận triển khai production hoặc cài trên iPhone.

## Khi cần báo lỗi

Bản local mới chỉ ghi loại lỗi/mã HTTP khi tải cấu hình và tổng quan trợ lý,
không in nguyên nội dung lỗi vào console. Log HTTP giữ request ID để đối chiếu.
Không gửi API key, mật khẩu hoặc hóa đơn riêng tư vào log hỗ trợ. Thay đổi đã
test local, chưa có nghĩa bản trên iPhone/production đã được cập nhật.

## AI-agent reference

Native submit uses `POST /ai/advisor/messages/async`, which returns the run ID
before provider work completes; synchronous submit remains a compatibility route.
Persistence locks conversation then owner-scoped run and revalidates generation/
status before insertion. Async submit and the late-insert fix are recorded as
deployed; cross-process provider interruption is now covered by an owner-scoped
persisted run-status probe (250ms nominal polling), while local cancellation
remains immediate.
[Cancellation work](../work/tickets/NATIVE-ADVISOR-CANCEL-20261015.md).

History wire contract: current page50, `before_seq` is exclusive, native displays
ascending sequence. The response keeps the message array in `data` and may add
`meta` with `conversation_id`, `generation`, `last_user_message_at`,
`expires_at` and server `server_time`; old clients may ignore `meta`. Native
stores this metadata only as stale-response context and never extends expiry
from a read. Local real-DB/API/native UI proof uses51synthetic messages:
50latest→one older→no latest jump→restart, persisted messages unchanged. This is
not live model/cancellation/race or full accessibility acceptance.
[History verification](../work/tickets/NATIVE-ADVISOR-HISTORY-E2E-20261015.md).

Local correction2026-10-15: structured-card money uses exact signed finance-safe
integers (absolute value<=9007199254740991). Missing/null/wrong-type/fractional/
nonfinite/overflow renders unavailable instead of fabricated zero/truncation.
Count and JSON fallback use exact integer conversion; malformed/unknown fallback
consults existing masking before formatting. Native range, unsafe goal percentage
and masked-fallback sensitivity have RED→GREEN proof. Final287/287PASS includes
279unit tests and8native chatUI regressions, money/count/large JSON and hosted malformed-card layout. This is
not full privacy/device acceptance; updated candidate not installed yet.
[Numeric/privacy work](../work/tickets/NATIVE-AI-CARD-NUMERIC-SAFETY-20261015.md).

Idle expiry: the 30-minute inactivity worker is recorded as deployed on
2026-10-16. The backend records the last accepted user question and applies an inclusive
30-minute server-authoritative idle boundary. At expiry, history/run access and
the next submit rotate the generation and clear advisor messages/events/facts;
financial rows and confirmed writes remain untouched. The worker repeats this
cleanup in bounded batches for owners who do not return. Reads, polling and
assistant output do not extend retention. A purged request ID returns `410
advisor_history_expired` instead of calling the provider. Native old history
decoders remain compatible with an empty page. When iOS receives the 410 (or a
terminal `purged` run), it clears the local conversation pointer and visible
messages, preserves the unsent composer text, and waits for an explicit resend;
it never retries automatically. History reads may include additive server expiry
metadata; the native “Chat mới” reset is now implemented through the owner-scoped
idempotent reset route. See the [lifecycle design](../superpowers/specs/2026-10-15-advisor-idle-reset-design.md)
and [implementation ticket](../work/tickets/NATIVE-ADVISOR-IDLE-RESET-20261015.md).

The native header exposes a destructive “Chat mới” action. After confirmation it
calls the reset route with a persisted idempotency key, clears only assistant
state on success, and leaves financial data untouched. Repeating the same key
replays the original reset result instead of deleting newer questions.

Platform: native iOS. API base: `/api/v1`. Auth: current owner JWT/session or
active full-owner API key, without per-key scopes; owner isolation still applies.
Native submit: `POST /ai/advisor/messages/async`; synchronous compatibility:
`POST /ai/advisor/messages`; history: `GET /ai/advisor/conversation/messages`;
reset: `POST /ai/advisor/conversation/reset` with a persisted `Idempotency-Key`;
run state: `GET /ai/advisor/runs/:id`; replay persisted progress events with
authenticated `GET /ai/advisor/runs/:id/events?after_seq=N` (or the
`Last-Event-ID` header); cancel: `POST /ai/advisor/runs/:id/cancel`. The event
endpoint is bounded replay, not token streaming, and does not re-run the model.
Keep client request IDs, owner isolation and explicit confirmation boundaries.
Local correction 2026-10-19: confirmation searches persisted history pages for
older proposal cards, not only the latest 50 messages. Owner/generation checks
and replay protection remain mandatory; model context remains bounded separately.
Do not automatically retry a possibly accepted request. No schema/endpoint change
accompanies this UI update. [Tool catalog](../architecture/AI_ADVISOR_TOOLS.md).

Chat layout: keyboard visible => shell tabs absent, no reserved tab height,
overview header action opens the existing source-backed sheet, explicit dismiss
restores navigation. Accessibility text sizes => shared full-label navigation
menu. Draft remains owner-session memory only; never auto-submit on navigation.

Wallet proposal validation: require supported type/name; optional absent balance
defaults0 and total flagtrue, but malformed present values fail closed. Money is
nonnegative exact integer <=9007199254740991; goal target/credit limit positive;
goal date valid YYYY-MM-DD; credit days1–31. Invalid card cannot confirm. A
confirmed wallet create reuses one `Idempotency-Key` for retries; the backend
mutation ledger replays the same owner/payload without creating a duplicate.
New proposal cards carry a server-generated `action_id` in the persisted
assistant message, so reopening the card keeps the same replay boundary. Older
cards without that field use a mounted-card fallback key.

Diagnostics: HTTP metadata = method/status/generated requestID; feature errors
use APIError.diagnosticDescription, class + optional numeric status only, unknown
errors = unexpected. Never log ProblemDetails fields, prompt/receipt/finance
payloads, credential headers or arbitrary Error reflection.
[Local proof and limitations](../work/tickets/NATIVE-FEATURE-LOG-PRIVACY-20261014.md).

Adaptive UI has local keyboard/draft/large-text regression evidence. Physical
landscape and accessibility acceptance remain distinct gates; do not infer
all-screen acceptance from a normal chat test. Refer to the dated validation
matrix rather than historical failed prototypes for current evidence.
