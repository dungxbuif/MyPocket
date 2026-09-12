import React from "react";
import { createRoot } from "react-dom/client";

import { FinancePrototypePage } from "./atomic/pages/FinancePrototypePage";
import "./styles.css";

createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <FinancePrototypePage />
  </React.StrictMode>,
);
