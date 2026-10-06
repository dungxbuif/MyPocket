import { Fragment, type ReactNode } from "react";
import { Heading } from "../atoms/Heading";
import { Text } from "../atoms/Text";

type MarkdownBlock =
  | { kind: "heading"; level: number; content: string }
  | { kind: "paragraph"; content: string }
  | { kind: "list"; ordered: boolean; items: string[] }
  | { kind: "table"; headers: string[]; rows: string[][] }
  | { kind: "code"; content: string };

function isTableSeparator(line: string): boolean {
  const cells = splitTableRow(line);
  return cells.length > 0 && cells.every(cell => /^:?-{3,}:?$/.test(cell.trim()));
}

function splitTableRow(line: string): string[] {
  const trimmed = line.trim().replace(/^\|/, "").replace(/\|$/, "");
  return trimmed.split("|").map(cell => cell.trim());
}

function parseBlocks(content: string): MarkdownBlock[] {
  const lines = content.replace(/\r\n?/g, "\n").split("\n");
  const blocks: MarkdownBlock[] = [];
  let index = 0;
  while (index < lines.length) {
    const line = lines[index].trim();
    if (!line) { index += 1; continue; }
    if (line.startsWith("```")) {
      const code: string[] = [];
      index += 1;
      while (index < lines.length && !lines[index].trim().startsWith("```")) code.push(lines[index++]);
      if (index < lines.length) index += 1;
      blocks.push({ kind: "code", content: code.join("\n") });
      continue;
    }
    const heading = /^(#{1,4})\s+(.+)$/.exec(line);
    if (heading) {
      blocks.push({ kind: "heading", level: Math.min(3, heading[1].length), content: heading[2] });
      index += 1;
      continue;
    }
    if (line.startsWith("|") && index + 1 < lines.length && isTableSeparator(lines[index + 1])) {
      const headers = splitTableRow(line);
      const rows: string[][] = [];
      index += 2;
      while (index < lines.length && lines[index].trim().startsWith("|")) {
        rows.push(splitTableRow(lines[index]));
        index += 1;
      }
      blocks.push({ kind: "table", headers, rows });
      continue;
    }
    const listMatch = /^(?:[-*]|\d+[.)])\s+(.+)$/.exec(line);
    if (listMatch) {
      const ordered = /^\d/.test(line);
      const items = [listMatch[1]];
      index += 1;
      while (index < lines.length) {
        const next = /^(?:[-*]|\d+[.)])\s+(.+)$/.exec(lines[index].trim());
        if (!next || /^\d/.test(lines[index].trim()) !== ordered) break;
        items.push(next[1]);
        index += 1;
      }
      blocks.push({ kind: "list", ordered, items });
      continue;
    }
    const paragraph = [line];
    index += 1;
    while (index < lines.length && lines[index].trim() && !/^(#{1,4})\s+/.test(lines[index].trim()) && !/^(?:[-*]|\d+[.)])\s+/.test(lines[index].trim()) && !lines[index].trim().startsWith("|")) {
      paragraph.push(lines[index].trim());
      index += 1;
    }
    blocks.push({ kind: "paragraph", content: paragraph.join(" ") });
  }
  return blocks;
}

function inline(content: string): ReactNode[] {
  const token = /(\*\*[^*]+\*\*|__[^_]+__|`[^`]+`|\[[^\]]+\]\(https?:\/\/[^\s)]+\))/g;
  const parts = content.split(token);
  return parts.filter(Boolean).map((part, index) => {
    if (/^(\*\*[^*]+\*\*|__[^_]+__)$/.test(part)) return <Text key={index} as="span" weight="semibold">{part.slice(2, -2)}</Text>;
    if (/^`[^`]+`$/.test(part)) return <code key={index} className="rounded bg-row px-1 font-mono">{part.slice(1, -1)}</code>;
    const link = /^\[([^\]]+)\]\(https?:\/\/[^\s)]+\)$/.exec(part);
    if (link) return <Text key={index} as="span" tone="action" weight="semibold">{link[1]}</Text>;
    return <Fragment key={index}>{part}</Fragment>;
  });
}

function heading(block: Extract<MarkdownBlock, { kind: "heading" }>, key: string): ReactNode {
  const level = block.level === 1 ? "h1" : block.level === 2 ? "h2" : "h3";
  return <Heading key={key} as={level} size={block.level === 1 ? "section" : "field"}>{inline(block.content)}</Heading>;
}

export function FinanceAssistantMarkdown({ content }: { content: string }) {
  const blocks = parseBlocks(content);
  return <div className="space-y-2" data-testid="finance-assistant-markdown">
    {blocks.map((block, index) => {
      const key = `${block.kind}-${index}`;
      if (block.kind === "heading") return heading(block, key);
      if (block.kind === "paragraph") return <Text key={key}>{inline(block.content)}</Text>;
      if (block.kind === "code") return <pre key={key} className="overflow-x-auto rounded-lg bg-row p-2 text-xs"><code>{block.content}</code></pre>;
      if (block.kind === "list") {
        const List = block.ordered ? "ol" : "ul";
        return <List key={key} className="list-inside space-y-1 text-sm text-ink">{block.items.map((item, itemIndex) => <li key={`${key}-${itemIndex}`}><Text as="span">{inline(item)}</Text></li>)}</List>;
      }
      return <div key={key} className="overflow-x-auto rounded-lg border border-line">
        <table className="min-w-full text-left text-xs"><thead className="bg-row"><tr>{block.headers.map((header, headerIndex) => <th key={`${key}-header-${headerIndex}`} className="px-2 py-2 font-semibold text-secondary"><Text as="span" size="xs" weight="semibold" tone="secondary">{inline(header)}</Text></th>)}</tr></thead>
          <tbody>{block.rows.map((row, rowIndex) => <tr key={`${key}-row-${rowIndex}`} className="border-t border-line">{block.headers.map((_, cellIndex) => <td key={`${key}-cell-${rowIndex}-${cellIndex}`} className="px-2 py-2"><Text as="span" size="xs">{inline(row[cellIndex] ?? "")}</Text></td>)}</tr>)}</tbody>
        </table>
      </div>;
    })}
  </div>;
}
