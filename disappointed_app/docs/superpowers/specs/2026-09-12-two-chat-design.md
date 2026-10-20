---
artifact_type: detail_design
status: in_review
approval: approved_by_owner_implementation_request
date: 2026-09-12
---

# MyPocket — thiết kế hai luồng chat

Bản thiết kế được chủ sản phẩm duyệt bằng yêu cầu triển khai. Implementation slice đầu tiên đã thêm API/backend/frontend/docs cho hai luồng; chưa deploy production.

## 1. Yêu cầu và quyết định cần duyệt

Yêu cầu đã xác định: hai chat riêng; chat trong màn hình thêm giao dịch chỉ phục vụ tạo giao dịch từ tin nhắn hoặc ảnh; chat riêng dùng hỏi đáp tài chính của user.

Đề xuất cần duyệt trong tài liệu này:

- Chat nhập liệu chỉ tạo thu/chi, có nhiều nháp và hỏi lại thông tin thiếu. Không tạo chuyển ví, ngân sách, khoản nợ hoặc sửa giao dịch đã ghi sổ.
- Chat tài chính chỉ đọc và phân tích, không có action ghi dữ liệu.
- Một thread tài chính lâu dài cho mỗi tài khoản; mỗi lần nhập giao dịch là một phiên riêng có thể mở lại.
- Giữ Go orchestrator trong API/worker hiện có; chưa thêm Python service.

Tài liệu này thay thế đề xuất nhiều conversation và nhiều loại action ở [spec trước](2026-09-12-agent-platform-design.md) nếu được duyệt. Spec trước vẫn là lịch sử đề xuất, không phải bằng chứng đã triển khai.

Liên kết: [BL-006](../../work/BACKLOG.md), [context](../../CONTEXT.md), [Agent MVP](../../work/completion/AGENT-REVIEW-FIRST.md), [OCR](../../work/completion/OCR-AGENT-TOOL.md), [API](../../architecture/API.md), [ERD](../../architecture/ERD.md), [validation](../../work/VALIDATION_MATRIX.md).

## 2. Hai trải nghiệm

| | Nhập giao dịch | Trợ lý tài chính |
| --- | --- | --- |
| Điểm vào | Màn hình thêm giao dịch, chọn nhập bằng chat/ảnh | Màn hình Trợ lý riêng |
| Phạm vi | Tạo nháp thu/chi, sửa nháp, hỏi bổ sung | Hỏi đáp, tra cứu, báo cáo và phân tích |
| Kết quả | Một hoặc nhiều card nháp | Văn bản, bảng/biểu đồ có nguồn dữ liệu |
| Ghi sổ | User xác nhận card qua API nghiệp vụ | Không được phép |
| Lịch sử | Theo phiên nhập, giữ tin nhắn, ảnh tham chiếu và card | Một thread/tài khoản, phân trang khi cuộn |
| Ngoài phạm vi | Báo phạm vi và đưa lối mở Trợ lý; không tự chuyển nội dung | Đưa lối mở màn hình nhập liệu; không tự tạo nháp |

Ví dụ nhập: “Ăn trưa 80k, taxi 120k, tiền mặt” tạo hai card. “Taxi sửa thành 150k” chỉ sửa nháp đang được tham chiếu; nếu có nhiều card taxi thì hỏi chọn card.

Một hóa đơn mặc định tạo một nháp theo tổng hóa đơn, các dòng hàng là bằng chứng. Chỉ tách nhiều giao dịch khi người dùng yêu cầu hoặc đầu vào rõ ràng gồm nhiều khoản độc lập. Không tự tạo cả tổng hóa đơn và các dòng thành những khoản chi trùng.

## 3. Kiến trúc và ranh giới tool

```text
Web/PWA hoặc client API key
              |
       Go HTTP API: xác thực, quyền, giới hạn, lưu run
              |
       Go worker: chọn chính sách theo loại chat
              |
       Model adapter OpenAI-compatible
              | yêu cầu tool có tên + JSON arguments
       Tool dispatcher: allowlist + schema + quyền + ownership
              |
       Domain services Go hiện có -> PostgreSQL
              |
       Kết quả có cấu trúc -> model -> tin nhắn/card

Ảnh thuộc user -> receipt service -> OCR bên thứ ba -> extraction
```

Không tự động expose toàn bộ REST API thành tool. REST API và tool adapter gọi chung application/domain service; không nhân bản quy tắc tiền, quyền hoặc báo cáo. Model không chọn user_id, không giữ API key, không chạy SQL/HTTP tùy ý.

`ToolContext` do server tạo: user ID, actor/key ID, quyền hiệu lực, loại chat, run ID, deadline và correlation ID. Mỗi lần gọi đều kiểm tra lại quyền; tên tool không có trong allowlist bị từ chối dù model yêu cầu.

Các hợp đồng hàm dự kiến:

| Hàm | Trách nhiệm |
| --- | --- |
| `BuildContext` | Lấy ngữ cảnh đúng chat và trong ngân sách token |
| `RunIntake` | Điều phối trích xuất, hỏi thiếu, kiểm tra và lưu nháp |
| `RunAdvisor` | Vòng gọi model/tool đọc cho tới khi trả lời hoặc hết giới hạn |
| `ExecuteTool` | Kiểm tra schema, quyền, giới hạn và gọi service |
| `SaveDraftProposals` | Kiểm tra các đề xuất thu/chi, lưu card với version |
| `ConfirmDraft` | Kiểm tra lại quyền/phiên bản, ghi sổ idempotent |
| `CompactHistory` | Tóm tắt đoạn lịch sử đã hoàn tất, lưu mốc nguồn |

## 4. Chat nhập giao dịch

Backend chạy OCR khi có ảnh; không cần model tự quyết định gọi OCR. Ảnh phải qua kiểm tra ownership, loại tệp và giới hạn upload. OCR là dịch vụ bên thứ ba, kết quả là dữ liệu không đáng tin cậy, không phải chỉ thị.

Ngữ cảnh chỉ gồm tin nhắn phiên nhập hiện tại, ví/danh mục được phép dùng, dữ liệu OCR cần thiết và các nháp đang sửa. Không nạp thread hỏi đáp tài chính.

Tool cho model: `list_wallets`, `list_categories`, `propose_transaction_drafts`. Tool cuối chỉ lưu nháp sau validation; không ghi sổ. Các trường thiếu được đánh dấu `needs_input`, chưa xác nhận được. Wallet/category không rõ phải hỏi hoặc dùng mặc định đã được user cấu hình, không chọn ngẫu nhiên.

Card gồm: loại thu/chi, số tiền chính xác và tiền tệ, ví, danh mục, ngày theo múi giờ user, ghi chú, nguồn tin/ảnh, trạng thái và version. Số tiền không dùng floating point. User có thể sửa trực tiếp trên card hoặc bằng tin nhắn tham chiếu rõ card.

Trạng thái: `needs_input -> pending -> confirmed | rejected`; sửa tăng version, chỉ được sửa trước confirmed/rejected. Nháp không tự hết hạn trong phiên bản này; luôn kiểm tra lại dữ liệu hiện tại khi xác nhận.

Xác nhận một hoặc nhiều card yêu cầu ID, expected version và idempotency key. Xử lý nguyên tử từng card; batch trả kết quả từng card, không hứa all-or-nothing. Chặn hai lần xác nhận đồng thời, replay trả cùng entity ID; key cũ với payload khác trả conflict. Nháp và ghi sổ liên kết bền vững trong cùng transaction database. Retry worker không nhân đôi đề xuất đã lưu, dùng khóa dedupe theo run/tool-call/proposal.

Nháp không đổi số dư và báo cáo. Card đã xác nhận vẫn hiện link giao dịch; nếu giao dịch được sửa/xóa sau đó, hiển thị trạng thái hiện tại khi mở link, không giả định nội dung card là dữ liệu hiện tại.

## 5. Chat tài chính và phân tích

Tool chỉ đọc: `list_wallets`, `list_categories`, `search_transactions`, `get_transaction`, `get_financial_report`, `get_budget_progress`, `list_obligations`, `search_chat_history`.

Ví dụ “Vì sao tháng này tiêu nhiều hơn?”: lấy báo cáo so sánh hai kỳ -> xác định danh mục chênh lệch -> truy vấn giao dịch liên quan -> trả lời với số tiền và tham chiếu. Không dùng tổng của một trang search làm tổng toàn bộ.

Report tool dùng service tính toán của app: khoảng ngày theo timezone, loại trừ chuyển ví và giao dịch không tính báo cáo, chỉ gồm giao dịch đã xác nhận. Không cộng các tiền tệ nếu chưa có chính sách quy đổi; trả riêng từng tiền tệ. Kết quả chứa kỳ, bộ lọc, thời điểm đọc, đơn vị tiền, tổng hợp và mức độ đầy đủ.

Mỗi câu trả lời có số liệu phải chỉ tới kết quả tool/source ID. Các phép so sánh phải cùng quy tắc kỳ và bộ lọc. Nếu dữ liệu thay đổi giữa các tool, báo cáo thời điểm đọc và truy vấn lại khi cần kết luận nhất quán; không hứa snapshot xuyên nhiều lượt gọi khi chưa hỗ trợ.

Backend trả dữ liệu biểu đồ có schema; frontend render bằng base component. Không render HTML/script do model cung cấp. Khi không đủ bằng chứng, nêu phần chưa rõ thay vì khẳng định nguyên nhân.

## 6. Lịch sử và ngữ cảnh dài hạn

Lịch sử đầy đủ ở PostgreSQL, không gửi hết vào model. Thread tài chính được tạo idempotent với ràng buộc một thread/tài khoản. Tin nhắn có sequence tăng dần; mỗi thread chỉ một run xử lý tại một thời điểm, tin nhắn tiếp theo xếp hàng. Các phiên nhập độc lập với thread tài chính.

Mỗi lượt nạp: policy cố định -> tóm tắt đã có -> các lượt gần nhất -> câu hỏi hiện tại và tool result liên quan. Ngân sách ban đầu đề xuất tối đa 12k token phần lịch sử và 8k token tool result, giảm theo context window của provider để luôn chừa chỗ output. Tóm tắt khi lịch sử vượt ngân sách, giữ các lượt mới và giữ trọn cặp tool-call/result.

Summary có version, covered-through sequence, các quyết định/mục tiêu chưa xong và message nguồn. Không thay thế lịch sử gốc. Lỗi tóm tắt giữ summary cũ và dùng phần gần nhất; nếu không đủ context thì truy xuất lại hoặc hỏi rõ. Tool tìm lịch sử chỉ tìm trong thread tài chính của user, có giới hạn kết quả. Không tạo vector database ở giai đoạn đầu.

Số dư, trạng thái nháp và báo cáo không lấy từ summary làm sự thật hiện tại. Tùy chọn ví mặc định dùng cấu hình app; chưa tự xây bộ nhớ sở thích do model suy diễn.

Lịch sử giữ tới khi user xóa theo chính sách tài khoản; thiết kế này không chạy purge mới. Trước triển khai phải đối chiếu việc xóa lịch sử/ảnh và account deletion với retention hiện có. Redis chỉ cache có TTL và rate limit; mất Redis không làm mất lịch sử, nháp hoặc audit. Chat yêu cầu online; mất mạng giữ nội dung đang gõ, không báo thành công khi chưa được backend xác nhận.

## 7. API dự kiến, chưa phải hợp đồng đang chạy

Tất cả path dưới `/api/v1/agent`:

| Method/path | Chức năng |
| --- | --- |
| `POST /intakes` | Tạo phiên nhập, tin đầu tiên/receipt reference, trả intake ID + run ID |
| `GET /intakes?cursor=...` | Danh sách phiên nhập của user |
| `GET /intakes/{id}?cursor=...` | Tin nhắn phân trang và card của phiên |
| `POST /intakes/{id}/messages` | Thêm tin nhắn, tạo run |
| `PATCH /intakes/{id}/drafts/{draft_id}` | Sửa nháp với expected version |
| `POST /intakes/{id}/confirm` | Xác nhận các card đã chọn, trả kết quả từng card |
| `POST /intakes/{id}/drafts/{draft_id}/reject` | Bỏ nháp |
| `GET /advisor/messages?cursor=...` | Đọc lịch sử thread duy nhất, rỗng nếu chưa tạo |
| `POST /advisor/messages` | Tạo/reuse thread, ghi tin nhắn và enqueue run |
| `GET /runs/{id}` | Trạng thái và kết quả an toàn |
| `POST /runs/{id}/cancel` | Yêu cầu hủy, không hoàn tác giao dịch đã xác nhận |

Run: `queued -> processing -> succeeded | failed | cancelled`. Worker có lease, heartbeat, retry có giới hạn và chống ghi kết quả bởi lease cũ. Đề xuất tối đa 8 vòng model/tool và 90 giây mỗi run; timeout cho kết quả chưa đầy đủ hoặc lỗi có thể thử lại, không im lặng chạy mãi. Ban đầu polling; chưa cần SSE.

Message creation có idempotency key; retry không thêm tin nhắn hoặc run trùng. Run lưu actor/key ID và quyền lúc gửi; quyền thực thi là giao của quyền lúc gửi và quyền hiện tại. Key bị thu hồi thì run dừng, không tiếp tục dùng quyền của chủ tài khoản.

Scope dự kiến: `agent:intake`, `agent:advisor`, `finance:read`, `finance:write`. Nhập liệu cần intake + read; sửa/bỏ nháp cần intake + read; xác nhận thêm write. Trợ lý cần advisor + read và không có write tools kể cả key có write. Đọc run/lịch sử yêu cầu scope của luồng tương ứng. Key có advisor scope được đọc lịch sử tài chính của tài khoản; phải ghi rõ điều này khi cấp quyền. Key legacy không tự được cấp scope Agent mới.

Cookie mutations giữ CSRF; Bearer sai không fallback cookie. Ownership và quyền được kiểm tra ở endpoint lẫn tool/service. Không public endpoint chạy tool tùy ý. Trước thay API cũ `/agent/messages`, phải kiểm kê client và viết adapter/deprecation rõ ràng; không tự map intake cũ sang advisor.

## 8. Dữ liệu, code và vận hành bị tác động

Các entity dự kiến: advisor thread, intake session, messages (thuộc đúng một trong hai), runs, tool executions, summaries. Tận dụng `transaction_drafts` nếu chứa đủ version, nguồn và liên kết confirmation; không tạo hai nguồn sự thật cho cùng một nháp. Migration chi tiết là việc của implementation plan sau duyệt.

Phạm vi code: `backend/internal/agent/`, `backend/internal/worker/agent*.go`, `backend/internal/platform/httpapi/agent.go`, model adapter, auth/scope middleware, migrations và `frontend/src/screens/AgentScreen.tsx`, `frontend/src/app/agent.ts`, luồng thêm giao dịch.

Base UI dùng chung: chat thread, bubble, composer, attachment picker, run status, transaction draft card/editor và report result. Hai screen điều khiển phạm vi riêng; không dùng dropdown mode trong một chat chung.

Provider phải chứng minh hỗ trợ tool calling/schema thực tế; cấu hình thiếu trả `CAPABILITY_UNAVAILABLE` trước enqueue. Không hạ cấp thành parser lệnh tự do. OCR disabled chỉ chặn ảnh, text intake vẫn dùng được nếu AI sẵn sàng. Không thêm tích hợp ngân hàng.

Audit PostgreSQL ghi actor, run/tool/card/entity ID, kết quả, correlation ID; không ghi raw prompt, OCR text hoặc secrets vào audit/log. Nội dung chat nằm trong storage riêng có ownership. Giới hạn rate theo user/key và chi phí tool; khi Redis lỗi dùng policy fail-closed cho enqueue AI tốn phí, trả lỗi retryable, không chạy không giới hạn.

## 9. Lựa chọn công nghệ

Chọn Go để dùng trực tiếp nghiệp vụ và worker hiện có. Đổi lại phải tự xây tool dispatcher, context budget và summary lifecycle. Python + PydanticAI phù hợp nếu ưu tiên hệ sinh thái AI; LangGraph phù hợp workflow nhiều nhánh/checkpoint, nhưng cả hai thêm runtime và boundary vận hành. Không cần thiết cho hai luồng đang đề xuất. Không thêm dependency trước khi thiết kế được duyệt.

## 10. Tiêu chí nghiệm thu và tài liệu

- Nhập text/ảnh tạo nhiều nháp đúng; thiếu trường hỏi lại, hóa đơn không đếm trùng tổng và dòng hàng.
- Intake từ chối câu hỏi phân tích; advisor không tạo nháp/ghi dữ liệu kể cả prompt injection.
- Reload, retry, hai thiết bị và worker restart không mất lịch sử hoặc tạo nháp/giao dịch trùng.
- Kiểm thử isolation hai user, thiếu scope, key revoked giữa run, foreign receipt/card và CSRF.
- Kiểm thử chỉnh card trong khi confirm, batch thành công một phần và stale version.
- Báo cáo đúng kỳ, timezone, transfer neutrality, report exclusion và nhiều tiền tệ; không lấy nháp vào báo cáo.
- Hội thoại dài kích hoạt summary vẫn tìm lại nguồn; cặp tool call/result hợp lệ; dữ liệu tài chính cũ không được coi là mới.
- Provider/OCR giả lập timeout, malformed tool arguments, retry và cancellation; smoke provider thật trước release.
- E2E desktop/mobile: hai điểm vào riêng, nhiều card, confirm, mở lại lịch sử và hỏi báo cáo sau ghi sổ.

Sau duyệt, implementation plan ghi fixtures và lệnh test chính xác; dự kiến backend integration/race (`rtk proxy go test -race ./...` trong backend với DB test riêng), frontend tests/build (`rtk npm test`, `rtk npm run build`), E2E và Docusaurus build theo scripts của repo. Chưa chạy các test runtime vì lượt này chỉ viết thiết kế.

Docs phải cập nhật cùng code: requirements, architecture/API/ERD, ADR cho hai luồng và scope, public Docusaurus hướng dẫn riêng từng chat, ví dụ API-key, lỗi/giới hạn; OpenAPI, llms.txt/llms-full.txt và skill API phải đồng bộ. Public docs hiện hành chưa đổi vì hành vi này chưa được triển khai. Context, backlog, validation matrix và release notes cập nhật theo bằng chứng thực tế, không đánh dấu tính năng hoàn thành từ bản thiết kế.

## 11. Review trước commit

Chủ sản phẩm xem bốn quyết định ở mục 1 và API/scope ở mục 7. Bản này chỉ là draft để review; không commit, push hoặc triển khai trước khi được yêu cầu tiếp.
