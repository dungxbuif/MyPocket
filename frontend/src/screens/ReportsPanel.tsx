import { useEffect, useRef, useState } from "react";
import type { WalletSummary } from "../app/finance";
import { queryReport, type DetailedReport, type ReportFilters, type ReportKind } from "../app/reportQueries";
import { Button } from "../components/ui/button";
import { Card } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Select } from "../components/ui/select";
import { OperationError, operationFailure, type OperationFailure } from "../components/feedback/OperationError";

export interface ReportsPanelProps {
  userID: string;
  online: boolean;
  privacyMasked: boolean;
  wallets: WalletSummary[];
  onClose: () => void;
}

const kinds: Array<{ value: ReportKind; label: string }> = [
  { value: "cash-flow", label: "Thu chi" },
  { value: "categories", label: "Danh mục" },
  { value: "daily", label: "Theo ngày" },
  { value: "comparison", label: "So sánh kỳ" },
  { value: "cumulative", label: "Lũy kế" },
];
const numbers = new Intl.NumberFormat("vi-VN", { maximumFractionDigits: 2 });

function currentMonth(): ReportFilters {
  const parts = new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Ho_Chi_Minh", year: "numeric", month: "2-digit" }).formatToParts(new Date());
  const year = parts.find(part => part.type === "year")!.value;
  const month = parts.find(part => part.type === "month")!.value;
  const lastDay = new Date(Date.UTC(Number(year), Number(month), 0)).getUTCDate();
  return { kind: "cash-flow", from: `${year}-${month}-01`, to: `${year}-${month}-${lastDay}`, walletID: "" };
}

function validDate(value: string) {
  const date = new Date(`${value}T00:00:00Z`);
  return /^\d{4}-\d{2}-\d{2}$/.test(value) && !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === value;
}

export function ReportsPanel({ userID, online, privacyMasked, wallets, onClose }: ReportsPanelProps) {
  const [filters, setFilters] = useState<ReportFilters>(currentMonth);
  const [result, setResult] = useState<{ userID: string; filters: ReportFilters; report: DetailedReport } | null>(null);
  const [loading, setLoading] = useState(false);
  const [failure, setFailure] = useState<OperationFailure | null>(null);
  const request = useRef(0);
  const controller = useRef<AbortController | null>(null);

  function cancel() {
    request.current += 1;
    controller.current?.abort();
    controller.current = null;
  }

  useEffect(() => {
    cancel();
    setResult(null);
    setLoading(false);
    setFailure(null);
    setFilters(currentMonth());
    return cancel;
  }, [userID]);

  useEffect(() => {
    if (!online) {
      cancel();
      setLoading(false);
    }
  }, [online]);

  function changeFilters(next: Partial<ReportFilters>) {
    cancel();
    setFilters(current => ({ ...current, ...next }));
    setResult(null);
    setLoading(false);
    setFailure(null);
  }

  async function load() {
    if (!online) return;
    cancel();
    setResult(null);
    setFailure(null);
    if (!validDate(filters.from) || !validDate(filters.to)) {
      setFailure({ message: "Chọn ngày bắt đầu và ngày kết thúc hợp lệ." });
      return;
    }
    if (filters.from > filters.to) {
      setFailure({ message: "Ngày bắt đầu phải trước hoặc bằng ngày kết thúc." });
      return;
    }
    const token = request.current;
    const nextController = new AbortController();
    controller.current = nextController;
    setLoading(true);
    try {
      const report = await queryReport(filters, nextController.signal);
      if (request.current === token) setResult({ userID, filters, report });
    } catch (error) {
      if (request.current === token) setFailure(operationFailure(error, "Không tải được báo cáo. Vui lòng thử lại."));
    } finally {
      if (request.current === token) setLoading(false);
    }
  }

  const visible = result?.userID === userID ? result : null;
  const report = visible?.report;
  const money = (value: number | undefined) => privacyMasked ? "••••••" : value === undefined ? "Không có dữ liệu" : `${numbers.format(value)} đ`;
  const percent = (value: number) => privacyMasked ? "••••••" : `${numbers.format(value)}%`;
  const cell = "px-3 py-2 text-right whitespace-nowrap";

  return (
    <section role="region" aria-label="Báo cáo chi tiết" className="content-stack">
      <div className="sub-header"><h1>Báo cáo chi tiết</h1><Button variant="outline" size="sm" onClick={onClose}>Đóng báo cáo</Button></div>
      <Card>
        <form className="content-stack" noValidate onSubmit={event => { event.preventDefault(); void load(); }}>
          <label className="form-row">Loại báo cáo<Select value={filters.kind} onChange={event => changeFilters({ kind: event.target.value as ReportKind })}>{kinds.map(kind => <option key={kind.value} value={kind.value}>{kind.label}</option>)}</Select></label>
          <label className="form-row">Từ ngày<Input type="date" value={filters.from} onChange={event => changeFilters({ from: event.target.value })} /></label>
          <label className="form-row">Đến ngày<Input type="date" value={filters.to} onChange={event => changeFilters({ to: event.target.value })} /></label>
          <label className="form-row">Ví báo cáo<Select value={filters.walletID} onChange={event => changeFilters({ walletID: event.target.value })}><option value="">Tất cả ví</option>{wallets.map(wallet => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</Select></label>
          <Button type="submit" loading={loading} disabled={!online}>Xem báo cáo</Button>
        </form>
      </Card>
      {!online ? <p role="status" className="offline-warning">Ngoại tuyến. {report ? "Số liệu đã tải có thể cũ và chưa phản ánh thay đổi chưa đồng bộ. Kết nối lại để cập nhật." : "Kết nối lại để tải báo cáo; chưa có số liệu cho bộ lọc này."}</p> : loading ? <p role="status">Đang tải báo cáo…</p> : null}
      <OperationError failure={failure} onRetry={online ? () => void load() : undefined} retryLabel="Thử lại báo cáo" busy={loading} />
      {!report && !loading && !failure && online ? <p>Chọn bộ lọc rồi nhấn Xem báo cáo.</p> : null}
      {report && visible ? <Card className="content-stack">
        <p>{kinds.find(kind => kind.value === visible.filters.kind)?.label} · {report.summary.from} – {report.summary.to} · {visible.filters.walletID ? wallets.find(wallet => wallet.id === visible.filters.walletID)?.name ?? "Ví đã chọn" : "Tất cả ví"}</p>
        <p>Cập nhật: {new Date(report.summary.generated_at).toLocaleString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh" })} (Asia/Ho_Chi_Minh)</p>
        <div className="overflow-x-auto"><table className="w-full" aria-label="Tổng hợp báo cáo"><tbody>
          <tr><th scope="row" className="text-left">Tổng thu</th><td className={cell}>{money(report.summary.income_vnd)}</td></tr>
          <tr><th scope="row" className="text-left">Tổng chi</th><td className={cell}>{money(report.summary.expense_vnd)}</td></tr>
          <tr><th scope="row" className="text-left">Thu nhập ròng</th><td className={cell}>{money(report.summary.net_income_vnd)}</td></tr>
          <tr><th scope="row" className="text-left">Chi trung bình mỗi ngày</th><td className={cell}>{money(report.summary.daily_average_vnd)}</td></tr>
        </tbody></table></div>
        {visible.filters.kind === "categories" ? (report.categories?.length ? <div className="overflow-x-auto"><table className="w-full" aria-label="Chi theo danh mục"><thead><tr><th scope="col">Danh mục</th><th scope="col">Số tiền</th><th scope="col">Tỷ trọng</th></tr></thead><tbody>{report.categories.map((category, index) => <tr key={category.category_id ?? `${category.category_name}-${index}`}><th scope="row" className="text-left">{category.category_name}</th><td className={cell}>{money(category.amount_vnd)}</td><td className={cell}>{percent(category.share_percent)}</td></tr>)}</tbody></table></div> : <p>Không có khoản chi theo danh mục trong kỳ.</p>) : null}
        {visible.filters.kind === "daily" || visible.filters.kind === "cumulative" ? (report.daily?.length ? <div className="overflow-x-auto"><table className="w-full" aria-label={visible.filters.kind === "daily" ? "Báo cáo theo ngày" : "Lũy kế theo ngày"}><thead><tr><th scope="col">Ngày</th><th scope="col">Thu</th><th scope="col">Chi</th><th scope="col">Thu nhập ròng</th>{visible.filters.kind === "cumulative" ? <th scope="col">Lũy kế</th> : null}</tr></thead><tbody>{report.daily.map(day => <tr key={day.date}><th scope="row">{day.date}</th><td className={cell}>{money(day.income_vnd)}</td><td className={cell}>{money(day.expense_vnd)}</td><td className={cell}>{money(day.net_income_vnd)}</td>{visible.filters.kind === "cumulative" ? <td className={cell}>{money(day.cumulative_net_vnd)}</td> : null}</tr>)}</tbody></table></div> : <p>Không có dữ liệu theo ngày trong kỳ.</p>) : null}
        {visible.filters.kind === "comparison" ? (report.prior ? <>
          <p>Kỳ trước: {report.prior.from} – {report.prior.to}</p>
          <div className="overflow-x-auto"><table className="w-full" aria-label="So sánh kỳ"><thead><tr><th scope="col">Chỉ số</th><th scope="col">Kỳ trước</th><th scope="col">Kỳ này</th><th scope="col">Thay đổi</th></tr></thead><tbody>
            <tr><th scope="row">Thu</th><td className={cell}>{money(report.prior.income_vnd)}</td><td className={cell}>{money(report.summary.income_vnd)}</td><td className={cell}>{privacyMasked ? "••••••" : report.summary.not_comparable || report.prior.income_vnd === 0 ? "Không thể so sánh" : percent(report.summary.income_change_percent ?? 0)}</td></tr>
            <tr><th scope="row">Chi</th><td className={cell}>{money(report.prior.expense_vnd)}</td><td className={cell}>{money(report.summary.expense_vnd)}</td><td className={cell}>{privacyMasked ? "••••••" : report.summary.not_comparable || report.prior.expense_vnd === 0 ? "Không thể so sánh" : percent(report.summary.expense_change_percent ?? 0)}</td></tr>
          </tbody></table></div>
        </> : <p>Không có dữ liệu kỳ trước để so sánh.</p>) : null}
      </Card> : null}
    </section>
  );
}
