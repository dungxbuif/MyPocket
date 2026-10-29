import React, { useState } from "react";
import { createRoot } from "react-dom/client";
import { BaseButton } from "../src/atomic/atoms/BaseButton";
import { SurfaceCard } from "../src/atomic/atoms/SurfaceCard";
import { SegmentedControl } from "../src/atomic/atoms/SegmentedControl";
import { BaseTextInput, BaseSelect, FormField } from "../src/atomic/atoms/FormField";
import { BaseBottomSheet } from "../src/atomic/molecules/BaseBottomSheet";
import { BudgetGauge, Progress } from "../src/atomic/atoms/Progress";
import { BaseCategoryTree } from "../src/atomic/molecules/BaseCategoryTree";
import "../src/styles.css";

function Fixture() {
  const [value, setValue] = useState("expense"), [sheet, setSheet] = useState(false);
  return <main className="mx-auto max-w-phone space-y-4 p-4">
    <SurfaceCard padding="md"><BaseButton id="primary">Lưu</BaseButton><BaseButton loading id="loading">Lưu</BaseButton><BaseButton variant="danger">Xóa</BaseButton></SurfaceCard>
    <SegmentedControl value={value} onChange={setValue} options={[{value:"expense",label:"Khoản chi"},{value:"income",label:"Khoản thu"},{value:"debt",label:"Vay/Nợ"}]} />
    <SurfaceCard padding="md"><FormField label="Tên ví"><BaseTextInput defaultValue="Tiền mặt" /></FormField><FormField label="Loại"><BaseSelect><option>Ví thường</option></BaseSelect></FormField></SurfaceCard>
    <BaseCategoryTree root={{id:"food",name:"Ăn uống",subtitle:"Áp dụng tất cả ví",isSystem:true,isEditable:true}} children={[{id:"coffee",name:"Cà phê",subtitle:"Hoạt động trong 1 ví",isSystem:false,isEditable:true}]} onSelect={() => setSheet(true)} />
    <SurfaceCard padding="md"><BudgetGauge value={65} label="Đã chi" /><Progress value={115} label="Ăn uống" danger /></SurfaceCard>
    <BaseButton id="open-sheet" onClick={() => setSheet(true)}>Mở chọn biểu tượng</BaseButton>
    {sheet ? <BaseBottomSheet title="Chọn biểu tượng" closeLabel="Đóng" onClose={() => setSheet(false)}><BaseButton onClick={() => setSheet(false)}>Chọn</BaseButton></BaseBottomSheet> : null}
  </main>;
}
createRoot(document.getElementById("root")!).render(<Fixture />);
