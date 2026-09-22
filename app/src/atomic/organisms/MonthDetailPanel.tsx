import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "@tanstack/react-router";

import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextArea, FormField } from "../atoms/FormField";
import { Heading } from "../atoms/Heading";
import { Progress } from "../atoms/Progress";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { PageBackHeader } from "../molecules/PageBackHeader";
import { deleteMonthNote, fetchMonthSummary, saveMonthNote, type MonthSummary } from "../../services/months";
import { isMonthKey, jarUsagePercent, monthLabel, monthOptions } from "../../services/monthJarLogic";
import { formatVND } from "../utils/format";

export function MonthDetailPanel({ month: routeMonth, masked, refreshKey = 0 }: { month: string; masked: boolean; refreshKey?: number }) {
  const navigate = useNavigate();
  const month = isMonthKey(routeMonth) ? routeMonth : undefined;
  const [summary, setSummary] = useState<MonthSummary | null>(null);
  const [draft, setDraft] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [noteError, setNoteError] = useState("");
  const [saving, setSaving] = useState(false);
  const dirty = summary !== null && draft !== summary.note;
  const monthChoices = useMemo(() => month ? monthOptions(month, 24, 1) : [], [month]);

  useEffect(() => {
    if (!month) {
      setError("Tháng không hợp lệ.");
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    fetchMonthSummary(month).then((value) => {
      if (cancelled) return;
      setSummary(value);
      setDraft(value.note);
      setError("");
      setNoteError("");
    }).catch(() => { if (!cancelled) setError("Không thể tải tổng kết tháng này."); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [month, refreshKey]);

  const changeMonth = (next: string) => {
    if (!isMonthKey(next) || next === month) return;
    if (dirty && !window.confirm("Ghi chú đang sửa chưa được lưu. Bỏ thay đổi và chuyển tháng?")) return;
    void navigate({ to: "/months/$month", params: { month: next } });
  };

  const saveNote = async () => {
    if (!month) return;
    setSaving(true);
    setNoteError("");
    try {
      if (draft.trim()) await saveMonthNote(month, draft);
      else await deleteMonthNote(month);
      const updated = await fetchMonthSummary(month);
      setSummary(updated);
      setDraft(updated.note);
    } catch {
      setNoteError("Không lưu được ghi chú. Nội dung đang nhập vẫn được giữ.");
    } finally {
      setSaving(false);
    }
  };

  const clearNote = async () => {
    if (!month || !summary?.note || !window.confirm(dirty ? "Bỏ ghi chú đang sửa và xóa ghi chú đã lưu của tháng này?" : "Xóa ghi chú của tháng này?")) return;
    setSaving(true);
    setNoteError("");
    try {
      await deleteMonthNote(month);
      const updated = await fetchMonthSummary(month);
      setSummary(updated);
      setDraft(updated.note);
    } catch {
      setNoteError("Không xóa được ghi chú. Hãy thử lại.");
    } finally {
      setSaving(false);
    }
  };

  const openJars = () => {
    if (month) void navigate({ to: "/jars", search: { month } });
  };

  return <>
    <PageBackHeader title="Tổng kết tháng" backTo="/" backLabel="Về tổng quan" />
    <div className="space-y-3 pb-4">
      {month ? <FormField label="Tháng"><BaseSelect aria-label="Chọn tháng tổng kết" value={month} onChange={(event) => changeMonth(event.target.value)}>{monthChoices.map((option) => <option key={option} value={option}>{monthLabel(option)}</option>)}</BaseSelect></FormField> : null}
      {loading ? <StatusMessage>Đang tải tổng kết...</StatusMessage> : null}
      {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      {!loading && !error && summary ? <>
        <SurfaceCard padding="md">
          <div className="flex items-start justify-between gap-3"><div><Heading as="h2" size="section">{monthLabel(summary.month)}</Heading><Text size="sm" tone="secondary">{summary.is_complete ? "Đã hoàn tất" : "Đang diễn ra"} · {summary.timezone}</Text></div><BaseButton variant="ghost" size="sm" onClick={openJars}>Hũ tháng</BaseButton></div>
          <div className="mt-4 grid grid-cols-2 gap-3">
            <Metric label="Thu" value={summary.income} masked={masked} />
            <Metric label="Chi" value={summary.expense} masked={masked} />
            <Metric label="Chênh lệch" value={summary.net} masked={masked} />
            <Metric label="Giao dịch" value={summary.transaction_count} masked={false} />
          </div>
          <Text size="xs" tone="secondary" className="mt-3">Kỳ báo cáo: {formatBoundary(summary.start_at, summary.timezone)} – {formatBoundary(summary.next_start_at, summary.timezone)} (đầu kỳ kế tiếp không tính).</Text>
        </SurfaceCard>

        <SurfaceCard padding="md">
          <Heading as="h2" size="section">Theo nhóm</Heading>
          {summary.categories.length === 0 ? <StatusMessage variant="plain">Chưa có giao dịch trong tháng.</StatusMessage> : <div className="mt-3 space-y-3">{summary.categories.map((row) => <div key={`${row.type}:${row.category_id ?? row.name}`} className="flex items-center justify-between gap-3"><div className="min-w-0"><Text>{row.name}</Text><Text size="xs" tone="secondary">{row.type === "income" ? "Thu" : "Chi"} · {row.count} giao dịch</Text></div><Text numeric weight="semibold">{masked ? "••••••" : formatVND(row.amount)}</Text></div>)}</div>}
        </SurfaceCard>

        <SurfaceCard padding="md">
          <div className="flex items-center justify-between gap-3"><Heading as="h2" size="section">Hũ chi tiêu</Heading><BaseButton variant="ghost" size="sm" onClick={openJars}>Quản lý hũ</BaseButton></div>
          {!summary.jar_summary || summary.jar_summary.items.length === 0 ? <StatusMessage variant="plain">Tháng này chưa có hũ hoặc khoản chi gắn hũ.</StatusMessage> : <div className="mt-3 space-y-4">{summary.jar_summary.items.map((item) => {
            const allocation = item.calculated_allocation ?? (item.allocation_mode === "fixed" ? item.allocation_amount : undefined);
            const usage = jarUsagePercent(item.spent, allocation);
            return <div key={item.jar_id}><div className="flex items-center justify-between gap-3"><Text>{item.name}</Text><Text numeric>{masked ? "••••••" : formatVND(item.spent)}</Text></div>{usage !== null ? <div className="mt-2"><Progress value={usage} label={`Mức sử dụng hũ ${item.name}`} danger={usage > 100} /><Text size="xs" tone={usage > 100 ? "danger" : "secondary"}>{Math.round(usage)}% phân bổ</Text></div> : null}</div>;
          })}</div>}
        </SurfaceCard>

        <SurfaceCard padding="md">
          <Heading as="h2" size="section">Ghi chú của tôi</Heading>
          <div className="mt-3"><FormField label="Ghi chú tháng"><BaseTextArea aria-label="Ghi chú tháng" rows={5} maxLength={5000} value={draft} onChange={(event) => setDraft(event.target.value)} placeholder="Điều bạn muốn ghi nhớ về tháng này" /></FormField><Text size="xs" tone="secondary">{draft.length}/5000</Text></div>
          {noteError ? <div className="mt-3"><StatusMessage tone="danger">{noteError}</StatusMessage></div> : null}
          <div className="mt-3 flex flex-wrap gap-2"><BaseButton loading={saving} disabled={!dirty} onClick={() => void saveNote()}>Lưu ghi chú</BaseButton><BaseButton variant="outline" disabled={!dirty || saving} onClick={() => setDraft(summary.note)}>Bỏ thay đổi</BaseButton>{summary.note ? <BaseButton variant="ghost" disabled={saving} onClick={() => void clearNote()}>Xóa ghi chú</BaseButton> : null}</div>
        </SurfaceCard>
      </> : null}
    </div>
  </>;
}

function Metric({ label, value, masked }: { label: string; value: number; masked: boolean }) {
  return <div><Text size="xs" tone="secondary">{label}</Text><Text numeric weight="bold">{masked ? "••••••" : label === "Giao dịch" ? String(value) : formatVND(value)}</Text></div>;
}

function formatBoundary(value: string, timezone: string): string {
  return new Intl.DateTimeFormat("vi-VN", { dateStyle: "short", timeStyle: "short", timeZone: timezone }).format(new Date(value));
}
