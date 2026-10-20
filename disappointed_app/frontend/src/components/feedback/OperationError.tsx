import { APIClientError } from '../../app/apiClient';
import { PillButton } from '../../app/components';

export type OperationFailure = { message: string; correlationID?: string };

export function operationFailure(error: unknown, message: string): OperationFailure {
  const id = error instanceof APIClientError ? error.correlationID : undefined;
  return { message, correlationID: id && /^[a-zA-Z0-9_.:-]{1,128}$/.test(id) ? id : undefined };
}

export function OperationError({ failure, onRetry, retryLabel = 'Tải lại', busy = false }: {
  failure: OperationFailure | null;
  onRetry?: () => void;
  retryLabel?: string;
  busy?: boolean;
}) {
  if (!failure) return null;
  return (
    <div className="offline-warning" role="alert">
      <p>{failure.message}</p>
      {failure.correlationID ? <p>Mã hỗ trợ: <code>{failure.correlationID}</code></p> : null}
      {onRetry ? <PillButton disabled={busy} onClick={onRetry}>{retryLabel}</PillButton> : null}
    </div>
  );
}
