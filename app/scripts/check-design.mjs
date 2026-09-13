import ts from 'typescript';
import { readFileSync, readdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

export function inspect(file, source) {
  const errors = [];
  const atom = file.includes('/atomic/atoms/');
  const tree = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
  const aliases = new Map();
  const variables = new Map();
  function collect(node) {
    if (ts.isImportSpecifier(node)) aliases.set(node.name.text, node.propertyName?.text ?? node.name.text);
    if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.initializer) variables.set(node.name.text, node.initializer);
    ts.forEachChild(node, collect);
  }
  collect(tree);
  function expanded(node, seen = new Set()) {
    if (!node) return '';
    if (ts.isIdentifier(node) && variables.has(node.text) && !seen.has(node.text)) return expanded(variables.get(node.text), new Set([...seen, node.text]));
    let result = node.getText(tree);
    ts.forEachChild(node, child => { if (ts.isIdentifier(child)) result += ' ' + expanded(child, seen); });
    return result;
  }
  const report = (node, rule) => errors.push(`${file}:${tree.getLineAndCharacterOfPosition(node.getStart(tree)).line + 1} ${rule}`);
  function visit(node) {
    if (!file.endsWith('/ui/theme.css') && (ts.isStringLiteralLike(node) || ts.isTemplateExpression(node)) && /#[\da-f]{3,8}\b|(?:rgb|hsl)a?\(/i.test(node.getText(tree))) report(node, 'Use a semantic theme token, not a literal color.');
    if ((ts.isJsxOpeningElement(node) || ts.isJsxSelfClosingElement(node)) && !atom && /^(button|input|select|textarea)$/.test(node.tagName.getText(tree))) report(node, 'Native controls belong to atoms; use or extend a base component.');
    if ((ts.isJsxOpeningElement(node) || ts.isJsxSelfClosingElement(node)) && !atom) {
      const tag = aliases.get(node.tagName.getText(tree)) ?? node.tagName.getText(tree);
      const cls = node.attributes.properties.find(attr => attr.name?.getText(tree) === 'className');
      const classes = expanded(cls?.initializer);
      if (/^(BaseButton|IconButton|IconBadge|BaseTextInput|BaseSelect|Heading|Text)$/.test(tag) && /\b(bg-|text-(?:ink|heading|secondary|muted|action|danger|card|xs|sm|base|lg|xl|[234]xl|\[)|font-|rounded|shadow|ring-|border-)/.test(classes)) report(node, 'Base visual overrides are forbidden: add a named variant/tone/size to the base.');
      if (/^(div|section|article)$/.test(tag) && /rounded-(?:xl|2xl|3xl|control|card)/.test(classes) && /bg-/.test(classes)) report(node, 'Use SurfaceCard for card surfaces.');
      if (/^(p|h1|h2|h3|h4)$/.test(tag)) report(node, 'Use Text or Heading for the shared type scale.');
      if (/\/(organisms|pages|templates)\//.test(file) && node.attributes.properties.some(attr => attr.name?.getText(tree) === 'style' && /\b(background|color|borderRadius|boxShadow|fontSize|fontFamily)\s*:/.test(expanded(attr.initializer)))) report(node, 'Screen inline visual styles belong to a shared base.');
    }
    ts.forEachChild(node, visit);
  }
  visit(tree);
  return errors;
}

function files(dir) { return readdirSync(dir, { withFileTypes: true }).flatMap(item => item.isDirectory() ? files(path.join(dir, item.name)) : [path.join(dir, item.name)]); }
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const root = fileURLToPath(new URL('../src', import.meta.url));
  const errors = files(root).filter(file => /\.tsx?$/.test(file)).flatMap(file => inspect(file, readFileSync(file, 'utf8')));
  const css = readFileSync(path.join(root, 'styles.css'), 'utf8');
  if (!css.includes('@import "./ui/theme.css"')) errors.push('styles.css must import the canonical theme.');
  if (/#[\da-f]{3,8}\b|(?:rgb|hsl)a?\(/i.test(css)) errors.push('styles.css must use theme variables for colors.');
  if (errors.length) { console.error(errors.join('\n')); process.exitCode = 1; }
  else console.log('Design guardrail passed: theme wired, no literal colors, native controls owned by atoms.');
}
