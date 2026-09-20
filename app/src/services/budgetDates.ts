export function budgetDateLabel(date:Date):string {
  return `${date.getFullYear()}-${String(date.getMonth()+1).padStart(2,"0")}-${String(date.getDate()).padStart(2,"0")}`;
}

// Keep persisted instants for untouched fields, including after a browser timezone change.
// Intentional edits use the local dates currently shown in the editor.
export function budgetDateInput(startLabel:string,endLabel:string,original?:{start_at:string;end_at:string}) {
  const start=new Date(`${startLabel}T00:00:00`),end=new Date(`${endLabel}T00:00:00`);
  end.setDate(end.getDate()+1);
  if(!Number.isFinite(start.getTime()) || !Number.isFinite(end.getTime()))throw new Error("Khoảng ngày không hợp lệ.");
  return {
    start_at:original && startLabel===budgetDateLabel(new Date(original.start_at))?original.start_at:start.toISOString(),
    end_at:original && endLabel===budgetDateLabel(new Date(Date.parse(original.end_at)-1))?original.end_at:end.toISOString(),
  };
}
