# R0 web/docs deployment — 2026-09-09

Status: in_review for authenticated docs/device acceptance. Redeployment, public HTTPS smoke checks and registry publication completed at 21:25 Vietnam on 2026-09-09 after owner requested “deploy laij ddiu”. Earlier connectivity/SSH blockers below are historical; infrastructure recovery was not performed or attributed to this agent. This is not physical-device acceptance or full R0 completion.

## Redeployment and public verification — 21:25 Vietnam

- Redeployed the same previously tested images, without rebuilding or adding application changes: `rtk proxy docker compose up -d --no-deps --pull never --force-recreate --wait --wait-timeout 45 api web` in `/Users/dungxbuif/production/mypocket`. Exit 0; web/API healthy. Worker retained its previous image and start time. No migrations, env changes, database mutations or orphan cleanup.
- Public connectivity had already recovered **before** this redeploy: `rtk proxy curl -sS -I --connect-timeout 5 --max-time 10 https://money.dungxbuif.com/docs/` returned HTTP/2 303 to `/`; registry `/v2/` returned its expected 401 authentication challenge. No DNS, Caddy, Rathole, TLS or SSH changes in this run.
- After rollout, `rtk proxy curl -fsS --max-time 10 https://money.dungxbuif.com/api/v1/health/ready` returned status/database ok. Public `/` serves `index-Dz4PSmjw.js` and `index-DL_iVDHZ.css`; public JS request returned 200 and 333903 bytes. `/docs/updates/2026-09-09` returns 303 to `/` with private/no-store when unauthenticated. TLS verification enabled throughout. Authenticated docs rendering still not tested; user must log in normally.
- Deployed docs release page exists. API/worker/migrate SHA-256 values match the earlier recorded binaries. No new frontend test run was needed for this unchanged-image redeployment; earlier build/unit evidence remains linked, not represented as rerun.
- Before publication, `rtk proxy docker manifest inspect` returned no-such-manifest for both exact release tags. Then `rtk proxy docker push registry.dungxbuif.com/mypocket-web:prod-2026.09.09.1` and the corresponding API docs tag both exited 0. Subsequent manifest inspection succeeded and config digests match the deployed local image IDs.
- Published web digest: `sha256:748242c933f2994784aeecf6ff986b49d28763c049eea553ba48573c1b439d47`; API docs digest: `sha256:e91caac8e66d6a5790a4ffee8679ab0ef074957b5fe1c4394a9654b6f1a71b82`. The older local-only statements below describe the first rollout, not current availability.
- Existing rollback tags remain unchanged. Physical Safari/PWA/receipt-provider UAT, full R0 and authenticated docs acceptance are not promoted by successful deployment.

Trace: [design](../phases/R0-WEB-DOCS-RELEASE-DETAIL_DESIGN.md), [backlog](../BACKLOG.md), [context](../../CONTEXT.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md), [Docusaurus update](../../../frontend/docs/docs/updates/2026-09-09.mdx).

## Deployed scope

At 2026-09-09 13:33 UTC (20:33 Vietnam), recreated only `mypocket-api` and `mypocket-web` using `/Users/dungxbuif/production/mypocket/docker-compose.yml`:

| Service | Previous tag | Deployed tag | Local image ID |
| --- | --- | --- | --- |
| web | prod-2026.09.08.6 | prod-2026.09.09.1 | sha256:7479f967fd61138d39399b035f0ccc885273370c6de9e1a3376478473aae0692 |
| api (docs-only derivative) | prod-2026.09.08.4 | prod-2026.09.09.1-docs | sha256:7bcfab3fc40ef1d864e6c2c9037052df3671858e94a55719763e113b487bf798 |

Image repository prefix is `registry.dungxbuif.com/mypocket-`. New images currently exist **locally only**; registry manifest inspection failed before publication and no push/remote overwrite was attempted. Do not rely on pulling these tags on another host until registry access is restored and publication is verified.

Worker retained container ID `7318cdcd2236755af185b82ff55d4c9213a46dec08c73032d6e8f5be18b0879c`, started 2026-09-08T16:37:35.953240167Z, old API tag. Migrate image unchanged and no migration command ran. Compose warned about old orphan postgres/redis containers; they were not removed or modified. No secrets/env/data/ports/security policy changed. API/worker/migrate executables remain byte-identical to the previous API image; unrelated dirty backend changes were not released.

## Fresh verification

| Check | Command / result |
| --- | --- |
| Frontend tests | From frontend: `rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run`: **100 pass, 11 files**, exit 0. |
| Web image | From frontend: `rtk proxy docker build -t registry.dungxbuif.com/mypocket-web:prod-2026.09.09.1 .`: typecheck + Vite pass; index-Dz4PSmjw.js / index-DL_iVDHZ.css. |
| Docs-only API image | From repo: `rtk proxy docker build -f backend/Dockerfile.docs-release --build-arg API_BASE=registry.dungxbuif.com/mypocket-api@sha256:869b518de8f61b357388f5163d952eb39e04c635a2d7d85145f76999652722fd -t registry.dungxbuif.com/mypocket-api:prod-2026.09.09.1-docs .`: Docusaurus production build passes; no broken-link warnings. Non-blocking warnings: git timestamps unavailable in build context; API_BASE requires explicit build argument. |
| Binary integrity | `rtk proxy docker exec mypocket-api sha256sum /app/api /app/worker /app/migrate` before and after rollout, plus `docker run --rm --network none --entrypoint sha256sum` on new image: all three match. Inherited user `app`, entrypoint `/app/api` match. |
| Rollout | From production directory: `rtk proxy docker compose up -d --no-deps --pull never api web`: exit 0. |
| Readiness | `rtk proxy curl -fsS --max-time 10 http://127.0.0.1:18083/api/v1/health/ready`: status ok, database ok. API Docker health subsequently healthy. |
| Web | `rtk proxy curl -fsS --max-time 10 http://127.0.0.1:18083/`: new JS/CSS names. `curl -fsS --max-time 10 -o /dev/null -w '%{http_code} %{size_download}\n' http://127.0.0.1:18083/assets/index-Dz4PSmjw.js` through RTK verifies new bundle served. |
| Docs | File `/app/docs/updates/2026-09-09/index.html` exists in deployed API; heading and canonical money.dungxbuif.com URL verified. Unauthenticated `/docs/` before rollout and `/docs/updates/2026-09-09` after rollout return 303 to `/` with `Cache-Control: private, no-store`, as expected. No authenticated production session was fabricated; authenticated end-to-end docs rendering not verified. |
| Diff | `rtk proxy git diff --check`: pass. Existing unrelated dirty edits preserved; no commit/staging. |

Binary SHA-256 values:

```text
api     1ffeff88280a56d4061bb7ab2c214e490c7bd805632486c4f6d55fb7f6dc8382
worker  97db3b58f48f19a7b958c61904782b9db1f1fb7178b5140f6f23eb9393c7d1df
migrate f938c98dc04f573a9a4087a201ca1011d390e63d074ca4f5f02f2bad5da72306
```

## Public access blocker

Expected docs URL: `https://money.dungxbuif.com/docs/`; release page: `/docs/updates/2026-09-09`. Sign in at the app first when connectivity is restored.

Before rollout, host curl returned LibreSSL SSL_ERROR_SYSCALL for both app/docs; browser navigation returned ERR_CONNECTION_CLOSED. `rtk proxy docker manifest inspect` for each new tag failed pinging the registry with EOF, so tag availability is unknown remotely. After rollout, a disposable `node:22-alpine` container using Node HTTPS with normal TLS verification returned ECONNRESET for app, docs and registry `/v2/`. DNS lookup returned 103.82.21.202 for the app. These observations establish a preexisting public-path failure from this host, not its underlying cause or a global outage. No ingress, DNS, certificate or firewall changes were attempted; resolving infrastructure beyond this bounded release requires a separate diagnostic/fix scope. TLS verification was not disabled.

## Rollback and remaining gates

### Follow-up public-path diagnosis — 2026-09-09 20:54 Vietnam

Owner requested “tiếp”; read-only investigation continued using the systematic-debugging workflow. No infrastructure/app/config fixes attempted (0 fix cycles).

- `rtk proxy curl -sS -I --connect-timeout 10 --max-time 15 https://money.dungxbuif.com/docs/`: same TLS connection failure. HTTP public root responds 404, so this is not simply a nonexistent DNS name.
- `rtk proxy dig @1.1.1.1 money.dungxbuif.com A +short`: 103.82.21.202, matching the local resolver. macOS `scutil --proxy` shows no enabled HTTP/HTTPS proxy.
- Production README identifies VPS → Rathole → Pi5 Caddy → Mac as the ingress path. Its saved Caddy cutover file predates MyPocket and cannot establish current live routing; it was not applied.
- `rtk proxy curl -fsS --max-time 10 http://127.0.0.1:18083/api/v1/health/ready`: status/database ok. Container-to-Mac LAN request `http://10.10.0.10:18083/` returns 200. Current app containers remain running; no restart.
- `rtk proxy curl -sS -I --connect-timeout 5 --max-time 10 --resolve money.dungxbuif.com:443:10.10.0.5 https://money.dungxbuif.com/docs/` cannot connect; registry via the same Pi5 route also cannot connect. TLS validation was not weakened.
- `rtk proxy nc -vz -G 3 10.10.0.5 443`: connection refused; port 22 succeeds. This locates a failing listener/path at the documented TLS edge, but does not prove Caddy is stopped or that its current port/config matches old docs.
- Read-only SSH attempt to inspect Caddy/Rathole failed: the configured 1Password SSH agent could not sign the `Pi` ED25519 key (agent communication failure), followed by too many authentication failures. A direct BatchMode attempt without SSH config returned permission denied. No password/key search, secret extraction, new credential, or security workaround was attempted. SSH attempt has exited.

Next required input: restore/unlock the existing 1Password SSH agent for Pi5 or provide the correct current SSH host alias. Then inspect live Caddy/Rathole status/listeners/config before designing any fix. Current public-access work is blocked on authenticated read-only access, not on a failing application build. Keep full R0/device acceptance open.

Old images verified present locally. Restore only API `prod-2026.09.08.4` and web `prod-2026.09.08.6` in the exact production compose, then `rtk proxy docker compose up -d --no-deps --pull never api web` from that directory. Rollback is prepared, not executed. No DB rollback required.

Docs review: Docusaurus now covers search/error/API-key/receipt/PWA behavior and known limits; removed misleading fixed-count intro endpoint list and corrected logout cache semantics/canonical URL. Prior API audit edits preserved and clearly distinguished from backend deployment. Requirements/ERD/API implementation/auth/ADR unchanged. Context/backlog/validation/changelog reconciled. Public HTTPS, remote image publication, authenticated docs rendering and physical Safari/PWA/receipt-provider UAT remain open; no acceptance sign-off inferred from the deployment request.
