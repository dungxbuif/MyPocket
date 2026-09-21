const HOLD_MS = 500;
type Scheduler = { schedule: (fn: () => void, delay: number) => ReturnType<typeof setTimeout>; clear: (id: ReturnType<typeof setTimeout>) => void };

export function createLongPress(onClick: () => void, onHold: () => void, scheduler: Scheduler = { schedule: (fn, delay) => setTimeout(fn, delay), clear: id => clearTimeout(id) }) {
  let timer: ReturnType<typeof setTimeout> | undefined;
  let suppressed = false;
  const stop = () => { if (timer !== undefined) scheduler.clear(timer); timer = undefined; };
  return {
    start() { stop(); suppressed = false; timer = scheduler.schedule(() => { timer = undefined; suppressed = true; onHold(); }, HOLD_MS); },
    release() { stop(); },
    cancel() { stop(); suppressed = true; },
    click(keyboard = false) { if (keyboard || !suppressed) onClick(); suppressed = false; },
  };
}
