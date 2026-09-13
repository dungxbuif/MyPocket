# MyPocket Design Contract

Đây là gateway UI bắt buộc. Đọc theo thứ tự: [system](system/DESIGN.md) → [tokens](system/TOKENS.md) → [base contracts](system/BASE_COMPONENTS.md) → [inventory](INDEX.md) → đặc tả component/màn đang làm.

## Base-first bắt buộc

1. Trước khi viết JSX, ghi component cần dùng và đường dẫn implementation vào ticket/detail design.
2. Có base phù hợp: compose bằng props/variant/tone/size. Thiếu: định nghĩa contract và bổ sung base trước, rồi mới dùng trong màn.
3. Native button/input/select/textarea chỉ nằm trong atoms. Text/heading, card, badge, progress và feedback phải dùng base tương ứng.
4. Không override màu, font, shape, border, shadow của base qua className. className chỉ dành cho layout của consumer (width, margin, alignment, placement). Không tạo wrapper base rỗng để chuyển nguyên CSS tự custom xuống một tầng.
5. Primitive màu/font/radius/shadow thuộc app/src/ui/theme.css; variants thuộc app/src/ui/variants.ts; mapping nghiệp vụ thuộc domainVariants.ts/categoryPresentation.ts.
6. Organism sở hữu fetching, state, điều hướng và composition. Molecule sở hữu bố cục tái sử dụng. Atom sở hữu hình thức và hành vi control.
7. Build bắt buộc qua npm run check:design. Thêm base phải thêm test cho behavior; không sửa guardrail để miễn trừ màn riêng.

## Chỉ giữ đặc tả

Ảnh PNG và HTML export đã được đọc trực quan/OCR, đối chiếu và cô đọng tại [SOURCE_EXTRACTION](system/SOURCE_EXTRACTION.md). Theo yêu cầu owner ngày 2026-09-13, docs/design giữ Markdown specification; không yêu cầu duy trì code.html/screen.png nữa. Bản gốc tra được trong Git ở commit 1bc013d. Browser fixture kiểm thử nằm trong app/tests, không phải một design source khác.

Khi implement đến màn nào, thêm/cập nhật screens/<screen>/README.md của màn đó: route, composition, states, event → effect, validation, quyền thao tác, copy, API dependency, proof và known gaps. Không suy đoán behavior chưa thấy trong ảnh thành quyết định sản phẩm. Xem [screen template](screens/README.md).

## Authority và verification

Quyết định owner hiện tại → product requirements cho nghiệp vụ → design contract chuẩn hóa cho UI → component implementation. Code hiện tại không tự ghi đè thiết kế; lệch contract phải sửa hoặc ghi quyết định được duyệt.
[ADR](../decisions/ADR-002-design-contract-enforcement.md) · [Work item](../work/tickets/UI-BASE-01-DETAIL_DESIGN.md) · [Validation](../work/VALIDATION_MATRIX.md).
