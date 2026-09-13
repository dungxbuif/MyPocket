import { test } from 'node:test';
import assert from 'node:assert/strict';
import { inspect } from './check-design.mjs';
test('rejects controls in screens and molecules', () => {
  for (const layer of ['organisms', 'molecules', 'pages']) {
    for (const tag of ['button', 'input', 'select', 'textarea']) assert.ok(inspect(`/src/atomic/${layer}/Example.tsx`, `const x = <${tag} />`).length);
  }
});
test('allows native controls only in atoms', () => assert.deepEqual(inspect('/src/atomic/atoms/BaseButton.tsx', 'const x = <button />'), []));
test('rejects hex and rgb in classes including templates', () => {
  assert.ok(inspect('/src/Test.tsx', 'const x = <div className="bg-[#006e1c]" />').length);
  assert.ok(inspect('/src/Test.tsx', 'const x = `shadow-[0_4px_20px_rgb(0_0_0/0.03)] ${extra}`').length);
});
test('ignores commented JSX and accepts semantic composition', () => assert.deepEqual(inspect('/src/Test.tsx', '// <button />\nconst x = <BaseButton className="w-full" />'), []));
test('rejects restyling bases and local cards', () => {
  assert.ok(inspect('/src/Test.tsx', 'const x = <BaseButton className="bg-action rounded-xl" />').length);
  assert.ok(inspect('/src/Test.tsx', 'const x = <div className="rounded-xl bg-card p-4" />').length);
  assert.deepEqual(inspect('/src/Test.tsx', 'const x = <Text tone="danger" size="sm" />'), []);
});
test('checks imported aliases and local class constants', () => assert.ok(inspect('/src/Test.tsx', 'import {BaseButton as Button} from "./BaseButton"; const look="bg-danger"; const x=<Button className={look}/>').length));
