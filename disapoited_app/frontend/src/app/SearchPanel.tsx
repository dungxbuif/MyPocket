import { useEffect, useState } from 'react';
import { Card } from './components';
import { searchRecords, type SearchResult } from './analytics';
import { SearchInputField } from '../components/inputs/SearchInputField';
import { OperationError, operationFailure, type OperationFailure } from '../components/feedback/OperationError';

type SearchState = {
  query: string;
  attempt: number;
} & ({ status: 'loading' } | { status: 'success'; results: SearchResult[] } | { status: 'error'; failure: OperationFailure });

export function SearchPanel({ userID, online, query, onQueryChange }: {
  userID: string;
  online: boolean;
  query: string;
  onQueryChange: (query: string) => void;
}) {
  const [state, setState] = useState<SearchState | null>(null);
  const [attempt, setAttempt] = useState(0);
  const canSearch = online && Boolean(userID) && query.trim().length >= 2;

  useEffect(() => {
    if (!canSearch) {
      setState(null);
      return;
    }
    let cancelled = false;
    setState({ query, attempt, status: 'loading' });
    void searchRecords(userID, query).then(results => {
      if (!cancelled) setState({ query, attempt, status: 'success', results });
    }).catch(error => {
      if (!cancelled) setState({ query, attempt, status: 'error', failure: operationFailure(error, 'Không thể tải kết quả tìm kiếm. Vui lòng thử lại.') });
    });
    // A closed panel, changed query/owner or connectivity invalidates the result.
    return () => { cancelled = true; };
  }, [query, userID, online, canSearch, attempt]);

  const current = canSearch && state?.query === query && state.attempt === attempt ? state : null;
  const loading = canSearch && (!current || current.status === 'loading');

  return (
    <Card className="search-panel" aria-label="Tìm kiếm">
      <SearchInputField autoFocus ariaLabel="Tìm kiếm giao dịch và ví" placeholder="Tìm ví, giao dịch, nhóm..." value={query} onChange={onQueryChange} />
      {!online ? <p className="notification-status" role="status">Tìm kiếm cần kết nối mạng.</p> : null}
      {loading ? <p className="notification-status" role="status">Đang tìm kiếm…</p> : null}
      <OperationError failure={current?.status === 'error' ? current.failure : null} retryLabel="Thử lại" onRetry={() => setAttempt(value => value + 1)} busy={loading} />
      {current?.status === 'success' ? (
        current.results.length === 0 ? <p className="notification-status" role="status">Không tìm thấy kết quả.</p> :
          current.results.map(item => <div className="search-result" key={`${item.kind}-${item.id}`}><strong>{item.label || 'Không có ghi chú'}</strong><small>{item.detail ?? item.kind}</small></div>)
      ) : null}
    </Card>
  );
}
