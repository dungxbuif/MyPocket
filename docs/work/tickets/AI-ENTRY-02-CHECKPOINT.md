# AI entry checkpoint — 2026-09-21

Owner requested committing existing work before correcting forms against the newly supplied Stitch references. This is a work-in-progress snapshot, not a completed feature.

- S3 adapter and opt-in live OCR/model tests exist. Earlier live checks reported PDF OCR and private S3 PUT/delete success; signed URL generation was tested but a real GET/body roundtrip was not checked.
- Model live extraction failed strict draft schema validation; unresolved.
- UI is partially edited; it still calls persistent session/message endpoints despite the owner's newer no-conversation requirement. It also contains an unused `newConversation` callback referencing removed copy. Batch save/delete, stored attachment linkage and PDF support in the app are incomplete.
- `rtk npm run build` on this checkpoint fails the design documentation guard because newly supplied HTML/PNG reference folders are not yet declared as visual-reference bundles. Existing AI UI tests also have an intentional failing save-all assertion and stale conversation expectations.
- Secrets stay in ignored `.env.local`, never staged. New owner-supplied `docs/design/stitch_my_pocket/` assets are left out of the old-code checkpoint and will be registered by the next UI work item.

Trace: [design](AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [backlog](../BACKLOG.md).
