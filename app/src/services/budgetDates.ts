export function budgetDateLabel(date:Date):string {
  return `${date.getFullYear()}-${String(date.getMonth()+1).padStart(2,"0")}-${String(date.getDate()).padStart(2,"0")}`;
}

export function budgetDateInput(startLabel:string,endLabel:string) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(startLabel) || !/^\d{4}-\d{2}-\d{2}$/.test(endLabel) || endLabel < startLabel) {
    throw new Error("Khoảng ngày không hợp lệ.");
  }
  return { start_date: startLabel, end_date: endLabel };
}
