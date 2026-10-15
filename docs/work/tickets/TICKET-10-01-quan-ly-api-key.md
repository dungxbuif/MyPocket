---
artifact_type: ticket
id: TICKET-10-01
status: in_review
owner: human
parent: TICKET-10
trace:
  parent: TICKET-10-ket-noi-cong-cu.md
  guide: README.md
---

# TICKET-10-01 — Tạo và thu hồi API key

Ticket lớn: [Kết nối công cụ cá nhân](TICKET-10-ket-noi-cong-cu.md).

## Mục tiêu và phạm vi

Đặt tên, tạo key, xem thông tin sử dụng và thu hồi quyền của công cụ.

## Tiêu chí nghiệm thu

- [x] Secret chỉ hiển thị khi tạo; danh sách sau đó không lộ lại secret.
- [x] Trong trạng thái secret một lần, người dùng có thể chủ động sao chép
  chính xác key vào clipboard và nhận xác nhận; không lưu hoặc log secret.
- [x] Thu hồi khiến công cụ không tiếp tục dùng key đó được.
- [x] Key không tạo/quản lý key khác; management routes JWT-only và luôn owner-scoped.

Native copy affordance design: [API key copy detail](API-KEY-COPY-DETAIL_DESIGN.md).

Local implementation: the Stage v1 baseline plus `POST/GET/DELETE /api/v1/api-keys`, `mpk_<lookup>.<random>` format, digest-only persistence, expiry/revocation, advisor scope enforcement and redacted best-effort Redis advisor access audit. Remaining before public release: durable audit delivery/alerting, self-key endpoints, browser/provider E2E and deployment proof.

Native copy UI and tests are implemented; owner review on the physical iPhone is
pending because Xcode has no signed-in provisioning account/profile. Changelog
and release-note entry are deferred until owner approval.
