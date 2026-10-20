# API key cho tích hợp cá nhân

Trạng thái 2026-10-20: chính sách toàn quyền đã triển khai cho REST.
Key có toàn quyền đọc và thay đổi dữ liệu của tài khoản tạo key, tương đương
JWT của tài khoản đó. Không còn chọn scope; key cũ còn hiệu lực cũng có toàn quyền.
Chỉ chia sẻ key với dịch vụ bạn tin tưởng. Thu hồi key không còn dùng.
MCP hiện là bản ứng viên local, chưa xác nhận triển khai production.

## Tạo và dùng key

Trong app: Tài khoản → hồ sơ → API keys → **+**, nhập tên rồi **Tạo**.
Secret chỉ hiện một lần; dùng **Sao chép API key** và lưu trong secret manager.
App dùng clipboard local-only, hết hạn sau5phút; không log/lưu secret vào docs.
Nếu quên secret, thu hồi và tạo key mới; không có API xem lại secret cũ.

Qua API, dùng JWT hoặc một API key còn hiệu lực của chính bạn:

```http
POST /api/v1/api-keys
Authorization: Bearer <JWT-or-user-key>
Content-Type: application/json

{"name":"Tích hợp cá nhân","expires_at":"2027-01-01T00:00:00Z"}
```

expires_at không bắt buộc; nếu gửi phải là RFC3339 trong tương lai. Không gửi
scopes. Server bỏ qua scopes do client cũ gửi; nó không giới hạn quyền. Response
vẫn có scopes:[] để app đã cài đọc được, không phải danh sách quyền.

Mọi user endpoint trước đây nhận JWT đều nhận:

```http
Authorization: Bearer <mpk-secret>
```

Bao gồm ví/giao dịch/chuyển ví, danh mục/ngân sách/hũ, báo cáo/định kỳ/du lịch/
đầu tư/tín dụng, AI entry/Assistant, Feedback, profile và quản lý key. Validation
nghiệp vụ, xác nhận thao tác AI và password khi đăng ký Face ID vẫn giữ nguyên.
Key chỉ thao tác tài khoản của nó; user_id/owner_id trong request không đổi owner.
Giao dịch cần chống trùng vẫn phải giữ Idempotency-Key ổn định theo
[hướng dẫn sync](offline-sync-api.md).

## Thu hồi và đăng xuất

- GET /api/v1/api-keys chỉ trả metadata của bạn, không trả secret/hash/lookup.
- DELETE /api/v1/api-keys/{id} thu hồi key của bạn; key hết hạn/đã thu hồi bị401.
- Trong app, key đã thu hồi không hiện lại khi mở lại danh sách. Secret vừa tạo
  và nút copy của nó được xóa khi thu hồi; API metadata vẫn giữ lịch sử thu hồi.
- POST /api/v1/auth/logout dùng Bearer API key sẽ thu hồi **chính key đó**,
  không cần refresh_token, không đăng xuất các phiên khác.
- JWT logout vẫn cần refresh_token và thu hồi family của phiên hiện tại.
- Key không refresh thành JWT. Login/refresh/device-login giữ credential exchange
  riêng. Dịch vụ integration phải giữ secret an toàn, không đặt trong URL/log.

User key không phải FEEDBACK_AGENT_TOKEN: không đọc feedback của tất cả user,
không truy cập route internal/admin chỉ dành cho service token.

## Kết nối MCP

Cấu hình client Streamable HTTP đến `<server-origin>/api/v1/mcp` với header
`Authorization: Bearer <JWT-or-mpk-secret>`. Không đưa secret vào URL.
Client dùng quy trình khởi tạo của SDK rồi `tools/list` để lấy tên và schema
công cụ hiện tại; chưa có luồng đăng nhập hay tạo phiên qua MCP.

MCP có cùng quyền của tài khoản sở hữu key, không có scope riêng. Công cụ ghi
thông thường như tạo ví/giao dịch/ngân sách thực thi trực tiếp theo API tương
ứng. Chỉ luồng bản nháp Assistant vẫn cần xác nhận action đã lưu phía server
với đúng hội thoại/thế hệ; không gửi lại payload tài chính để thay thế bản nháp.
Giữ khóa chống trùng trong `params._meta.idempotency_key` cho API hỗ trợ receipt,
không đặt vào `arguments`; MCP không tự thử lại hay tự tạo khóa.

[Hướng dẫn MCP cho người và AI agent](../ai/mcp-api-parity.md) có ví dụ placeholder,
giới hạn file/base64, trạng thái async, đọc lại sự kiện, tải tài liệu, CAS và lỗi.
Không lưu/log link tải đã ký, nội dung tài chính hoặc secret trả về một lần.

## AI-agent reference

| Property | Contract |
| --- | --- |
| Authentication | Authorization: Bearer mpk_…; never JWT fallback for key-looking values |
| Authority | full same-owner user operations; no scopes or per-key permissions |
| Identity | server derives owner from validated key row; request cannot impersonate another owner |
| Create | POST /api/v1/api-keys, name + optional future RFC3339 expires_at; secret returned once |
| Legacy | persisted scopes ignored; response scopes:[] is compatibility only |
| Lifecycle | digest comparison + expiry/revocation check each request; long advisor runs revalidate |
| Logout | POST /api/v1/auth/logout with API key revokes that key only |
| Session | JWT retains Redis session validation; API keys do not auto-refresh |
| Service boundary | dedicated internal/agent tokens remain separate, not granted to user keys |
| MCP | local `/api/v1/mcp` candidate; official Go SDK v1.8.0; `tools/list` supplies exact input schemas |
| MCP writes | ordinary user writes direct; existing Assistant action confirmation still required |
| Deployment | REST full-owner policy deployed; MCP rollout/owner-authenticated staging acceptance remain open |

[API reference](../architecture/API.md), [auth](../architecture/AUTH_API.md),
[work/proof](../work/tickets/API-KEY-FULL-ACCESS-20261014.md).
