// Small arithmetic grammar: no eval, variables, functions or executable input.
export function calculateAmount(expression: string): string | null {
  if (!/^\d+(?:[+−×÷-]\d+)*$/.test(expression)) return null;
  const tokens = expression.match(/\d+|[+−×÷-]/g)!;
  let term = Number(tokens[0]), total = 0, sign = 1;
  for (let i = 1; i < tokens.length; i += 2) {
    const value = Number(tokens[i + 1]), op = tokens[i];
    if (!Number.isSafeInteger(value)) return null;
    if (op === "×") term *= value;
    else if (op === "÷") { if (!value) return null; term /= value; }
    else { total += sign * term; term = value; sign = op === "+" ? 1 : -1; }
    if (!Number.isFinite(term) || Math.abs(term) > Number.MAX_SAFE_INTEGER) return null;
  }
  const result = total + sign * term;
  return Number.isSafeInteger(result) && result >= 0 ? String(result) : null;
}
