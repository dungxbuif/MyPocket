import type { ReactNode } from "react";
import { ArrowLeft } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Heading } from "../atoms/Heading";
import { IconButton } from "../atoms/IconButton";

type PageBackHeaderProps = {
  title: string;
  backTo: string;
  backLabel: string;
  trailing?: ReactNode;
};

export function PageBackHeader({ title, backTo, backLabel, trailing }: PageBackHeaderProps) {
  return <header className="grid grid-cols-[40px_1fr_40px] items-center pt-3">
    <Link to={backTo} aria-label={backLabel}><IconButton label={backLabel} className="rounded-full shadow-[0_2px_8px_rgb(0_0_0/0.06)]"><ArrowLeft size={19} /></IconButton></Link>
    <Heading as="h1" size="section" className="text-center text-[19px] text-slate-900">{title}</Heading>
    <span className="grid place-items-center">{trailing}</span>
  </header>;
}
