---
artifact_type: decision_contract
id: FEEDBACK-20261019-PRODUCT-HELP-RAG-CONTRACT
status: local_handbook_v1
updated: 2026-10-19
---

# Product help and release-aware retrieval contract

Feedback `1a59aa1b` requests broader, release-aware AI help. The current
`get_product_help` implementation is intentionally deterministic: it searches
a small source-controlled Vietnamese handbook, returns a version and stable
source IDs, and returns `not_found` instead of inventing an answer. It never
reads owner data or accepts arbitrary URLs.

## Source contract

Every indexed section must have a stable ID, handbook version, title, excerpt,
supported release range, owner (`product` or `engineering`), and last-reviewed
date. Retrieval may return only sections from an enabled source manifest. A
release note is not authoritative for financial behavior unless it links to a
current product contract.

## Refresh contract

Until an authenticated source and refresh job are selected, the checked-in
handbook remains the source of truth. A future refresh must be atomic and
versioned, preserve the previous manifest for rollback, reject malformed or
prompt-injected content, and expose the active version in each response. It
must not fetch model/provider instructions from retrieved text.

## Answer and privacy contract

- Navigation questions call `get_product_help` and cite returned source IDs.
- Unknown or unsupported features produce an explicit abstention.
- Retrieval is independent of `Principal.OwnerID`; no transaction, amount,
  note, token, credential or arbitrary external URL may enter the index.
- Live provider grounding remains opt-in and uses synthetic tool results only.

## Current proof

`TestAdvisorProductHelpIsVersionedAndSourceIdentifiable`,
`TestAdvisorProductHelpNoMatchDoesNotDumpHandbook`, and
`TestAdvisorProductHelpRejectsBoundsAndUnknownFields` cover the local boundary;
the opt-in live probe covers provider tool routing without accessing user data.
External release feeds and semantic/vector retrieval are not enabled until an
authoritative source and refresh owner are approved.
