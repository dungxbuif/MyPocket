import type { ReactNode } from "react";

export function InlineControlRow({ leading, children }: { leading: ReactNode; children: ReactNode }) {
  return <div className="flex items-center gap-4 px-3 py-3.5"><span className="grid h-11 w-11 shrink-0 place-items-center">{leading}</span><div className="min-w-0 flex-1">{children}</div></div>;
}
