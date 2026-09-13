import type { ReactNode } from "react";
import { X } from "lucide-react";
import { Heading } from "../atoms/Heading";
import { IconButton } from "../atoms/IconButton";

export function BaseBottomSheet({ title, closeLabel, onClose, children }: { title: string; closeLabel: string; onClose: () => void; children: ReactNode }) {
  return <div className="fixed inset-0 z-40 flex items-end justify-center bg-slate-950/30 px-0" role="presentation" onMouseDown={onClose}>
    <section role="dialog" aria-modal="true" aria-label={title} className="max-h-[88vh] w-full max-w-[430px] overflow-y-auto rounded-t-[32px] bg-white px-4 pb-[calc(env(safe-area-inset-bottom)+20px)] pt-3 shadow-[0_-10px_30px_rgb(15_23_42/0.12)]" onMouseDown={(event) => event.stopPropagation()}>
      <div className="mx-auto mb-3 h-1 w-10 rounded-full bg-slate-200" />
      <header className="mb-4 grid grid-cols-[1fr_auto_1fr] items-center"><span /><Heading as="h2" size="section" className="text-center">{title}</Heading><span className="justify-self-end"><IconButton label={closeLabel} onClick={onClose}><X size={18} /></IconButton></span></header>
      {children}
    </section>
  </div>;
}
