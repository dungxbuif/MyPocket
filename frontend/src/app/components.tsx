import { forwardRef, type ButtonHTMLAttributes, type HTMLAttributes, type InputHTMLAttributes, type ReactNode, type TextareaHTMLAttributes } from "react";

export function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(" ");
}

// For existing class-based controls: shared native semantics without restyling.
export function ActionButton(props: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button type="button" {...props} />;
}

export const InputControl = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(function InputControl(props, ref) {
  return <input ref={ref} {...props} />;
});

export const TextAreaControl = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(function TextAreaControl(props, ref) {
  return <textarea ref={ref} {...props} />;
});

export function Card({ className, ...props }: HTMLAttributes<HTMLElement>) {
  return <section className={cx("card", className)} {...props} />;
}

export function ActionCard({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button type="button" className={cx("card w-full text-left disabled:cursor-not-allowed", className)} {...props} />;
}

export function PillButton({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button className={cx("pill-button", className)} type="button" {...props} />;
}

export function IconButton({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button className={cx("plain-icon", className)} type="button" {...props} />;
}

export function SheetFrame({
  title,
  label,
  leading,
  trailing,
  className,
  headerClassName,
  footer,
  children,
}: {
  title: string;
  label?: string;
  leading: ReactNode;
  trailing: ReactNode;
  className?: string;
  headerClassName?: string;
  footer?: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="sheet-backdrop">
      <section className={cx("transaction-sheet", footer != null && "sheet-with-footer", className)} role="dialog" aria-modal="true" aria-label={label ?? title}>
        <header className={headerClassName}>
          {leading}
          <h2>{title}</h2>
          {trailing}
        </header>
        {footer != null ? <><div className="sheet-scroll-body">{children}</div>{footer}</> : children}
      </section>
    </div>
  );
}

export function SectionTitle({ title, action }: { title: string; action?: ReactNode }) {
  return (
    <div className="section-title">
      <h2>{title}</h2>
      {action}
    </div>
  );
}
