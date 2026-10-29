# Nhập số trên iPhone

Các ô dùng cùng kiểu hiển thị nhưng kiểm tra theo đúng loại dữ liệu:

| Loại | Ví dụ | Quy tắc |
| --- | --- | --- |
| Số tiền VND | `15000` | Số nguyên; định dạng tiền không thay đổi giá trị lưu |
| Khối lượng đầu tư | `1,25` hoặc `1.25` | Số dương, tối đa8 chữ số thập phân; không trộn dấu phân cách |
| Chu kỳ định kỳ | `2` | Số kỳ nguyên dương; nhập `1000` nếu muốn1000 kỳ, không dùng `1.000` |
| Tỷ lệ phân bổ hũ | `12,50` | Lớn hơn0, tối đa100%, tối đa2 chữ số thập phân |

App không tự xóa ký tự sai để biến nội dung thành một số khác. Nếu dán nội dung
không hợp lệ, sửa lại trước khi lưu. Giá/phí đầu tư là số tiền, không phải khối lượng.

## AI-agent contract

Native `MPNumberInputKind.normalized` accepts ASCII digits and one decimal
separator (comma or dot), returns a string without floating-point conversion.
Empty remains empty; malformed syntax returns nil. Integer fields reject
fractions/grouping, percentage syntax accepts0…100; feature validators impose
strict positivity and domain scale. Native portfolio request encodes normalized
`quantity` with dot. Backend payload types/rules and money parsers unchanged.
