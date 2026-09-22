import assert from "node:assert/strict";
import test from "node:test";
import { createServer } from "vite";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
try {
  const { validateEntryFiles } = await server.ssrLoadModule("/src/services/aiEntryLogic.ts");
  const { BaseFileUpload } = await server.ssrLoadModule("/src/atomic/atoms/BaseFileUpload.tsx");
  test("accepts twenty files and rejects the twenty-first", () => {
    const valid = Array.from({ length: 20 }, (_, index) => ({ name: `receipt-${index}.png`, type: "image/png", size: 1024 }));
    assert.equal(validateEntryFiles(valid), "");
    assert.match(validateEntryFiles([...valid, { name: "receipt-20.png", type: "image/png", size: 1024 }]), /20/);
  });
  test("rejects duplicate files instead of silently duplicating them", () => {
    assert.match(validateEntryFiles([{ name: "same.png", type: "image/png", size: 1024 }, { name: "same.png", type: "image/png", size: 1024 }]), /trùng/);
  });
  test("renders an accessible drop zone", () => {
    const html = renderToStaticMarkup(React.createElement(BaseFileUpload, { label: "Chứng từ", onFiles() {}, maxFiles: 20 }));
    assert.match(html, /data-dropzone/);
    assert.match(html, /Kéo thả/);
  });
} finally {
  await server.close();
}
