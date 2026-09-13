import type { ReactNode } from "react";
import { ArrowLeft } from "lucide-react";
import { BaseLink } from "../atoms/BaseLink";
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
    <BaseLink to={backTo} label={backLabel}><ArrowLeft size={19} /></BaseLink>
    <Heading as="h1" size="screen" className="text-center">{title}</Heading>
    <span className="grid place-items-center">{trailing}</span>
  </header>;
}
