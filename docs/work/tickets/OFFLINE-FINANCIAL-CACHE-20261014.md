# Account-scoped last-known financial cache

Status: in_progress. Next approved offline runtime prerequisite, not full feature.

Before code: store an atomic encrypted snapshot of server-confirmed transactions,
wallets, categories and account timezone with fetchedAt. No pending projection is
included; all transactions must match the snapshot's authenticated owner. Namespace
by exact normalized API base URL and owner, hashed paths, AES-GCM with namespace
AAD, per-namespace Keychain key, backup-excluded protected directory and atomic
file replacement. No tokens/provider keys. Missing file returns nil; corrupt,
wrong key/lost key/unsupported schema fail closed without deleting/replacing
financial data. Only successful explicit snapshot writes replace known cache.

Cache fallback is allowed only for transient transport/408/429/5xx, never auth,
ownership mismatch, validation or malformed server responses. FetchedAt is shown
as stale, not used to imply freshness. No cached capability authorizes a send.
Wire/app composition will inject this store into transaction/overview loading;
do not claim runtime availability before that wiring and UI proof.

TDD first: persistence/restart, owner/server isolation, ciphertext/no plaintext,
wrong-key/corrupt/version denial without destructive recovery, foreign-owned rows
refused, nil on missing and controlled fallback classes. Full native regression,
guards/project/diff, docs human+agent/current ledger updates. No API/schema/prod
deployment or signed install in cache prerequisite slice.

Runtime read integration before code: inject optional cache into TransactionsViewModel
and TransactionsView. Auth composition creates protected ApplicationSupport cache
only when using its default API environment; custom API clients must inject a
matching cache explicitly, never write into production namespace by assumption.
Bind read requests to ownerID. On transient primary transaction/wallet failure,
restore entire matching snapshot, publish cachedAt and clear generic load error.
Auth/ownership errors clear rows and never fall back. Successful full reference
load saves one snapshot; supplementary metadata failures do not overwrite cache.
Use existing MPStatusMessage and MPButton retry above totals, with fetchedAt text;
cached totals remain last-known confirmed, never pending sums. UI registry and
public guide updated; visual/DynamicType/VoiceOver proof remains a separate gate.

Runtime modelRED90809 followed by70417 stillfails for503{} while offline transport
fallback passes. Root cause: ProblemDetails has all optional fields; {} decodes
successfully with nilstatus, APIClient throws problem(nil) instead of preserving
HTTP503. Normalize problem.status to authoritative HTTP status for JSON and
multipart responses; body status cannot override transport classification. Add
boundary regression for missing/contradicting body status before fix. This changes
client recovery only, not backend API or permissions.

## Local checkpoint

Encrypted atomic snapshot and default Auth→ledger composition implemented.
Custom API clients require matching cache injection. Owner-bound transient-only
fallback publishes cachedAt; shared MPStatusMessage/MPButton show stale+retry.
Complete metadata saves snapshot, incomplete references preserve existing cache.
Default cache setup failure leaves online auth usable, no offline fallback.
RuntimeRED95867→236unitsGREEN; modelRED90809/70417 identified503{} issue;
HTTPstatusRED41489→normalization at JSON/multipart boundary. Final13769exit0,
239/239PASS0skip(238units+actualAPIkeyUI),
/tmp/mypocket-cache-final-20261014.xcresult. ActualledgerUI17267exit0,1/1PASS0skip,
/tmp/mypocket-cache-ledger-ui-20261014.xcresult. Twelve synthetic server rows,
weekly/custom/cashflow/scroll checked then APIcleanup. Screenshot59D42ADA inspected
under /tmp/mypocket-cache-ledger-visual-20261014: normalsharedrows/navbar intact;
Feedback covers lastamount, knownblockeddesign/no fourthplacementguess. Synthetic
lastknowncache retains fixture rows after cleanup by design, not authoritative
server state. Base/guardtests/project/diffpass; XcodeGen includednewfiles.
PublicMarkdown guide is human+agent surface, no backendwire/schema/Swagger/ERD
change. ui-ux-pro-max targeted retry matched loading/error recovery guidance.

Remaining: stalebanner realUI/DynamicType/VoiceOver/landscape, overview/report
cache, fulloutbox/forms/runtime/foregroundevents/authlifecycle/refetch/networkE2E.
Only savedvalidsession cache read is enabled; expiredaccess+offline refresh
restore not newly enabled. No prod/sign/install/release/fullgoalcompletion.
