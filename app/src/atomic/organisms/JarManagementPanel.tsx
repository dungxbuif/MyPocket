import { useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "@tanstack/react-router";

import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { Heading } from "../atoms/Heading";
import { Progress } from "../atoms/Progress";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { PageBackHeader } from "../molecules/PageBackHeader";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";
import { fetchJarCumulative, fetchJarMonth, createJar, removeJarMonth, updateJarMonth, type JarCumulative, type JarItem, type JarMonthSummary, type JarConfigInput } from "../../services/jars";
import { isMonthKey, jarUsagePercent, monthLabel, monthOptions, resolveMonthKey } from "../../services/monthJarLogic";
import { formatVND } from "../utils/format";

type JarDraft = { jarID?: string; name: string; mode: JarItem["allocation_mode"]; amount: string };

export function JarManagementPanel({ masked, refreshKey = 0 }: { masked: boolean; refreshKey?: number }) {
  const timezone = useAccountTimezone();
  const location = useLocation();
  const navigate = useNavigate();
  const requestedMonth = new URLSearchParams(location.searchStr).get("month") ?? undefined;
  const [month, setMonth] = useState(() => resolveMonthKey(requestedMonth, timezone));
  const [summary, setSummary] = useState<JarMonthSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [reload, setReload] = useState(0);
  const [draft, setDraft] = useState<JarDraft | null>(null);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");
  const [selectedJarID, setSelectedJarID] = useState("");
  const [cumulative, setCumulative] = useState<JarCumulative | null>(null);
  const [cumulativeLoading, setCumulativeLoading] = useState(false);
  const [cumulativeError, setCumulativeError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchJarMonth(month).then((value) => {
      if (cancelled) return;
      setSummary(value);
      setError("");
      setSelectedJarID((current) => value.jars.some((jar) => jar.jar_id === current) ? current : value.jars[0]?.jar_id ?? "");
    }).catch(() => { if (!cancelled) setError("Không thể tải dữ liệu hũ tháng này."); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [month, refreshKey, reload]);

  useEffect(() => {
    if (!selectedJarID) {
      setCumulative(null);
      setCumulativeError("");
      return;
    }
    let cancelled = false;
    setCumulativeLoading(true);
    fetchJarCumulative(selectedJarID, month).then((value) => {
      if (cancelled) return;
      setCumulative(value);
      setCumulativeError("");
    }).catch(() => {
      if (cancelled) return;
      setCumulative(null);
      setCumulativeError("Chưa có dữ liệu cộng dồn cho hũ này.");
    }).finally(() => { if (!cancelled) setCumulativeLoading(false); });
    return () => { cancelled = true; };
  }, [selectedJarID, month, reload]);

  const monthChoices = useMemo(() => monthOptions(month, 24, 1), [month]);
  const changeMonth = (next: string) => {
    if (!isMonthKey(next)) return;
    setMonth(next);
    void navigate({ to: "/jars", search: { month: next } });
  };

  const openCreate = () => {
    setFormError("");
    setDraft({ name: "", mode: "none", amount: "" });
  };
  const openEdit = (item: JarItem) => {
    setFormError("");
    const amount = item.allocation_mode === "fixed" ? item.allocation_amount : item.allocation_mode === "percent" ? item.allocation_percent : undefined;
    setDraft({ jarID: item.jar_id, name: item.name, mode: item.allocation_mode, amount: amount == null ? "" : String(amount) });
  };

  const saveDraft = async () => {
    if (!draft || !draft.name.trim()) {
      setFormError("Vui lòng nhập tên hũ.");
      return;
    }
    const numericAmount = Number(draft.amount);
    if ((draft.mode === "fixed" || draft.mode === "percent") && (!Number.isFinite(numericAmount) || numericAmount <= 0 || (draft.mode === "percent" && numericAmount > 100))) {
      setFormError(draft.mode === "percent" ? "Tỷ lệ phải lớn hơn 0 và không quá 100%." : "Mức phân bổ phải lớn hơn 0.");
      return;
    }
    const input: JarConfigInput = {
      month,
      name: draft.name.trim(),
      allocation_mode: draft.mode,
      ...(draft.mode === "fixed" ? { allocation_amount: Math.round(numericAmount) } : {}),
      ...(draft.mode === "percent" ? { allocation_percent: numericAmount } : {}),
    };
    setSaving(true);
    setFormError("");
    try {
      if (draft.jarID) await updateJarMonth(draft.jarID, input);
      else await createJar(input);
      setDraft(null);
      setReload((value) => value + 1);
    } catch {
      setFormError("Không lưu được hũ. Hãy kiểm tra kết nối rồi thử lại.");
    } finally {
      setSaving(false);
    }
  };

  const removeCurrentMonth = async (item: JarItem) => {
    if (!window.confirm(`Gỡ “${item.name}” khỏi tháng ${monthLabel(month)}? Giao dịch cũ vẫn được giữ.`)) return;
    try {
      await removeJarMonth(item.jar_id, month);
      setReload((value) => value + 1);
    } catch {
      setError("Không gỡ được hũ khỏi tháng này. Hãy thử lại.");
    }
  };

  return (
    <>
      <PageBackHeader title="Hũ chi tiêu" backTo="/" backLabel="Về tổng quan" />
      <div className="space-y-3 pb-4">
        <FormField label="Tháng">
          <BaseSelect aria-label="Chọn tháng cho hũ" value={month} onChange={(event) => changeMonth(event.target.value)}>
            {monthChoices.map((option) => <option key={option} value={option}>{monthLabel(option)}</option>)}
          </BaseSelect>
        </FormField>

        {loading ? <StatusMessage>Đang tải hũ tháng này...</StatusMessage> : null}
        {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
        {!loading && !error && summary ? <>
          <SurfaceCard padding="md">
            <Heading as="h2" size="section">Tổng quan tháng</Heading>
            <div className="mt-3 grid grid-cols-2 gap-3">
              <Metric label="Thu thực tế" value={summary.actual_income} masked={masked} />
              <Metric label="Chi thực tế" value={summary.total_spent} masked={masked} />
              <Metric label="Đã phân bổ" value={summary.total_allocated} masked={masked} />
              <Metric label="Chi chưa gắn hũ" value={summary.unassigned_spent} masked={masked} />
            </div>
            {summary.over_income || summary.over_one_hundred_percent ? <StatusMessage tone="danger">Tổng phân bổ vượt thu thực tế. Đây là cảnh báo, không chặn giao dịch.</StatusMessage> : null}
          </SurfaceCard>

          <div className="flex items-center justify-between gap-3">
            <Heading as="h2" size="section">Hũ tháng</Heading>
            <BaseButton variant="secondary" size="sm" onClick={openCreate}>Thêm hũ</BaseButton>
          </div>
          {summary.items.length === 0 ? <SurfaceCard padding="md"><StatusMessage variant="plain">Chưa có hũ cho tháng này. Chi tiêu vẫn được ghi nhận bình thường.</StatusMessage></SurfaceCard> : summary.items.map((item) => {
            const allocation = item.calculated_allocation ?? (item.allocation_mode === "fixed" ? item.allocation_amount : undefined);
            const usage = jarUsagePercent(item.spent, allocation);
            return <SurfaceCard key={item.jar_id} padding="md">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0"><Text weight="bold">{item.name}</Text><Text size="sm" tone="secondary">{allocation == null ? "Chưa đặt phân bổ" : `Phân bổ ${masked ? "••••••" : formatVND(allocation)}`}</Text></div>
                <Text numeric weight="bold">{masked ? "••••••" : formatVND(item.spent)}</Text>
              </div>
              {usage !== null ? <div className="mt-3"><Progress value={usage} label={`Mức sử dụng hũ ${item.name}`} danger={usage > 100} /><Text size="xs" tone={usage > 100 ? "danger" : "secondary"}>{Math.round(usage)}% mức phân bổ</Text></div> : null}
              {item.active ? <div className="mt-3 flex gap-2"><BaseButton variant="outline" size="sm" onClick={() => openEdit(item)}>Sửa tháng này</BaseButton><BaseButton variant="ghost" size="sm" onClick={() => void removeCurrentMonth(item)}>Gỡ khỏi tháng</BaseButton></div> : <Text size="xs" tone="secondary">Cấu hình lịch sử; giao dịch vẫn được giữ.</Text>}
            </SurfaceCard>;
          })}

          <SurfaceCard padding="md">
            <Heading as="h2" size="section">Cộng dồn</Heading>
            {summary.jars.length === 0 ? <StatusMessage variant="plain">Tạo hũ để xem chi tiêu cộng dồn.</StatusMessage> : <>
              <FormField label="Chọn hũ">
                <BaseSelect aria-label="Chọn hũ xem cộng dồn" value={selectedJarID} onChange={(event) => setSelectedJarID(event.target.value)}>
                  {summary.jars.map((jar) => <option key={jar.jar_id} value={jar.jar_id}>{jar.name}</option>)}
                </BaseSelect>
              </FormField>
              {cumulativeLoading ? <StatusMessage>Đang tải báo cáo cộng dồn...</StatusMessage> : null}
              {cumulativeError ? <StatusMessage variant="plain">{cumulativeError}</StatusMessage> : null}
              {cumulative ? <>
                <div className="mt-3 grid grid-cols-2 gap-3"><Metric label="Chi thực tế" value={cumulative.total_spent} masked={masked} /><Metric label="Đã phân bổ" value={cumulative.total_allocated} masked={masked} /></div>
                <Text size="xs" tone="secondary">{cumulative.from_month} – {cumulative.to_month}. Không cộng dồn số dư.</Text>
                <div className="mt-3 space-y-2">{cumulative.months.map((row) => <div key={row.month} className="flex items-center justify-between gap-3"><Text size="sm">{monthLabel(row.month)}</Text><Text size="sm" numeric>{masked ? "••••••" : formatVND(row.spent)}</Text></div>)}</div>
              </> : null}
            </>}
          </SurfaceCard>
        </> : null}
      </div>

      {draft ? <BaseBottomSheet title={draft.jarID ? "Sửa hũ tháng" : "Thêm hũ"} closeLabel="Đóng" onClose={() => setDraft(null)} presentation="form" closingDisabled={saving} footer={<BaseButton className="w-full" loading={saving} onClick={() => void saveDraft()}>{draft.jarID ? "Lưu thay đổi" : "Tạo hũ"}</BaseButton>}>
        <div className="space-y-4">
          <FormField label="Tên hũ"><BaseTextInput value={draft.name} maxLength={80} onChange={(event) => setDraft({ ...draft, name: event.target.value })} placeholder="Ví dụ: Ăn uống" /></FormField>
          <FormField label="Phân bổ"><BaseSelect value={draft.mode} onChange={(event) => setDraft({ ...draft, mode: event.target.value as JarItem["allocation_mode"], amount: "" })}><option value="none">Không đặt</option><option value="fixed">Số tiền cố định</option><option value="percent">Phần trăm thu thực tế</option></BaseSelect></FormField>
          {draft.mode !== "none" ? <FormField label={draft.mode === "fixed" ? "Số tiền (VND)" : "Tỷ lệ (%)"}><BaseTextInput type="number" min="0" max={draft.mode === "percent" ? 100 : undefined} step={draft.mode === "percent" ? "0.01" : "1"} inputMode="decimal" value={draft.amount} onChange={(event) => setDraft({ ...draft, amount: event.target.value })} /></FormField> : null}
          {formError ? <StatusMessage tone="danger">{formError}</StatusMessage> : null}
        </div>
      </BaseBottomSheet> : null}
    </>
  );
}

function Metric({ label, value, masked }: { label: string; value: number; masked: boolean }) {
  return <div><Text size="xs" tone="secondary">{label}</Text><Text numeric weight="bold">{masked ? "••••••" : formatVND(value)}</Text></div>;
}
