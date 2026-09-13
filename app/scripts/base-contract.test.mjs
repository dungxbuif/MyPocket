import { createServer } from 'vite';
import { strict as assert } from 'node:assert';
import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';

const server = await createServer({ server: { middlewareMode: true }, appType: 'custom' });
try {
  const { BaseButton } = await server.ssrLoadModule('/src/atomic/atoms/BaseButton.tsx');
  const { SurfaceCard } = await server.ssrLoadModule('/src/atomic/atoms/SurfaceCard.tsx');
  const { Progress, BudgetGauge } = await server.ssrLoadModule('/src/atomic/atoms/Progress.tsx');
  const { BaseTextInput } = await server.ssrLoadModule('/src/atomic/atoms/FormField.tsx');
  const { WalletEditorForm, EMPTY_WALLET_FORM } = await server.ssrLoadModule('/src/atomic/molecules/WalletEditorForm.tsx');
  const { BaseDonutChart } = await server.ssrLoadModule('/src/atomic/molecules/BaseCharts.tsx');
  const html = (component, props) => renderToStaticMarkup(React.createElement(component, props));
  const loading = html(BaseButton, { loading:true, children:'Save' });
  assert.match(loading, /disabled=""/); assert.match(loading, /aria-busy="true"/); assert.match(loading, /bg-brand/);
  const flat = html(SurfaceCard, {elevation:'flat',children:'x'}); assert.match(flat,/shadow-none/); assert.doesNotMatch(flat,/shadow-card/);
  assert.match(html(Progress,{value:115,label:'Food'}),/aria-valuenow="100"/);
  assert.match(html(Progress,{value:-10,label:'Food'}),/aria-valuenow="0"/);
  assert.match(html(BudgetGauge,{value:115,label:'Food'}),/stroke-danger/);
  const inline = html(BaseTextInput,{variant:'inline','aria-label':'Name'}); assert.doesNotMatch(inline,/border-line/); assert.match(inline,/focus-visible:outline-action/);
  const wallet = html(WalletEditorForm,{state:EMPTY_WALLET_FORM,saving:true,onChange(){},onSubmit(){},onCancel(){}});
  assert.equal((wallet.match(/disabled=""/g) ?? []).length,7); // four inputs, select, save and cancel
  assert.match(html(BaseDonutChart,{shares:[],center:'No data'}),/var\(--color-line\)/);
  console.log('Base contracts passed: loading, single elevation, clamped progress, danger gauge, inline input focus.');
} finally { await server.close(); }
