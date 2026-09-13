import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { BaseButton } from "../atoms/BaseButton";
import { useState } from "react";
import { CalendarDays, Camera, Utensils, Wallet, X } from "lucide-react";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { FormSelectorRow } from "../molecules/FormSelectorRow";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
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
    <BaseBottomSheet title="Thêm giao dịch" closeLabel="Đóng thêm giao dịch" onClose={onClose}>
        <div className="mb-4 flex justify-end">
          <BaseButton variant="primary" size="md" className="" onClick={onClose}>
            Lưu
          </BaseButton>
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
        <SurfaceCard padding="md" className="mt-4">
          <Text size="xs" weight="semibold" tone="secondary" className="uppercase tracking-[0.12em]">Số tiền</Text>
          <div className="mt-1 flex items-end justify-between">
            <Text numeric size="4xl" weight="bold" className="">{amount ? formatVND(Number(amount.replace(/\D/g, ""))) : "0 đ"}</Text>
            <BaseButton variant="row" size="row" className="">VND</BaseButton>
          </div>
        </SurfaceCard>
        <div className="mt-3 grid grid-cols-3 gap-2">
          {["50.000", "150.000", "500.000"].map((chip) => (
            <BaseButton variant="chip" size="md" key={chip} onClick={() => setAmount(chip.replace(/\D/g, ""))} className="">
              {chip}
            </BaseButton>
          ))}
        </div>
        <SurfaceCard padding="sm" className="mt-3">
          <FormSelectorRow label="Danh mục" value="Ăn uống" icon={Utensils} />
          <FormSelectorRow label="Ví" value="Tiền mặt" icon={Wallet} />
          <FormSelectorRow label="Ngày" value="Hôm nay" icon={CalendarDays} />
          <FormSelectorRow label="Hóa đơn" value="Thêm ảnh" icon={Camera} />
        </SurfaceCard>
        <div className="mt-3 grid grid-cols-3 gap-2">
          {keys.map((key) => (
            <BaseButton variant="key" size="lg" key={key} onClick={() => append(key)} className="">
              {key === "back" ? "⌫" : key}
            </BaseButton>
          ))}
          <BaseButton variant="danger" size="md" onClick={() => append("clear")} className="col-span-2">
            Xóa
          </BaseButton>
          <BaseButton variant="primary" size="md" onClick={onClose} className="">
            Xong
          </BaseButton>
        </div>
    </BaseBottomSheet>
  );
}
