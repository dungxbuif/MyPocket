# Hỏi trợ lý tài chính trên iPhone

Bản cài thử ngày15/10/2026 đã được cài và mở trên iPhone của chủ tài khoản,
kết nối backend production, không đặt lại dữ liệu. Chủ tài khoản đang kiểm thử;
đây chưa phải xác nhận mọi lỗi trong tài liệu đã được xử lý. Các cảnh báo local/
production bên dưới vẫn áp dụng, đặc biệt bản sửa backend sau hủy chưa deploy.

Mở tab **Trợ lý** (tab chính của shell) để hỏi về chi tiêu hoặc tìm giao dịch. Trợ lý
không nằm trong **Tài khoản/More**. Chọn câu gợi ý để điền
vào ô chat, sửa lại nếu cần rồi nhấn mũi tên **Gửi**.

- Tổng quan thu nhập/chi tiêu nằm trên cùng. Chạm để xem số liệu và nguồn.
- Khi mở bàn phím, thanh tab tạm ẩn để dành chỗ cho hội thoại. Nút biểu đồ
  trên header mở cùng Tổng quan; nút bàn phím đóng bàn phím và khôi phục tabs.
- Ở cỡ chữ trợ năng, dùng menu chọn màn hình có nhãn đầy đủ thay thanh bốn tab.
  Đóng bàn phím trước khi đổi màn; câu hỏi chưa gửi vẫn giữ trong phiên hiện tại.
- Ô nhập mở rộng tối đa bốn dòng; nội dung dài hơn cuộn bên trong ô.
- Khi có mã lượt chạy, **Dừng xử lý** gửi yêu cầu hủy. Hiện POST chat đồng bộ
  có thể chưa trả mã trong lúc đang chờ model, nên nút Dừng chưa sẵn sàng ở
  giai đoạn đó. Bản local đã sửa và kiểm thử lỗi lưu câu trả lời sau khi hủy
  trong lúc chờ khóa hội thoại; backend production chưa nhận bản sửa này.
  Chưa nên coi thao tác hủy là bảo đảm cho mọi tình huống kết quả đến muộn.
- **Chat mới** hủy best-effort các provider run đang chạy trong process khi
  reset được xác nhận; reset replay không hủy run mới. Provider work ở process
  khác vẫn dựa vào generation/lease guard để từ chối ghi muộn.
- Backend lưu lịch sử. Cuộn lên đầu hội thoại và chọn **Tải tin cũ** khi có.
  Tải tin cũ giữ vùng đang đọc, không tự đưa bạn về tin mới nhất. Mở lại app sẽ
  tải trang mới nhất; bạn vẫn có thể tải các trang cũ.
- Dịch vụ không khả dụng: vẫn soạn nháp được nhưng chưa gửi được.
  Lỗi không tự động gửi lại; nội dung chưa được tiếp nhận vẫn giữ trong ô.
- Bản sửa local mới áp dụng chế độ che số dư cho cả nhánh thẻ lạ/sai định dạng,
  thay dữ liệu thô bằng thông báo nội dung ẩn. Bản này chưa cài lại lên iPhone;
  không coi bản cài trước đó đã có sửa lỗi này hay đã nghiệm thu riêng tư toàn app.
- Số tiền thiếu hoặc không hợp lệ hiện **Chưa có dữ liệu**, không tự đổi thành0
  hoặc cắt phần lẻ. Số0 hợp lệ và số dư âm vẫn được giữ; nhánh tiến độ chỉ vẽ
  khi có số tiền/mục tiêu hợp lệ. Đây cũng là bản sửa local đang kiểm chứng.

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
chưa đủ dữ liệu thay vì tự đổi thành ví thường hay số dư0. Đây là thay đổi local;
đã nằm trong bản cài thử iPhone15/10, chưa được chủ tài khoản nghiệm thu. Xác nhận UI không phải bảo đảm chống
tạo lại cùng đề xuất sau khi tải lại lịch sử; action bền vững hiện dùng
conversation/generation binding và mutation ledger, đã kiểm thử ở backend.
Feedback tự né ô nhập. Thêm giao dịch qua tab Giao dịch hoặc nút cộng ở Tổng quan.

### Hướng dẫn trong ứng dụng

Khi câu hỏi liên quan tới cách dùng My Pocket, trợ lý có thể trả về thẻ
**Hướng dẫn trong ứng dụng** bằng tool đọc `get_product_help`. Thẻ chỉ hiển thị
những đoạn trong handbook nhúng, có phiên bản và mã nguồn ổn định để đối chiếu;
không chứa dữ liệu tài chính của chủ tài khoản, không truy cập URL bên ngoài và
không được dùng để tự suy đoán chính sách. Nếu không có đoạn phù hợp, thẻ trả
trạng thái “Chưa tìm thấy hướng dẫn phù hợp” thay vì đổ toàn bộ handbook vào
cuộc trò chuyện. Đây là bộ tra cứu xác định, phiên bản local v1.1, chưa phải
RAG/live-docs hay bằng chứng provider production.

Bản local đã vượt kiểm thử màn ngang/bàn phím, nháp và menu ở cỡ chữ lớn.
Đã có bản cài thử iPhone15/10, chưa phát hành/nghiệm thu hay chứng nhận trợ năng toàn app.
[Phạm vi/bằng chứng](../work/tickets/NATIVE-CHAT-ADAPTIVE-20261014.md).

## Khi cần báo lỗi

Bản local mới chỉ ghi loại lỗi/mã HTTP khi tải cấu hình và tổng quan trợ lý,
không in nguyên nội dung lỗi vào console. Log HTTP giữ request ID để đối chiếu.
Không gửi API key, mật khẩu hoặc hóa đơn riêng tư vào log hỗ trợ. Thay đổi đã
test local, chưa có nghĩa bản trên iPhone/production đã được cập nhật.

## AI-agent reference

Cancellation limitation2026-10-15: synchronous submit returns run_id after provider
work; native onCancel is unavailable before that ID. Approved local PostgreSQL
fix locks conversation then owner-scoped run and revalidates generation/status
before insertion. Controlled-lock regression now passes, alongside owner/new-run
and provider-context gates4/4. Backend production is not updated; broader
completion/cancel atomicity and multi-process interruption remain separate.
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

Idle expiry:30-minute inactivity behavior is implemented locally but not enabled
in production. The backend records the last accepted user question and applies an inclusive
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
Submit: `POST /ai/advisor/messages`; history: `GET /ai/advisor/conversation/messages`;
reset: `POST /ai/advisor/conversation/reset` with a persisted `Idempotency-Key`;
run state: `GET /ai/advisor/runs/:id`; cancel: `POST /ai/advisor/runs/:id/cancel`.
Keep client request IDs, owner isolation and explicit confirmation boundaries.
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

Adaptive UI status: not accepted. Actual local API landscape+keyboard test has
zero-height conversation; fake-session case7pt. Largest Dynamic Type input
control checks pass, navigation visual overflow remains. Do not infer all-screen
or accessibility acceptance from normal chat regression. No adaptive chrome
change has shipped; design review pending.
