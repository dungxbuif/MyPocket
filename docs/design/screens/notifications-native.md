# Native notifications screen

Status: in_progress. The notifications route keeps the existing owner-scoped
list and `markNotificationRead` mutation. Populated notifications are grouped
into “Chưa đọc” and “Đã đọc” sections while preserving API order within each
section.

Each item uses a shared reference surface, compact bell identity, title/body/
date typography and an icon-sized “Đánh dấu đã đọc” control. Unread state is
represented by the action-colored bell/dot; read items remain visible for
history and have no write control. Empty, loading, retry and refresh behavior
is unchanged.

No API/model/schema changes are introduced. The presentation grouping is pure
and covered by `CreditNotificationsReferenceTests`; parent integration must
run signed simulator tests and a populated screenshot with both unread and
read items.
