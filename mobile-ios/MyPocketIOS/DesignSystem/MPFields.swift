import SwiftUI
import UIKit

/// Conversational input, not a note editor. Features retain the draft and own
/// request lifecycle; the base owns keyboard focus and compact touch geometry.
struct MPChatComposer: View {
    @Binding var text: String
    let isBusy: Bool
    let isEnabled: Bool
    let onSend: () -> Void
    var onCancel: (() -> Void)?
    var dismissKeyboardRequest = 0
    @FocusState private var focused: Bool

    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            TextField("Hỏi về tiền của bạn...", text: $text, axis: .vertical)
                .lineLimit(1...4)
                .font(MPFont.body)
                .foregroundStyle(MPColor.ink)
                .focused($focused)
                .disabled(isBusy)
                .padding(.vertical, 10)
                .padding(.leading, 12)
                .accessibilityLabel("Tin nhắn cho trợ lý")
                .accessibilityIdentifier("assistant-composer-input")
            Button {
                if let onCancel { onCancel() }
                else { onSend() }
            } label: {
                Image(systemName: onCancel == nil ? "arrow.up" : "stop.fill")
                    .font(MPFont.bodyEmphasis)
                    .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
                    .foregroundStyle(.white)
                    .background(MPColor.action, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
            }
            .buttonStyle(.plain)
            .disabled(onCancel == nil && (!isEnabled || isBusy || text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty))
            .opacity(onCancel == nil && (!isEnabled || isBusy || text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty) ? 0.5 : 1)
            .accessibilityLabel(onCancel == nil ? "Gửi" : "Dừng xử lý")
            .accessibilityIdentifier("assistant-composer-action")
        }
        .padding(4)
        .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        .overlay {
            RoundedRectangle(cornerRadius: MPMetrics.cornerRadius)
                .stroke(focused ? MPColor.action : MPColor.line, lineWidth: MPMetrics.inputBorderWidth)
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("assistant-composer")
        .onChange(of: dismissKeyboardRequest) { _, _ in focused = false }
    }
}

private struct MPFormControlModifier: ViewModifier {
    let isFocused: Bool

    func body(content: Content) -> some View {
        content
            .font(MPFont.body)
            .foregroundStyle(MPColor.ink)
            .padding(.horizontal, MPMetrics.inputHorizontalPadding)
            .frame(minHeight: MPMetrics.controlHeight)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
            .overlay {
                RoundedRectangle(cornerRadius: MPMetrics.cornerRadius)
                    .stroke(isFocused ? MPColor.accent : MPColor.line, lineWidth: isFocused ? MPMetrics.inputFocusWidth : MPMetrics.inputBorderWidth)
            }
            .shadow(color: isFocused ? MPColor.successLine.opacity(0.9) : .clear, radius: isFocused ? 3 : 0)
    }
}

struct MPTextField: View {
    let title: String
    @Binding var text: String
    var placeholder: String = ""
    var error: String?
    var accessibilityIdentifier: String?
    var keyboardType: UIKeyboardType
    @FocusState private var isFocused: Bool

    init(title: String, text: Binding<String>, placeholder: String = "", error: String? = nil, accessibilityIdentifier: String? = nil, keyboardType: UIKeyboardType = .default) {
        self.title = title
        self._text = text
        self.placeholder = placeholder
        self.error = error
        self.accessibilityIdentifier = accessibilityIdentifier
        self.keyboardType = keyboardType
    }

    private var clearButtonAccessibilityIdentifier: String {
        if let accessibilityIdentifier {
            return "clear-\(accessibilityIdentifier)"
        }
        return "clear-field-\(title)"
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(MPFont.label)
                .foregroundStyle(MPColor.ink)
            HStack(spacing: 8) {
                Group {
                    if let accessibilityIdentifier {
                        MPUikitTextField(
                            text: $text,
                            isFocused: $isFocused,
                            placeholder: placeholder,
                            accessibilityLabel: title,
                            accessibilityIdentifier: accessibilityIdentifier,
                            keyboardType: keyboardType
                        )
                        .frame(minWidth: 0, maxWidth: .infinity, minHeight: MPMetrics.controlHeight, maxHeight: MPMetrics.controlHeight, alignment: .leading)
                    } else {
                        MPUikitTextField(
                            text: $text,
                            isFocused: $isFocused,
                            placeholder: placeholder,
                            accessibilityLabel: title,
                            accessibilityIdentifier: nil,
                            keyboardType: keyboardType
                        )
                        .frame(minWidth: 0, maxWidth: .infinity, minHeight: MPMetrics.controlHeight, maxHeight: MPMetrics.controlHeight, alignment: .leading)
                    }
                }
                if !text.isEmpty {
                    Button {
                        // Resign first so UIKit's active editing buffer cannot
                        // overwrite the cleared binding on the next run loop.
                        isFocused = false
                        text = ""
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundStyle(MPColor.muted)
                            .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("Xóa \(title)")
                    .accessibilityIdentifier(clearButtonAccessibilityIdentifier)
                }
            }
            .frame(minWidth: 0, maxWidth: .infinity, alignment: .leading)
            .modifier(MPFormControlModifier(isFocused: isFocused))
            .accessibilityLabel(title)
            if let error {
                Text(error)
                    .font(MPFont.bodySmall)
                    .foregroundStyle(MPColor.danger)
                    .accessibilityLabel("Lỗi: \(error)")
            }
        }
    }
}

@MainActor
private struct MPUikitTextField: UIViewRepresentable {
    @Binding var text: String
    let isFocused: FocusState<Bool>.Binding
    let placeholder: String
    let accessibilityLabel: String
    let accessibilityIdentifier: String?
    let keyboardType: UIKeyboardType

    func makeCoordinator() -> Coordinator {
        Coordinator(text: $text, isFocused: isFocused)
    }

    func makeUIView(context: Context) -> UITextField {
        let field = UITextField(frame: .zero)
        field.borderStyle = .none
        field.keyboardType = keyboardType
        field.backgroundColor = .clear
        field.font = MPFont.uiBody
        field.textColor = UIColor(MPColor.ink)
        field.placeholder = placeholder
        field.accessibilityLabel = accessibilityLabel
        field.accessibilityIdentifier = accessibilityIdentifier
        field.text = text
        field.delegate = context.coordinator
        field.addTarget(context.coordinator, action: #selector(Coordinator.valueChanged(_:)), for: .editingChanged)
        field.addTarget(context.coordinator, action: #selector(Coordinator.editingDidBegin(_:)), for: .editingDidBegin)
        field.addTarget(context.coordinator, action: #selector(Coordinator.editingDidEnd(_:)), for: .editingDidEnd)
        field.inputAccessoryView = context.coordinator.makeInputAccessoryView(for: field)
        return field
    }

    func updateUIView(_ field: UITextField, context: Context) {
        field.keyboardType = keyboardType
        if field.text != text {
            field.text = text
        }
        if field.placeholder != placeholder {
            field.placeholder = placeholder
        }
        if field.accessibilityLabel != accessibilityLabel {
            field.accessibilityLabel = accessibilityLabel
        }
        if field.accessibilityIdentifier != accessibilityIdentifier {
            field.accessibilityIdentifier = accessibilityIdentifier
        }
        // Do not force first-responder state from SwiftUI updates. Text input
        // events can update the binding per keystroke; forcing focus here can
        // resign the native field after the first character.
    }

    @MainActor
    final class Coordinator: NSObject, UITextFieldDelegate {
        private var text: Binding<String>
        private var isFocused: FocusState<Bool>.Binding
        private weak var field: UITextField?

        init(text: Binding<String>, isFocused: FocusState<Bool>.Binding) {
            self.text = text
            self.isFocused = isFocused
        }

        @objc func valueChanged(_ sender: UITextField) {
            text.wrappedValue = sender.text ?? ""
        }

        func textField(_ textField: UITextField, shouldChangeCharactersIn range: NSRange, replacementString string: String) -> Bool {
            guard let current = textField.text, let swiftRange = Range(range, in: current) else {
                return true
            }
            text.wrappedValue = current.replacingCharacters(in: swiftRange, with: string)
            return true
        }

        @objc func editingDidBegin(_ sender: UITextField) {
            isFocused.wrappedValue = true
        }

        @objc func editingDidEnd(_ sender: UITextField) {
            isFocused.wrappedValue = false
        }

        func makeInputAccessoryView(for field: UITextField) -> UIToolbar {
            self.field = field
            let toolbar = UIToolbar(frame: CGRect(x: 0, y: 0, width: 0, height: 44))
            toolbar.items = [
                UIBarButtonItem(systemItem: .flexibleSpace),
                UIBarButtonItem(title: "Xong", style: .done, target: self, action: #selector(done(_:)))
            ]
            toolbar.sizeToFit()
            return toolbar
        }

        @objc private func done(_ sender: UIBarButtonItem) {
            isFocused.wrappedValue = false
            field?.resignFirstResponder()
        }
    }
}

/// Money, quantity, count and percentage share field appearance but not parsers.
/// Keep raw binding text so an invalid paste cannot silently become a valid sum.
struct MPNumberField: View {
    let title: String
    @Binding var text: String
    let kind: MPNumberInputKind
    var placeholder: String = ""
    var accessibilityIdentifier: String?

    var body: some View {
        MPTextField(
            title: title, text: $text, placeholder: placeholder,
            error: !text.isEmpty && kind.normalized(text) == nil ? validationMessage : nil,
            accessibilityIdentifier: accessibilityIdentifier,
            keyboardType: kind == .integer ? .numberPad : .decimalPad
        )
    }

    private var validationMessage: String {
        switch kind {
        case .integer: return "Nhập số nguyên không âm."
        case .decimal: return "Nhập số thập phân, dùng dấu phẩy hoặc dấu chấm."
        case .percentage: return "Nhập tỷ lệ từ 0 đến 100%."
        }
    }
}

struct MPSecureField: View {
    let title: String
    @Binding var text: String
    var placeholder: String = ""
    var accessibilityIdentifier: String?
    @FocusState private var isFocused: Bool

    init(title: String, text: Binding<String>, placeholder: String = "", accessibilityIdentifier: String? = nil) {
        self.title = title
        self._text = text
        self.placeholder = placeholder
        self.accessibilityIdentifier = accessibilityIdentifier
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(MPFont.label)
                .foregroundStyle(MPColor.ink)
            Group {
                if let accessibilityIdentifier {
                    SecureField(placeholder, text: $text)
                        .accessibilityIdentifier(accessibilityIdentifier)
                } else {
                    SecureField(placeholder, text: $text)
                }
            }
            .focused($isFocused)
            .frame(minWidth: 0, maxWidth: .infinity, alignment: .leading)
            .modifier(MPFormControlModifier(isFocused: isFocused))
            .accessibilityLabel(title)
        }
    }
}

@MainActor
private struct MPNumberPadTextField: UIViewRepresentable {
    enum Presentation: Equatable {
        case control
        case transactionAmount
    }

    @Binding var text: String
    let placeholder: String
    let accessibilityLabel: String
    var presentation: Presentation = .control
    var accessibilityIdentifier: String?

    private var displayValue: String {
        guard presentation == .transactionAmount, let amount = Int64(text) else { return text }
        let formatter = NumberFormatter()
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.numberStyle = .decimal
        formatter.groupingSeparator = "."
        formatter.groupingSize = 3
        formatter.maximumFractionDigits = 0
        return formatter.string(from: NSNumber(value: amount)) ?? text
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(text: $text, presentation: presentation)
    }

    func makeUIView(context: Context) -> UITextField {
        let field = UITextField(frame: .zero)
        field.borderStyle = .none
        field.keyboardType = .numberPad
        field.placeholder = placeholder
        field.accessibilityLabel = accessibilityLabel
        field.accessibilityIdentifier = accessibilityIdentifier
        field.text = displayValue
        let leftInset = presentation == .control ? MPMetrics.inputHorizontalPadding : 0
        field.leftView = UIView(frame: CGRect(x: 0, y: 0, width: leftInset, height: 1))
        field.leftViewMode = leftInset > 0 ? .always : .never
        context.coordinator.applyStyle(to: field, focused: false, presentation: presentation)
        field.addTarget(context.coordinator, action: #selector(Coordinator.valueChanged(_:)), for: .editingChanged)
        field.addTarget(context.coordinator, action: #selector(Coordinator.editingDidBegin(_:)), for: .editingDidBegin)
        field.addTarget(context.coordinator, action: #selector(Coordinator.editingDidEnd(_:)), for: .editingDidEnd)
        field.inputAccessoryView = context.coordinator.makeInputAccessoryView(for: field)
        return field
    }

    func updateUIView(_ field: UITextField, context: Context) {
        context.coordinator.presentation = presentation
        if field.text != displayValue {
            field.text = displayValue
            if field.isFirstResponder, let end = field.position(from: field.beginningOfDocument, offset: field.text?.utf16.count ?? 0) {
                field.selectedTextRange = field.textRange(from: end, to: end)
            }
        }
        if field.accessibilityLabel != accessibilityLabel {
            field.accessibilityLabel = accessibilityLabel
        }
        if field.accessibilityIdentifier != accessibilityIdentifier {
            field.accessibilityIdentifier = accessibilityIdentifier
        }
        context.coordinator.applyStyle(to: field, focused: field.isFirstResponder, presentation: presentation)
    }

    @MainActor
    final class Coordinator: NSObject {
        private var text: Binding<String>
        var presentation: MPNumberPadTextField.Presentation
        private weak var field: UITextField?

        init(text: Binding<String>, presentation: MPNumberPadTextField.Presentation) {
            self.text = text
            self.presentation = presentation
        }

        @objc func valueChanged(_ sender: UITextField) {
            let value = sender.text ?? ""
            text.wrappedValue = presentation == .transactionAmount
                ? value.filter(\.isNumber)
                : value
        }

        @objc func editingDidBegin(_ sender: UITextField) {
            applyStyle(to: sender, focused: true, presentation: presentation)
        }

        @objc func editingDidEnd(_ sender: UITextField) {
            applyStyle(to: sender, focused: false, presentation: presentation)
        }

        func applyStyle(to field: UITextField, focused: Bool, presentation: MPNumberPadTextField.Presentation) {
            field.backgroundColor = UIColor(MPColor.surface)
            field.textColor = UIColor(MPColor.ink)
            switch presentation {
            case .control:
                field.font = MPFont.uiBody
                field.layer.cornerRadius = MPMetrics.cornerRadius
                field.layer.borderWidth = focused ? MPMetrics.inputFocusWidth : MPMetrics.inputBorderWidth
                field.layer.borderColor = UIColor(focused ? MPColor.accent : MPColor.line).cgColor
                field.layer.shadowColor = UIColor(MPColor.successLine).cgColor
                field.layer.shadowOpacity = focused ? 0.9 : 0
                field.layer.shadowRadius = focused ? 3 : 0
            case .transactionAmount:
                field.backgroundColor = .clear
                field.font = UIFont.systemFont(ofSize: 34, weight: .semibold)
                field.layer.cornerRadius = 0
                field.layer.borderWidth = 0
                field.layer.borderColor = UIColor.clear.cgColor
                field.layer.shadowOpacity = 0
                field.layer.shadowRadius = 0
            }
        }

        func makeInputAccessoryView(for field: UITextField) -> UIToolbar {
            self.field = field
            let toolbar = UIToolbar(frame: CGRect(x: 0, y: 0, width: 0, height: 44))
            toolbar.items = [
                UIBarButtonItem(systemItem: .flexibleSpace),
                UIBarButtonItem(title: "Xong", style: .done, target: self, action: #selector(done(_:)))
            ]
            toolbar.sizeToFit()
            return toolbar
        }

        @objc private func done(_ sender: UIBarButtonItem) {
            field?.resignFirstResponder()
        }
    }
}

/// Integer-safe calculator used by every VND amount field. The parser is
/// deliberately small and never evaluates executable input: only digits and
/// the four supported arithmetic operators are accepted, with multiplication
/// and division taking precedence over addition and subtraction.
enum MPAmountCalculator {
    static let maximumExpressionLength = 80

    static func calculate(_ expression: String) -> Int64? {
        let characters = Array(expression)
        guard !characters.isEmpty, characters.count <= maximumExpressionLength else { return nil }

        var numbers: [Int64] = []
        var operators: [Character] = []
        var current = ""
        for character in characters {
            if character.isNumber {
                current.append(character)
                continue
            }
            guard "+−-×÷".contains(character), !current.isEmpty,
                  let value = Int64(current) else { return nil }
            numbers.append(value)
            operators.append(character)
            current = ""
        }
        guard !current.isEmpty, let last = Int64(current), numbers.count == operators.count else { return nil }
        numbers.append(last)

        var term = numbers[0]
        var total: Int64 = 0
        var sign: Int64 = 1
        for index in operators.indices {
            let value = numbers[index + 1]
            switch operators[index] {
            case "×":
                let result = term.multipliedReportingOverflow(by: value)
                guard !result.overflow else { return nil }
                term = result.partialValue
            case "÷":
                guard value != 0, !(term == Int64.min && value == -1), term % value == 0 else { return nil }
                term /= value
            case "+", "−", "-":
                let signedTerm: Int64
                if sign == 1 {
                    signedTerm = term
                } else {
                    let negated = term.multipliedReportingOverflow(by: -1)
                    guard !negated.overflow else { return nil }
                    signedTerm = negated.partialValue
                }
                let subtotal = total.addingReportingOverflow(signedTerm)
                guard !subtotal.overflow else { return nil }
                total = subtotal.partialValue
                term = value
                sign = operators[index] == "+" ? 1 : -1
            default:
                return nil
            }
        }

        let signedTerm: Int64
        if sign == 1 {
            signedTerm = term
        } else {
            let negated = term.multipliedReportingOverflow(by: -1)
            guard !negated.overflow else { return nil }
            signedTerm = negated.partialValue
        }
        let result = total.addingReportingOverflow(signedTerm)
        guard !result.overflow, result.partialValue >= 0 else { return nil }
        return result.partialValue
    }

    static func display(_ expression: String) -> String {
        expression.isEmpty ? "0" : expression
    }
}

/// Shared calculator surface for integer VND amounts. Consumers bind only the
/// final amount; keypad state, arithmetic validation and key geometry stay in
/// this DesignSystem base.
struct MPAmountKeypad: View {
    @Binding var amount: String
    let title: String
    @Environment(\.dismiss) private var dismiss
    @State private var expression: String
    @State private var errorMessage: String?

    private static let keys: [[String]] = [
        ["C", "÷", "×", "⌫"],
        ["7", "8", "9", "−"],
        ["4", "5", "6", "+"],
        ["1", "2", "3", "="],
        ["0", "000", ",", "XONG"]
    ]

    init(amount: Binding<String>, title: String = "Bàn phím số") {
        self._amount = amount
        self.title = title
        self._expression = State(initialValue: amount.wrappedValue)
    }

    var body: some View {
        VStack(spacing: 12) {
            MPReferenceScreenHeader(title: title, closeLabel: "Hủy", onClose: { dismiss() })
            MPSurfaceCard(style: .reference) {
                VStack(alignment: .leading, spacing: 8) {
                    MPText("Số tiền", style: .secondary)
                    MPText(Self.displayExpression(expression), style: .display)
                        .lineLimit(1)
                        .minimumScaleFactor(0.55)
                        .accessibilityIdentifier("amount-keypad-display")
                    if let errorMessage {
                        MPStatusMessage(text: errorMessage, tone: .danger)
                    }
                }
            }
            LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 8), count: 4), spacing: 8) {
                ForEach(Self.keys.flatMap { $0 }, id: \.self) { key in
                    MPButton(
                        style: key == "XONG" ? .primary : .key,
                        fullWidth: true,
                        action: { press(key) }
                    ) {
                        if key == "⌫" {
                            Image(systemName: "delete.left")
                                .accessibilityLabel("Xóa số cuối")
                        } else {
                            Text(key)
                        }
                    }
                    .accessibilityLabel(key == "⌫" ? "Xóa số cuối" : key)
                    .accessibilityIdentifier("amount-key-\(key == "⌫" ? "backspace" : key.lowercased())")
                }
            }
            .padding(.horizontal, MPMetrics.pagePadding)
            Spacer(minLength: 0)
        }
        .padding(MPMetrics.pagePadding)
        .background(MPColor.canvas.ignoresSafeArea())
        .onAppear {
            expression = amount
            errorMessage = nil
        }
    }

    private static func displayExpression(_ value: String) -> String {
        value.isEmpty ? "0" : value
    }

    private func press(_ key: String) {
        errorMessage = nil
        switch key {
        case "C":
            expression = ""
        case "⌫":
            expression.removeLast()
        case "=":
            evaluate(commit: false)
        case "XONG":
            evaluate(commit: true)
        case ",":
            // VND writes are integer-only. Keep the visual reference key, but
            // fail at evaluation instead of silently rounding money.
            expression.append(",")
        default:
            guard expression.count < MPAmountCalculator.maximumExpressionLength else { return }
            expression.append(contentsOf: key)
        }
    }

    private func evaluate(commit: Bool) {
        guard let result = MPAmountCalculator.calculate(expression) else {
            errorMessage = "Nhập phép tính hợp lệ, kết quả phải là số nguyên không âm."
            return
        }
        expression = String(result)
        if commit {
            amount = String(result)
            dismiss()
        }
    }
}

struct MPAmountField: View {
    enum Presentation: Equatable {
        case standard
        case transaction
    }

    let title: String
    @Binding var text: String
    var error: String?
    var showsCalculator: Bool = true
    var presentation: Presentation = .standard
    var accessibilityIdentifier: String?
    @State private var calculatorPresented = false

    var body: some View {
        Group {
            if presentation == .transaction {
                HStack(alignment: .center, spacing: 10) {
                    Text("VND")
                        .font(MPFont.label)
                        .foregroundStyle(MPColor.secondary)
                        .padding(.horizontal, 10)
                        .frame(minHeight: 40)
                        .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
                        .overlay {
                            RoundedRectangle(cornerRadius: MPMetrics.cornerRadius)
                                .stroke(MPColor.line, lineWidth: MPMetrics.inputBorderWidth)
                        }
                        .accessibilityIdentifier("\(title)-currency")
                    VStack(alignment: .leading, spacing: 1) {
                        Text(title)
                            .font(MPFont.bodySmall)
                            .foregroundStyle(MPColor.secondary)
                        MPNumberPadTextField(
                            text: $text,
                            placeholder: "0",
                            accessibilityLabel: title,
                            presentation: .transactionAmount,
                            accessibilityIdentifier: accessibilityIdentifier
                        )
                        .frame(maxWidth: .infinity, minHeight: 50)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    if showsCalculator {
                        MPIconButton(systemName: "calculator", label: "Mở bàn phím số") {
                            calculatorPresented = true
                        }
                        .accessibilityIdentifier("amount-calculator-\(title)")
                    }
                }
                .padding(.vertical, 4)
            } else {
                VStack(alignment: .leading, spacing: 6) {
                    Text(title)
                        .font(MPFont.label)
                        .foregroundStyle(MPColor.ink)
                    HStack(alignment: .firstTextBaseline, spacing: 8) {
                        MPNumberPadTextField(
                            text: $text,
                            placeholder: "0",
                            accessibilityLabel: title,
                            accessibilityIdentifier: accessibilityIdentifier
                        )
                            .frame(maxWidth: .infinity, minHeight: MPMetrics.controlHeight)
                        MPText("VND", style: .secondary)
                            .accessibilityIdentifier("\(title)-currency")
                        if showsCalculator {
                            MPIconButton(systemName: "calculator", label: "Mở bàn phím số") {
                                calculatorPresented = true
                            }
                            .accessibilityIdentifier("amount-calculator-\(title)")
                        }
                    }
                }
            }
            if let error {
                Text(error)
                    .font(MPFont.bodySmall)
                    .foregroundStyle(MPColor.danger)
                    .accessibilityLabel("Lỗi: \(error)")
            }
        }
        .mpFullScreenCover(isPresented: $calculatorPresented) {
            MPAmountKeypad(amount: $text)
        }
    }
}

/// Reference transaction row used by high-frequency forms. The row owns the
/// icon/title/value/chevron geometry so selectors and inline actions do not
/// drift between screens.
struct MPFormRowLabel: View {
    let title: String
    let iconName: String
    let value: String
    var isEnabled: Bool = true
    var showsDisclosure: Bool = true

    var body: some View {
        HStack(spacing: 16) {
            Image(systemName: iconName)
                .font(MPFont.bodyEmphasis)
                .foregroundStyle(isEnabled ? MPColor.ink : MPColor.muted)
                .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(MPFont.body)
                    .foregroundStyle(isEnabled ? MPColor.ink : MPColor.muted)
                    .lineLimit(1)
                if !value.isEmpty {
                    Text(value)
                        .font(MPFont.bodySmall)
                        .foregroundStyle(isEnabled ? MPColor.secondary : MPColor.muted)
                        .lineLimit(1)
                }
            }
            Spacer(minLength: 8)
            if showsDisclosure {
                Image(systemName: "chevron.right")
                    .font(MPFont.labelSmall)
                    .foregroundStyle(isEnabled ? MPColor.secondary : MPColor.muted)
            }
        }
        .frame(minHeight: MPMetrics.controlHeight)
        .contentShape(Rectangle())
    }
}

/// Tappable semantic form row. Keep the action outside the label so a row can
/// be used as a Menu label without making the menu trigger itself a nested
/// button.
struct MPFormRow: View {
    static let minimumHitTarget = MPMetrics.hitTarget

    let title: String
    let iconName: String
    let value: String
    var isEnabled: Bool = true
    var accessibilityIdentifier: String?
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            MPFormRowLabel(title: title, iconName: iconName, value: value, isEnabled: isEnabled)
        }
        .buttonStyle(.plain)
        .disabled(!isEnabled)
        .accessibilityLabel(value.isEmpty ? title : "\(title): \(value)")
        .accessibilityIdentifier(accessibilityIdentifier ?? "form-row-\(title)")
    }
}

struct MPTextEditor: View {
    let title: String
    @Binding var text: String
    var placeholder: String = ""
    @FocusState private var isFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(MPFont.label)
                .foregroundStyle(MPColor.ink)
            ZStack(alignment: .topLeading) {
                if text.isEmpty {
                    Text(placeholder)
                        .font(MPFont.body)
                        .foregroundStyle(MPColor.secondary)
                        .padding(12)
                }
                TextEditor(text: $text)
                    .focused($isFocused)
                    .frame(minHeight: 110)
                    .padding(8)
                    .scrollContentBackground(.hidden)
            }
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
            .overlay {
                RoundedRectangle(cornerRadius: MPMetrics.cornerRadius)
                    .stroke(isFocused ? MPColor.accent : MPColor.line, lineWidth: isFocused ? MPMetrics.inputFocusWidth : MPMetrics.inputBorderWidth)
            }
            .shadow(color: isFocused ? MPColor.successLine.opacity(0.9) : .clear, radius: isFocused ? 3 : 0)
            .accessibilityLabel(title)
        }
    }
}

struct MPOptionalDateField: View {
    let title: String
    @Binding var value: String
    var timeZone: String = TimeZone.current.identifier
    let accessibilityIdentifier: String

    private var zone: TimeZone { TimeZone(identifier: timeZone) ?? .current }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            MPCheckboxRow(title: title, isSelected: !value.isEmpty, accessibilityIdentifier: "optional-date-toggle-\(accessibilityIdentifier)") {
                value = value.isEmpty ? MPFormat.todayISO(timeZone: zone) : ""
            }
            if value.isEmpty {
                MPText("Không giới hạn ngày", style: .secondary)
            } else {
                MPDateField(title: title, date: Binding(
                    get: { MPFormat.parseCalendarDate(value, timeZone: zone) ?? Date() },
                    set: { value = MPFormat.calendarDateString($0, timeZone: zone) }
                ), timeZone: timeZone, showsTitle: false)
            }
        }
    }
}

struct MPDateField: View {
    let title: String
    @Binding var date: Date
    var timeZone: String = TimeZone.current.identifier
    var showsTitle: Bool
    @State private var calendarPresented = false
    @State private var calendarSelection: Date

    init(title: String, date: Binding<Date>, timeZone: String = TimeZone.current.identifier, showsTitle: Bool = true) {
        self.title = title
        self._date = date
        self.timeZone = timeZone
        self.showsTitle = showsTitle
        _calendarSelection = State(initialValue: date.wrappedValue)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            if showsTitle {
                Text(title)
                    .font(MPFont.label)
                    .foregroundStyle(MPColor.ink)
            }
            MPButton(style: .plain, fullWidth: true) {
                calendarSelection = date
                calendarPresented = true
            } label: {
                HStack(spacing: 10) {
                    Image(systemName: "calendar")
                        .foregroundStyle(MPColor.action)
                    Text(MPDateStepper.displayLabel(for: date, timeZone: timeZone))
                        .font(MPFont.body)
                        .foregroundStyle(MPColor.ink)
                    Spacer(minLength: 0)
                    Image(systemName: "chevron.right")
                        .font(MPFont.label)
                        .foregroundStyle(MPColor.secondary)
                }
                .padding(.horizontal, MPMetrics.inputHorizontalPadding)
                .frame(minHeight: MPMetrics.controlHeight)
                .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
                .overlay {
                    RoundedRectangle(cornerRadius: MPMetrics.cornerRadius)
                        .stroke(MPColor.line, lineWidth: MPMetrics.inputBorderWidth)
                }
            }
            .accessibilityLabel(title)
            .accessibilityValue(MPDateStepper.displayLabel(for: date, timeZone: timeZone))
            .accessibilityIdentifier("date-field-\(title)")
        }
        .mpFullScreenCover(isPresented: $calendarPresented) {
            MPDateCalendarSheet(
                date: $calendarSelection,
                timeZone: timeZone,
                onCancel: { calendarPresented = false },
                onDone: {
                    date = calendarSelection
                    calendarPresented = false
                }
            )
        }
    }
}

/// Account-local date control matching the Stitch transaction editor. The
/// date label is exposed as a pure helper so contract tests can pin the
/// timezone/locale behavior without rendering a view.
struct MPDateStepper: View {
    let date: Date
    let timeZone: String
    let onPrevious: () -> Void
    let onNext: () -> Void
    let onSelect: ((Date) -> Void)?
    var accessibilityIdentifier: String = "date-stepper"
    @State private var calendarPresented = false
    @State private var calendarSelection: Date

    init(
        date: Date,
        timeZone: String,
        onPrevious: @escaping () -> Void,
        onNext: @escaping () -> Void,
        onSelect: ((Date) -> Void)? = nil,
        accessibilityIdentifier: String = "date-stepper"
    ) {
        self.date = date
        self.timeZone = timeZone
        self.onPrevious = onPrevious
        self.onNext = onNext
        self.onSelect = onSelect
        self.accessibilityIdentifier = accessibilityIdentifier
        _calendarSelection = State(initialValue: date)
    }

    static func displayLabel(for date: Date, timeZone: String) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.timeZone = TimeZone(identifier: timeZone) ?? .current
        formatter.dateFormat = "EEEE, dd/MM/yyyy"
        return formatter.string(from: date).capitalized(with: formatter.locale)
    }

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: "calendar")
                .font(MPFont.bodyEmphasis)
                .foregroundStyle(MPColor.ink)
                .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
            MPIconButton(systemName: "chevron.left", label: "Ngày trước", action: onPrevious)
                .accessibilityIdentifier(accessibilityIdentifier)
                .background(MPColor.row, in: Circle())
            MPButton(style: .plain, fullWidth: true) {
                calendarSelection = date
                calendarPresented = true
            } label: {
                Text(Self.displayLabel(for: date, timeZone: timeZone))
                    .font(MPFont.body)
                    .foregroundStyle(MPColor.accent)
                    .frame(maxWidth: .infinity, minHeight: MPMetrics.hitTarget)
                    .background(MPColor.row, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
                    .lineLimit(1)
                    .minimumScaleFactor(0.8)
            }
            .accessibilityLabel("Ngày giao dịch")
            .accessibilityIdentifier("\(accessibilityIdentifier)-calendar-button")
            MPIconButton(systemName: "chevron.right", label: "Ngày sau", action: onNext)
                .background(MPColor.row, in: Circle())
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier(accessibilityIdentifier)
        .mpFullScreenCover(isPresented: $calendarPresented) {
            MPDateCalendarSheet(
                date: $calendarSelection,
                timeZone: timeZone,
                onCancel: { calendarPresented = false },
                onDone: {
                    onSelect?(calendarSelection)
                    calendarPresented = false
                }
            )
        }
    }
}

struct MPCalendarDay: Identifiable, Equatable {
    let date: Date
    let key: String
    let dayNumber: String
    let isInDisplayedMonth: Bool

    var id: String { key }
}

/// Pure month/grid rules for the Stitch calendar surface. Date-only values are
/// represented at local noon to avoid DST midnight rollovers when they cross
/// the API boundary.
enum MPCalendarGrid {
    static let weekdayLabels = ["Thứ 2", "Thứ 3", "Thứ 4", "Thứ 5", "Thứ 6", "Thứ 7", "CN"]

    static func calendar(timeZone: TimeZone) -> Calendar {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        calendar.firstWeekday = 2
        calendar.minimumDaysInFirstWeek = 4
        return calendar
    }

    static func localNoon(_ date: Date, timeZone: TimeZone) -> Date {
        let calendar = calendar(timeZone: timeZone)
        let components = calendar.dateComponents([.year, .month, .day], from: date)
        return calendar.date(from: DateComponents(
            timeZone: timeZone,
            year: components.year,
            month: components.month,
            day: components.day,
            hour: 12
        )) ?? date
    }

    static func monthStart(for date: Date, timeZone: TimeZone) -> Date {
        let calendar = calendar(timeZone: timeZone)
        let local = localNoon(date, timeZone: timeZone)
        return calendar.date(from: calendar.dateComponents([.year, .month], from: local)) ?? local
    }

    static func dateKey(_ date: Date, timeZone: TimeZone) -> String {
        let formatter = DateFormatter()
        formatter.calendar = calendar(timeZone: timeZone)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = timeZone
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.string(from: date)
    }

    static func datePreservingClock(_ source: Date, on day: Date, timeZone: TimeZone) -> Date {
        let calendar = calendar(timeZone: timeZone)
        let clock = calendar.dateComponents([.hour, .minute, .second], from: source)
        var target = calendar.dateComponents([.year, .month, .day], from: day)
        target.hour = clock.hour ?? 12
        target.minute = clock.minute ?? 0
        target.second = clock.second ?? 0
        return calendar.date(from: target) ?? day
    }

    static func monthTitle(for date: Date, timeZone: TimeZone) -> String {
        let formatter = DateFormatter()
        formatter.calendar = calendar(timeZone: timeZone)
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.timeZone = timeZone
        formatter.dateFormat = "MMM yyyy"
        return formatter.string(from: monthStart(for: date, timeZone: timeZone)).lowercased(with: formatter.locale)
    }

    static func days(for date: Date, timeZone: TimeZone) -> [MPCalendarDay] {
        let calendar = calendar(timeZone: timeZone)
        let month = monthStart(for: date, timeZone: timeZone)
        let weekday = calendar.component(.weekday, from: month)
        let leadingDays = (weekday - calendar.firstWeekday + 7) % 7
        let gridStart = calendar.date(byAdding: .day, value: -leadingDays, to: month) ?? month
        let displayedMonth = calendar.component(.month, from: month)
        return (0..<42).compactMap { offset in
            guard let rawDate = calendar.date(byAdding: .day, value: offset, to: gridStart) else { return nil }
            let localDate = localNoon(rawDate, timeZone: timeZone)
            let components = calendar.dateComponents([.month, .day], from: localDate)
            guard let day = components.day else { return nil }
            return MPCalendarDay(
                date: localDate,
                key: dateKey(localDate, timeZone: timeZone),
                dayNumber: String(day),
                isInDisplayedMonth: components.month == displayedMonth
            )
        }
    }
}

/// Shared Stitch-style calendar surface opened by account-local date controls.
/// The feature owns only the selected-date effect; month arithmetic, day grid,
/// optional hour/minute selection, title, cancel/done layout and accessibility
/// IDs stay centralized here.
struct MPDateCalendarSheet: View {
    @Binding var date: Date
    let timeZone: String
    let includesTime: Bool
    let onCancel: () -> Void
    let onDone: () -> Void
    @State private var visibleMonth: Date

    init(
        date: Binding<Date>,
        timeZone: String,
        includesTime: Bool = false,
        onCancel: @escaping () -> Void,
        onDone: @escaping () -> Void
    ) {
        self._date = date
        self.timeZone = timeZone
        self.includesTime = includesTime
        self.onCancel = onCancel
        self.onDone = onDone
        _visibleMonth = State(initialValue: MPCalendarGrid.monthStart(
            for: date.wrappedValue,
            timeZone: TimeZone(identifier: timeZone) ?? .current
        ))
    }

    private var resolvedTimeZone: TimeZone {
        TimeZone(identifier: timeZone) ?? .current
    }

    private var calendar: Calendar {
        MPCalendarGrid.calendar(timeZone: resolvedTimeZone)
    }

    private var days: [MPCalendarDay] {
        MPCalendarGrid.days(for: visibleMonth, timeZone: resolvedTimeZone)
    }

    private var selectedKey: String {
        MPCalendarGrid.dateKey(date, timeZone: resolvedTimeZone)
    }

    private var todayKey: String {
        MPCalendarGrid.dateKey(Date(), timeZone: resolvedTimeZone)
    }

    var body: some View {
        ZStack {
            MPColor.ink.opacity(0.34)
                .ignoresSafeArea()
            MPSurfaceCard(style: .reference) {
                VStack(spacing: 16) {
                    HStack(spacing: 12) {
                        MPButton(style: .plain, fullWidth: false, action: onCancel) {
                            Text("Hủy")
                                .font(MPFont.body)
                                .foregroundStyle(MPColor.ink)
                        }
                        .padding(.horizontal, 12)
                        .background(MPColor.row, in: Capsule())
                        .accessibilityIdentifier("screen-close")
                        Spacer(minLength: 0)
                        MPText(includesTime ? "Chọn ngày và giờ" : "Chọn ngày", style: .title)
                            .accessibilityIdentifier(includesTime ? "screen-title-Chọn ngày và giờ" : "screen-title-Chọn ngày")
                        Spacer(minLength: 0)
                        Color.clear.frame(width: 44, height: MPMetrics.hitTarget)
                            .accessibilityHidden(true)
                    }

                    HStack(spacing: 12) {
                        MPIconButton(systemName: "chevron.left", label: "Tháng trước") {
                            visibleMonth = calendar.date(byAdding: .month, value: -1, to: visibleMonth) ?? visibleMonth
                        }
                        .background(MPColor.row, in: Circle())
                        .accessibilityIdentifier("date-calendar-previous")
                        Spacer(minLength: 0)
                        MPText(MPCalendarGrid.monthTitle(for: visibleMonth, timeZone: resolvedTimeZone), style: .title)
                            .foregroundStyle(MPColor.action)
                            .accessibilityIdentifier("date-calendar-month-title")
                        Spacer(minLength: 0)
                        MPIconButton(systemName: "chevron.right", label: "Tháng sau") {
                            visibleMonth = calendar.date(byAdding: .month, value: 1, to: visibleMonth) ?? visibleMonth
                        }
                        .background(MPColor.row, in: Circle())
                        .accessibilityIdentifier("date-calendar-next")
                    }

                    LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 4), count: 7), spacing: 4) {
                        ForEach(MPCalendarGrid.weekdayLabels, id: \.self) { label in
                            Text(label)
                                .font(MPFont.bodySmall.weight(.medium))
                                .foregroundStyle(MPColor.ink)
                                .frame(maxWidth: .infinity, minHeight: 30)
                        }
                    }

                    LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 4), count: 7), spacing: 4) {
                        ForEach(days) { day in
                            let isSelected = day.key == selectedKey
                            let isToday = day.key == todayKey
                            Button {
                                select(day)
                            } label: {
                                Text(day.dayNumber)
                                    .font(MPFont.body)
                                    .foregroundStyle(isSelected ? MPColor.surface : (day.isInDisplayedMonth ? MPColor.ink : MPColor.muted))
                                    .frame(maxWidth: .infinity, minHeight: 42)
                                    .background {
                                        Circle()
                                            .fill(isSelected ? MPColor.brand : (isToday ? MPColor.row : .clear))
                                            .frame(width: 40, height: 40)
                                    }
                                    .overlay {
                                        if isToday && !isSelected {
                                            Circle()
                                                .stroke(MPColor.brand, lineWidth: 1.5)
                                                .frame(width: 38, height: 38)
                                        }
                                    }
                            }
                            .buttonStyle(.plain)
                            .accessibilityLabel("Ngày (day.key)")
                            .accessibilityIdentifier("date-calendar-day-\(day.key)")
                            .accessibilityAddTraits(isSelected ? .isSelected : [])
                        }
                    }
                    .accessibilityIdentifier("date-calendar-picker")

                    if includesTime {
                        DatePicker("Giờ", selection: $date, displayedComponents: [.hourAndMinute])
                            .datePickerStyle(.wheel)
                            .labelsHidden()
                            .frame(height: 116)
                            .clipped()
                            .environment(\.timeZone, resolvedTimeZone)
                            .accessibilityIdentifier("date-time-calendar-time")
                    }

                    HStack {
                        MPButton("Hủy", style: .secondary, fullWidth: false, action: onCancel)
                            .accessibilityIdentifier("date-calendar-cancel")
                        Spacer(minLength: 8)
                        MPButton("Xong", action: onDone)
                            .accessibilityIdentifier("date-calendar-done")
                    }
                }
            }
            .padding(.horizontal, 20)
        }
        .background(MPColor.canvas.ignoresSafeArea())
    }

    private func select(_ day: MPCalendarDay) {
        guard includesTime else {
            date = day.date
            return
        }
        date = MPCalendarGrid.datePreservingClock(date, on: day.date, timeZone: resolvedTimeZone)
    }
}

struct MPDateTimeField: View {
    let title: String
    @Binding var date: Date
    let timeZone: String
    @State private var pickerPresented = false
    @State private var selection: Date

    init(title: String, date: Binding<Date>, timeZone: String = TimeZone.current.identifier) {
        self.title = title
        self._date = date
        self.timeZone = timeZone
        _selection = State(initialValue: date.wrappedValue)
    }

    static func displayLabel(for date: Date, timeZone: String) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.timeZone = TimeZone(identifier: timeZone) ?? .current
        formatter.dateFormat = "EEEE, dd/MM/yyyy HH:mm"
        return formatter.string(from: date).capitalized(with: formatter.locale)
    }

    var body: some View {
        MPButton(style: .plain, fullWidth: true) {
            selection = date
            pickerPresented = true
        } label: {
            HStack(spacing: 10) {
                Image(systemName: "calendar.badge.clock")
                    .foregroundStyle(MPColor.ink)
                    .frame(width: MPMetrics.hitTarget, height: MPMetrics.hitTarget)
                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .font(MPFont.bodySmall)
                        .foregroundStyle(MPColor.secondary)
                    Text(Self.displayLabel(for: date, timeZone: timeZone))
                        .font(MPFont.body)
                        .foregroundStyle(MPColor.accent)
                        .lineLimit(1)
                        .minimumScaleFactor(0.8)
                }
                Spacer(minLength: 8)
                Image(systemName: "chevron.right")
                    .foregroundStyle(MPColor.secondary)
            }
            .padding(.horizontal, 12)
            .frame(maxWidth: .infinity, minHeight: MPMetrics.controlHeight, alignment: .leading)
            .background(MPColor.row, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .accessibilityLabel(title)
        .accessibilityIdentifier("datetime-field-\(title)")
        .mpFullScreenCover(isPresented: $pickerPresented) {
            MPDateCalendarSheet(
                date: $selection,
                timeZone: timeZone,
                includesTime: true,
                onCancel: { pickerPresented = false },
                onDone: {
                    date = selection
                    pickerPresented = false
                }
            )
        }
    }
}

struct MPDateRangePicker: View {
    @Binding var start: Date
    @Binding var end: Date
    let timeZone: String

    init(start: Binding<Date>, end: Binding<Date>, timeZone: String = TimeZone.current.identifier) {
        self._start = start
        self._end = end
        self.timeZone = timeZone
    }

    var body: some View {
        VStack(spacing: 8) {
            MPDateField(title: "Từ ngày", date: $start, timeZone: timeZone)
            MPDateField(title: "Đến ngày", date: $end, timeZone: timeZone)
        }
    }
}

/// Keeps wire values stable while presenting the semantic labels used by the
/// Vietnamese web UI. Unknown values intentionally fall back to the raw value
/// so a newly-added backend enum is still visible and selectable.
enum MPSelectionLabels {
    static func display(for value: String) -> String {
        switch value {
        case "expense": return "Chi"
        case "income": return "Thu"
        case "basic": return "Thông thường"
        case "goal": return "Tiết kiệm"
        case "credit": return "Tín dụng"
        case "none": return "Không đặt"
        case "fixed": return "Số tiền cố định"
        case "percent": return "Phần trăm thu nhập"
        case "daily": return "Mỗi ngày"
        case "weekly": return "Mỗi tuần"
        case "monthly": return "Mỗi tháng"
        case "yearly": return "Mỗi năm"
        case "buy": return "Mua"
        case "sell": return "Bán"
        case "purchase": return "Mua hàng"
        case "refund": return "Hoàn tiền"
        case "fee": return "Phí"
        case "interest": return "Lãi"
        case "payment": return "Thanh toán"
        case "paid": return "Đã thanh toán"
        case "partial": return "Đã trả một phần"
        case "overdue": return "Đã quá hạn"
        case "due": return "Đang đến hạn"
        case "not_configured": return "Chưa cấu hình"
        case "open": return "Mở"
        case "triaged": return "Đã phân loại"
        case "in_progress": return "Đang xử lý"
        case "rejected": return "Từ chối"
        case "increase": return "Tăng"
        case "decrease": return "Giảm"
        case "bug": return "Lỗi"
        case "feature": return "Tính năng"
        case "improvement": return "Cải thiện"
        case "finance:read": return "Đọc tài chính"
        case "advisor:read": return "Đọc trợ lý"
        case "advisor:chat": return "Chat trợ lý"
        case "feedback:read": return "Đọc phản hồi"
        case "feedback:write": return "Gửi phản hồi"
        default: return value
        }
    }

    static func creditKind(_ value: String?) -> String {
        guard let value else { return "Giao dịch tín dụng" }
        return display(for: value)
    }

    static func feedbackStatus(_ value: String) -> String {
        switch value {
        case "open": return "Mở"
        case "triaged": return "Đã phân loại"
        case "in_progress": return "Đang xử lý"
        case "fixed": return "Đã xử lý"
        case "rejected": return "Từ chối"
        default: return value
        }
    }
}

struct MPWalletSelector: View {
    let title: String
    let options: [String]
    @Binding var selection: String

    var body: some View {
        Menu {
            ForEach(options, id: \.self) { option in
                Button(MPSelectionLabels.display(for: option)) { selection = option }
            }
        } label: {
            HStack {
                Text(title)
                Spacer()
                Text(MPSelectionLabels.display(for: selection))
                    .foregroundStyle(MPColor.secondary)
                Image(systemName: "chevron.up.chevron.down")
                    .font(MPFont.labelSmall)
            }
            .font(MPFont.body)
            .frame(minHeight: MPMetrics.controlHeight)
            .padding(.horizontal, 12)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel("\(title): \(MPSelectionLabels.display(for: selection))")
    }
}

struct MPSelectionOption: Identifiable, Hashable {
    let id: String
    let label: String
}

/// Selector dành cho dữ liệu có identity thật. Không dùng label làm khóa vì
/// người dùng có thể tạo hai ví hoặc danh mục trùng tên.
struct MPOptionSelector: View {
    let title: String
    let options: [MPSelectionOption]
    @Binding var selectionID: String
    var emptyLabel: String = "Chưa chọn"
    var iconName: String = "chevron.right"
    var accessibilityIdentifier: String?
    var isEnabled: Bool = true

    private var selectedLabel: String {
        options.first(where: { $0.id == selectionID })?.label ?? emptyLabel
    }

    var body: some View {
        Menu {
            ForEach(options) { option in
                Button(option.label) { selectionID = option.id }
            }
        } label: {
            MPFormRowLabel(
                title: title,
                iconName: iconName,
                value: selectedLabel,
                isEnabled: isEnabled && !options.isEmpty,
                showsDisclosure: isEnabled && !options.isEmpty
            )
        }
        .foregroundStyle(MPColor.ink)
        .disabled(options.isEmpty || !isEnabled)
        .accessibilityLabel("\(title): \(selectedLabel)")
        .accessibilityIdentifier(accessibilityIdentifier ?? "form-row-\(title)")
    }
}

/// Menu primitive for selectors whose visual label is a reference capsule or
/// another domain surface rather than the standard form-row label.
struct MPMenuSelector<Label: View>: View {
    let options: [MPSelectionOption]
    let onSelect: (String) -> Void
    private let label: () -> Label

    init(
        options: [MPSelectionOption],
        onSelect: @escaping (String) -> Void,
        @ViewBuilder label: @escaping () -> Label
    ) {
        self.options = options
        self.onSelect = onSelect
        self.label = label
    }

    var body: some View {
        Menu {
            ForEach(options) { option in
                Button(option.label) { onSelect(option.id) }
            }
        } label: {
            label()
        }
        .foregroundStyle(MPColor.ink)
    }
}

/// Compact wallet-scope control used by ordinary ledger screens. It keeps the
/// centered capsule anatomy from the web ledger while delegating menu
/// interaction and wallet identity to shared design-system components.
struct MPWalletScopeSelector: View {
    let wallets: [WalletModel]
    @Binding var selectionID: String
    var centered: Bool = true

    private var selectedWallet: WalletModel? {
        wallets.first(where: { $0.id == selectionID })
    }

    private var selectedName: String {
        selectedWallet?.name ?? "Tổng cộng"
    }

    private var options: [MPSelectionOption] {
        [MPSelectionOption(id: "__all__", label: "Tổng cộng")]
            + wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }
    }

    var body: some View {
        HStack {
            if centered { Spacer(minLength: 0) }
            MPMenuSelector(options: options) { selectionID = $0 } label: {
                HStack(spacing: 8) {
                    walletIcon
                    MPText(selectedName, style: .emphasis)
                        .lineLimit(1)
                        .truncationMode(.tail)
                        .frame(maxWidth: 176, alignment: .leading)
                    MPIcon(systemName: "chevron.down", color: MPColor.ink, size: 16)
                }
                .padding(.horizontal, 12)
                .frame(minHeight: 44)
                .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.buttonCornerRadius))
                .contentShape(Rectangle())
            }
            .accessibilityLabel("Ví: \(selectedName)")
            .accessibilityIdentifier("transaction-wallet-selector")
            if centered { Spacer(minLength: 0) }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    private var walletIcon: some View {
        if let selectedWallet {
            MPWalletScopeGlyph(type: selectedWallet.type)
        } else {
            MPWalletScopeGlyph(type: nil)
        }
    }
}

/// Base selector for a boolean product decision that must remain explicit in
/// a review form. The wire value stays a Bool while the UI uses localized
/// labels and the same selector surface as other option fields.
struct MPBooleanSelector: View {
    let title: String
    @Binding var value: Bool
    var trueLabel: String = "Có"
    var falseLabel: String = "Không"

    private var selectionID: Binding<String> {
        Binding(
            get: { value ? "true" : "false" },
            set: { value = $0 == "true" }
        )
    }

    var body: some View {
        MPOptionSelector(
            title: title,
            options: [
                MPSelectionOption(id: "true", label: trueLabel),
                MPSelectionOption(id: "false", label: falseLabel)
            ],
            selectionID: selectionID
        )
    }
}

/// Multi-select selector for relationships that may apply to more than one
/// wallet. An empty selection means "all wallets", matching the API contract.
struct MPMultiOptionSelector: View {
    let title: String
    let options: [MPSelectionOption]
    @Binding var selectionIDs: Set<String>
    var allLabel: String = "Tất cả"

    private var selectedLabel: String {
        if selectionIDs.isEmpty { return allLabel }
        return "Đã chọn \(selectionIDs.count)"
    }

    var body: some View {
        Menu {
            ForEach(options) { option in
                Button {
                    if selectionIDs.contains(option.id) {
                        selectionIDs.remove(option.id)
                    } else {
                        selectionIDs.insert(option.id)
                    }
                } label: {
                    HStack {
                        Text(option.label)
                        Spacer()
                        if selectionIDs.contains(option.id) {
                            Image(systemName: "checkmark")
                        }
                    }
                    .font(MPFont.body)
                }
            }
            if !selectionIDs.isEmpty {
                Divider()
                Button("Bỏ chọn (tất cả ví)") { selectionIDs.removeAll() }
            }
        } label: {
            HStack {
                Text(title)
                Spacer()
                Text(selectedLabel)
                    .foregroundStyle(MPColor.secondary)
                Image(systemName: "checklist")
                    .font(MPFont.labelSmall)
            }
            .font(MPFont.body)
            .frame(minHeight: MPMetrics.controlHeight)
            .padding(.horizontal, 12)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel("\(title): \(selectedLabel)")
    }
}

struct MPCategoryIconOption: Identifiable, Hashable {
    let id: String
    let label: String
    let systemName: String
    let tone: String
}

enum MPCategoryIconCatalog {
    static let options: [MPCategoryIconOption] = [
        MPCategoryIconOption(id: "tag", label: "Nhãn", systemName: "tag.fill", tone: "categorySlate"),
        MPCategoryIconOption(id: "expense_food", label: "Ăn uống", systemName: "fork.knife", tone: "categoryOrange"),
        MPCategoryIconOption(id: "expense_bills", label: "Hóa đơn", systemName: "receipt", tone: "categoryBlue"),
        MPCategoryIconOption(id: "expense_shopping", label: "Mua sắm", systemName: "bag.fill", tone: "categoryViolet"),
        MPCategoryIconOption(id: "expense_family", label: "Gia đình", systemName: "house.fill", tone: "categoryRose"),
        MPCategoryIconOption(id: "expense_transport", label: "Di chuyển", systemName: "car.fill", tone: "categorySky"),
        MPCategoryIconOption(id: "expense_health", label: "Sức khỏe", systemName: "heart.text.square.fill", tone: "categoryRed"),
        MPCategoryIconOption(id: "expense_entertainment", label: "Giải trí", systemName: "gamecontroller.fill", tone: "categoryFuchsia"),
        MPCategoryIconOption(id: "expense_savings", label: "Tiết kiệm", systemName: "banknote.fill", tone: "categoryLime"),
        MPCategoryIconOption(id: "expense_travel", label: "Du lịch", systemName: "airplane", tone: "categoryCyan")
    ]

    static func option(for key: String) -> MPCategoryIconOption {
        options.first(where: { $0.id == key }) ?? options[0]
    }
}

/// Shared category identity picker. The feature layer owns only the selected
/// API key; presentation, labels, and the reference grid stay in the design
/// system so web/native category forms cannot drift independently.
struct MPCategoryIconSelector: View {
    let title: String
    @Binding var selection: String
    var accessibilityIdentifier: String = "category-icon-selector"
    @State private var isPresented = false

    private var selected: MPCategoryIconOption { MPCategoryIconCatalog.option(for: selection) }

    var body: some View {
        MPButton(style: .secondary, fullWidth: true) {
            isPresented = true
        } label: {
            HStack(spacing: 10) {
                MPTransactionGlyph(systemName: selected.systemName, tone: selected.tone)
                    .scaleEffect(0.72)
                    .frame(width: 34, height: 34)
                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .font(MPFont.label)
                    Text(selected.label)
                        .font(MPFont.bodySmall)
                        .foregroundStyle(MPColor.secondary)
                }
                Spacer(minLength: 0)
                Image(systemName: "chevron.right")
                    .font(MPFont.labelSmall)
                    .foregroundStyle(MPColor.secondary)
            }
        }
        .accessibilityLabel("\(title): \(selected.label)")
        .accessibilityIdentifier(accessibilityIdentifier)
        .mpFullScreenCover(isPresented: $isPresented) {
            MPCategoryIconPickerSheet(title: title, selection: $selection) {
                isPresented = false
            }
        }
    }
}

private struct MPCategoryIconPickerSheet: View {
    let title: String
    @Binding var selection: String
    let onClose: () -> Void

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPReferenceScreenHeader(title: "Chọn biểu tượng", closeLabel: "Đóng", onClose: onClose)
                MPSurfaceCard(style: .reference) {
                    LazyVGrid(columns: [GridItem(.adaptive(minimum: 86), spacing: 12)], spacing: 12) {
                        ForEach(MPCategoryIconCatalog.options) { option in
                            MPButton(
                                style: option.id == selection ? .segmentedSelected : .outline,
                                fullWidth: true,
                                action: {
                                    selection = option.id
                                    onClose()
                                }
                            ) {
                                VStack(spacing: 6) {
                                    MPTransactionGlyph(systemName: option.systemName, tone: option.tone)
                                        .scaleEffect(0.82)
                                        .frame(width: 42, height: 42)
                                    Text(option.label)
                                        .font(MPFont.labelSmall)
                                        .lineLimit(1)
                                }
                            }
                            .accessibilityLabel(option.label)
                            .accessibilityIdentifier("category-icon-\(option.id)")
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .safeAreaInset(edge: .bottom) { MPBottomNavigationInset() }
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Chọn biểu tượng cho \(title)")
    }
}
