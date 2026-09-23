# BaseFab hold activation

Consumer: [AI entry](../../pages/assistant/README.md). Implementation: `app/src/atomic/atoms/BaseNavigation.tsx` and `longPress.ts`.

Optional onLongPress extends BaseFab without changing ordinary callers. Primary pointer down starts 500 ms hold; pointer up before threshold retains normal click. Hold fires once and consumes release click. Any pointer move, pointercancel, lost capture, element/window blur or unmount cancels the timer and pointer click. Keyboard click still invokes ordinary onClick; the manual editor exposes a separate labeled AI button. Native context menu is suppressed only for hold-enabled FAB. Shared FAB focus/tokens/56px target unchanged. Proof: `node scripts/ai-entry.test.mjs` and browser fixture.
