# Состояние WiFi Share

## Governance migration — 2026-08-24

- Security stage объединён в `prompts/STAGES.md`; project overlay — PASS.
- `go test ./...` — PASS с изолированными TMP/GOCACHE; frontend production build — PASS.
- Frontend lint — FAIL: repository использует ESLint 9, но не содержит `eslint.config.js`; автоматическая config migration не выполнялась.
- Репозиторий находится в `${PROJECTS_ROOT}/wifi-share` (локальный default: `~/wifi-share`); dependency-manager migration интегрирована и опубликована в `main`.

## Актуализация: Этап A — базовая безопасность

- Реализованы и покрыты Go-тестами: private/loopback bind и явный opt-in `0.0.0.0` для доверенной LAN, запрет неявного wildcard/port `0`, изоляция `root` и `data`, лимиты request/file/count/concurrency, квота 10 GiB и Windows free-space reserve, пяти минутные read/write timeouts, атомарная загрузка через скрытый temporary directory, stale cleanup, collision protection, CSP/`nosniff` и attachment для всех типов кроме консервативного inline allowlist.
- Cookie-мутаторы требуют exact `Origin`; login получает rate limit, TTL cleanup и bounded in-memory registry. Cookie получает `Secure` при HTTPS.
- UI отображает подтверждённый bind address через `/api/health`; документация требует явного private LAN address для доступа из Wi-Fi.
- Проверки: `go test ./...` проходит; clean restore и `pnpm build` в `web/` проходят. `pnpm lint` блокирован отсутствующим `eslint.config.*` для ESLint 9 (pre-existing tooling gap). Живой multi-device Wi-Fi E2E имеет статус `BLOCKED_BY_BACKEND_WIFI_SHARE` без стенда из двух устройств.
- Dependency manager migration 2026-08-24: frontend использует `pnpm@11.23.0`, единственный `web/pnpm-lock.yaml` и global virtual store. Точечные overrides `brace-expansion@5.0.9`, `js-yaml@4.3.1`, `nanoid@3.3.18` устранили все найденные advisories; `pnpm audit` — PASS.
- Frontend build сохраняет предупреждение о chunk `611.37 kB`; это pre-existing performance debt, не ошибка dependency restore.

Этап A реализован в рабочей ветке и ожидает ручной E2E-проверки LAN; к этапу B не переходить без отдельной задачи.

## Текущий baseline

- Go `net/http` backend и React/TypeScript/Vite frontend.
- Одна динамически выбираемая общая папка; клиент использует непрозрачные ID, а сервер проверяет выход за корень и внешние символические ссылки.
- Анонимное чтение и воспроизведение; мутации требуют общего пароля и 24-часовой `HttpOnly`/`SameSite=Strict` сессии в памяти.
- Потоковая отдача с `Range`, MIME и ETag через `http.ServeContent`.
- SQLite в WAL-режиме для журнала операций, Windows Recycle Bin, системный трей и плейлист текущей папки.
- Имена загрузок проверяются, а коллизии отклоняются через `O_EXCL`.

## Подтверждённые пробелы

- `ParseMultipartForm(64 << 20)` ограничивает память multipart, но не является общим лимитом тела запроса или файла.
- Загрузка пишет сразу под конечным именем. Обработка ошибки удаляет partial file, однако crash процесса может оставить его видимым.
- Нет квот, ограничения конкурентных загрузок, контроля свободного места, checksum, timeout и отмены операции.
- HTML/SVG и другое активное содержимое может отдаваться `inline`; полной политики изоляции и CSP пока нет.
- Запрет пересечения `root` и `data` документирован, но не проверяется при запуске и смене общей папки.
- Вход не имеет throttling/rate limit; CSRF/Origin-модель для cookie-аутентификации не зафиксирована тестами.
- Значение `:8080` слушает все интерфейсы, а UI ещё не объясняет фактическую сетевую экспозицию и правила firewall.

## Следующий этап

`Этап A — Базовая безопасность` из `docs/ROADMAP.md` и `prompts/01-security-baseline.md`. До его завершения возобновляемые передачи, сопряжение, обнаружение, WebSocket и эксперименты производительности не являются текущей реализацией.

## Источники истины

- реализованные архитектурные решения: `ARCHITECTURE.mdx` и код;
- будущие требования: `docs/SPECIFICATION.md`;
- порядок: `docs/ROADMAP.md`;
- угрозы и проверки: `docs/SECURITY.md` и `docs/TESTING.md`.
