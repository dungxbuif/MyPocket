import type { ReactNode } from "react";
import { useModalFocus } from "./useModalFocus";
export function BaseModal({ label, onClose, children }: { label: string; onClose: () => void; children: ReactNode }) {
  const panel = useModalFocus(onClose);
  return <div className="fixed inset-0 z-50 flex items-center justify-center bg-heading/40 p-5" onMouseDown={event => { event.stopPropagation(); if (event.target === event.currentTarget) onClose(); }}>
    <section ref={panel} role="dialog" aria-modal="true" aria-label={label} tabIndex={-1} className="max-h-[85dvh] w-full max-w-sm overflow-y-auto rounded-3xl bg-card p-5 shadow-raised">{children}</section>
  </div>;
}
