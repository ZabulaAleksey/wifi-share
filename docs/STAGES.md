# Этапы WiFi Share

- Stage ID: WFS-SECURITY-BASELINE

Единственный владелец текущего этапа, статуса, blockers, evidence и NEXT.
Исторические утверждения AI-файлов сверены с кодом и сохранены в
`docs/notes/legacy-ai-state-evidence.md`; они не являются текущим статусом.

## WFS-SECURITY-BASELINE — Этап A: базовая безопасность

- Status: partial
- NEXT: WFS-SECURITY-BASELINE-VERIFY
- Blockers: frontend lint не запускается: ESLint 9 не находит
  `eslint.config.*`; `pnpm install --frozen-lockfile` в isolated clone
  завершился `ERR_PNPM_IGNORED_BUILDS` для `esbuild`, после чего
  проверены локальные binaries. Живой обмен по Wi-Fi между двумя устройствами,
  firewall и экспозиция private LAN ещё не подтверждены.
- Evidence: `go test ./...` — PASS на Go 1.25.14 с task-local cache;
  `.bin/tsc -b` и `.bin/vite build` — PASS; `.bin/eslint .` — FAIL
  из-за отсутствующего config. Go security regression tests в
  `cmd/wifi-share` и `internal/app` проходят; браузерная сборка сама по себе
  не подтверждает живой LAN E2E. `docs/SPECIFICATION.md`,
  `docs/SECURITY.md`, `docs/TESTING.md`, ROADMAP и код сверены с прежним
  AI_STATUS на GitHub `main` `ccb6e64`; ранний список «подтверждённых
  пробелов» в AI_STATUS противоречит более позднему Stage A implementation
  и не переносится как live blocker без проверки.

Scope: закрыть tooling gate для ESLint, восстановить frozen install с
явным разрешением обязательного `esbuild` build script по ADR-000,
выполнить негативные security checks и контролируемый обмен между двумя
доверенными устройствами на private LAN. Изменение сетевой policy и
новые transfer capabilities не входят в этот срез.

DoD: `go test ./...`, frozen frontend restore, lint/build и security
regression checks PASS; manual client → backend сценарий с private bind,
login/Origin, upload/download и отказом для неразрешённого доступа
зафиксирован без секретов. До всех gates статус не повышать.

### Действия пользователя и интеграция

- `USER-WFS-LAN-E2E` — `PENDING`, condition: есть два доверенных устройства
  в одной private LAN и безопасный тестовый файл. Действие: после готового
  lint/restore gate запустить сервер на конкретном private IP, убедиться,
  что UI показывает тот же bind address, открыть клиент со второго
  устройства, выполнить login, upload/download тестового файла, проверить
  отказ cookie mutation при чужом Origin и отсутствие доступа вне
  разрешённой папки. Ожидаемое evidence: адреса без credentials,
  результаты запросов и локальный test log с PASS/FAIL; это разблокирует
  terminal LAN E2E gate для Stage A. Не публиковать сервер в internet.
- `USER-WFS-STAGES-INTEGRATION` — `PENDING`, condition: isolated
  `feature/docs-stages-canonical` commit/push, compatibility matrix и
  проверки готовы. Действие: разрешить merge этой точной ветки в `main`
  после review сохранённых AI-фактов и unchanged product code/tests.
  Ожидаемое evidence: clean main ancestry и GitHub read-back с
  `docs/STAGES.md` без `prompts/STAGES.md`, `docs/AI_PLAN.md`,
  `docs/AI_STATUS.md`; canonical adapter возвращает `partial` и
  `WFS-SECURITY-BASELINE-VERIFY`. Это разблокирует единый state owner.

## WFS-TRANSFER-RELIABILITY — Этап B: надёжная передача

- Status: planned
- Depends on: WFS-SECURITY-BASELINE completed
- NEXT: deferred until Stage A terminal gates pass
- Evidence: будущие требования и порядок принадлежат
  `docs/SPECIFICATION.md` и `docs/ROADMAP.md`; старый STAGES предлагал
  peer/session flow как Этап 2, но security-first ADR-013 и ROADMAP
  требуют сначала завершить Stage A.

Scope: resumable sessions, chunks, checksum, retry и safe finalization
с live backend/frontend проверкой. DoD определится по SPEC и Stage A evidence
перед запуском этапа; текущая implementation claim отсутствует.
