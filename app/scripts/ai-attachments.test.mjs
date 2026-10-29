import assert from "node:assert/strict";
import test from "node:test";
import { createServer } from "vite";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
try {
  const { validateEntryFiles } = await server.ssrLoadModule("/src/services/aiEntryLogic.ts");
  const { BaseFileUpload } = await server.ssrLoadModule("/src/atomic/atoms/BaseFileUpload.tsx");
  const { AssistantComposer } = await server.ssrLoadModule("/src/atomic/molecules/AssistantComposer.tsx");
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
  test("renders selected files as removable thumbnail items", () => {
    const files = [
      { name: "receipt.png", type: "image/png", size: 1024 },
      { name: "statement.pdf", type: "application/pdf", size: 2048 },
    ];
    const html = renderToStaticMarkup(React.createElement(AssistantComposer, {
      text: "", files, disabled: false, uploadDisabled: false, loading: false, canSend: true,
      hint: "Tệp riêng tư", onTextChange() {}, onFiles() {}, onClearFiles() {}, onSubmit() {},
    }));
    assert.equal((html.match(/data-file-thumbnail=/g) ?? []).length, 2);
    assert.match(html, /aria-label="Bỏ tệp receipt\.png"/);
    assert.match(html, /aria-label="Bỏ tệp statement\.pdf"/);
    assert.match(html, /data-thumbnail-kind="pdf"/);
  });
  test("renders a processing status and keeps attachments visible while loading", () => {
    const html = renderToStaticMarkup(React.createElement(AssistantComposer, {
      text: "", files: [{ name: "receipt.png", type: "image/png", size: 1024 }], disabled: true,
      uploadDisabled: true, loading: true, canSend: false, hint: "Đang đọc chứng từ",
      onTextChange() {}, onFiles() {}, onClearFiles() {}, onSubmit() {},
    }));
    assert.match(html, /data-processing-state/);
    assert.match(html, /Đang đọc chứng từ/);
    assert.match(html, /data-file-thumbnail=/);
    assert.match(html, /aria-busy="true"/);
  });
} finally {
  await server.close();
}
