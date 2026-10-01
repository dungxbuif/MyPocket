export function categoryMatchRank(rootName: string, childNames: string[], query: string): number {
  const normalized = query.toLocaleLowerCase("vi").trim();
  if (!normalized) return 0;
  if (rootName.toLocaleLowerCase("vi").includes(normalized)) return 0;
  if (childNames.some((name) => name.toLocaleLowerCase("vi").includes(normalized))) return 1;
  return 2;
}
