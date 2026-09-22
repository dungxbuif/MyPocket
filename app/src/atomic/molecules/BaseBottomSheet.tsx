import type { ReactNode } from "react";
import { X } from "lucide-react";
import { Heading } from "../atoms/Heading";
import { IconButton } from "../atoms/IconButton";
import { BaseButton } from "../atoms/BaseButton";
import { useModalFocus } from "../atoms/useModalFocus";
export function BaseBottomSheet({ title, closeLabel, onClose, children, presentation = "default", headerAction, closingDisabled = false, footer }: { title: string; closeLabel: string; onClose: () => void; children: ReactNode; presentation?: "default" | "form"; headerAction?: ReactNode; closingDisabled?: boolean; footer?: ReactNode }) {
  const close = () => { if (!closingDisabled) onClose(); };
  const panel = useModalFocus(close);
  return <div className="fixed inset-0 z-40 flex items-end justify-center bg-heading/30" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) close(); }}>
    <section ref={panel} tabIndex={-1} role="dialog" aria-modal="true" aria-label={title} className={`max-h-[92dvh] w-full max-w-[430px] rounded-t-3xl ${presentation === "form" ? "h-[92dvh] bg-canvas" : "bg-card"} flex flex-col px-4 pt-3 shadow-sheet`}>
      <div className="mx-auto mb-3 h-1 w-10 shrink-0 rounded-full bg-line" />
      <header className="mb-5 grid shrink-0 grid-cols-[1fr_auto_1fr] items-center">{presentation === "form" ? <BaseButton variant="chip" size="sm" disabled={closingDisabled} className="justify-self-start" onClick={close}>{closeLabel}</BaseButton> : <span />}<Heading as="h2" size="section" className="text-center">{title}</Heading><span className="justify-self-end">{presentation === "form" ? headerAction : <IconButton label={closeLabel} onClick={close}><X size={18} /></IconButton>}</span></header>
      <div className="min-h-0 flex-1 overflow-y-auto pb-5">{children}</div>
      {footer ? <footer className="shrink-0 bg-canvas py-3 pb-[calc(env(safe-area-inset-bottom)+16px)]">{footer}</footer> : <div className="shrink-0 pb-[calc(env(safe-area-inset-bottom)+16px)]" />}
    </section>
  </div>;
}
