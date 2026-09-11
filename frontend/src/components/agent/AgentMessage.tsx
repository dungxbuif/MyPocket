import { Card } from "../ui/card";

export function AgentMessage({ role, text }: { role: "user" | "agent"; text: string }) {
  return <Card className={role === "user" ? "agent-message agent-message-user" : "agent-message"} aria-label={role === "user" ? "Tin nhắn của bạn" : "Trả lời của trợ lý"}>
    <small>{role === "user" ? "Bạn" : "MyPocket Agent"}</small>
    <p>{text}</p>
  </Card>;
}
