import type { ButtonHTMLAttributes, HTMLAttributes, ReactNode } from "react";

export function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(" ");
}

// For existing class-based controls: shared native semantics without restyling.
export function ActionButton(props: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button type="button" {...props} />;
}

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
