import { useEffect, useRef, type ReactNode } from "react";
import { X } from "lucide-react";
import { Heading } from "../atoms/Heading";
import { IconButton } from "../atoms/IconButton";
import { BaseButton } from "../atoms/BaseButton";

export function BaseBottomSheet({ title, closeLabel, onClose, children, presentation = "default", headerAction, closingDisabled = false }: { title: string; closeLabel: string; onClose: () => void; children: ReactNode; presentation?: "default" | "form"; headerAction?: ReactNode; closingDisabled?: boolean }) {
  const panel = useRef<HTMLElement>(null);
  const close = useRef(onClose);
  close.current = onClose;
  useEffect(() => {
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const overflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const focusable = () => Array.from(panel.current?.querySelectorAll<HTMLElement>('button:not(:disabled), a[href], input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex="0"]') ?? []);
    (focusable()[0] ?? panel.current)?.focus();
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") { event.preventDefault(); close.current(); }
      if (event.key !== "Tab") return;
      const items = focusable(); const first = items[0], last = items.at(-1);
      if (!first || !last) { event.preventDefault(); panel.current?.focus(); return; }
      if (event.shiftKey && (document.activeElement === first || document.activeElement === panel.current)) { event.preventDefault(); last.focus(); }
      else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
    };
    document.addEventListener("keydown", onKey);
    return () => { document.removeEventListener("keydown", onKey); document.body.style.overflow = overflow; previous?.focus(); };
  }, []);
  return <div className="fixed inset-0 z-40 flex items-end justify-center bg-heading/30 px-0" role="presentation" onMouseDown={onClose}>
    <section ref={panel} tabIndex={-1} role="dialog" aria-modal="true" aria-label={title} className={`max-h-[88vh] w-full max-w-[430px] overflow-y-auto rounded-t-3xl ${presentation === "form" ? "h-[88dvh] bg-canvas" : "bg-card"} px-4 pb-[calc(env(safe-area-inset-bottom)+20px)] pt-3 shadow-sheet`} onMouseDown={(event) => event.stopPropagation()}>
      <div className="mx-auto mb-3 h-1 w-10 rounded-full bg-line" />
      <header className="mb-4 grid grid-cols-[1fr_auto_1fr] items-center">{presentation === "form" ? <BaseButton variant="chip" size="sm" disabled={closingDisabled} className="justify-self-start" onClick={onClose}>{closeLabel}</BaseButton> : <span />}<Heading as="h2" size="section" className="text-center">{title}</Heading><span className="justify-self-end">{presentation === "form" ? headerAction : <IconButton label={closeLabel} onClick={onClose}><X size={18} /></IconButton>}</span></header>
      {children}
    </section>
  </div>;
}
