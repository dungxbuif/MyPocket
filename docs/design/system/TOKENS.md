# Token contract

Nguồn primitive duy nhất: app/src/ui/theme.css, được import từ app/src/styles.css. Named component classes nằm ở app/src/ui/variants.ts; app/src/atomic/atoms/tokens.ts chỉ re-export để tương thích import cũ.

| Role | Token | Giá trị | Dùng cho |
| --- | --- | --- | --- |
| CTA | brand / brand-hover | #4caf50 / #006e1c | primary button, FAB |
| Action | action / action-hover | #059669 / #047857 | link, positive amount, focus |
| Selected/progress | accent | #10b981 | checked, ring, progress |
| Canvas/card | canvas / card | #f6f8fa / #ffffff | nền màn / card |
| Text | heading / ink | #0f172a / #1e293b | heading / body |
| Metadata | secondary / muted | #64748b / #94a3b8 | nội dung phụ / chú thích |
| Divider/row | line / row | #f1f5f9 / #f8fafc | border / hover và muted card |
| Danger | danger / danger-hover | #f43f5e / #e11d48 | âm, lỗi, xóa |
| Danger background | danger-soft / danger-line | #fff1f2 / #ffe4e6 | feedback |
| Success background | success-soft / success-line | #ecfdf5 / #d1fae5 | positive badge |
| Warning | warning / warning-soft / warning-line | #d97706 / #fffbeb / #fef3c7 | cảnh báo |
| Card/control radius | radius-card / radius-control | 12px / 12px | corners hiện hành |
| Badge radius | radius-badge | 16px | compact visual badge khi dùng |
| Card shadow | shadow-card | 0 4px 20px, black 3% | subtle elevation |

Category palette dùng các named tones trong IconBadge; business mapping chọn tone, không cung cấp màu. Các màu orange/teal/red còn có token riêng trong theme; extended category palette lấy từ Tailwind tại variant registry, không ở screen.

Text dùng Text/Heading với size/weight/tone. Số tiền dùng numeric; scale text 10/11/12/14/16/18/20/24/30/36px. Kích thước 10–11px chỉ cho metadata phụ; không dùng cho input/hành động.
Layout spacing theo Tailwind scale; kích thước đặc thù tree/chart phải được base sở hữu và đặc tả. Không khai báo thêm màu literal trong TS/TSX/CSS ngoài theme. Không thêm font/theme song song.
