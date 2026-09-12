import { useState } from "react";
import { CalendarDays, Camera, Utensils, Wallet, X } from "lucide-react";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { FormSelectorRow } from "../molecules/FormSelectorRow";
import { formatVND } from "../utils/format";

type AddType = "expense" | "income" | "transfer" | "debt";

export function QuickAddSheet({ onClose }: { onClose: () => void }) {
  const [type, setType] = useState<AddType>("expense");
  const [amount, setAmount] = useState("250000");
  const keys = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "+", "0", "back"];

  function append(input: string) {
    setAmount((current) => {
      if (input === "clear") return "";
      if (input === "back") return current.slice(0, -1);
      if (input === "+") return current;
      return `${current}${input}`.replace(/^0+(?=\d)/, "");
    });
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/35">
      <section className="w-full max-w-[430px] rounded-t-[28px] bg-[#fbf9f9] p-4 pb-[calc(env(safe-area-inset-bottom)+16px)]">
        <div className="mb-4 flex items-center justify-between">
          <button className="rounded-full p-2 text-[#3f4a3c]" onClick={onClose}>
            <X size={20} />
          </button>
          <p className="font-bold">Thêm giao dịch</p>
          <button className="rounded-full bg-[#006e1c] px-4 py-2 text-sm font-bold text-white" onClick={onClose}>
            Lưu
          </button>
        </div>
        <SegmentedControl
          value={type}
          onChange={setType}
          options={[
            { value: "expense", label: "Chi" },
            { value: "income", label: "Thu" },
            { value: "transfer", label: "Chuyển" },
            { value: "debt", label: "Vay/Nợ" },
          ]}
        />
        <div className="mt-4 rounded-3xl bg-white p-4">
          <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[#3f4a3c]">Số tiền</p>
          <div className="mt-1 flex items-end justify-between">
            <p className="money text-4xl font-bold">{amount ? formatVND(Number(amount.replace(/\D/g, ""))) : "0 đ"}</p>
            <button className="rounded-full bg-[#f5f3f3] px-3 py-1 text-sm font-bold">VND</button>
          </div>
        </div>
        <div className="mt-3 grid grid-cols-3 gap-2">
          {["50.000", "150.000", "500.000"].map((chip) => (
            <button key={chip} onClick={() => setAmount(chip.replace(/\D/g, ""))} className="rounded-full bg-white py-2 text-sm font-bold">
              {chip}
            </button>
          ))}
        </div>
        <div className="mt-3 rounded-3xl bg-white p-2">
          <FormSelectorRow label="Danh mục" value="Ăn uống" icon={Utensils} />
          <FormSelectorRow label="Ví" value="Tiền mặt" icon={Wallet} />
          <FormSelectorRow label="Ngày" value="Hôm nay" icon={CalendarDays} />
          <FormSelectorRow label="Hóa đơn" value="Thêm ảnh" icon={Camera} />
        </div>
        <div className="mt-3 grid grid-cols-3 gap-2">
          {keys.map((key) => (
            <button key={key} onClick={() => append(key)} className="h-14 rounded-2xl bg-white text-xl font-bold">
              {key === "back" ? "⌫" : key}
            </button>
          ))}
          <button onClick={() => append("clear")} className="col-span-2 h-14 rounded-2xl bg-[#ffdad6] font-bold text-[#93000a]">
            Xóa
          </button>
          <button onClick={onClose} className="h-14 rounded-2xl bg-[#006e1c] font-bold text-white">
            Xong
          </button>
        </div>
      </section>
    </div>
  );
}

