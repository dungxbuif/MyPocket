import { useEffect, useRef, type ButtonHTMLAttributes, type ReactNode, type MouseEvent } from "react";
import { UI_CLASSES } from "./tokens";
import { createLongPress } from "./longPress";

export function BaseNavigationItem({ active, children, className = "", ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { active: boolean; children: ReactNode }) {
  return <button {...props} type="button" aria-current={active ? "page" : undefined} className={`${UI_CLASSES.focus} flex min-h-13 flex-col items-center justify-center gap-1 text-xs font-bold ${active ? "text-action" : "text-secondary"} ${className}`}>{children}</button>;
}
export function BaseFab({ children, className = "", onLongPress, onClick, ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { onLongPress?: () => void }) {
  const callbacks = useRef({ onLongPress, onClick });
  callbacks.current = { onLongPress, onClick };
  const clickEvent = useRef<MouseEvent<HTMLButtonElement> | null>(null);
  const pointer = useRef<number | null>(null);
  const press = useRef(createLongPress(() => { if (clickEvent.current) callbacks.current.onClick?.(clickEvent.current); }, () => callbacks.current.onLongPress?.()));
  useEffect(() => {
    const cancel = () => { pointer.current = null; press.current.cancel(); };
    window.addEventListener("blur", cancel);
    return () => { window.removeEventListener("blur", cancel); cancel(); };
  }, []);
  return <button {...props} type="button"
    onClick={event => {
      if (!onLongPress) { onClick?.(event); return; }
      clickEvent.current = event; press.current.click(event.detail === 0); clickEvent.current = null;
    }}
    onPointerDown={event => {
      if (!onLongPress || event.button !== 0 || !event.isPrimary) return;
      pointer.current = event.pointerId;
      event.currentTarget.setPointerCapture(event.pointerId);
      press.current.start();
    }}
    onPointerMove={() => { if (pointer.current !== null) press.current.cancel(); }}
    onPointerUp={() => { pointer.current = null; press.current.release(); }}
    onPointerCancel={() => { pointer.current = null; press.current.cancel(); }}
    onLostPointerCapture={() => { if (pointer.current !== null) { pointer.current = null; press.current.cancel(); } }}
    onBlur={() => { pointer.current = null; press.current.cancel(); }}
    onContextMenu={event => { if (onLongPress) event.preventDefault(); }}
    className={`${UI_CLASSES.focus} grid h-14 w-14 place-items-center rounded-full bg-brand text-card shadow-fab disabled:opacity-50 ${className}`}>{children}</button>;
}
