# Learning Log — WiFi Share

## 2026-08-10 — Переносимый слепок проекта

### Задача

Создать Markdown-контекст для передачи в ChatGPT: функция продукта, фактическая архитектура, API, состояние разработки, MCP, hooks, skills и субагенты.

### Что исследовали

README и старые ADR сравнили с Go backend, React frontend, tests, local config shape и внешним WiFi Share preset. Это позволило отделить готовые функции от roadmap.

### Основные команды

`rg --files` — карта source-controlled файлов без runtime/toolchain artifacts.

`git status --short` и `git log` — проверка чистого baseline и версии исходников.

`go test ./cmd/... ./internal/...` — unit/integration-level проверка Go packages.

`npm run build --prefix web` — TypeScript и production Vite build.

`npm run lint --prefix web` — проверка frontend rules; выявила отсутствующий ESLint flat config.

### Как была устроена проблема

Корневой `ARCHITECTURE.mdx` одновременно описывал текущий код и будущие pairing, thumbnails, WebSocket и resumable transfer. External AI preset также существовал отдельно от repo, поэтому наличие файлов могло быть ошибочно принято за активные hooks/agents.

### Что изменили

Созданы current-state `PROJECT_SNAPSHOT.md`, `ARCHITECTURE.md`, `DECISIONS.md`, `DESIGN.md`, `AI_STATUS.md`, `AI_PLAN.md` и `ROADMAP.md`. Runtime-код, config, база и shared files не менялись.

### Почему выбран такой подход

Один подробный слепок удобен для передачи в ChatGPT. Короткие канонические документы дают устойчивые точки входа для hooks и будущих сессий, не дублируя весь отчёт.

### Что пошло не так

- Исходный cwd указывал на несвязанный Arduino-проект `Cube`; поиск по workspace нашёл правильный `C:\MAMP\htdocs\wifi-share`.
- Первая полная Go-проверка превысила 120 секунд во время первичной компиляции; повторный целевой запуск с большим timeout прошёл.
- Frontend lint не дошёл до анализа кода: ESLint 9 не нашёл `eslint.config.*`.

### Проверки

- Go tests: PASS.
- Frontend build: PASS, warning о chunk ~611 kB.
- Frontend lint: FAIL по отсутствующей конфигурации; зафиксировано как known issue.
- Документы дополнительно проверяются Markdown/secret scan и `git diff --check` перед commit.

### Как повторить вручную

1. Выполнить `rg --files` и исключить ignored runtime data.
2. Сопоставить заявленные функции README/ADR с routes и frontend calls.
3. Проверить `.codex`/`.agents` отдельно в repo, preset и user config.
4. Никогда не вставлять в отчёт local password, database content или ShareDir paths/content.
5. Запустить tests/build/lint и записать как успешные, так и неуспешные проверки.

### Что стоит изучить

Go `http.ServeContent`, multipart limits, atomic file publishing, same-origin active content, ESLint flat config и distinction между AI preset и активной Codex configuration.
