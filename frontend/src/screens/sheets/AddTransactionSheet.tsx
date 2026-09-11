import * as React from "react";
import { Wallet, Users, CalendarDays, List, ImagePlus, MapPin, BriefcaseBusiness, Bell } from "lucide-react";
import {
  createTransaction,
  type CategorySummary,
  type Transaction,
  type WalletSummary,
} from "../../app/finance";
import { buildTransactionInput, calendarDateInHoChiMinh } from "../../app/transactionInput";
import { createObligation, type ObligationDirection } from "../../app/planning";
import { uploadFile } from "../../app/receipts";
import { queueReceiptUpload } from "../../offline/db";

export interface AddTransactionSheetProps {
  categories: CategorySummary[];
  wallets: WalletSummary[];
  readOnly: boolean;
  onCreated: (transaction: Transaction) => void;
  onDebtCreated: () => void;
  onClose: () => void;
}

export function quickAddTypeLabel(type: "expense" | "income" | "debt") {
  switch (type) {
    case "expense":
      return "Khoản chi";
    case "income":
      return "Khoản thu";
    case "debt":
      return "Vay/nợ";
  }
}

export function AddTransactionSheet({
  categories,
  wallets,
  readOnly,
  onCreated,
  onDebtCreated,
  onClose,
}: AddTransactionSheetProps) {
  const [type, setType] = React.useState<"expense" | "income" | "debt">("expense");
  const [amount, setAmount] = React.useState("");
  const [note, setNote] = React.useState("");
  const [sourceWalletID, setSourceWalletID] = React.useState(wallets[0]?.id ?? "");
  const [categoryID, setCategoryID] = React.useState("");
  const [excludedFromReports, setExcludedFromReports] = React.useState(false);
  const [debtDirection, setDebtDirection] = React.useState<ObligationDirection>("borrowed");
  const [counterparty, setCounterparty] = React.useState("");
  const [dueOn, setDueOn] = React.useState(() => calendarDateInHoChiMinh());
  const [occurredOn, setOccurredOn] = React.useState(() => calendarDateInHoChiMinh());
  const [saving, setSaving] = React.useState(false);
  const [receiptFile, setReceiptFile] = React.useState<File | null>(null);
  const [showDetails, setShowDetails] = React.useState(false);

  const filteredCategories = categories.filter(
    (category) => category.kind === (type === "income" ? "income" : "expense")
  );
  const chosenCategoryID = type !== "debt" ? categoryID || filteredCategories[0]?.id : "";
  const canSave =
    !readOnly &&
    wallets.length > 0 &&
    Number(amount) > 0 &&
    (type === "debt" ? counterparty.trim() !== "" && dueOn !== "" : Boolean(sourceWalletID));

  React.useEffect(() => {
    if (!sourceWalletID && wallets[0]) setSourceWalletID(wallets[0].id);
  }, [sourceWalletID, wallets]);

  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      if (type === "debt") {
        await createObligation({
          direction: debtDirection,
          principal_vnd: Number(amount),
          counterparty: counterparty.trim(),
          due_on: dueOn,
          note,
        });
        onDebtCreated();
        onClose();
        return;
      }
      const receipt = receiptFile && navigator.onLine ? await uploadFile(receiptFile) : undefined;
      const transaction = await createTransaction(
        buildTransactionInput({
          type,
          amount,
          sourceWalletID,
          categoryID: chosenCategoryID,
          note,
          excludedFromReports,
          occurredOn,
          receiptObjectID: receipt?.id,
        })
      );
      if (receiptFile && !navigator.onLine) {
        await queueReceiptUpload({
          transaction_id: transaction.id,
          file: receiptFile,
          filename: receiptFile.name,
          content_type: receiptFile.type,
        });
      }
      onCreated(transaction);
      onClose();
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label="Thêm Giao Dịch">
        <header>
          <button className="pill-button" type="button" onClick={onClose}>
            Hủy
          </button>
          <h2>Thêm Giao Dịch</h2>
          <span />
        </header>

        {readOnly ? (
          <p className="offline-warning">Offline storage chưa sẵn sàng. Mở mạng lại để lưu giao dịch.</p>
        ) : null}

        <div className="transaction-form-card">
          <div className="segmented sheet-segmented">
            {(["expense", "income", "debt"] as const).map((option) => (
              <button
                className={type === option ? "active" : ""}
                type="button"
                key={option}
                onClick={() => setType(option)}
              >
                {quickAddTypeLabel(option)}
              </button>
            ))}
          </div>

          <label className="amount-row">
            <span>VND</span>
            <input
              aria-label="Số tiền"
              type="text"
              inputMode="numeric"
              value={amount}
              onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))}
              placeholder="0"
            />
          </label>

          {type === "debt" ? (
            <div className="segmented debt-segmented">
              <button
                type="button"
                className={debtDirection === "borrowed" ? "active" : ""}
                onClick={() => setDebtDirection("borrowed")}
              >
                Tôi vay
              </button>
              <button
                type="button"
                className={debtDirection === "lent" ? "active" : ""}
                onClick={() => setDebtDirection("lent")}
              >
                Tôi cho vay
              </button>
            </div>
          ) : null}

          {type !== "debt" ? (
            <label className="sheet-row">
              <Wallet />
              <select
                aria-label="Ví nguồn"
                value={sourceWalletID}
                onChange={(event) => setSourceWalletID(event.target.value)}
              >
                {wallets.map((wallet) => (
                  <option key={wallet.id} value={wallet.id}>
                    {wallet.name}
                  </option>
                ))}
              </select>
            </label>
          ) : (
            <label className="sheet-row">
              <Users />
              <input
                aria-label="Đối tác"
                value={counterparty}
                onChange={(event) => setCounterparty(event.target.value)}
                placeholder="Người liên quan"
              />
            </label>
          )}

          {type !== "debt" ? (
            <label className="sheet-row">
              <span className="dot-icon" />
              <select
                aria-label="Nhóm"
                value={chosenCategoryID}
                onChange={(event) => setCategoryID(event.target.value)}
              >
                <option value="">Chọn nhóm</option>
                {filteredCategories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </select>
            </label>
          ) : (
            <label className="sheet-row">
              <CalendarDays />
              <input
                aria-label="Ngày đến hạn"
                type="date"
                value={dueOn}
                onChange={(event) => setDueOn(event.target.value)}
              />
            </label>
          )}

          {type !== "debt" ? (
            <label className="sheet-row">
              <List />
              <input
                aria-label="Ghi chú"
                value={note}
                onChange={(event) => setNote(event.target.value)}
                placeholder="Ghi chú"
              />
            </label>
          ) : null}

          {type !== "debt" ? (
            <label className="date-row">
              <CalendarDays />
              <input
                aria-label="Ngày giao dịch"
                type="date"
                value={occurredOn}
                onChange={(event) => setOccurredOn(event.target.value)}
              />
            </label>
          ) : null}

          {type !== "debt" ? (
            <label className="exclude-row">
              <input
                type="checkbox"
                checked={excludedFromReports}
                onChange={(event) => setExcludedFromReports(event.target.checked)}
              />
              <span>Không tính vào báo cáo</span>
            </label>
          ) : null}

          <button
            className="details-trigger"
            type="button"
            onClick={() => setShowDetails((current) => !current)}
            aria-expanded={showDetails}
          >
            {showDetails ? "Ẩn chi tiết" : "Thêm chi tiết"}
          </button>

          {showDetails ? (
            <div className="details-panel">
              <label className="sheet-row">
                <List />
                <input
                  aria-label="Ghi chú chi tiết"
                  value={note}
                  onChange={(event) => setNote(event.target.value)}
                  placeholder="Ghi chú"
                />
              </label>
              <div className="sheet-row muted">
                <Users />
                <span>Với</span>
              </div>
              <div className="sheet-row muted">
                <MapPin />
                <span>Đặt vị trí</span>
              </div>
              <div className="sheet-row muted">
                <BriefcaseBusiness />
                <span>Chọn sự kiện</span>
              </div>
              <div className="sheet-row muted">
                <Bell />
                <span>Đặt nhắc nhở</span>
              </div>
              <label className="image-row" htmlFor="receipt-image">
                <ImagePlus size={22} />
                {receiptFile ? receiptFile.name : "Thêm Hình Ảnh"}
                <input
                  id="receipt-image"
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  onChange={(event) => setReceiptFile(event.target.files?.[0] ?? null)}
                  hidden
                />
              </label>
            </div>
          ) : null}
        </div>

        <div className="save-bar">
          <button type="button" disabled={!canSave || saving} onClick={() => void save()}>
            {saving ? "Đang lưu" : "Lưu"}
          </button>
          <button type="button" className="receipt">
            <ImagePlus size={24} />
          </button>
        </div>
      </section>
    </div>
  );
}
