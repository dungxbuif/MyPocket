import type { ReactNode } from 'react';
import { ActionButton } from '../../app/components';

export function UnavailableAction({ icon, label, reason }: { icon: ReactNode; label: string; reason: string }) {
  return <ActionButton className="sheet-row" disabled>
    {icon}
    <span className="min-w-0">{label}<small className="block text-xs text-gray-500">{reason}</small></span>
  </ActionButton>;
}
