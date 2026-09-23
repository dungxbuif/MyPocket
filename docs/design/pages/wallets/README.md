# Quản lý ví — đặc tả màn hình chuẩn

## Cập nhật implementation mới nhất

BaseSelect đã thay tương tác picker loại khi tạo. Danh sách đã dùng WalletSelectionList: Đóng/Chọn Ví/Sửa, Tổng cộng, hai nhóm tính/không tính vào tổng và card hành động. Sửa bật chế độ chọn ví để edit; xóa nằm trong edit. Chọn bất kỳ ví nào mở WalletDetailPanel với số dư, dòng tiền và lịch sử giao dịch; ví tín dụng chỉ xem lịch sử và không mở editor thu/chi thường. Goal hỗ trợ target_date qua API và UI. Sổ giao dịch có select đầu màn hình cho `Tổng cộng` hoặc từng ví; ví thường dùng cùng layout/kỳ với `Tổng cộng` và chỉ lọc giao dịch của ví được chọn. Xem [savings](../savings/README.md).

## Quyết định hiện hành

Owner chốt dùng **FormField + BaseSelect** chọn loại khi tạo: Ví thường / Ví tiết kiệm / Ví tín dụng. Cho phép đổi trong draft, lưu xong loại bất biến ở UI và API. Không cần sheet chọn loại. Tài liệu này gom danh sách, thêm và sửa ví; không đánh đồng thiết kế đích với implementation hiện tại.

## Bố cục và base

| Vùng | Base | Contract |
| --- | --- | --- |
| Header quản lý | PageBackHeader hoặc BaseLink + Heading/Text | Quay lại Tài khoản, tiêu đề, Thêm ví |
| Danh sách ví | SurfaceCard, IconBadge, Text, BaseButton | Tên, nhãn loại tiếng Việt, số dư thực, sửa/xóa; chỉ dữ liệu owner |
| Thêm ví | BaseBottomSheet presentation=form | Hủy / Thêm Ví / Lưu; focus trap, Escape, khôi phục focus |
| Loại ví | FormField + BaseSelect | Chỉ chọn trong draft tạo; không dùng ảnh Chọn Ví làm picker loại |
| Thông tin chung | SurfaceCard, Divider, BaseTextInput, IconBadge | Tên, VND cố định, số dư đầu phù hợp loại; không ghi chú khi tạo |
| Trường theo loại | FormField + BaseTextInput | Target/limit dương theo loại; ngày/sao kê chưa hỗ trợ phải ghi rõ |
| Không tính vào tổng | BaseSwitch + SurfaceCard + Text | checked = is_in_total false; mặc định tính vào tổng |
| Trạng thái | StatusMessage | Empty plain; lỗi alert; giữ draft khi lỗi |

Màu, font, bo góc theo [tokens](../../system/TOKENS.md) và [base contracts](../../system/BASE_COMPONENTS.md), không copy CSS/device frame từ ảnh. Liên kết dịch vụ hiện disabled với giải thích chưa hỗ trợ.

## Hành vi riêng từng loại

| Loại | Khi tạo | Sau khi mở ví — thiết kế nghiệp vụ | Hiện trạng |
| --- | --- | --- | --- |
| Thường | Tên, VND, số dư đầu, tính vào tổng | Số dư và lịch sử thu/chi; chuyển ví/điều chỉnh là luồng riêng | CRUD, màn chi tiết và thu/chi có |
| Tiết kiệm | Trường chung, mục tiêu > 0, hạn tùy chọn | Đã có, còn thiếu, tiến độ theo số dư thật; nạp/rút cập nhật tiến độ; không tự tính lãi ngân hàng | Có target/ngày mục tiêu trong form/API, tiến độ và lịch sử thực trong chi tiết; visual UAT còn thiếu |
| Tín dụng | Tên, VND, hạn mức > 0; thiết kế tiếp số dư sao kê gần nhất, ngày sao kê/hạn trả | Dư nợ, hạn mức khả dụng, mua/hoàn tiền/phí, thanh toán và sao kê; không dùng ledger thu/chi thường | Core CRUD/hạn mức và màn chi tiết đọc lịch sử đã có; ledger tín dụng chuyên biệt vẫn chưa triển khai |

Không suy diễn số dư sao kê bằng dư nợ hiện tại. Màn chi tiết theo loại cần contract và proof riêng trước khi coi hoàn thành; các trường dữ liệu mới cần detail design API/schema.

Ví tín dụng hiện **mới có đặc tả nghiệp vụ cấp cao** ở [SPEC](../../../requirements/SPEC.md) và [BUSINESS_RULES](../../../requirements/BUSINESS_RULES.md): hạn mức, dư nợ, sao kê, ngày đến hạn, mua/hoàn/phí/lãi/trả nợ và tránh tính chi hai lần. Chưa có đặc tả màn chi tiết được duyệt cho kỳ sao kê, hạn trả, phân bổ thanh toán/trả dư và các API/schema tương ứng; vì vậy màn đang có chỉ là danh sách lịch sử đọc được, không được gắn nhãn hoàn thiện hay dùng thu/chi thường để ghi thẻ.

Du lịch hiện là `Travel Mode`/sự kiện gắn giao dịch theo [TICKET-05-03](../../../work/tickets/TICKET-05-03-su-kien-travel-mode.md), không phải loại ví thứ tư. Chưa có màn/contract UI riêng cho chuyến đi được phê duyệt và chưa có runtime Travel Mode; không hiển thị nhãn “ví du lịch” hoặc hứa rằng số dư ví thuộc riêng chuyến đi. Khi thiết kế màn chuyến đi, cần phân biệt tổng thu/chi của giao dịch liên kết với số dư của các ví nguồn.

## Sự kiện và trạng thái

- Vào màn: tải dữ liệu thật; loading → ready/empty/error; không thay lỗi bằng mock.
- Thêm: draft mới, basic, số dư 0, tính vào tổng bật. Đổi select hiện trường phù hợp và chỉ gửi dữ liệu liên quan loại đã chọn.
- Lưu: tên không trống, tiền nguyên VND hợp lệ, target/limit dương khi cần; khóa gửi lặp. Thành công đóng sheet và refresh; thất bại giữ draft và lỗi trong sheet.
- Hủy/Escape/backdrop: bỏ draft khi không đang lưu; không autosave.
- Sửa: loại và opening_balance chỉ đọc. Metadata sửa theo loại; ghi chú cũ trong edit vẫn giữ. Điều chỉnh số dư không phải sửa opening_balance.
- Xóa: xác nhận xóa vĩnh viễn, nêu số giao dịch và ảnh hưởng; refresh sau thành công. Liên kết chuyển ví/tín dụng cần thiết kế riêng trước mở rộng.
- Field có label, select dùng keyboard native, modal dùng focus trap/return focus. Server kiểm tra quyền sở hữu và validation cuối cùng.

## Phân biệt Chọn Ví và quản lý ví

[Ảnh Chọn Ví](../../ch_n_v_wallet_selector/README.md) chọn ví đã tồn tại hoặc Tổng cộng cho một màn tiêu thụ; không chọn loại. Runtime dùng lựa chọn này trong màn quản lý ví và bộ lọc đầu sổ giao dịch; [Ảnh Thêm Ví](../../th_m_v_add_wallet/README.md) là tham chiếu bố cục form.

## Nguồn Money Lover đã kiểm tra

- [Tạo/sửa ví](https://moneylover.zendesk.com/hc/en-us/articles/34974522779417-Create-edit-archive-and-delete-wallets): form theo loại, ảnh thao tác; tín dụng có hạn mức, số dư sao kê, ngày sao kê và hạn trả.
- [Tiết kiệm](https://moneylover.zendesk.com/hc/en-us/articles/36976160006809-How-to-manage-your-financial-goals-savings-effectively): mục tiêu, hạn, đã có/còn thiếu, tiến độ, dòng tiền; không tính lãi ngân hàng.
- [Tín dụng](https://moneylover.zendesk.com/hc/en-us/articles/37006112179609-How-to-manage-credit-cards-effectively): số dư/dư nợ, hạn trả, hạn mức khả dụng và dòng tiền.
- [Báo cáo](https://moneylover.zendesk.com/hc/en-us/articles/34533897137177-Report-Definition-and-Usage): thông tin riêng theo loại nằm trong báo cáo khi chọn ví; chưa đủ bằng chứng về dashboard độc lập khi mở ví.

Khóa loại sau tạo là quyết định MyPocket, không gán cho Money Lover. Tài liệu Money Lover có nhắc chuyển basic sang linked. Nguồn tham khảo không tự phê duyệt thay đổi nghiệp vụ/API MyPocket.

## Acceptance và khoảng trống

Runtime dùng WalletTypePicker cho loại ví và WalletDetailPanel cho mọi ví. Proof cần bổ sung visual UAT cho detail basic/goal/credit, chọn `Tổng cộng`/ví trong sổ giao dịch, create/edit khóa loại, validation theo loại, switch tổng, save/error/cancel và keyboard/mobile. Chạy check:design, test:design, build khi môi trường frontend có Node.

[Backlog](../../../work/BACKLOG.md) · [Validation](../../../work/VALIDATION_MATRIX.md) · [Ticket](../../../work/tickets/TICKET-01-02-quan-ly-vi.md) · [Release](../../../releases/CHANGELOG.md).

## Bản triển khai trước quyết định select — chỉ để đối chiếu

Route `/account/wallets`. [Work/design](../../../work/tickets/UI-WALLET-02-DETAIL_DESIGN.md), [reference](../../th_m_v_add_wallet/README.md).

Add opens one BaseBottomSheet, canvas/form presentation, with Hủy / Thêm Ví / Lưu header. SurfaceCard groups IconBadge + inline title input, fixed VND, and opening balance. Divider separates rows. No note. A row opens the three-type picker; goal requires positive target and credit positive limit. Service linking is disabled and labeled unavailable. BaseSwitch “Không tính vào tổng” inversely maps to is_in_total.

Type picker uses shared row buttons, IconBadge, Text and selected check. Selecting returns to the preserved draft; back/Escape returns without saving. One dialog owns focus trap, scroll lock and restore. Save is unavailable for blank name, invalid balance, invalid type-specific amount or pending request. Successful API save closes and refreshes; failure keeps draft and displays alert inside sheet. Full cancel discards draft. Edit continues existing immutable type/opening-balance behavior.

No new public API, currency selector, icon persistence or bank-link integration. Use shared tokens even where export uses different colors or larger corners. Empty wallet messages remain plain. Proof: shared SSR contracts, design regressions and production build; user visual acceptance pending.
