---
artifact_type: detail_design
id: JARS-AS-WALLET-CARDS-20261103
status: in_progress
owner: shared
human_fields:
  - approval
  - decisions
ai_fields:
  - analysis
  - screen_specs
  - verification_plan
backlog: ../BACKLOG.md
requirements: ../../requirements/BUSINESS_RULES.md#jar--hũ-theo-tháng-và-cộng-dồn
adr: ../../decisions/ADR-029-jar-cumulative-plan-presentation.md
screens: ../../design/pages/jars/README.md
---

# Hũ hiển thị như ví: kế hoạch còn lại cộng dồn, chưa phân bổ, và AI gán hũ từ ghi chú

## Quyết định của owner (2026-11-03)

1. Gộp khái niệm "hũ" và "quỹ" làm một, giữ tên **Hũ**. Hũ vẫn là lựa chọn tuỳ chọn trên một khoản chi; khoản chi vẫn thuộc ví. Hũ là một bộ lọc/góc nhìn khác của chi tiêu.
2. Màn Hũ hiển thị các hũ **như danh sách ví**: mỗi thẻ hũ cho thấy toàn bộ số tiền liên quan. Giao dịch chưa gắn hũ nằm ở thẻ **Chưa phân bổ**.
3. Giữ **mức ngân sách theo tháng** của hũ. Sang tháng sau, phần chưa tiêu hết **vẫn thấy được** dưới dạng kế hoạch còn lại cộng dồn.
4. Thẻ "Chưa phân bổ" hiện số **tháng này** là chính, cộng dồn là phụ.
5. "Còn có thể chi" tính từ **tháng đầu có cấu hình** của hũ. Chưa cho chọn mốc bắt đầu.
6. Dòng đối chiếu so với **tổng các ví tính vào tổng** (`is_in_total = true`, bỏ ví tín dụng).
7. Người dùng có thể **trò chuyện với Trợ lý để phân loại giao dịch vào hũ** dựa trên ghi chú và lịch sử; Trợ lý đề xuất, người dùng xác nhận.

## Điều không đổi (bảo vệ sự ổn định)

- Không đổi schema, không có migration, không sửa giao dịch cũ.
- Hũ không giữ tiền. Số trên thẻ hũ là **kế hoạch**, không phải số dư ví. Không bao giờ dùng chữ "số dư hũ".
- Thu nhập không gắn hũ ở lát này. Mức theo % tổng thu thực tế tháng giữ nguyên (JAR-04).
- Vượt mức tháng, vượt thu, tổng % > 100 chỉ cảnh báo, không chặn lưu (JAR-05).
- Tổng chi theo hũ + Chưa phân bổ = tổng chi thường của tháng (JAR-09).

## Ngữ nghĩa và công thức

```text
Mức tháng m            = phân bổ của hũ ở tháng m (cố định hoặc % × thu thực tế tháng m); không có khi tháng không đặt mức
Chi tháng m            = tổng chi thường hợp lệ gắn hũ trong tháng m
Dư tháng m             = Mức tháng m − Chi tháng m          (chỉ khi có mức)
Còn có thể chi (đến m) = Σ_{k = tháng đầu có cấu hình .. m} (Mức k − Chi k) với các tháng có mức
                         − Σ Chi k của các tháng không có mức   (chi không có mức vẫn trừ vào kế hoạch)
Chưa phân bổ tháng m   = chi thường hợp lệ của tháng m không gắn hũ
Đối chiếu              = Σ Còn có thể chi của các hũ đang hoạt động  so với  Σ số dư ví có is_in_total = true
```

Ví dụ owner: mức 15.000.000 đ/tháng. Tháng 10 chi 10.000.000 đ → Dư +5.000.000 đ. Tháng 11 mức vẫn 15.000.000 đ, thẻ hũ ghi **Còn có thể chi 20.000.000 đ** và dòng chú thích:

> Kế hoạch tháng này: 15.000.000 đ · Dư từ tháng trước: +5.000.000 đ · Còn có thể chi theo kế hoạch: 20.000.000 đ

Nếu tháng 11 chi 17.000.000 đ: thẻ ghi "Còn 3.000.000 đ", mức tháng vẫn 15.000.000 đ và cảnh báo "Chi vượt mức tháng 2.000.000 đ". Hai con số sống song song: mức tháng để giữ kỷ luật, còn có thể chi để biết mình đang ở đâu.

Nếu Σ Còn có thể chi = 20.000.000 đ mà ví tính vào tổng chỉ còn 12.000.000 đ: dòng đối chiếu ghi "Kế hoạch vượt tiền thực có 8.000.000 đ". Đây là điều kiện bắt buộc để hũ được trình bày giống ví mà không gây hiểu nhầm.

## Đặc tả từng màn

Bases theo [BASE_COMPONENTS](../../design/system/BASE_COMPONENTS.md). Màn chỉ làm layout, state, gọi API và copy nghiệp vụ; không tạo màu/bo góc/control riêng.

### A. Màn Hũ (Tài khoản → Hũ chi tiêu) — làm lại

| Vùng | Base | Hành vi |
| --- | --- | --- |
| Header | `MPSelectionScreenHeader` | Tựa "Hũ chi tiêu"; trái Đóng; phải Sửa/Xong như màn Ví |
| Dòng đối chiếu | `MPStatusMessage` (info/danger) | "Kế hoạch còn lại của các hũ: X · Tiền trong ví: Y". Khi X > Y: tone danger "Kế hoạch vượt tiền thực có Z". Luôn hiện khi có ít nhất một hũ có mức |
| Chọn tháng | `MPOptionSelector` hiện có | Tháng `YYYY-MM` theo timezone tài khoản; đổi tháng tải lại toàn màn; cộng dồn tính đến tháng đã chọn |
| Danh sách hũ | `MPWalletSelectionList` + `MPWalletRow` (biến thể `.jar`) + `MPProgressBar` + `MPStatusBadge` | Mỗi hũ một dòng: tên; số chính **Còn có thể chi** (cộng dồn); dòng phụ "Tháng này 10.000.000 / 15.000.000 đ"; thanh tiến độ chi/mức tháng; badge "Vượt mức" khi chi > mức; hũ không đặt mức: số chính là "Đã chi cộng dồn" và dòng phụ "Chưa đặt mức" (JAR-08). Chạm mở màn chi tiết hũ |
| Thẻ Chưa phân bổ | cùng `MPWalletRow`, glyph trung tính | Số chính: chi chưa gắn hũ **tháng này**; dòng phụ: "Cộng dồn: …". Chạm mở sổ giao dịch với bộ lọc `Chưa gắn hũ` của tháng đó. Luôn hiện, kể cả 0 |
| Hành động | `MPButton` | "Thêm hũ" luôn hiện; chế độ Sửa chạm dòng mở editor; gỡ khỏi tháng nằm trong editor có xác nhận |
| Editor | `MPEditorScaffold`, `MPAmountField`, `MPNumberField(.percentage)` | Giữ nguyên luồng hiện có; thêm dòng xem trước "Còn có thể chi sau khi lưu: …" |
| Trạng thái | `MPLoadingPresentationPolicy`, `MPEmptyView`, `MPErrorView` | Vẽ từ cache đọc-trước rồi tải nền (local-first). Rỗng: "Chưa có hũ tháng này" + Thêm hũ; Chưa phân bổ vẫn hiện |

Bỏ khối "Cộng dồn" chọn hũ ở cuối màn hiện tại: dữ liệu đó chuyển lên thẻ và màn chi tiết.

### B. Màn chi tiết hũ — mới

| Vùng | Base | Hành vi |
| --- | --- | --- |
| Header | `MPCenteredScreenHeader` + `MPOverflowMenu` | Tên hũ; menu: Sửa mức tháng, Gỡ khỏi tháng, Hỏi Trợ lý về hũ này |
| Ba số | `MPMetricGrid` (3 `MPMetricItem`, hỗ trợ masked) | Còn có thể chi · Mức tháng này · Đã chi tháng này |
| Chú thích | `MPText(.secondary)` | Đúng câu ở mục Ngữ nghĩa: kế hoạch tháng · dư từ tháng trước · còn có thể chi. Khi tháng không đặt mức: "Tháng này chưa đặt mức; chi vẫn trừ vào kế hoạch còn lại" |
| Diễn biến tháng | `MPFinanceCard` + `MPGroupedRow` | Mỗi tháng: mức, chi, dư, luỹ kế; ghi "Mức có ở N/M tháng" (JAR-10) |
| Giao dịch của hũ | `MPTransactionRow` trong `MPSurfaceCard(.grouped)` | Mặc định tháng đang xem; nút "Xem tất cả" mở sổ giao dịch với bộ lọc hũ này; chạm dòng mở editor giao dịch |
| Chuyển sang Trợ lý | `MPButton(.secondary)` "Phân loại bằng Trợ lý" | Mở Assistant toàn màn với ngữ cảnh hũ (xem mục AI) |

### C. Sổ giao dịch

| Thay đổi | Base | Hành vi |
| --- | --- | --- |
| Bộ lọc hũ | `MPWalletScopeSelector`-style capsule thứ hai, hoặc mở rộng `MPOptionSelector` | Tuỳ chọn: Tất cả · từng hũ của tháng · Chưa gắn hũ. Kết hợp được với bộ lọc ví và kỳ. Cần API `GET /transactions?jar_id=<id>` và `jar_id=unassigned` |
| Chip hũ trên dòng | `MPStatusBadge(.muted)` trong `MPTransactionRow` | Chỉ với khoản chi có hũ; không chip cho thu/chuyển ví |
| Ô chọn hũ trong form | `MPOptionSelector` hiện có + `MPText(.secondary)` | Dưới ô chọn, hiện "Hũ Ăn uống còn có thể chi 20.000.000 đ" (hoặc "sẽ vượt mức tháng 2.000.000 đ" tone danger). Chỉ thông tin, không chặn. Dữ liệu từ `GET /jars?month=` đã tải sẵn cho tháng của giao dịch |

### D. Tổng quan

Thẻ tháng thêm hai dòng trong `MPGroupedRow`: "Hũ: còn có thể chi tổng X" và "Chưa phân bổ tháng này: Y". Nút "Xem hũ tháng" giữ nguyên. Dữ liệu từ cùng `GET /jars?month=`.

### E. Tháng (month detail)

Phần "bối cảnh hũ" dùng cùng dòng thẻ như màn Hũ nhưng chỉ đọc; cộng dồn tính **đến tháng đang xem**, không phải tháng hiện tại. Chạm thẻ mở màn chi tiết hũ ở tháng đó.

### F. Báo cáo

"Hũ cộng dồn" giữ dữ liệu, đổi nhãn "Chênh lệch" → "Kế hoạch còn lại"; thêm dòng "Chưa phân bổ" để tổng chi theo hũ luôn khớp tổng chi tháng. Không thêm biểu đồ mới ở lát này.

### G. Trợ lý AI — phân loại giao dịch vào hũ từ ghi chú

Hiện trạng đã kiểm: công cụ `search_transactions` trả về `note` và có bộ lọc `note_contains`, nên Trợ lý **đọc được ghi chú** khi tìm trực tiếp. Điều thiếu là: (1) không có công cụ nào để gán hũ cho giao dịch, chỉ có đọc tiến độ hũ và đề xuất tạo/sửa hũ; (2) ngữ cảnh báo cáo gửi sang Trợ lý đã bị rút gọn tên/ghi chú; (3) không có hướng dẫn trong prompt để coi ghi chú là bằng chứng mục đích. Vì vậy người dùng thấy Trợ lý "chỉ đọc mã giao dịch".

Thiết kế:

| Thành phần | Nội dung |
| --- | --- |
| Công cụ mới `propose_assign_jar` | Vào: `month`, danh sách `{transaction_id, jar_id, reason}`; `jar_id` rỗng = bỏ gán. Ra: thẻ xác nhận `jar_assignment_confirmation` liệt kê từng dòng: ghi chú, số tiền, ngày, danh mục, hũ đề xuất và lý do (trích ghi chú). Tối đa 50 dòng một thẻ. Chỉ khoản chi thường; thu/chuyển ví/tín dụng bị loại ngay ở server với lý do |
| Xác nhận | Người dùng bỏ từng dòng hoặc đổi hũ ngay trên thẻ, rồi "Áp dụng". Server áp bằng `PATCH /transactions/{id}` (`jar_id`) từng dòng với `expected_version`; dòng nào xung đột thì báo và giữ các dòng khác. Không ghi trực tiếp từ công cụ (đúng quy tắc AI không tự ghi) |
| Prompt | Thêm vào handbook: "Mục đích của một khoản chi nằm ở ghi chú và tên danh mục, không phải ở mã. Khi phân loại hũ, trích dẫn ghi chú làm lý do; nếu ghi chú trống thì dựa vào danh mục và các giao dịch tương tự đã có hũ; không đoán khi không có bằng chứng — xếp vào 'cần hỏi lại'." |
| Ngữ cảnh hũ | Nút "Phân loại bằng Trợ lý" gửi `jar_context` (id hũ, tháng) giống `report_context` hiện có: server nạp lại bằng ID, kèm danh sách khoản chi chưa gắn hũ của tháng (có ghi chú). Câu hỏi mẫu điền sẵn: "Gợi ý hũ cho các khoản chi chưa phân bổ tháng này" |
| Học từ lịch sử | Trước khi đề xuất, Trợ lý dùng `search_transactions` với `note_contains` theo từ khoá trong ghi chú để xem các khoản tương tự đã gắn hũ nào; lý do ghi "3 khoản 'Grab đi làm' trước đây gắn hũ Di chuyển" |
| Ghi chú có hashtag | Nếu ghi chú chứa `#tên-hũ` khớp tên hũ (bỏ dấu, chữ thường) thì đề xuất với độ tin cậy cao. Không tự tạo hũ từ hashtag; thiếu hũ thì dùng `propose_create_jar` hiện có |
| Giới hạn | Trợ lý không chuyển tiền, không đổi mức hũ khi phân loại, không gán hũ cho thu nhập. Tất cả thay đổi đi qua thẻ xác nhận |

Thẻ xác nhận mới tái dùng `MPAssistantCard` + `MPTransactionRow` + `MPOptionSelector` (đổi hũ từng dòng) + `MPButton` "Áp dụng"/"Bỏ qua"; trạng thái từng dòng bằng `MPStatusBadge`.

## Thay đổi API (chỉ thêm, không đổi hợp đồng cũ)

| Endpoint | Thêm |
| --- | --- |
| `GET /jars?month=` | Mỗi item: `cumulative_allocated`, `cumulative_spent`, `remaining_plan` (= còn có thể chi), `allocation_months`, `months_in_range`, `first_config_month`, `month_variance`; summary: `remaining_plan_total`, `unassigned_cumulative`, `wallets_in_total_balance`, `plan_exceeds_wallets_by` |
| `GET /transactions` | Bộ lọc `jar_id=<id>` và `jar_id=unassigned` (chỉ áp lên chi thường) |
| `GET /jars/{id}/report` | Thêm `remaining_plan` luỹ kế theo tháng trong `months[]`; đổi nhãn tài liệu "variance" → kế hoạch còn lại |
| Assistant | Công cụ `propose_assign_jar`, phần `jar_assignment_confirmation`, đầu vào `jar_context` cho `POST /ai/advisor/messages(/async)`; áp dụng qua `POST /ai/advisor/actions/{id}/confirm` hiện có |

Tất cả owner-scoped như hiện nay. Không có migration.

## Quy tắc nghiệp vụ cần ghi thêm (BUSINESS_RULES)

- JAR-12: "Còn có thể chi" là kế hoạch còn lại cộng dồn từ tháng đầu có cấu hình đến tháng đang xem; chi ở tháng không đặt mức vẫn trừ; được hiển thị trên thẻ hũ nhưng không phải tiền trong ví và không cấp thêm mức cho tháng mới (làm rõ JAR-07/JAR-11, không thay thế).
- JAR-13: Màn hũ phải hiển thị dòng đối chiếu tổng kế hoạch còn lại với tổng số dư các ví tính vào tổng; vượt thì cảnh báo.
- JAR-14: "Chưa phân bổ" là chi thường hợp lệ không gắn hũ, hiển thị theo tháng và cộng dồn; luôn có mặt để tổng theo hũ khớp tổng chi.
- AI-JAR-01: Trợ lý chỉ đề xuất gán hũ qua thẻ xác nhận; lý do phải dẫn từ ghi chú, danh mục hoặc giao dịch tương tự; không gán cho thu nhập/chuyển ví/tín dụng.

## Tiến độ thực hiện

- iOS Assistant: thẻ `jar_assignment_confirmation` + `jar_context` handoff từ chi tiết hũ đã làm (475 unit test xanh, guard xanh). Chưa có ảnh chụp thẻ trên runtime thật (cần mô hình).

- Giai đoạn 0 xong: ba commit (backend, iOS local-first, docs).
- Giai đoạn 1 — **backend xong**: `GET /jars?month=` có các trường cộng dồn và tổng đối chiếu; `GET /transactions?jar_id=<id>|unassigned`; test repository cho ví dụ 15/10/20, vượt mức, đối chiếu ví, cô lập tháng. Số liệu thật trên fixture đã đọc lại.
- Giai đoạn 1 — **iOS đã code, chưa nghiệm thu bằng mắt**: model có trường cộng dồn (optional, payload cũ vẫn decode), base mới `MPJarPlanRow`, `JarPlanPresentation` (tiêu đề/số chính/chú thích/đối chiếu; âm hiển thị "Vượt kế hoạch"), màn Hũ dạng ví + thẻ Chưa phân bổ + dòng đối chiếu, màn chi tiết hũ. Unit test 465 qua; chưa có UI test/ảnh chụp vì tài khoản UI-test không có hũ.
- Còn lại giai đoạn 1: UI test với hũ tạo qua API, đối chiếu ảnh chụp; thay `MPJarPlanRow` bằng biến thể của `MPWalletRow` nếu owner muốn giống hệt ví.
- Giai đoạn 3 — **backend xong (chưa nghiệm thu bằng mắt/iOS)**: công cụ `propose_assign_jar`, phần `jar_assignment_confirmation`, áp dụng từng dòng với `expected_version` qua `assign_jar`, `jar_context`, prompt/handbook, `jar_id`/`jar_name` trong `search_transactions`. Test: validation, xung đột một dòng, confirm lặp, owner scope `jar_context`, hash. Thẻ iOS và đo tỷ lệ đúng trên 20 khoản: xem VALIDATION_MATRIX.
- Giai đoạn 2 chưa làm.

## Giai đoạn và thứ tự

| Giai đoạn | Phạm vi | Bằng chứng kết thúc |
| --- | --- | --- |
| 0 | Nghiệm thu và commit local-first + các sửa lỗi đang dở | Owner chấp nhận |
| 1 | API `GET /jars?month=` mở rộng + bộ lọc `jar_id`; màn Hũ dạng ví, thẻ Chưa phân bổ, dòng đối chiếu; màn chi tiết hũ | Ví dụ 15/10/20 đúng trên simulator với fixture; test repository cộng dồn; UI test đổi tháng |
| 2 | Sổ giao dịch: bộ lọc hũ, chip, dòng "còn có thể chi" trong form; Tổng quan; Tháng; Báo cáo đổi nhãn | Unit + UI test bộ lọc; tổng theo hũ + chưa phân bổ = tổng chi |
| 3 | Trợ lý: `propose_assign_jar`, thẻ xác nhận, `jar_context`, prompt | Chạy thật với fixture: ≥ 80% đề xuất đúng trên 20 khoản có ghi chú rõ; 0 ghi trực tiếp; xung đột version được báo |

## Kế hoạch kiểm chứng

- Backend: test cộng dồn với tháng không mức, tháng đầu, đổi timezone; test bộ lọc `jar_id`/`unassigned`; test đối chiếu chỉ cộng ví `is_in_total`; test `propose_assign_jar` loại thu/chuyển ví, giới hạn 50, áp dụng với `expected_version` và xung đột từng dòng.
- Native: unit test presentation (số chính/phụ, badge, chú thích); UI test màn Hũ với fixture tạo hũ 15 triệu, chi 10 triệu, đổi sang tháng sau thấy 20 triệu; guard base component; UI test bộ lọc sổ giao dịch.
- Trợ lý: bộ 20 khoản chi có ghi chú tiếng Việt (có/không hashtag), đo tỷ lệ đề xuất đúng và số dòng "cần hỏi lại".
- Local-first: màn Hũ và chi tiết hũ vẽ từ cache đọc-trước, không loading chặn khi quay lại.

## Điểm mở

- Có cần "reset mốc cộng dồn" (ví dụ từ đầu năm) không: để sau, theo quyết định 5.
- Thu nhập gán hũ (mô hình phong bì đầy đủ): chỉ xét sau khi owner dùng giai đoạn 1–2 vài tuần.
- Web (`app/`) là lưu trữ, không cập nhật trong lát này.

## Trace

[Backlog](../BACKLOG.md) · [ADR-029](../../decisions/ADR-029-jar-cumulative-plan-presentation.md) · [Spec màn Hũ](../../design/pages/jars/README.md) · [Sổ giao dịch](../../design/pages/transactions/README.md) · [Tổng quan](../../design/pages/overview/README.md) · [Tháng](../../design/pages/month-detail/README.md) · [Báo cáo](../../design/pages/reports/README.md) · [Trợ lý](../../design/pages/finance-assistant/README.md) · [Quy tắc](../../requirements/BUSINESS_RULES.md) · [API](../../architecture/API.md) · [Validation](../VALIDATION_MATRIX.md) · [Context](../../CONTEXT.md)
- Giai đoạn 2 — **iOS đã code và nghiệm thu bằng ảnh cho màn Hũ và chi tiết hũ**: sổ giao dịch có bộ lọc hũ (Tất cả · hũ · Chưa gắn hũ), chip tên hũ trên dòng chi, dòng gợi ý "còn có thể chi / sẽ vượt mức" dưới ô chọn hũ (chỉ thông tin); Tổng quan có hai dòng hũ (cache-first, tải nền); màn Tháng dùng từ "Còn có thể chi" và dòng "Chưa phân bổ"; Báo cáo iOS không có khối hũ nên không đổi. Dọn UX: ẩn nút "+" ở chi tiết giao dịch, Việt hoá "Chế độ du lịch"/"Nhật ký thay đổi", "Xem tất cả N ví", đổi tiêu đề nhóm "Ví của tôi" thành "Tài sản", đồng nhất ô tìm danh mục. UI test `testJarsScreenShowsCumulativePlanLikeWalletCards` qua. Chưa chụp ảnh riêng: bộ lọc hũ, gợi ý trong editor, thẻ tháng Tổng quan, màn Tháng.
- Trợ lý gắn hũ — bộ gợi ý tất định (2026-11-03): `JarSuggester` (hashtag `#ten-hu`, ghi chú lặp lại trong lịch sử cùng một hũ) chạy trong `jar_context`, trả `suggested_assignments` và `needs_clarification`; server tự dựng thẻ `jar_assignment_confirmation` khi mô hình không tạo được thẻ; chạm giới hạn vòng công cụ trả kết quả đã gom hoặc thông báo rõ. Đo trực tiếp: xem "Live Assistant jar-assignment check" trong VALIDATION_MATRIX.

- Tiếp tục cùng một hũ sang tháng sau bằng tay (2026-11-04): `PUT /jars/{jar_id}/months/{month}` nay là upsert theo chủ sở hữu (tạo cấu hình tháng chưa có, kích hoạt lại tháng đã gỡ, giữ nguyên `jar_id`; hũ lạ/không tồn tại vẫn 404). Không thêm route mới. Test: `TestUpdateMonthConfigIsOwnerScopedUpsert`.
