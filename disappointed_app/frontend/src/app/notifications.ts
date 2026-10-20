import { apiFetch } from "./apiClient";

export type NotificationNotice = {
  id: string;
  kind: string;
  title: string;
  body: string;
  source_type: string;
  source_id: string;
  read_at?: string;
  created_at: string;
};

export async function loadNotifications() {
  const response = await apiFetch<{ notifications: NotificationNotice[] }>("/api/v1/notifications?limit=30");
  return response.notifications;
}

export async function markNotificationRead(id: string) {
  await apiFetch(`/api/v1/notifications/${encodeURIComponent(id)}/read`, { method: "PATCH" });
}

export async function subscribeToPush() {
  if (!navigator.onLine) throw new Error("offline");
  if (!("serviceWorker" in navigator) || !("PushManager" in window) || !("Notification" in window)) throw new Error("unsupported");
  const permission = await Notification.requestPermission();
  if (permission !== "granted") throw new Error("denied");
  const registration = await navigator.serviceWorker.ready;
  const key = import.meta.env.VITE_WEB_PUSH_PUBLIC_KEY;
  if (!key) throw new Error("unconfigured");
  const subscription = await registration.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: decodeBase64URL(key) });
  const json = subscription.toJSON();
  await apiFetch("/api/v1/push-subscriptions", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ endpoint: json.endpoint, expiration_time: json.expirationTime, keys: json.keys }) });
}

function decodeBase64URL(value: string) {
  const padded = value.replace(/-/g, "+").replace(/_/g, "/") + "===".slice((value.length + 3) % 4);
  const raw = atob(padded);
  return Uint8Array.from(raw, (char) => char.charCodeAt(0));
}
