# Исторические AI facts WiFi Share

Это evidence archive, а не второй plan/status owner. Полные исходные bytes
восстанавливаются из Git `main` `ccb6e64` / parent миграционного commit.

| Original source | SHA256 |
|---|---|
| `prompts/STAGES.md` | `6e428d6c97ba1c95fd179c9f4cc97a74a423db6044faae3508e9fda97704d354` |
| `docs/AI_PLAN.md` | `6c703327ea9587d1bfe190c86fe074e2a0b575e5c0434d19857b81b7bde72a4b` |
| `docs/AI_STATUS.md` | `61396eccba229af3e91c9c3cdedd93bbc9c24ecdaf06ce3e7eab8f56a925c938` |

AI_PLAN фиксировал 2026-08-24 миграцию npm → pnpm как DONE и предлагал
governance/smoke без отдельного test evidence. AI_STATUS фиксировал
security baseline: fail-closed private/loopback bind с явным opt-in wildcard,
root/data isolation, request/file/concurrency/resource limits и quota,
temporary upload с cleanup/atomic rename, CSP/`nosniff`/attachment,
Origin check, login rate limit/TTL и `Secure` cookie для HTTPS.
Исторический документ также сообщал `go test` и frontend build PASS,
frontend lint FAIL из-за отсутствия ESLint 9 config, multi-device Wi-Fi
E2E pending и frontend chunk warning 611.37 kB. Эти результаты повторены
изолированно только в пределах, описанных в выбранном STAGES record.

Ниже в том же AI_STATUS сохранён более ранний baseline и список
«подтверждённых пробелов»: неполные request/file limits, видимые partial
uploads, отсутствие quota/concurrency/free-space/timeouts, inline active
content, unchecked root/data overlap, отсутствие login/CSRF policy и
неявный wildcard bind. Список конфликтует с более поздней implementation
и текущим кодом; его нельзя считать live open work без новой проверки.

Старый STAGES объединял Security Stage 1 с garbled Stage A prompt и
ставил peer/session flow в Stage 2. `prompts/README.md` называл
`01-security-baseline.md` текущим prompt, однако этого файла не было в
Git `main` `ccb6e64`. Security-first ADR-013 и ROADMAP определяют живой
порядок; старые launcher instructions сохранены здесь для восстановления.
