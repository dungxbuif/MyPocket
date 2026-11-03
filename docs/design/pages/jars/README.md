# Quản lý hũ

## Thiết kế đích 2026-11-03 — hũ hiển thị như ví (planned)

Nguồn: [detail design](../../../work/tickets/JARS-AS-WALLET-CARDS-20261103-DETAIL_DESIGN.md), [ADR-029](../../../decisions/ADR-029-jar-cumulative-plan-presentation.md). Trạng thái: `ready`, chưa implement; phần "Composition" lịch sử bên dưới mô tả bản đang chạy.

### Màn Hũ

| Vùng | Base | Contract |
| --- | --- | --- |
| Header | `MPSelectionScreenHeader` | "Hũ chi tiêu"; Đóng trái, Sửa/Xong phải như màn Ví |
| Dòng đối chiếu | `MPStatusMessage` info/danger | "Kế hoạch còn lại của các hũ: X · Tiền trong ví: Y"; X > Y → danger "Kế hoạch vượt tiền thực có Z". Y = Σ số dư ví `is_in_total` |
| Chọn tháng | `MPOptionSelector` | `YYYY-MM` theo timezone tài khoản; cộng dồn tính đến tháng chọn |
| Dòng hũ | `MPWalletSelectionList` + `MPWalletRow(.jar)` + `MPProgressBar` + `MPStatusBadge` | Số chính "Còn có thể chi" (cộng dồn); dòng phụ "Tháng này chi / mức"; tiến độ chi/mức tháng; badge "Vượt mức"; hũ không mức: số chính "Đã chi cộng dồn", phụ "Chưa đặt mức" |
| Chưa phân bổ | `MPWalletRow` glyph trung tính | Chính: chi chưa gắn hũ tháng này; phụ "Cộng dồn: …"; chạm mở sổ giao dịch lọc `Chưa gắn hũ`; luôn hiện |
| Hành động | `MPButton` | Thêm hũ luôn hiện; Sửa: chạm dòng mở editor; gỡ khỏi tháng có xác nhận trong editor |
| Editor | `MPEditorScaffold`, `MPAmountField`, `MPNumberField(.percentage)` | Như hiện tại + dòng xem trước "Còn có thể chi sau khi lưu" |
| Trạng thái | `MPLoadingPresentationPolicy`, `MPEmptyView`, `MPErrorView` | Vẽ từ cache đọc-trước, tải nền; rỗng vẫn hiện Chưa phân bổ và Thêm hũ |

Khối "Cộng dồn" chọn hũ ở cuối màn hiện tại bị bỏ; dữ liệu chuyển lên thẻ và màn chi tiết.

### Màn chi tiết hũ (mới)

| Vùng | Base | Contract |
| --- | --- | --- |
| Header | `MPCenteredScreenHeader` + `MPOverflowMenu` | Tên hũ; Sửa mức tháng · Gỡ khỏi tháng · Hỏi Trợ lý về hũ này |
| Ba số | `MPMetricGrid` ×3 (masked) | Còn có thể chi · Mức tháng này · Đã chi tháng này |
| Chú thích | `MPText(.secondary)` | "Kế hoạch tháng này: 15.000.000 đ · Dư từ tháng trước: +5.000.000 đ · Còn có thể chi theo kế hoạch: 20.000.000 đ"; tháng không mức: "Tháng này chưa đặt mức; chi vẫn trừ vào kế hoạch còn lại" |
| Diễn biến | `MPFinanceCard` + `MPGroupedRow` | Mỗi tháng: mức, chi, dư, luỹ kế; "Mức có ở N/M tháng" |
| Giao dịch | `MPSurfaceCard(.grouped)` + `MPTransactionRow` | Tháng đang xem; "Xem tất cả" mở sổ lọc theo hũ; chạm dòng mở editor |
| Trợ lý | `MPButton(.secondary)` "Phân loại bằng Trợ lý" | Mở Assistant toàn màn với `jar_context` |

### Copy

"Còn có thể chi", "Mức tháng này", "Đã chi tháng này", "Dư từ tháng trước", "Chưa phân bổ", "Kế hoạch vượt tiền thực có". Cấm: "số dư hũ".

### API

`GET /api/v1/jars?month=` (thêm `remaining_plan`, `cumulative_*`, `allocation_months`, `months_in_range`, summary `remaining_plan_total`, `unassigned_cumulative`, `wallets_in_total_balance`, `plan_exceeds_wallets_by`), `GET /api/v1/transactions?jar_id=<id>|unassigned`, `GET /api/v1/jars/{id}/report` (thêm `remaining_plan` luỹ kế).


## Task 4 ownership boundary

The native implementation remains in
`Features/Secondary/SecondaryFeatureViews.swift`. This organization slice owns
that composition, its focused tests, and this contract; jar headers, rows,
spacing, or colors must not be duplicated in another file.

Native override (2026-10-14): Tài khoản → Hũ chi tiêu. Add/edit replaces the list
with `MPEditorScaffold`, one scroll and pinned Save. `MPAmountField` handles fixed
money and `MPNumberField(.percentage)` handles percentages. Pending disables
repeated saves/cancel; failures retain the draft. Historical web presentation
below is not the native UI authority. Hủy does not save; reopening the same jar
within the same month/screen resumes the input. Changing month clears target and
draft, even when returning to the original month. Leaving/restarting does not
guarantee draft persistence. The owned unit contract verifies mounted
same-month reuse and rehydration after the month scope changes; it does not
claim process-durable persistence.

Native percentage allocation uses `MPNumberField(.percentage)` and shared
comma→dot normalization, preserving raw invalid drafts. Generic input range is
0…100; jar validation still requires positive allocation with at most2 decimal
places. Fixed allocation remains a monetary field. No API type/schema changes.
[Number input guide](../../../guides/numeric-input-ios.md).

The current native route remains the first-class Hũ entry from Overview,
Tài khoản and Tháng, with month scope, API-backed spend/allocation, cumulative
history, shared transaction policy and explicit remove confirmation. No backend
or jar semantics were changed in the UI-density integration pass.

Route `/jars`; implements [TICKET-04](../../../work/tickets/TICKET-04-hu-chi-tieu.md) and [CORE-03](../../../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md). Status: implementation in progress.

## Composition

| Region | Base | Behavior |
| --- | --- | --- |
| Header/month choice | `PageBackHeader`, `BaseSelect`, `FormField`, `Text` | Back returns to the prior app route; month value is `YYYY-MM` in the saved account timezone. |
| Monthly figures | `SurfaceCard`, `Text`, `Progress`, `StatusMessage` | Displays actual income, ordinary expense, allocated total, unassigned spend, over-allocation and overspend advisories from the API. |
| Jar rows | `SurfaceCard`, `Text`, `Progress`, `BaseButton` | Each saved monthly config has its own name/allocation and API-derived actual spend. Inactive historical configs with transactions remain readable, not editable. |
| Add/edit form | `BaseBottomSheet`, `FormField`, `BaseTextInput`, `BaseSelect`, `BaseButton`, `StatusMessage` | User selects name, allocation mode (`Không đặt`, fixed VND, percent of actual included income), and the corresponding amount. |
| Cumulative view | `SurfaceCard`, `BaseSelect`, `Text`, `Progress`, `StatusMessage` | Select a stable jar from the owner's full jar catalog and view all-history totals/month breakdown; only configured months contribute allocation/difference. No carryover balance is implied. |

All styling, controls, typography, progress and cards are owned by the named bases. Screen classes are layout-only. No screen-local chart, currency palette, or clickable `div` is allowed.

## Events and effects

| Event | Effect |
| --- | --- |
| Enter `/jars` | Fetch selected month summary from the authenticated API; the server initializes that month once from the nearest earlier initialized configuration. |
| Change month | Fetch that month's actual config and ledger-derived spend. Historical months are viewable. |
| Add jar | POST one stable jar identity and the selected month's configuration; refresh month and cumulative options. |
| Edit monthly config | PUT the current month only; prior/future snapshots and the stable identity are not renamed. |
| Remove from month | Confirm, then deactivate only this month's config; keep stable ID, transaction links, and other snapshots. |
| Select a jar for cumulative | Fetch the stable-identity aggregate through the selected month, with all-history default range. |
| Save/API failure | Keep the draft, show retryable error, and do not optimistically change totals. |

## States and validation

- Loading/error: explicit status; no sample jars or fabricated zero allocations.
- Empty month: explain there are no jars configured and provide `Thêm hũ`; unassigned expenses remain valid and visible in summary.
- Ready: display real config and spend; warning states are advisory and never block ledger saves.
- Saving: disable duplicate actions and preserve draft until success.
- Allocation is optional. Fixed VND and percent are non-negative/positive according to API validation; percentages are bounded by backend rules. No allocation is represented as `Không đặt`.
- Only ordinary expense transactions can be assigned; income, transfer-out, and out-of-scope rows do not count as jar spend.
- Removing a monthly config never deletes a jar or linked ledger rows.

## Copy and API

Use `Hũ tháng`, `Chi chưa gắn hũ`, `Phân bổ`, `Chi thực tế`, `Cộng dồn`, `Tháng này`, and advisory copy such as `Tổng phân bổ vượt thu thực tế` / `Chi vượt mức hũ`. Never call cumulative difference “số dư hũ”. Money is VND; percentage is based on actual report-included ordinary income.

APIs: `/api/v1/jars?month=YYYY-MM`, `POST /api/v1/jars`, `PUT/DELETE /api/v1/jars/{jarID}/months/{YYYY-MM}`, and `/api/v1/jars/{jarID}/report?to=YYYY-MM` (optional `from` narrows the range; omission means all history). All access is owner-scoped.

## Proof and gaps

Implementation update: `/jars` now uses the authenticated APIs for month totals, create/edit/remove, prior-month configuration, stable-identity cumulative totals and month breakdown. Empty collections render as real empty states. Manual and AI transaction editors load choices for the transaction's account-local month and submit the optional `jar_id`.

Automated proof: PostgreSQL tests cover timezone rebucketing, empty-array response contracts, concurrent one-time copy of active config, and owner isolation; UI tests cover expense-only selection, hidden income/transfer-out controls, and account-local month derivation. Browser UAT must still verify create/edit/remove, month copy, manual and AI assignment, and cumulative history with owner-approved test data. No mock data, carryover, notifications, or AI inference is in scope.

## Organization reference composition — 2026-10-26

The monthly jar list is one shared grouped surface with `MPGroupedRow` and
`MPInsetDivider`. Monthly totals and cumulative history use `MPFinanceCard`;
section spacing comes from `JarsReferencePresentation` rather than local
literals. Add/edit remains the existing `MPEditorScaffold` flow.

No allocation, month-copy, stable-jar identity, removal, report, masking, or API
behavior changes. Shared metrics now provide row geometry. Focused unit tests
prove through the editor-session contract that cancel/reopen retains a draft
only for the same jar and month, while a month change requires rehydration. The
session is consumed by `JarsView`; this is mounted in-memory editor state,
not persistence across navigation, relaunch, or process death. The rendered
grouped hierarchy and accessibility key still require later UI acceptance.

Review remediation evidence (2026-10-27): the four owned reference test classes
passed 24/24 and the normally signed `MyPocketIOSTests` target passed 422/422 on
the iPhone 16 iOS 18.5 simulator. The Release warnings-as-errors build,
base-component guard, and project verification also passed.
