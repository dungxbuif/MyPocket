---
artifact_type: detail_design
id: AI-CHAT-UNIFIED-20261031
status: draft
owner: shared
approval: approach_accepted_full_design_pending_review
---

# Trợ lý tài chính: chat, chứng từ và hành động trong cùng hội thoại

## Quyết định và phạm vi

Owner đã chọn đính kèm chứng từ, xem/sửa/duyệt giao dịch ngay trong hội thoại.
Thiết kế bao gồm backend Go và FE iOS SwiftUI. Web được ghi nhận là archived tại
[design gateway](../../design/README.md); không mở lại dự án web trong lát này.
API additive giữ client cũ hoạt động. Đây là thiết kế đề xuất, chưa phải chức năng
đã triển khai hoặc bằng chứng production. Cần review thiết kế trước implementation.

Nguồn hình: `/Users/dungxbuif/Downloads/stitch_my_pocket (1).zip` gồm `code.html`,
`screen.png`, `DESIGN.md`. Nguồn chức năng: advisor/entry handler, tool registry,
native AssistantViews, FinanceAPI và các screen/base contract hiện tại.

Mục tiêu: người dùng hỏi về tiền, xem bằng chứng và duyệt thao tác ngay trong một
luồng; mở lại app vẫn biết request và đề xuất đang ở đâu. Giữ màu hiện tại, light
mode, chữ/section gọn, vùng hội thoại rộng. Feedback bubble giữ trên các màn khác,
ẩn riêng Assistant theo chỉ đạo mới của owner.

## Các phương án

1. Chat mở màn nhập AI riêng: tái sử dụng nhanh nhưng ngắt ngữ cảnh. Không chọn.
2. Một hội thoại tích hợp, tái sử dụng engine nhập AI và domain handlers: owner
   đã chọn. Cần thêm binding/history/progress và đồng bộ trạng thái UI.
3. Viết lại toàn bộ AI thành engine đa phương thức mới: nhiều rủi ro và trùng
   domain logic. Không chọn.

Không thuộc đợt này: voice, tích hợp ngân hàng, cashback/ưu đãi, sức khỏe 84/100,
đánh giá mức A, dự báo chắc chắn, tự ghi sổ không xác nhận, thay model/provider,
đa hội thoại có sidebar, triển khai production. Chip chỉ xuất hiện khi capability
và dữ liệu hỗ trợ. Không sao chép số liệu/đánh giá minh họa trong Stitch.

## Luồng chức năng

| Ý định | Kết quả trong chat | Hành động |
| --- | --- | --- |
| Phân tích/so sánh chi tiêu | Kết luận ngắn, số liệu, phạm vi ví, kỳ và nguồn | Hỏi tiếp, mở báo cáo |
| Tìm giao dịch | Tối đa 5 row đầu, tổng số khớp và phân trang | Xem thêm, mở đúng giao dịch |
| Ví/hũ/ngân sách/mục tiêu | Card tiến độ cùng số liệu domain thực | Xem chi tiết, yêu cầu đề xuất thay đổi |
| Tạo/sửa/xoá dữ liệu | Preview có loại thao tác và các trường thay đổi | Sửa bản đề xuất, xác nhận, bỏ qua |
| Ảnh/PDF/XLSX hỗ trợ | Attachment + tiến độ + batch proposal | Sửa, lưu từng dòng, lưu các dòng hợp lệ, bỏ qua |
| Hướng dẫn sản phẩm | Câu trả lời và nguồn hướng dẫn | Đi tới chức năng tương ứng |

Ví dụ: “Ăn trưa 85k, ví tiền mặt” → thiếu danh mục/hũ thì dùng rule hiện có hoặc
hỏi lại khi chưa đủ; preview thể hiện rõ 85.000 đ, ví, ngày, danh mục/hũ → người
dùng sửa hoặc lưu → trạng thái xác nhận từ server và link giao dịch thật.

Ví dụ chứng từ: đính kèm 3 ảnh → thấy từng file và lỗi riêng → gửi một request
→ theo dõi tiến độ → xem các dòng đề xuất → sửa một dòng bằng transaction editor
base → lưu các dòng đã kiểm tra → giữ kết quả từng dòng trong lịch sử chat.

## Anatomy iOS và base mapping

Màn chính chỉ giữ header gọn, transcript chiếm vùng còn lại, composer neo trên
bàn phím/navigation. Không thêm dashboard hoặc insight feed riêng dưới transcript.
Chào/gợi ý/tóm tắt đều là nội dung trong luồng; thông tin tổng quan dài mở sheet.

| Vùng Stitch / trạng thái | Base tái sử dụng hoặc mở rộng |
| --- | --- |
| Header bot/menu | MPScreenHeader, MPIcon, MPOverflowMenu; trạng thái theo service thực |
| Chat toàn vùng | MPChatSurface (phải tự sở hữu VStack và layout), MPChatMessage, MPMarkdown |
| Gợi ý câu hỏi | MPButton variant chip; gợi ý theo capability, không ghi đè draft |
| Ô nhập/đính kèm/gửi/dừng | MPChatComposer + MPDocumentPicker + MPAttachmentThumb |
| Số liệu/so sánh/tiến độ | MPMetricGrid, finance comparison/progress bases |
| Preview giao dịch | MPSurfaceCard + MPTransactionRow; field/row giống Thêm giao dịch |
| Sửa proposal | Tách shared transaction fields/editor binding từ form hiện tại; không fork validation |
| Review hành động | MPAssistantCard + shared confirmation/action row; dùng variant theo action |
| Loading/lỗi/phục hồi | MPStatusMessage, MPLoadingView, MPButton |

Screen chỉ compose dữ liệu/action; token typography, padding, radius và variant
thuộc DesignSystem. Không bọc thêm bubble quanh card đã có surface. Heading
section nhỏ hơn số tiền. Composer tăng chiều cao theo nội dung nhưng không chiếm
hết transcript khi bàn phím mở. Tệp đính kèm có preview/remove trước gửi.

Khi đọc tin cũ: không tự kéo xuống; có nút Tin mới. Load older giữ anchor. Khi
đang ở cuối: theo dõi tin mới. Chuyển tab giữ draft; mở sheet sửa không mất vị trí.
Privacy masking áp dụng cả nội dung/card/attachment metadata nhạy cảm và VoiceOver.
Nguồn, phạm vi và cảnh báo thiếu dữ liệu vẫn đọc được.

## Backend: tái sử dụng và phần cần bổ sung

Đã có: `/ai/advisor/messages/async`, run GET/events/cancel, lịch sử/reset,
`/actions/{id}/confirm`; `/ai/entry/process`, request recovery, proposal patch/
approve/reject; tools đọc và đề xuất owner-scoped.

Thiếu: binding entry task vào conversation/generation, typed attachment/batch part,
lịch sử trạng thái proposal có thể khôi phục, capability cho loại card/action iOS.
Native renderer hiện có read cards và confirmation ví/hũ/ngân sách; cần bổ sung
transaction create/update/delete/recategorize, transfer, category create/update.

### Contract additive đề xuất

- Giữ JSON text submit hiện tại. Thêm endpoint multipart
  `POST /ai/advisor/entries` với `client_request_id`, `conversation_id`,
  `generation`, `text`, `files`; trả `202` cùng task ID/request ID và binding.
  Việc nhận request phải được lưu bền trước khi trả accepted; OCR/model chạy sau.
- Task `GET /ai/advisor/entries/{id}` trả trạng thái, stage nếu đo được, file
  metadata, batch reference và terminal error. GET phải đủ để phục hồi khi SSE
  không khả dụng. Sự kiện entry dùng hạ tầng replay hiện có nhưng có namespace/
  task identity rõ; không gửi entry ID vào advisor run API hiện tại.
- Adapter gọi lại pipeline nhập hiện tại; không tạo pipeline OCR/normalize/save
  thứ hai. Plain text chỉ submit một lane; FE không gửi cùng ý định vào cả entry
  và advisor rồi tạo hai bộ đề xuất.
- History thêm typed `entry_batch` part tham chiếu batch ID và version, không
  chép proposal thành các bản độc lập. Có batch read trả trạng thái authoritative
  để hydrate sau relaunch. Metadata tệp chỉ dùng private authorized access.
- Capabilities thêm version, supported part kinds, action kinds, attachment
  formats/limits và entry integration flag. FE ẩn control chưa hỗ trợ; unknown
  part hiển thị fallback đọc được và không cho thao tác ghi không hiểu.

Tên endpoint/DTO trên là đề xuất để review; cần contract tests/OpenAPI trước FE
integration. Chưa khẳng định endpoint đã tồn tại.

### Dữ liệu và lifecycle

Thêm binding bền owner + conversation + generation + client request + entry request.
Unique theo owner/conversation/generation/client request; request fingerprint bắt
retry khác payload trả conflict. Task lưu status, event sequence và terminal result.
Không giả định lifecycle entry cũ tự đáp ứng history chat: migration riêng phải
chốt retention, replay và ownership của task/attachments trước implementation.

Task states: accepted → processing → review_ready; hoặc failed/cancelled.
Stage chỉ thông báo việc backend thật sự đo được: upload/đọc tài liệu/phân tích/
kiểm tra đề xuất. Không hiển thị phần trăm hoặc token stream giả.
Proposal states: pending → applying → applied/rejected; expired/conflict là trạng
thái có thể cần tạo lại preview. Applied phải giữ resource ID và kết quả xác nhận
để retry/mở lại chat không làm xuất hiện nút Lưu một lần nữa.

Edit entry proposal tiếp tục dùng PATCH/approve hiện có, thêm revision check nếu
hiện trạng chưa có. Edit advisor action tạo replacement proposal được server
validate và bind generation; proposal cũ superseded. Confirm không nhận arbitrary
financial fields từ FE. Không có đường tắt save bỏ qua proposal/canonical rules.

Huỷ tác vụ chỉ dừng xử lý chưa lưu; không undo giao dịch đã xác nhận. Reset hội
thoại ngăn completion muộn chèn lại part và vô hiệu hoá pending action generation
cũ. Chứng từ/task cũ tuân retention đã chốt; không tự xoá giao dịch được duyệt.

Lưu nhiều là các quyết định từng dòng với kết quả riêng; không tuyên bố atomic
batch. Retry các dòng chưa rõ bằng identity cũ; không gửi lại dòng applied.
Chuyển ví đi qua canonical atomic transfer, không ghép hai income/expense riêng.
Giữ rule category, jar eligibility, account timezone, inactive/stale reference,
mixed XLSX/OCR restrictions và private attachment limits hiện có.

## Hành vi lỗi và cảm giác tốc độ

Trong FE, tin gửi xuất hiện tức thời với trạng thái Đang gửi và stable local ID.
Server accepted mới chuyển sang Đang xử lý; gửi không chắc chắn giữ draft/request
identity và GET recovery, không tạo POST mới. Đổi mạng/relaunch không nhân đôi tin.
Lỗi file hiển thị đúng file; kết quả từng proposal độc lập. Bỏ qua một đề xuất phải
có trạng thái từ backend khi cần lưu qua relaunch; không chỉ ẩn local.

Đo riêng tap→local feedback, submit→accepted, OCR, provider, validate và confirm→
resource persisted. Tiêu chí UX: phản hồi local trong 100ms trên thiết bị kiểm thử;
API timeouts/budgets theo đo p50/p95, không hứa tốc độ model chưa có evidence.
Tạo giao dịch thường chậm được điều tra riêng tại FinancialMutationQueue.submit /
finishPersisted và canonical write/refresh; không quy hết cho AI hoặc thay đổi
ledger semantics bằng optimistic balance. Phân biệt đã giữ trên máy/chờ xác nhận/
đã ghi sổ nếu chọn cải thiện queue UX sau đo.

## Các lát triển khai và nghiệm thu dự kiến

1. Contract + capability matrix + base/UI state spec; chốt schema/retention và ADR.
   Dựng fixture minh hoạ cả trạng thái đọc, review, applying, applied, error.
2. Hoàn thiện text/read/action trong chat: native renderer đủ 12 action kinds được
   capability công bố, generic typed confirmation adapter và server status hydrate.
3. Entry task binding + async acceptance/recovery + migration + replay; kiểm thử
   pipeline thật với fixture provider và disposable DB/storage trước FE hookup.
4. Attachment/batch/review trong chat; shared editor, save-one/save-valid/skip;
   giữ entry route độc lập làm đường truy cập hiện có dùng cùng engine/base.
5. Đo latency, failure recovery và review visual toàn app dùng base đã thay đổi;
   owner kiểm từng flow trên simulator có data, rồi device UAT. Release riêng.

Không chia agents sửa chung AssistantViews hoặc cùng migration. Có thể song song
contract/backend task và shared base fixtures sau khi version/schema đã chốt;
tích hợp native phụ thuộc backend contract/fixtures ổn định.

## Verification và rủi ro

- Backend contract/DB: foreign owner, payload mismatch, expired generation,
  concurrent confirm/edit, duplicate approve, response loss/replay, reset/cancel
  racing completion, stale domain version, worker restart/lease recovery.
- Batch: partial success, duplicate documents, unsupported type, limits, ambiguous
  amount/date/category/transfer; ledger không thay đổi trước confirm.
- Native: submit giữ draft, append attachments, multiline keyboard, unknown part,
  scroll anchor, hydrate applied state, partial retry, privacy, đủ hit target.
- Visual: chụp màn trống/có nhiều tin/card/batch/bàn phím/lỗi; small và regular
  iPhone; large Dynamic Type; light mode. So cùng data với Stitch và base guideline.
- Backend `go test ./internal/usecase ./internal/controller/http
  ./internal/infrastructure/repository` tại backend, fixture DB theo test suite;
  native xcodebuild focused unit/API-backed UI, base/project guards và diff check.
- Phải đọc xcresult failure trước kết luận. UI test lỗi hiện tại từ đợt styling
  chưa có root cause xác nhận; không gọi đó là runner-only hoặc visual pass.

Rủi ro chính: entry retention khác chat retention; retry làm trùng proposal;
confirm lost response; UI card stale; unsupported card có nút ghi; panel layout
không thực sự xếp transcript/composer trong VStack. Các ca trên là acceptance
gates, không chỉ checklist styling.

## Trace và reconciliation

[Backlog](../../work/BACKLOG.md) · [Phase native](../../work/phases/PHASE-002-native-ios-swiftui.md)
· [Advisor ticket](../../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md)
· [Entry ticket](../../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md)
· [API](../../architecture/API.md) · [Base contracts](../../design/system/BASE_COMPONENTS.md)
· [Assistant spec](../../design/pages/finance-assistant/README.md)
· [Entry spec](../../design/pages/assistant/README.md)
· [Validation](../../work/VALIDATION_MATRIX.md) · [Context](../../CONTEXT.md)
· [Changelog](../../releases/CHANGELOG.md).

Thiết kế này supersede phương án composer né bubble 52pt của
[draft cũ](../../work/tickets/NATIVE-CHAT-COMPOSER-20261031-DETAIL_DESIGN.md) riêng
trên Assistant; giữ các yêu cầu density/icon/base khác.

Trước implementation: full design review; sau duyệt tạo execution plan và ADR
cho conversation-entry binding/lifecycle, reconcile API/ERD/requirements.
Sau implementation: ghi command evidence vào validation, cập nhật screen/base
spec, context/backlog/changelog và owner UAT. Chưa có test runtime mới trong bước
thiết kế; docs review kiểm tra source/contract và diff whitespace.
