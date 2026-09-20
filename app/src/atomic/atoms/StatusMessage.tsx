import type { ReactNode } from "react";
import { SurfaceCard } from "./SurfaceCard";
import { Text } from "./Text";

export function StatusMessage({ children, tone = "muted", variant = "card" }: { children: ReactNode; tone?: "muted" | "danger"; variant?: "card" | "plain" }) {
  if (variant === "plain") return <Text role={tone === "danger" ? "alert" : "status"} tone={tone === "danger" ? "danger" : "secondary"}>{children}</Text>;
  return <SurfaceCard padding="md" tone={tone === "danger" ? "danger" : "default"} role={tone === "danger" ? "alert" : "status"}><Text tone={tone === "danger" ? "danger" : "secondary"}>{children}</Text></SurfaceCard>;
}
