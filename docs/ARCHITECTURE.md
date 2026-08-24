# Архитектура WiFi Share

Проект состоит из Go backend (`cmd/`, `internal/`, `shared/`) и React/Vite frontend в `web/`. Конфигурация и локальные данные отделены от безопасного примера конфигурации.

Сетевые и файловые границы должны проверяться до обработки данных; LAN-only предположение не заменяет аутентификацию и авторизацию.

## Контракт зависимостей

- Источник истины (Source of truth): `go.mod` + `go.sum` для backend и `web/package.json` + `web/pnpm-lock.yaml` + `web/pnpm-workspace.yaml` для frontend.
- Канонические менеджеры — Go modules и `pnpm@11.23.0`; lock/checksum-графы независимы.
- Чистое восстановление (Clean restore): выполнить `go mod download` и `go mod verify`; во `web/` удалить только disposable `node_modules` и выполнить `pnpm install --frozen-lockfile`.
- Общие Go module/build caches и pnpm content/global virtual store разрешены; `web/node_modules` и build outputs пересоздаваемы.
- `.tools` считается project toolchain/runtime asset с отдельным lifecycle и не удаляется как dependency cache без подтверждённого mapping; пользовательские файлы и runtime config также исключены.
- Проверки: `go test ./...` и frontend production build; `pnpm lint` остаётся blocked из-за pre-existing отсутствующего `eslint.config.*` для ESLint 9.
