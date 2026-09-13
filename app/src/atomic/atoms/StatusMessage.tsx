import type { ReactNode } from "react";
import { SurfaceCard } from "./SurfaceCard";
import { Text } from "./Text";

export function StatusMessage({ children, tone = "muted" }: { children: ReactNode; tone?: "muted" | "danger" }) {
  return <SurfaceCard padding="md" tone={tone === "danger" ? "danger" : "default"} role={tone === "danger" ? "alert" : "status"}><Text tone={tone === "danger" ? "danger" : "secondary"}>{children}</Text></SurfaceCard>;
}
