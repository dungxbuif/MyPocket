import { useEffect, useState } from "react";

export function useOnlineStatus() {
  const [online, setOnline] = useState(() => window.navigator.onLine);

  useEffect(() => {
    const updateOnline = () => setOnline(true);
    const updateOffline = () => setOnline(false);
    window.addEventListener("online", updateOnline);
    window.addEventListener("offline", updateOffline);
    return () => {
      window.removeEventListener("online", updateOnline);
      window.removeEventListener("offline", updateOffline);
    };
  }, []);

  return online;
}
