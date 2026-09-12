export function formatVND(value: number) {
  const sign = value < 0 ? "-" : "";
  return `${sign}${Math.abs(value).toLocaleString("vi-VN")} đ`;
}

export function ratioPercent(value: number, total: number) {
  if (total <= 0) return 0;
  return Math.min(100, Math.round((value / total) * 100));
}

