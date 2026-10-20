# Gửi giao dịch có khóa chống trùng

Trạng thái: server contract đã deploy trong API/worker image
`release-1.1.0-20261015-2212`; ứng dụng iPhone có outbox/sync nhưng vẫn fail-closed
cho tới khi owner-authenticated capability payload được kiểm chứng.
Capability phải được đọc bằng phiên owner trước mỗi drain. Nếu server trả 404,
`enabled:false` hoặc thiếu operation/version/meta thì app phải giữ hàng đợi,
không gửi tự động. Không giả định server cũ hỗ trợ chỉ vì nó chấp nhận header.

## Trạng thái kiểm chứng — 2026-10-16

Bộ kiểm tra HTTP offline/sync hiện có bao phủ capability, receipt metadata,
replay cùng khóa, xung đột khi đổi nội dung, retry chuyển ví đồng thời, ngữ cảnh
chuyến đi cố định và cô lập owner qua JWT/API key. Regression chuyển ví cũng
kiểm tra một receipt bền vững cho nhiều lần retry cùng khóa. Lần chạy backend
local này không có `TEST_DATABASE_URL`/`TEST_REDIS_URL`, nên các ca cần
PostgreSQL/Redis bị bỏ qua; đây không phải bằng chứng cho UAT production hoặc
thiết bị thật. Cần owner-authenticated capability payload và signed-device UAT
trước khi coi việc gửi offline đã sẵn sàng.

## Xem giao dịch khi mất mạng — bản native đang phát triển

Màn Giao dịch lưu bản dữ liệu đã tải thành công dưới dạng mã hóa, riêng theo tài
khoản và server. Khi lỗi mạng/timeout/408/429/5xx, app có thể hiện bản gần nhất
kèm thời điểm tải và nút **Tải lại**. Đây là số liệu đã xác nhận trước đó, chưa
phản ánh thay đổi mới; không cộng các khoản đang chờ gửi vào tổng tiền.

Màn Ví dùng cùng snapshot owner-scoped và read-through cache: khi mở lại màn,
app dựng danh sách từ bản lưu trước trong lúc refresh máy chủ chạy nền. Vì vậy
refresh không thay toàn bộ nội dung bằng màn hình loading; spinner toàn màn chỉ
dùng khi chưa có dữ liệu nào để hiển thị. Khi đang refresh, dữ liệu cũ vẫn được
giữ nguyên và chỉ có trạng thái nhẹ báo chưa cập nhật. Cache được mã hóa, gắn
với cả owner và server; đổi tài khoản không thể đọc lại bản của tài khoản cũ.

Lỗi đăng nhập, phân quyền hoặc response không hợp lệ không dùng cache để che lỗi.
Cache mất key/hỏng sẽ không tự bị xóa hay ghi đè. Cần phiên đã lưu còn dùng được;
chưa bổ sung đăng nhập offline khi phải refresh token. Overview và Reports hiện
cũng đọc được snapshot report server-authoritative khi lỗi mạng tạm thời; Reports
chỉ dùng snapshot nếu đúng tháng đang chọn và có đủ cả summary/insider, không tự
tính lại số liệu. Hàng đợi ghi offline đã có contract local và sender được
đóng gói trong bản native; việc drain thật vẫn chờ owner-authenticated capability
UAT để tránh gửi vào backend không tương thích.

## Cách dùng

### Sửa bản nháp chuyển ví bị từ chối — kiểm chứng local

Trong hàng đợi của sổ giao dịch thông thường, mở menu ba chấm → **Sửa bản nháp**.
Kiểm tra và thay đổi thông tin trước khi xác nhận; nếu chưa thay đổi thì không
gửi lại. Bản sửa hợp lệ dùng khóa mới. Cặp chuyển ví ghi cùng nhau, cập nhật hai
số dư nhưng không tính vào thu/chi báo cáo. Nháp bị từ chối vẫn còn sau restart.
Server capability đã deploy; hàng đợi trong ví tiết kiệm và UI native vẫn cần
kiểm chứng riêng.

Agent proof: aggregate/basic ledger controlled422→edited22500 passed; two
request keys, exactly one paired transfer, correct balances and unchanged monthly
income/expense/count. [Evidence](../work/tickets/OFFLINE-RECOVERY-E2E-20261014.md).
Do not generalize to savings scope, cached reports or deployed capabilities.

Ví tiết kiệm — bản local: hàng đợi dùng chung tài khoản cũng xuất hiện trong ví
tiết kiệm. Fault-injected 422 đã được kiểm chứng qua restart: số dư/tiến độ giữ
nguyên khi pending, sửa nháp thành 22.500 đ commit đúng một cặp chuyển ví,
reload 122.500 đ / 877.500 đ và không đổi báo cáo tháng. Artifact 1/1:
`/tmp/mypocket-offline-savings-20261015e/Logs/Test/Test-MyPocketIOS-2026.10.15_20-37-31-+0700.xcresult`.
Đây vẫn là bằng chứng local; native sender và production owner UAT chưa bật.
[Evidence](../work/tickets/OFFLINE-GOAL-QUEUE-20261014.md).

### Log chẩn đoán trên iOS

Bản local mới không in body phản hồi lỗi hoặc URL/query của yêu cầu JSON vào log
HTTP DEBUG. Log giữ method, mã HTTP và request ID để đối chiếu; thông báo lỗi trên
UI vẫn giữ hành vi cũ. Đừng gửi token, API key, ảnh hóa đơn hoặc nội dung giao
dịch vào log hỗ trợ. Log lỗi ở Trợ lý/Ngân sách/fixture đăng nhập cũng dùng bộ
mô tả an toàn chung: chỉ loại lỗi và mã HTTP nếu có, không in lỗi tùy ý. Vẫn
chưa phải kiểm chứng mọi nhánh lỗi trên thiết bị hoặc đã rollout production.

Agent contract: native JSON HTTP DEBUG diagnostics = method + optional HTTP
status + locally generated requestID only; no request path/query/body, response
body, authorization/idempotency headers. Typed error delivery is unchanged.
[Proof and remaining scope](../work/tickets/NATIVE-HTTP-LOG-PRIVACY-20261014.md).
Feature diagnostics = bounded APIError class + optional numeric status;
unknown Error = unexpected. [Feature proof](../work/tickets/NATIVE-FEATURE-LOG-PRIVACY-20261014.md).

### Native create — đang kiểm chứng local

Form tạo thu/chi và chuyển ví mới giữ một key cố định mỗi lần mở form, lưu bản
nháp mã hóa trước khi gửi. Receipt đúng mới xác nhận đã ghi trên server. Nếu
chưa xác nhận, phần **Chờ xác nhận** trong Giao dịch tách khỏi số dư/báo cáo.
Server không quảng bá capability thì giữ bản nháp, không thử một POST online
khác. Production hiện đã có route capability; client vẫn phải nhận response
owner-authenticated `enabled:true` trước khi bật sender.

Khoản thu/chi phải chọn danh mục trước khi lưu nháp, kể cả khi sửa nháp bị từ
chối. Thiếu danh mục không tạo nháp mới và không dừng nháp gốc. Chuyển ví dùng
hai ví và cặp giao dịch do server tạo, không yêu cầu danh mục thu/chi từ người dùng.
Nháp cũ đã lưu không bị tự sửa nội dung hoặc đổi key khi cập nhật app. Replay
giữ nguyên yêu cầu để server xác nhận receipt cũ hoặc trả lỗi nghiệp vụ có thể
xử lý; điều kiện tạo nháp mới không được làm mất khả năng phục hồi nháp cũ.

Sửa/xóa, tín dụng, điều chỉnh vẫn online-only. Bản nháp giữ ngày/chuyến đã chọn,
không tự gắn chuyến active mới khi gửi muộn; có thể gắn chuyến sau trong chi tiết.
Ledger/tổng quan tải lại sau receipt+CAS. Nút Đồng bộ hiện tôn trọng lịch retry;
UI xử lý lỗi và lifecycle tự gửi đã nối trong bản local. API/schema không đổi;
chưa deploy/cài iPhone, chưa coi toàn bộ offline đã hoàn chỉnh.

### Khi nào app tự đồng bộ?

App đang mở và có đường mạng sẽ gửi theo thứ tự nháp. Mở lại app hoặc mạng trở
lại sẽ kiểm tra hỗ trợ/quyền rồi tiếp tục; lỗi tạm thời có lịch thử lại, không gửi
dồn liên tục. Lịch của dòng đầu được tôn trọng trước các dòng chờ sau. Nếu server
chưa quảng bá hỗ trợ, app giữ nháp và kiểm tra lại với khoảng nghỉ ít nhất30giây.

Khi app xuống nền hoặc không hoạt động, timer và drain tạm dừng; không có job
tự ghi tài chính trong nền. Một lượt gửi chưa rõ kết quả có thể còn hiện **Đang
xác nhận** đến khi lease60giây hết hạn; sau đó gửi lại cùng key, không tạo khoản
mới. Tạm dừng không phải logout. Logout khóa nháp theo tài khoản; mở app/mạng trở
lại không tự mở khóa lỗi đăng nhập. Cần đăng nhập lại đúng tài khoản.

### Phục hồi bản nháp — đang kiểm chứng local

Trong **Giao dịch → Chờ xác nhận**, mỗi dòng hiện loại, số tiền, ghi chú và trạng
thái. Mở **menu ba chấm** của đúng bản nháp để xem các thao tác phù hợp bên dưới.
Menu không dùng được trong lúc đồng bộ; bản đang gửi không hiện menu.

- **Thử lại ngay** chỉ dành cho lỗi tạm thời: dùng lại key và nội dung cũ, không
  tạo thêm ý định giao dịch.
- **Sửa bản nháp** dành cho conflict hoặc server từ chối. Sửa thông tin rồi chọn
  **Xác nhận bản sửa**: app lưu bản mới bằng key mới và dừng bản cũ trong cùng một
  transaction local. Không tự sửa giao dịch đã tồn tại trên server. Mở form rồi
  lưu lại mà không thay đổi nội dung không được coi là bản sửa mới.
- **Bỏ bản nháp** chỉ dừng hàng đợi trên thiết bị, không xóa giao dịch trên server.
  Nếu lần gửi trước mất kết nối, giao dịch có thể đã được lưu; đọc cảnh báo và tải
  lại dữ liệu trước khi tạo một khoản thay thế.

Các thao tác không đổi chủ tài khoản. Bản đang gửi hoặc đã được xác nhận không
được bỏ bằng lệnh local này. E2E local đã kiểm chứng qua UI thật và đọc lại API:

- Lỗi503 hoặc mất response sau commit: restart/retry giữ một key, không tạo trùng.
- Nháp422/409: không đổi nội dung thì không tạo key mới; đổi số tiền rồi xác nhận
  tạo đúng một giao dịch mới, giữ ngày đã chọn.
- Bỏ nháp: Hủy cảnh báo giữ nguyên nháp; xác nhận chỉ dừng local, kể cả khi giao
  dịch đã commit. Restart không khôi phục nháp đã bỏ hoặc xóa giao dịch server.
- Chuyển ví khi503/mất response: đúng một cặp, số dư đổi một lần. Tổng thu, tổng
  chi và số giao dịch trong báo cáo tháng không tăng vì cặp chuyển ví.

Đây là kiểm chứng bản local, không phải đã phát hành. Vẫn còn cache/tài khoản/
thiết bị thật và accessibility cần kiểm chứng; bubble Feedback còn có thể che
nội dung. [Bằng chứng/phạm vi](../work/tickets/OFFLINE-RECOVERY-E2E-20261014.md).

Trước khi gửi hàng đợi, gọi `GET /api/v1/sync/capabilities` bằng JWT/session hoặc API key của tài khoản:

```json
{
  "data": {
    "contract_version": 1,
    "enabled": true,
    "operations": ["transaction.create", "transfer.create"],
    "receipt_metadata": "/meta/mutation",
    "idempotency_key_required_for_offline": true,
    "legacy_unkeyed_online": true,
    "transaction_create_context": "explicit_travel_v1"
  },
  "meta": null
}
```

Chỉ cho phép thao tác đã biết và được quảng bá khi enabled=true,version1,
metadata/key policy đúng. 404, thiếu/sai field, version lạ hoặc disabled phải giữ
bản nháp/hàng đợi, không thử POST để đoán hỗ trợ. Kiểm tra mới trước mỗi đợt gửi;
câu trả lời luôn có `Cache-Control: no-store`; client/proxy không được lưu capability
để tái sử dụng sau rollback. Không dùng capability cache cũ để gửi sau khi server rollback.
Endpoint này chỉ
công bố implementation đã cấu hình, không bảo đảm DB/network đang khỏe.
Backend, engine, create/queue, recovery và foreground scheduler đã nối local;
full E2E/thiết bị thật và rollout vẫn còn gate riêng.

Transport iOS mới giữ metadata chống trùng, kiểm tra tài khoản của hàng đợi
trước/sau request. Refresh cũ không ghi đè đăng nhập mới hay khôi phục phiên đã
logout. Không chuyển ý định chờ sang tài khoản khác. Đây mới là bảo vệ HTTP/
session; chưa phải offline đã phát hành cho người dùng.

Adapter iOS cũng kiểm tra receipt và dữ liệu trả về trước khi xác nhận: key,
operation, trạng thái applied, tài khoản, ví, số tiền, loại và thời điểm phải
khớp ý định đã lưu. Chuyển ví phải có đúng hai dòng cùng transfer_id và đều
không tính báo cáo. Response thiếu hoặc sai receipt chưa được coi là thành công;
giữ nguyên key/nội dung để kiểm tra lại. Adapter chỉ trả ID đã xác nhận, không
tự đổi trạng thái hàng đợi. Coordinator/engine đã nối create và hàng đợi local,
nhưng recovery/scheduler/E2E chưa hoàn chỉnh; chưa phát hành cho người dùng.

Engine chỉ gửi lần lượt, kiểm tra capability mới mỗi đợt và ghi trạng thái sau
receipt hợp lệ. Mất mạng/timeout/408/429/5xx lên lịch retry cùng key với backoff
lũy thừa có jitter, tối đa15phút; không gửi liên tục. Lỗi400/403/404/422 dừng ở
rejected,409ởconflict. Hết phiên/đổi tài khoản khóa hàng đợi; chỉ đăng nhập lại
đúng tài khoản mới tiếp tục. Logout hủy drain, response về muộn không được tự
đánh dấu thành công. Foreground scheduler đã nối local; full recovery UI/cache/
report verification vẫn đang triển khai, chưa phải tính năng đã deploy.

Riêng `transaction.create` còn cần `transaction_create_context` đúng bằng
`explicit_travel_v1`. Server cũ thiếu field hoặc trả version lạ không được dùng
để gửi khoản thu/chi offline. Chuyển ví không tự gắn chuyến đi nên vẫn dùng gate
version/key/metadata riêng, không suy diễn quyền từ field của giao dịch thường.

Khi chuẩn bị khoản thu/chi, lưu ngày giờ cụ thể, `auto_link_travel:false` và
`travel_event_id` đã chọn (hoặc không có nếu không thuộc chuyến nào). Field false
ngăn server dùng chuyến active mới khi nhận yêu cầu muộn. Một chuyến đã chọn vẫn
phải thuộc tài khoản; chuyến đã bị xóa khiến ý định mới bị từ chối. Replay của
ý định đã thành công vẫn trả receipt gốc, không ghi lại hay gắn sang chuyến khác.
Không sửa policy/key/payload sau lần gửi đầu.

Request online bỏ field, gửi null hoặc true giữ tự gắn chuyến active như trước.
Field này chỉ dành cho tạo mới, không phải cách gỡ liên kết giao dịch cũ; dùng
`PATCH /transactions/{id}/travel` để đổi/gỡ chuyến cho giao dịch đã lưu.

1. Tạo một UUID cho một ý định lưu giao dịch. Lưu cả key và nội dung trước lần gửi
   đầu tiên. Không tạo key khác chỉ vì mất mạng hoặc timeout.
2. Gửi `Idempotency-Key` cùng request tới `POST /api/v1/transactions` hoặc
   `POST /api/v1/transactions/transfer`, sử dụng JWT/session hoặc API key còn hiệu lực.
3. Nếu kết quả chưa rõ, gửi lại **cùng key và cùng nội dung**. Không thêm giao
   dịch thủ công khác để thay thế khi chưa biết kết quả.
4. Sau thành công, tải lại số dư/giao dịch/báo cáo. Kết quả replay là bằng chứng
   thao tác ban đầu đã áp dụng, không có nghĩa dữ liệu chưa bị xóa/sửa sau đó.

Ví dụ response có key:

```json
{
  "data": {"id": "transaction-id"},
  "meta": {
    "mutation": {
      "operation": "transaction.create",
      "idempotency_key": "2b8c6c98-0000-4000-8000-000000000001",
      "status": "applied",
      "replayed": false
    }
  }
}
```

`data` ở ví dụ được rút gọn: response thật giữ đầy đủ giao dịch. Chuyển ví vẫn
trả `data` là mảng hai dòng, metadata ở cùng vị trí `meta.mutation`.
Lần ghi đầu trả201/replayed=false; retry trả200/replayed=true và ID ban đầu.
Ngày tự điền, ví/danh mục và chuyến đi đã chọn ở lần đầu không được tính lại.
Ví/danh mục bị thay đổi/xóa sau đó không khiến retry tạo giao dịch mới.

Trong giai đoạn tương thích, request **không có header** vẫn chạy luồng online
cũ với `meta:null`, không có bảo vệ retry. Không dùng nhánh này cho offline hoặc
tự động gửi lại. Việc bắt buộc key toàn bộ endpoint là gate rollout còn mở.

Code local mới hỗ trợ user API key toàn quyền tài khoản, không scope, theo quyết
định của chủ app. Key cũ cũng có quyền ghi/replay sau rollout; không cần tạo lại.
Local PostgreSQL/Redis HTTP tests xác nhận JWT và key cùng owner đều ghi/replay
được, nhưng không đọc chéo receipt hay ví của tài khoản khác. Key hết hạn/thu hồi
và JWT/session hết hiệu lực bị401 trước khi replay. **Chưa deploy prod**;
kiểm tra phiên bản dịch vụ trước khi dùng. [Hướng dẫn key](api-keys.md).

## Lỗi và phục hồi

Code iOS mới đã nối hàng đợi mã hóa, sync engine và form tạo thu/chi/chuyển ví
trong bản local. Lịch tự gửi khi mạng trở lại/background và các màn xử lý lỗi
chưa kiểm chứng đầy đủ, nên chưa thể coi app hoàn toàn offline. Key mã hóa giữ
riêng trong Keychain trên thiết bị, không đồng bộ
iCloud; mất key hoặc gỡ app có thể làm mất khả năng đọc bản chưa gửi. Không xóa
dữ liệu app để thử khắc phục một yêu cầu chưa rõ đã lưu hay chưa.

Outbox local đã cố định policy false và giữ chuyến/ngày trong payload mã hóa.
Form tạo và engine đã dùng policy này local; ngữ cảnh nhập hoàn toàn từ cache
khi mất mạng vẫn cần kiểm chứng riêng. Không dùng mặc định online (thiếu field)
cho giao dịch gửi muộn.

- `400 INVALID_IDEMPOTENCY_KEY`: header rỗng, nhiều giá trị, quá128byte hoặc ký tự
  không hợp lệ. Sửa request; không có ghi tài chính từ yêu cầu bị từ chối này.
- `409 IDEMPOTENCY_KEY_REUSED`: cùng owner/operation/key nhưng nội dung khác.
  Không tự tạo key mới; người dùng phải xem lại ý định và xác nhận sửa/gửi mới.
- `503 mutation_in_progress` hoặc `mutation_unavailable`: không fallback sang
  ghi không key. Giữ nguyên ý định/key; kiểm tra dịch vụ trước khi retry.
- Lỗi nghiệp vụ/xác thực vẫn theo API hiện tại. Không tự retry mọi400/403/404/409.

## AI-agent reference

| Contract | Value |
| --- | --- |
| Operations | transaction.create, transfer.create |
| Capability | GET /api/v1/sync/capabilities, JWT/session or owner key, contract_version1, `Cache-Control: no-store`; old404/disabled/unknown blocks sends |
| Key | one ASCII value,1–128bytes, stable per intent; offline client uses UUID |
| Scope | authenticated owner + operation + key, never body owner_id |
| Hash | decoded typed JSON before generated IDs/time/travel defaults; JSON order/whitespace ignored |
| Result | same object/array data shape; meta.mutation contains operation/idempotency_key/status/replayed |
| Safety | completed receipt lookup precedes mutable catalog validation; ledger+receipt commit atomically |
| Legacy | absent header retains unprotected online path, not offline-ready |
| Auth | JWT + live Redis session OR active full-owner key; both bind authenticated owner |
| Native store | SQLite schema1, AES-GCM, per-server/account device-only Keychain key; default native create/queue wired in the installed Release build, sender remains capability-gated |
| Local validation | ordinary income/expense requires category identifier at enqueue/replacement; optional jar/trip remain optional; transfer has no ordinary category input; existing ciphertext is not migrated/deleted |
| Compatibility | validateNewIntent applies only at creation/replacement; leased frozen replay keeps previous validation/key/body and lets server resolve original receipt or business rejection |
| Lifecycle | default AppRoot owner+scene and NWPathMonitor→coordinator; active+reachable only; one FIFO due/lease timer; inactive/network loss suspends generation without auth-block; logout separately blocks auth |
| Scheduler | exact persisted retry/lease deadline, min0.25s timer floor, unsupported pending capability cooldown at least30s; blockedAuth has no auto wake; no background financial-write job |
| Recovery | stable key on restart/retry; expired lease only; stale lease outcome refused; never regenerate missing key over existing ciphertext; changed rejected/conflict replacement uses a new key atomically |
| Recovery UI | compact type/amount/note/status; MPOverflowMenu exposes eligible retry/edit/discard, disabled while busy; discard requires uncertain-server-result confirmation |
| Recovery proof | local real native UI+API:422/409 changed-only replacement/newkey/date; rejected/committed discard remains local after restart; transfer503/commitdrop one paired write, unchanged monthly report aggregates |
| Native session | expectedOwnerID local-only; identity before/after request; refresh CAS preserves later login/logout |
| Travel | transaction.create requires capability explicit_travel_v1 + stored auto_link_travel=false + explicit trip/no-trip and fixed occurred_at; legacy omitted/null hash unchanged |
| Outstanding | authenticated capability payload, full cache/recovery/transfer/session UI verification and physical-device UAT; no full offline availability claim until those gates pass |

[API/OpenAPI](../architecture/API.md), [verification](../work/tickets/OFFLINE-CREATE-RECEIPTS-20261014.md).
