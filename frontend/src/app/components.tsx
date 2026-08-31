import type { ButtonHTMLAttributes, HTMLAttributes, ReactNode } from "react";

export function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(" ");
}

export function Card({ className, ...props }: HTMLAttributes<HTMLElement>) {
  return <section className={cx("card", className)} {...props} />;
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
  children,
}: {
  title: string;
  label?: string;
  leading: ReactNode;
  trailing: ReactNode;
  className?: string;
  headerClassName?: string;
  children: ReactNode;
}) {
  return (
    <div className="sheet-backdrop">
      <section className={cx("transaction-sheet", className)} role="dialog" aria-modal="true" aria-label={label ?? title}>
        <header className={headerClassName}>
          {leading}
          <h2>{title}</h2>
          {trailing}
        </header>
        {children}
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
