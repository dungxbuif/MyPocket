import {
  ArrowDownLeft,
  ArrowUpRight,
  BriefcaseBusiness,
  Camera,
  Car,
  CircleDollarSign,
  Coffee,
  CreditCard,
  Landmark,
  Plane,
  ReceiptText,
  ShoppingBag,
  Target,
  Utensils,
  Wallet,
  type LucideIcon,
} from "lucide-react";

export type WalletKind = "cash" | "bank" | "credit" | "goal";
export type TransactionKind = "expense" | "income" | "transfer" | "debt";

export type MockWallet = {
  id: string;
  name: string;
  kind: WalletKind;
  balance: number;
  subtitle: string;
  icon: LucideIcon;
};

export type MockTransaction = {
  id: string;
  title: string;
  category: string;
  wallet: string;
  amount: number;
  kind: TransactionKind;
  occurredLabel: string;
  note?: string;
  icon: LucideIcon;
};

export type MockBudget = {
  id: string;
  name: string;
  spent: number;
  limit: number;
  icon: LucideIcon;
};

export type MockGoal = {
  id: string;
  name: string;
  saved: number;
  target: number;
  icon: LucideIcon;
};

export type MockCategoryNode = {
  id: string;
  name: string;
  amount: number;
  icon: LucideIcon;
  children?: Array<{ id: string; name: string; amount: number; icon: LucideIcon }>;
};

export const wallets: MockWallet[] = [
  { id: "cash", name: "Tiền mặt", kind: "cash", balance: 2450000, subtitle: "Chi tiêu hằng ngày", icon: Wallet },
  { id: "tcb", name: "Techcombank", kind: "bank", balance: 18450000, subtitle: "Lương và thanh toán", icon: Landmark },
  { id: "credit", name: "Thẻ tín dụng", kind: "credit", balance: -3250000, subtitle: "Sao kê ngày 22", icon: CreditCard },
  { id: "travel", name: "Quỹ du lịch", kind: "goal", balance: 7200000, subtitle: "Đà Lạt tháng 11", icon: Plane },
];

export const transactions: MockTransaction[] = [
  { id: "t1", title: "Cơm trưa văn phòng", category: "Ăn uống", wallet: "Tiền mặt", amount: -68000, kind: "expense", occurredLabel: "Hôm nay", icon: Utensils },
  { id: "t2", title: "Lương tháng 9", category: "Thu nhập", wallet: "Techcombank", amount: 28000000, kind: "income", occurredLabel: "Hôm nay", icon: ArrowDownLeft },
  { id: "t3", title: "Cà phê gặp khách", category: "Cà phê", wallet: "Tiền mặt", amount: -95000, kind: "expense", occurredLabel: "Hôm qua", note: "Gợi ý AI: công việc", icon: Coffee },
  { id: "t4", title: "Trả góp laptop", category: "Nợ/Vay", wallet: "Thẻ tín dụng", amount: -1200000, kind: "debt", occurredLabel: "Hôm qua", icon: CreditCard },
  { id: "t5", title: "Chuyển sang quỹ du lịch", category: "Chuyển ví", wallet: "Techcombank", amount: -500000, kind: "transfer", occurredLabel: "Thứ 5", icon: ArrowUpRight },
  { id: "t6", title: "Grab đi họp", category: "Đi lại", wallet: "Techcombank", amount: -124000, kind: "expense", occurredLabel: "Thứ 5", icon: Car },
];

export const budgets: MockBudget[] = [
  { id: "food", name: "Ăn uống", spent: 3850000, limit: 5200000, icon: Utensils },
  { id: "shopping", name: "Mua sắm", spent: 2450000, limit: 2400000, icon: ShoppingBag },
  { id: "coffee", name: "Cà phê", spent: 640000, limit: 900000, icon: Coffee },
  { id: "transport", name: "Đi lại", spent: 890000, limit: 1600000, icon: Car },
];

export const goals: MockGoal[] = [
  { id: "travel-goal", name: "Đà Lạt cuối năm", saved: 7200000, target: 12000000, icon: Plane },
  { id: "laptop-goal", name: "Laptop làm việc", saved: 13500000, target: 26000000, icon: Target },
];

export const categoryTree: MockCategoryNode[] = [
  {
    id: "food",
    name: "Ăn uống",
    amount: 3850000,
    icon: Utensils,
    children: [
      { id: "lunch", name: "Cơm trưa", amount: 1680000, icon: Utensils },
      { id: "coffee", name: "Cà phê", amount: 640000, icon: Coffee },
    ],
  },
  {
    id: "shopping",
    name: "Mua sắm",
    amount: 2450000,
    icon: ShoppingBag,
    children: [
      { id: "clothes", name: "Quần áo", amount: 1350000, icon: ShoppingBag },
      { id: "gear", name: "Đồ làm việc", amount: 1100000, icon: BriefcaseBusiness },
    ],
  },
  {
    id: "finance",
    name: "Tài chính",
    amount: 1200000,
    icon: CircleDollarSign,
    children: [{ id: "debt", name: "Trả góp", amount: 1200000, icon: CreditCard }],
  },
];

export const insights = [
  "Ăn uống đang thấp hơn tuần trước 12%. Nếu giữ nhịp này, cuối tháng còn khoảng 2.1 triệu để chuyển vào quỹ du lịch.",
  "AI gợi ý 3 giao dịch cà phê nên gắn nhãn Công việc để báo cáo cá nhân chính xác hơn.",
  "Sắp tới hạn trả thẻ tín dụng trong 6 ngày. Cần giữ tối thiểu 3.25 triệu trong Techcombank.",
];

export const reportBars = [38, 54, 42, 76, 61, 88, 49, 66, 72, 45, 58, 91];

export const categoryShares = [
  { name: "Ăn uống", value: 42, color: "var(--color-action)" },
  { name: "Mua sắm", value: 26, color: "var(--color-danger)" },
  { name: "Đi lại", value: 17, color: "var(--color-secondary)" },
  { name: "Khác", value: 15, color: "var(--color-accent)" },
];

export const quickActions = [
  { label: "OCR hóa đơn", icon: Camera },
  { label: "Khoản nợ", icon: CreditCard },
  { label: "Mục tiêu", icon: Target },
  { label: "Báo cáo", icon: ReceiptText },
];
