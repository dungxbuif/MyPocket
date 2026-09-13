Dựa trên việc phân tích các luồng Money Lover và đối chiếu implementation hiện tại, hệ thống giao diện được tổ chức thành 7 nhóm base component. Danh sách 28 thành phần dưới đây là inventory thiết kế; không được hiểu là tất cả đã hoàn thiện trong code. Cấu trúc thư mục và quy tắc đồng bộ hai chiều được ghi tại [Design Source README](./README.md); product-level rules nằm ở [Financial Clarity System](./system/DESIGN.md).

1. Điều hướng & Cấu trúc khung (Navigation & Layout) — 4 components
   Top Bar / Screen Header:
   Dạng chính: Tiêu đề trang + nút điều hướng quay lại (Quay lại / icon mũi tên) + icon hành động (Tìm kiếm, Cài đặt, Lightbulb).
   Dạng Modal/Sheet: Nút Huỷ / Đóng góc trái, tiêu đề ở giữa, nút hành động phụ (như Sửa) góc phải.
   Bottom Navigation Bar: Thanh 4 tab (Tổng quan, Sổ giao dịch, Ngân sách, Tài khoản) tích hợp nút Floating Action Button (+) nổi bật ở chính giữa.
   Segmented Control / Tab Switcher:
   Dạng pill viền bo tròn 2 lựa chọn (Tuần / Tháng).
   Dạng 3 lựa chọn cho luồng giao dịch (Khoản chi / Khoản thu / Vay/Nợ).
   Sub-header Bar / Scope Selector: Thanh chọn phạm vi tài khoản/ví dạng pill thả xuống (Tổng cộng ⌄, bộ lọc tháng 08/2026 ⌄).
2. Thành phần nhập liệu & Điều khiển (Form Controls & Inputs) — 6 components
   Amount Display Input: Khung hiển thị số tiền lớn, gắn kèm badge đơn vị tiền tệ (VND / đ) và nút xoá nhanh (icon ×).
   Selector Row Item (Form Cell): Hàng nhập liệu chạm mở chọn danh mục/ví/sự kiện, gồm: Icon dẫn đầu + Nhãn placeholder/giá trị + Mũi tên điều hướng (>).
   Date Picker Selector Row: Hàng hiển thị ngày tháng tích hợp nút mũi tên chuyển nhanh ngày trước/sau (< Chủ Nhật, 23/08/2026 >).
   Switch Row (Toggle Cell): Khung cài đặt bật/tắt (Toggle switch) kèm tiêu đề và chú thích phụ (ví dụ: Lặp lại ngân sách này, Không tính vào báo cáo).
   Primary Action Button (CTA): Nút bo góc tròn lớn full-width cố định ở đáy màn hình (Lưu, Đăng ký ngay, Đặt ngân sách).
   Custom Numeric Keypad (Bàn phím số & máy tính): Bàn phím số tích hợp phím tính toán (+, -, ×, ÷), các nút gợi ý tiền nhanh (50.000, 150.000, 500.000) và nút XONG.
3. Hiển thị danh sách & Hạng mục (List Items & Category Cells) — 4 components
   Wallet List Item: Hiển thị thẻ ngân hàng/ví tiền mặt (Logo ngân hàng + Tên ví + Số dư phân biệt màu âm/dương).
   Transaction List Item: Hàng giao dịch (Avatar icon tròn danh mục + Tên khoản chi/thu + Ghi chú + Số tiền định dạng màu đỏ/xanh).
   Category Tree Item: Hàng chọn nhóm chi tiêu (Icon màu + Tên nhóm + Số ví áp dụng + Chevron >), hỗ trợ phân cấp nhóm cha/nhóm con.
   Section Date Header: Hàng tiêu đề phân nhóm theo ngày tháng trong sổ giao dịch (Ngày lớn 15 + Thứ/tháng + Tổng thu/chi trong ngày).
4. Hiển thị dữ liệu & Thẻ thông tin (Cards & Data Display) — 5 components
   Overview Metric Card (Total Balance Card): Thẻ tổng số dư với nút ẩn/hiện số tiền (icon con mắt) và nhãn phụ.
   Budget Progress Card: Thẻ ngân sách từng nhóm gồm: Icon + Tên nhóm + Hạn mức + Thanh tiến độ 2 màu (linear progress bar) + Trạng thái mốc thời gian (Hôm nay).
   Circular Gauge / Budget Semi-Circle Meter: Đồng hồ đo bán nguyệt hiển thị ngân sách tổng ("Số tiền bạn có thể chi" + tiến độ vạch màu xanh).
   Stat Metric Card (Comparison Tile): Thẻ thông số so sánh (ví dụ: Chi tiêu trung bình, Tỷ lệ tăng trưởng % so với tháng trước với badge đỏ/xanh).
   Promo / Notice Banner: Banner tiếp thị bo góc viền màu xanh lá (7 ngày Dùng Miễn Phí / Banner dữ liệu mẫu màu xanh dương kèm nút Tắt).
5. Biểu đồ dữ liệu (Charts & Visualizations) — 3 components
   Bar Comparison Chart: Biểu đồ cột đơn/đôi so sánh chi tiêu theo mốc thời gian (Tuần này vs Tuần trước).
   Donut Category Chart: Biểu đồ tròn thể hiện tỷ trọng chi tiêu có icon nhóm nằm ở trung tâm hoặc chú thích bên dưới.
   Trend Line & Gradient Bar Chart: Biểu đồ xu hướng chi tiêu theo chuỗi ngày kết hợp đường kẻ và dải màu gradient.
6. Cửa sổ bật lên & Lớp phủ (Modals, Overlays & Popovers) — 4 components
   Action Sheet / Bottom Modal: Khung danh sách hành động trượt từ dưới lên (như modal Thêm giao dịch bằng ảnh gồm 3 nút chọn).
   Calendar Date Picker Modal: Bảng lịch chọn ngày dạng lưới tháng (Header chuyển tháng + Lưới thứ 2 - CN + Vòng tròn highlight ngày chọn).
   Context Dropdown Menu / Popover: Menu thả xuống từ icon góc trên (gồm các tác vụ: Khoảng thời gian, Xem theo nhóm, Điều chỉnh số dư, Đồng bộ ví...).
   Wallet Selection Bottom Sheet: Sheet danh sách chọn ví có tích chọn radio xanh lá (✓) và nút thêm ví mới.
7. Huy hiệu & Phản hồi (Badges & Micro-elements) — 2 components
   Status & Percentage Badge: Huy hiệu bo tròn thể hiện tỷ lệ % (ví dụ: badge tròn viền đỏ 214%, badge pill xanh ↓ 100%).
   Category Icon Badge: Khung icon tròn đường kính cố định (~40px) với màu nền riêng biệt cho từng loại chi tiêu (Ăn uống, Cà phê, Hoá đơn, Mua sắm...).
   Tổng kết: 28 base components là phạm vi mục tiêu của design system. Code hiện tại mới có các component trong `app/src/atomic/`; các component chưa có phải được tạo ở layer base trước khi dùng.

### Quy tắc triển khai bắt buộc

Luồng bổ sung để review: [Tài khoản → Quản lý nhóm → Sửa nhóm](./system/DESIGN.md#account--quản-lý-nhóm), gồm nhóm cha, trường “category” cần làm rõ và ví áp dụng. Liên kết công việc: [TICKET-01-03](../work/tickets/TICKET-01-03-tong-vi-danh-muc.md). Chưa implement màn này.

- Luôn tìm component tương ứng trong `app/src/atomic/atoms/` và `app/src/atomic/molecules/` trước khi viết JSX mới.
- Nếu chưa có component phù hợp, tạo hoặc cập nhật base component trước; không copy markup/style vào từng screen.
- Các organism chỉ được lắp ghép base components và xử lý composition/state của màn hình.
- Không tạo biến thể chỉ khác màu, padding, icon badge hoặc progress markup nếu có thể mở rộng base component bằng props/variant.
- Khi cập nhật base component, rà toàn bộ consumer để giữ giao diện hiện tại đồng nhất.

Bạn có muốn bắt đầu bằng việc:

Tạo một trang Design System / Component Library dựng sẵn toàn bộ 28 components này dưới dạng UI Kit?
Hay ưu tiên dựng trước nhóm Form Controls & Keypad chuyên sâu cho luồng thêm giao dịch / thêm ngân sách?

---

# TÀI LIỆU YÊU CẦU SẢN PHẨM (PRODUCT REQUIREMENT DOCUMENT - PRD)

## DỰ ÁN: ỨNG DỤNG QUẢN LÝ TÀI CHÍNH CÁ NHÂN & THIẾT KẾ HỆ THỐNG UI/UX (FINANCIAL CLARITY)

---

### 1. TỔNG QUAN DỰ ÁN (EXECUTIVE SUMMARY)

- **Tên dự án:** Financial Clarity App (Tái thiết kế & Hệ thống hóa UI/UX ứng dụng quản lý chi tiêu cá nhân thông minh)
- **Cảm hứng & Tham chiếu gốc:** Ứng dụng Money Lover kết hợp ngôn ngữ thiết kế hiện đại Financial Clarity System.
- **Mục tiêu sản phẩm:**
   - Đơn giản hóa hành vi ghi chép chi tiêu thường nhật dưới 5 giây.
   - Cung cấp góc nhìn phân tích tài chính sâu sắc (Money Insider, Budgeting, Cashflow Forecast) giúp người dùng kiểm soát hạn mức và tích lũy hiệu quả.
   - Chuẩn hóa toàn bộ hệ sinh thái **Base Components** và **Design Tokens** thành bộ tài liệu có khả năng tái sử dụng (Design System-driven development).

---

### 2. MỤC TIÊU SẢN PHẨM & CHỈ SỐ ĐO LƯỜNG (GOALS & KPIS)

#### 2.1. Mục tiêu kinh doanh & người dùng

- Giảm tỷ lệ bỏ dở khi nhập giao dịch (Drop-off Rate tại Form thêm giao dịch) xuống dưới **8%**.
- Tăng tần suất người dùng mở app kiểm tra tiến độ ngân sách hàng ngày (**DAU/MAU tăng > 25%**).
- Tối ưu hóa hiệu suất lập trình giao diện nhờ bộ linh kiện giao diện chuẩn (Base Components) thống nhất 100% giữa Design và Code.

#### 2.2. Chỉ số thành công (Success Metrics)

- **Tốc độ ghi nhận chi tiêu (Time to log expense):** Trung bình ≤ 4 giây nhờ bàn phím số tích hợp máy tính mini và gợi ý danh mục thông minh.
- **Mức độ gắn kết với Báo cáo (Report Retention):** 65% người dùng xem biểu đồ so sánh xu hướng tuần/tháng ít nhất 1 lần/tuần.

---

### 3. ĐỐI TƯỢNG NGƯỜI DÙNG & CHÂN DUNG (TARGET AUDIENCE & PERSONAS)

1. **Persona 1 - Người đi làm / Nhân viên văn phòng (22 - 35 tuổi):**
   - _Hành vi:_ Chi tiêu nhiều khoản nhỏ hàng ngày (cà phê, cơm trưa, đi lại, hóa đơn tiện ích).
   - _Nỗi đau (Pain points):_ Ngại nhập form nhiều trường phức tạp; dễ quên các khoản chi lắt nhắt; cuối tháng thường hụt tiền mà không rõ nguyên nhân.
2. **Persona 2 - Người quản lý tài chính gia đình / Freelancer:**
   - _Hành vi:_ Quản lý cùng lúc nhiều ví/tài khoản (Tiền mặt, Techcombank, Thẻ tín dụng ACB Platinum,...); theo dõi ngân sách định mức nghiêm ngặt.
   - _Nỗi đau (Pain points):_ Khó quản lý chi tiêu phân tầng chi tiết; khó tính toán số tiền an toàn còn lại có thể tiêu trong ngày.

---

### 4. KIẾN TRÚC THÔNG TIN & LUỒNG NGƯỜI DÙNG (INFORMATION ARCHITECTURE)

```
[ Financial Clarity Mobile App ]
│
├── 1. Tab Tổng Quan (Dashboard / Overview)
│   ├── Tổng số dư toàn ví (Eye toggle che/hiển thị số dư)
│   ├── Thẻ danh sách ví thành viên (Ví tín dụng, Techcombank, Tiền mặt)
│   ├── Thẻ nhanh Money Insider (Cảnh báo chi tiêu tăng vọt, gợi ý tiết kiệm)
│   ├── Widget Báo cáo chi tiêu 2 kỳ gần nhất (Tuần này vs Tuần trước)
│   └── Danh sách giao dịch phát sinh gần đây
│
├── 2. Tab Sổ Giao Dịch (Transactions Log)
│   ├── Bộ lọc thời gian (Theo ngày, tuần, tháng, khoảng tùy chọn)
│   ├── Xem nhóm theo Danh mục / Theo Ví
│   └── Phân nhóm ngày giao dịch (Header ngày, tổng thu/chi trong ngày)
│
├── 3. Thao tác Nhanh (+) (Quick Add FAB)
│   ├── Form thêm giao dịch (Khoản chi / Khoản thu / Vay-Nợ)
│   ├── Bàn phím số chuyên dụng (Numeric Keypad + Máy tính cơ bản + Quick Chips)
│   └── Phân loại danh mục đa tầng (Nested Category Selector)
│
├── 4. Tab Ngân Sách (Budgets Management)
│   ├── Đồng hồ đo ngân sách bán nguyệt (Semi-Circle Gauge Meter)
│   ├── Chỉ số: "Số tiền bạn có thể chi mỗi ngày", số ngày đến cuối kỳ
│   └── Danh sách thanh tiến độ ngân sách danh mục (Ăn uống, Mua sắm, v.v.)
│
└── 5. Báo Cáo & Xu Hướng (Financial Reports & Insights)
    ├── Biểu đồ so sánh 2 kỳ (Bar Comparison)
    ├── Biểu đồ đường lũy kế tháng (Cumulative Trend Curve)
    ├── Biểu đồ cột bóc tách chi tiêu từng ngày (Daily Breakdown)
    └── Biểu đồ Donut phân bổ tỷ trọng cơ cấu chi phí
```

---

### 5. ĐẶC TẢ TÍNH NĂNG CHÍNH (FEATURE SPECIFICATIONS)

#### F1. Nhập Giao Dịch Nhanh (Fast Transaction Logging)

- **Mô tả:** Cho phép tạo giao dịch với ít bước chạm nhất có thể.
- **Yêu cầu chi tiết:**
   - Tab phân loại 3 trạng thái: _Khoản chi_ (mặc định), _Khoản thu_, _Cho vay/Đi vay_.
   - Ô số tiền lớn, font rõ nét, tự động định dạng dấu phân cách hàng nghìn (`250.000 đ`).
   - Hỗ trợ Keypad số tích hợp máy tính (+, -, ×, ÷, C, Xong) và phím tắt tiền mặt nhanh (`50.000`, `150.000`, `500.000`).
   - Lựa chọn ví chi trả, ngày thực hiện, ghi chú ngắn, đính kèm ảnh hóa đơn và công tắc _"Không tính vào báo cáo"_.

#### F2. Cấu Trúc Danh Mục Đa Tầng (Nested Category Architecture)

- **Mô tả:** Nhóm danh mục chi tiêu khoa học với liên kết nhánh cây trực quan.
- **Yêu cầu chi tiết:**
   - Phân cấp Cha - Con: Danh mục mẹ (Anchor) và các danh mục con trực thuộc.
   - Trục nối nhánh cây (_Tree Branch Line_): Xuất phát từ lề chân khối text nội dung của mục mẹ (`left: ~66px`), bo cong 90° nhẹ nhàng dẫn vào tâm icon mục con với khoảng cách lề `pl-10` đến `pl-12` (chuẩn hóa gọn gàng).
   - Trạng thái mở rộng/thu gọn danh mục linh hoạt.

#### F3. Đồng Hồ Đo Tiến Độ & Dự Báo Ngân Sách (Budget Gauge & Safe-to-Spend)

- **Mô tả:** Trực quan hóa giới hạn ngân sách tháng và tính toán định mức an toàn.
- **Yêu cầu chi tiết:**
   - Vòng cung đo bán nguyệt SVG mềm mại với con trỏ chỉ vị trí chi tiêu thực tế.
   - Hiển thị nổi bật: _"Số tiền bạn có thể chi"_ (Safe-to-spend).
   - 3 chỉ số vệ tinh: _Tổng ngân sách_, _Tổng đã chi_, _Số ngày còn lại đến hết chu kỳ_.
   - Thẻ tiến độ ngân sách riêng cho từng mục thiết yếu kèm nhãn định mức ngày.

#### F4. Bộ Sưu Tập Báo Cáo & Biểu Đồ Thống Kê (Financial Charts Suite)

- **Mô tả:** 4 mẫu biểu đồ phân tích chuyên sâu:
   1. _Biểu đồ cột so sánh 2 kỳ:_ Làm nổi bật biến động tăng/giảm phần trăm (`↑ 92%`).
   2. _Biểu đồ đường xu hướng lũy kế:_ So sánh đường tích lũy thực tế với đường trung bình 3 tháng trước, kèm Tooltip chi tiết theo từng mốc ngày.
   3. _Biểu đồ cột phân bổ theo ngày:_ Bóc tách chi tiêu từ ngày 1 đến ngày 31, phát hiện các đỉnh chi tiêu đột biến.
   4. _Biểu đồ Donut cơ cấu:_ Thể hiện % tỷ trọng chi tiêu của các danh mục, tâm donut hiển thị danh mục chiếm tỷ trọng cao nhất.

---

### 6. QUY CHUẨN HỆ THỐNG THIẾT KẾ (DESIGN SYSTEM SPECIFICATIONS)

Hệ thống tuân thủ **Financial Clarity System**:

| Danh mục Token           | Giá trị chuẩn                                               | Mục đích áp dụng                                       |
| :----------------------- | :---------------------------------------------------------- | :----------------------------------------------------- |
| **Typography**           | `Manrope`, sans-serif (Font chữ trung tính, dễ đọc số liệu) | Toàn bộ giao diện Mobile/Desktop                       |
| **Primary Color**        | `#22c55e` (Emerald / Green 500)                             | Nút hành động chính (CTA), Icon ví, Tín hiệu tích cực  |
| **Expense / Alert**      | `#ef4444` (Rose / Red 500)                                  | Cảnh báo vượt ngân sách, Cột chi phí kỳ hiện tại       |
| **Past Reference**       | `#fecdd3` (Pastel Rose 200)                                 | Cột đối ứng so sánh kỳ trước                           |
| **Background / Neutral** | `#fbf9f9` (Surface), `#ffffff` (Card Container)             | Nền sạch, tạo cảm giác nhẹ nhàng, thông thoáng         |
| **Text Primary**         | `#0f172a` (Slate 900) / `#1e293b` (Slate 800)               | Tiêu đề, số tiền chính                                 |
| **Text Secondary**       | `#64748b` (Slate 500) / `#94a3b8` (Slate 400)               | Nhãn phụ, mô tả ví, chú thích                          |
| **Corner Radius**        | `rounded-3xl` (24px) / `rounded-2xl` (16px)                 | Bo góc thẻ Card mềm mại đặc trưng                      |
| **Touch Target**         | Chiều cao hàng tối thiểu 44px – 48px                        | Chuẩn Human Interface Guidelines (HIG) cho iOS/Android |

---

### 7. DANH MỤC BASE COMPONENTS VÀ MỨC ĐỘ HIỆN THỰC

Các component dưới đây là phạm vi thiết kế; implementation hiện tại đã có một phần dưới dạng atoms/molecules và mock organisms:

1. **Amount Display Input & Numeric Keypad (Component #1):** Đã có bản mock trong `QuickAddSheet`; chưa phải base component độc lập.
2. **Budget Meter & Progress Cards (Component #2):** Đã có trong `BudgetsPanel` và `BudgetProgressItem`; progress primitive dùng chung chưa tách riêng.
3. **Multi-purpose Base Cards System (Component #3):** Showcase chuẩn nằm tại [`atoms/base-cards`](./atoms/base-cards/). `SurfaceCard` hỗ trợ radius, padding và elevation; `IconBadge` hỗ trợ size, shape và tone; `BaseButton` hỗ trợ variant, size và loading; các molecule wallet/transaction dùng mapping variant global.
4. **Transaction Form Card & Row Items (Component #4):** `FormSelectorRow` và `QuickAddSheet` đã có; form card vẫn là composition cục bộ.
5. **Financial Charts Suite (Component #5):** Mock chart hiện nằm trong `OverviewPanel`/`ReportsPanel`; chưa có chart primitives.
6. **Nested Category Cards (Component #6):** `CategoryTreeCard` đã có và đang dùng trong Reports.

---

### 8. LỘ TRÌNH TRIỂN KHAI & PHÁT TRIỂN TIẾP THEO (NEXT STEPS & ROADMAP)

- **Giai đoạn 1 (Hiện tại - UI Foundations):** Đã hoàn tất đóng gói toàn bộ Base Components và màn hình báo cáo chi tiết.
- **Giai đoạn 2 (Modal & Overlays):**
   - Dựng _Calendar Date Picker Modal_ (Lưới chọn ngày tháng giao dịch chuẩn iOS).
   - Dựng _Wallet Switcher & Management Sheet_ (Bottom Sheet chọn/quản lý ví).
   - Dựng _Action Context Sheet_ (Thực đơn thao tác phụ cho từng giao dịch).
- **Giai đoạn 3 (Assembly & Flow Testing):** Lắp ghép các Base Components vào các luồng màn hình hoàn chỉnh (End-to-End User Flow).
