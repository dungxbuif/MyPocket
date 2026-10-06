import { strict as assert } from "node:assert";
import { createServer } from "vite";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
try {
  const { FinanceAssistantMarkdown } = await server.ssrLoadModule("/src/atomic/molecules/FinanceAssistantMarkdown.tsx");
  const input = "## Hũ chi tiêu\n\n| Hũ | Đã chi | Hạn mức |\n| --- | ---: | ---: |\n| Ăn uống | 1.000.000 ₫ | 2.000.000 ₫ |\n\n- Theo dõi trong tháng này\n- Không tính chuyển ví";
  const markup = renderToStaticMarkup(React.createElement(FinanceAssistantMarkdown, { content: input }));
  assert.match(markup, /<h2[^>]*>Hũ chi tiêu<\/h2>/);
  assert.match(markup, /<table/);
  assert.match(markup, /<th[^>]*>[\s\S]*Hũ[\s\S]*<\/th>/);
  assert.match(markup, /<td[^>]*>[\s\S]*Ăn uống[\s\S]*<\/td>/);
  assert.match(markup, /<li[^>]*>[\s\S]*Theo dõi trong tháng này[\s\S]*<\/li>/);
  assert.doesNotMatch(markup, /\| Hũ \|/);
  console.log("Finance Assistant markdown rendering contract passed.");
} finally {
  await server.close();
}
