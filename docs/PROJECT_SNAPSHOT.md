# WiFi Share — переносимый слепок проекта

> Сформировано: 2026-08-10. Базовая версия исходного кода: `7471752` (`origin/main`, коммит `Add tray controls and folder playlists`). Рабочая ветка документации: `docs/project-snapshot`.

Этот файл предназначен для передачи в ChatGPT или другому разработчику как стартовый контекст для развития и debugging. Он описывает фактический код, а запланированные возможности помечает отдельно. Секреты, содержимое общей папки, локальная SQLite и значение пароля намеренно не включены.

## 1. Назначение проекта

WiFi Share — локальное desktop/web-приложение для доступа к файлам компьютера с других устройств в той же сети. На компьютере запускается Go HTTP-сервер и Windows tray; телефон, планшет, телевизор или другой компьютер открывает React-интерфейс по LAN-адресу.

Фактически реализовано:

- просмотр содержимого одной выбранной общей папки и вложенных каталогов;
- поиск по текущему открытому каталогу;
- открытие/скачивание файлов и HTTP Range для медиа через `http.ServeContent`;
- аудио/видео-плеер, очередь медиа текущей папки и автоматический переход к следующему элементу;
- пауза плееров в других вкладках того же браузерного профиля через `BroadcastChannel`;
- пароль полного доступа и 24-часовая in-memory session cookie;
- загрузка файлов с клиентским прогрессом, создание папок, переименование и удаление;
- перемещение удаляемых объектов в системную корзину Windows;
- SQLite-журнал операций;
- Windows tray: открыть приложение, показать URL, выбрать общую папку, открыть конфиг, завершить сервер;
- адаптивный тёмный web UI для desktop/mobile.

Не реализовано, хотя часть пунктов упоминается как целевая архитектура:

- pairing по PIN или QR, управление устройствами и ролями;
- WebSocket-события и автоматическое обновление листинга;
- discovery устройств/сервера;
- chunked/resumable upload, checksum и восстановление передачи;
- миниатюры WebP/FFmpeg и серверный thumbnail cache;
- PWA, autostart, installer, firewall setup;
- серверная координация единственного плеера между разными устройствами;
- ZIP на лету и API восстановления из корзины;
- UI текстового редактора. Backend `PUT content` существует, но frontend его не вызывает.

## 2. Технологический стек

| Область | Технология | Роль |
|---|---|---|
| Backend | Go 1.25+, фактически проверено на Go 1.26.5 | HTTP API, файловые операции, сессии, tray runtime |
| HTTP | стандартный `net/http` ServeMux | маршруты, static SPA, Range/conditional content |
| База | `modernc.org/sqlite` 1.38.2 | локальный audit log, WAL |
| Windows tray | `github.com/getlantern/systray` 1.2.2 | управление запущенным приложением |
| Frontend | React 19.1 + TypeScript 5.8 | файловый UI и локальное состояние |
| Server state | TanStack Query 5 | загрузка и invalidation листинга/auth state |
| Media | Vidstack 1.12 | аудио/видео controls |
| Icons | Lucide React | иконки файлов и действий |
| Build | Vite 7, npm | dev proxy и production bundle |

Модуль Go: `github.com/local/wifi-share`.

## 3. Архитектура выполнения

```text
Телефон / браузер / телевизор
          |
          | HTTP, обычно LAN :8080/:8081
          v
  Go net/http server (cmd/wifi-share)
          |
          +-- /api/auth/* ----------------> in-memory sessions
          |
          +-- /api/files* ----------------> filesystem adapter
          |                                  |
          |                                  +--> выбранный ShareDir
          |                                  +--> Windows Recycle Bin
          |
          +-- audit(action, item_id) ------> data/wifi-share.db (SQLite WAL)
          |
          +-- /* --------------------------> web/dist React SPA

Windows tray
    |-- выбирает новую ShareDir через SetShareDir + RWMutex
    |-- сохраняет root в config.local.json
    |-- открывает URL/config и выполняет graceful shutdown

Вкладки одного браузера
    +-- BroadcastChannel("wifi-share-media") --> пауза другого плеера
```

Ключевые границы доверия:

1. LAN-клиент нельзя считать доверенным, даже в домашней Wi-Fi сети.
2. Клиент никогда не должен передавать абсолютный путь, пригодный для прямого использования.
3. `config.local.json`, `data/` и выбранный `ShareDir` являются локальными runtime-данными, а не частью исходников.
4. Незалогиненный клиент имеет read-only доступ к именам и содержимому всех файлов внутри `ShareDir` — это сознательная текущая модель продукта.

## 4. Карта репозитория

```text
wifi-share/
├── cmd/wifi-share/
│   ├── main.go              # config -> app.New -> URL -> platform runtime
│   ├── config.go            # localConfig, load/save, LAN IPv4 URL discovery
│   ├── tray_windows.go      # Windows systray, folder picker, shutdown
│   └── tray_other.go        # blocking HTTP server на других OS
├── internal/app/
│   ├── app.go               # HTTP API, filesystem safety, SQLite, static SPA
│   ├── auth.go              # login/logout/session middleware
│   ├── runtime.go           # Shutdown и динамическая смена ShareDir
│   └── trash_*.go           # Windows Recycle Bin / unsupported adapter
├── web/
│   ├── src/App.tsx          # главный файловый интерфейс
│   ├── src/api.ts           # типизированный REST client + XHR upload progress
│   ├── src/MediaViewer.tsx  # Vidstack + BroadcastChannel
│   ├── src/styles.css       # фактическая design system и responsive rules
│   └── vite.config.ts       # Vite, ES2019, `/api` -> localhost:8080
├── config.example.json      # безопасный шаблон настроек
├── config.local.json        # локальный секрет, ignored, не передавать
├── data/                    # SQLite runtime, ignored
├── shared/                  # пример общей папки, ignored
├── ARCHITECTURE.mdx         # исходные ADR и часть целевого плана
├── docs/                    # каноническая документация и этот слепок
├── go.mod / go.sum
└── web/package*.json
```

Runtime/сборочные артефакты `.tools/`, `web/node_modules/`, `web/dist/`, `data/`, `shared/`, `*.exe` исключены через `.gitignore`.

## 5. Backend: основные типы и жизненный цикл

### Запуск

`cmd/wifi-share/main.go`:

1. Обязательно читает `config.local.json`; без файла или непустого пароля процесс завершается.
2. CLI-флаги `-addr`, `-root`, `-data`, `-web` перекрывают соответствующие поля config.
3. Создаёт `app.App`.
4. Печатает `localhost` и найденные активные не-loopback IPv4 URL.
5. На Windows запускает сервер в goroutine и блокируется на systray; на других OS вызывает `ListenAndServe` напрямую.

### `app.Config`

```go
type Config struct {
    Address  string
    ShareDir string
    DataDir  string
    WebDir   string
    Password string
}
```

### `app.App`

Хранит конфигурацию, текущий абсолютный `root`, `RWMutex` для смены папки, SQLite connection, `http.Server`, router, map активных сессий и platform trash adapter.

`New`:

- требует непустой пароль;
- нормализует `ShareDir`, создаёт `ShareDir` и `DataDir`, если их нет;
- открывает `data/wifi-share.db`;
- включает SQLite WAL;
- создаёт таблицу `operations`;
- регистрирует маршруты;
- оборачивает router в request logging и security headers;
- настраивает `ReadHeaderTimeout=10s`, `IdleTimeout=90s`.

### Идентификаторы файлов

- корень имеет специальный ID `root`;
- остальные ID — Base64 URL без padding от относительного пути с `/`;
- `decodeID` запрещает абсолютный путь и начальный `..`;
- `resolve` повторно строит путь под текущим root, проверяет `filepath.Rel` и уже существующие symlink через `EvalSymlinks`;
- имя отдельного файла/папки проходит `safeName`: пустые строки, `.`, `..`, `/`, `\` и составные пути запрещены.

## 6. REST API

Все ответы ошибок имеют вид `{"error":"message"}`. JSON decoder ограничен 1 MiB и запрещает неизвестные поля.

| Метод и маршрут | Auth | Поведение |
|---|---:|---|
| `GET /api/health` | нет | `{"status":"ok"}` |
| `GET /api/auth/status` | нет | сообщает состояние cookie-session |
| `POST /api/auth/login` | нет | проверяет пароль, создаёт 32 random bytes token на 24 часа |
| `DELETE /api/auth/session` | нет | удаляет текущую session и cookie |
| `GET /api/files?parent={id}` | нет | список текущей папки, folders first, затем case-insensitive name |
| `GET /api/files/{id}/content` | нет | inline content, MIME по расширению, Range и Last-Modified conditional requests через `ServeContent` |
| `POST /api/files/{id}/upload` | да | multipart field `files`, создание без overwrite (`O_EXCL`) |
| `POST /api/files/{id}/folders` | да | создаёт один каталог |
| `PUT /api/files/{id}/content` | да | заменяет существующий файл через temp + rename, заявленный лимит 5 MiB |
| `PATCH /api/files/{id}` | да | переименование, root запрещён |
| `DELETE /api/files/{id}` | да | системная корзина, root запрещён |

Авторизация:

- пароль хешируется SHA-256 только на время constant-time compare; в config он остаётся plaintext;
- session token хранится только в памяти и теряется при рестарте;
- cookie: `HttpOnly`, `SameSite=Strict`, `Path=/`, expiry 24 часа;
- `Secure` отсутствует, потому что текущий продукт работает по локальному HTTP;
- rate limit/lockout отсутствуют.

## 7. Данные и состояние

### `config.local.json`

```json
{
  "address": ":8080",
  "root": "C:\\path\\to\\shared",
  "data": "./data",
  "web": "./web/dist",
  "password": "REDACTED"
}
```

Локально на момент слепка:

- config существует и ignored;
- пароль задан, значение не читалось в отчёт;
- address настроен на `0.0.0.0:8081`;
- root, data и web существуют;
- data не находится внутри root.

### SQLite

Реально используется только таблица:

```sql
operations(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  action TEXT NOT NULL,
  item_id TEXT NOT NULL,
  details TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
)
```

Actions: `upload`, `create-folder`, `rename`, `edit`, `recycle`. API чтения аудита, retention и UI отсутствуют. Настройки, токены и thumbnail metadata в SQLite пока не хранятся, несмотря на более широкое описание в старом `ARCHITECTURE.mdx`.

## 8. Frontend и пользовательские потоки

`App.tsx` держит breadcrumbs, строку поиска, активный media item, login dialog и список upload progress. TanStack Query кэширует `files` и `auth`; query stale time — 5 секунд, при фокусе окна выполняется refetch.

Основные сценарии:

1. Открытие папки добавляет breadcrumb и запрашивает её listing.
2. Обычный файл открывается в новой вкладке через `/content`.
3. Audio/video открывается в fixed media dock.
4. Media queue строится из всех audio/video текущего listing; после `ended` выбирается следующий.
5. Вход обновляет query cache, после чего показываются upload/create/rename/delete controls.
6. Каждый выбранный файл отправляется отдельным параллельным XHR; progress показывается отдельно.
7. После мутации invalidates только listing текущей папки.

UI не показывает backend error для create/rename/delete; ошибки login и upload обработаны отдельно.

## 9. Фактический дизайн UI

- тёмная зелёно-графитовая тема;
- фон `#07100e`, surfaces `#0c1714`/`#111f1b`, accent mint `#80e5bc`;
- system/Inter-like sans stack;
- desktop layout: fixed sidebar 245 px + content;
- file panel — табличная сетка Name / Modified / Size / Actions;
- типы файлов различаются цветными Lucide icons;
- breakpoint 900 px превращает sidebar в верхнюю строку;
- breakpoint 650 px скрывает table header/date/size и текст action buttons;
- состояния: loading, error, empty/search-empty, read-only notice, modal login, upload progress, media dock;
- светлая тема отсутствует.

Каноническое описание: `docs/DESIGN.md`.

## 10. MCP, hooks, skills и субагенты

Важно различать файлы в самом репозитории, внешний AI preset и реально установленную конфигурацию пользователя.

### В самом репозитории

- нет `AGENTS.md`, `.codex/` и `.agents/`;
- нет project-local MCP server, hook или custom-agent runtime;
- приложение WiFi Share само не зависит от MCP и не вызывает AI API.

### Внешний project preset

Preset расположен вне репозитория:

`C:\MAMP\htdocs\ai-dev-team-codex-kit\projects\wifi-share`

Он задаёт:

- `multi_agent = true`;
- `hooks = true`;
- максимум 4 concurrent agent threads;
- project rule: `git push` только после prompt/approval;
- server override: LAN peers недоверенные, пути клиента нельзя использовать напрямую, bind должен быть явным, transfer должен быть resumable или clean-fail.

Определены три специализированных субагента, но их TOML не установлен в пользовательский `~/.codex/agents` и не скопирован в репозиторий:

| Агент | Назначение | Effort |
|---|---|---|
| `network_protocol_engineer` | discovery, transfer, reconnect, chunking, backpressure, checksum, throughput | high |
| `wifi_security_engineer` | LAN exposure, pairing, auth, path/filename/content safety, resource caps | high |
| `windows_network_engineer` | Windows startup, firewall, interfaces, practical installation | medium |

Preset также ссылается на глобальные роли `architect`, `planner`, `reviewer`, `test_engineer`, `security_reviewer`, `performance_engineer` и built-in `explorer`; они не являются частью исходников WiFi Share.

Локальный skill `wifi-share-stage` существует только в preset. Его workflow: выбрать ровно один roadmap stage, прочитать status/plan/architecture, подключить минимум специалистов, разделить ownership, выполнить тесты/review и обновить документацию. В текущем репозитории skill не установлен.

### MCP

Preset рекомендует:

- Context7 — библиотечная документация;
- GitHub — repository/PR context;
- Playwright — только для web UI;
- Chrome DevTools — только для browser/network diagnosis;
- database/filesystem MCP не требуются.

Безопасная проверка глобального пользовательского config на этой машине нашла секции Context7, GitHub и Chrome DevTools. Это machine-local состояние, не переносимая конфигурация проекта. Project Playwright MCP не обнаружен; браузерное тестирование может предоставляться отдельным plugin/tool.

### Hooks

В библиотеке AI Dev Team есть пример `hooks.json`:

- `SessionStart` запускает `session_context.py` и подмешивает первые части `docs/AI_STATUS.md`, `docs/AI_PLAN.md`, `docs/ARCHITECTURE.md`;
- `SubagentStart` делает то же с меньшим лимитом;
- `PreToolUse` для Bash запускает `guard_destructive.py`, блокирующий небольшой deny-list (`git reset --hard`, `git clean -f`, force push, root delete и т. п.).

На момент слепка `~/.codex/hooks.json` и оба hook-скрипта не установлены. Следовательно, это библиотечный шаблон, а не подтверждённо активные hooks текущего проекта.

## 11. Проверенное состояние

Выполнено 2026-08-10:

```text
go version
go test ./cmd/... ./internal/...
npm run build --prefix web
npm run lint --prefix web
```

Результаты:

- Go 1.26.5, оба Go package test suites прошли;
- Node 22.23.1, npm 10.9.8;
- TypeScript + Vite production build прошёл;
- Vite предупредил о главном JS chunk около 611.23 kB minified / 187.84 kB gzip;
- `npm run lint` не запускается: ESLint 9.39.5 не находит `eslint.config.js|mjs|cjs`;
- frontend unit/integration/E2E tests отсутствуют.

Первый `go test ./...` был остановлен по 120-секундному timeout во время первичной компиляции зависимостей; повторный целевой запуск завершился успешно примерно за 42 секунды.

## 12. Известные риски и долги

### Высокий приоритет

1. **Активный пользовательский HTML обслуживается inline на том же origin.** Файл `.html` получает `text/html` и `Content-Disposition: inline`. Открытый после login вредоносный файл способен выполнять same-origin запросы с session cookie. Нужны безопасная download policy для active content, отдельный origin или строгий sandbox/CSP.
2. **Upload не имеет общего server-side size limit.** `ParseMultipartForm(64 << 20)` задаёт memory threshold, но не ограничивает тело запроса; недоверенный LAN peer может расходовать диск/время. Нужен `http.MaxBytesReader`, quotas и timeouts.
3. **Upload пишет сразу в финальное имя.** Во время копирования виден частичный файл; crash оставляет partial file, а повторная попытка получает conflict из-за `O_EXCL`. Нужен temp file + fsync/close + atomic rename и cleanup.
4. **Сервис слушает все интерфейсы.** Локально задано `0.0.0.0:8081`; TLS, pairing, login rate limit и firewall automation отсутствуют. Риск зависит от профиля Windows Firewall и сети.

### Средний приоритет

1. Раздельность `DataDir` и `ShareDir` только документирована, но `app.New` её не валидирует.
2. Лимит editor body доверяет `Content-Length`; после `LimitReader(5 MiB + 1)` число байтов не проверяется. Backend также не ограничивает edit MIME/text type, а UI editor отсутствует.
3. Delete на non-Windows всегда возвращает ошибку, хотя Go server собирается кроссплатформенно.
4. Нет brute-force protection, session cleanup ticker, logout-all/device list или CSRF token. `SameSite=Strict` уменьшает CSRF-риск, но не решает same-origin active-content проблему.
5. Нет pagination/virtualization; большая папка полностью читается, сортируется и рендерится.
6. Все выбранные uploads стартуют параллельно без concurrency limit/backpressure.
7. Ошибки create folder/rename/delete не показываются пользователю.
8. Tray URL вычисляется только на старте и не обновляется после смены сети.
9. Audit log не имеет просмотра, лимита размера или retention.

### Качество и документация

1. Lint script сломан из-за отсутствующего flat config ESLint 9.
2. Нет frontend tests и E2E LAN/browser scenario.
3. Backend tests покрывают auth, базовую path validation, config и tray-independent runtime, но почти не покрывают upload/content/Range/rename/audit/symlink cases.
4. Старый `ARCHITECTURE.mdx` смешивает реализованные и будущие возможности; этот слепок является более точной картой текущего состояния.
5. Production bundle требует code splitting, главным образом вокруг media stack.

## 13. Рекомендуемый порядок развития

1. Закрыть data-integrity/security baseline: temp/atomic uploads, body limits, active-content isolation, enforced root/data separation, login throttling и тесты.
2. Восстановить `npm run lint`, добавить frontend unit tests и минимальный browser E2E.
3. Формализовать pairing/session model и bind/firewall onboarding.
4. Добавить resumable chunks, checksums, retry/backpressure и crash recovery.
5. Добавить WebSocket/file watching для актуального listing.
6. Реализовать discovery/QR UX.
7. Затем thumbnails, PWA/packaging/autostart и performance benchmarks.

Не начинать discovery или декоративные функции до устранения partial upload и same-origin active content: это базовые гарантии сохранности и безопасности.

## 14. Быстрый запуск и debugging

### Production-like локальный запуск

```powershell
Copy-Item .\config.example.json .\config.local.json
notepad .\config.local.json
npm ci --prefix web
npm run build --prefix web
go test ./cmd/... ./internal/...
go build -o wifi-share.exe ./cmd/wifi-share
.\wifi-share.exe
```

### Dev mode

```powershell
# terminal 1
go run ./cmd/wifi-share

# terminal 2
npm install --prefix web
npm run dev --prefix web
```

Vite слушает `0.0.0.0:5173` и proxy `/api` на `localhost:8080`. Если backend перенесён на 8081, dev proxy нужно синхронизировать или backend запускать с `-addr :8080`.

### Типичные симптомы

| Симптом | Сначала проверить |
|---|---|
| `config.local.json is missing` | скопирован ли example и задан ли непустой password |
| UI отвечает 503 | существует ли `web/dist/index.html`, выполнен ли `npm run build` |
| Телефон не открывает URL | один ли Wi-Fi, правильный LAN IPv4, bind address, Windows Firewall/private profile |
| Mutation возвращает 401 | login cookie, рестарт сервера, expiry 24h |
| Upload возвращает 409 | файл/папка с таким именем уже существует или остался partial file |
| Media не перематывается | ответ `/content`, `Range`, MIME и поддержка codec браузером |
| Delete 500 на Linux/macOS | system trash adapter реализован только для Windows |
| `npm run lint` падает до анализа | отсутствует ESLint flat config |

## 15. Готовый prompt для следующей ChatGPT-сессии

```text
Ниже приложен docs/PROJECT_SNAPSHOT.md проекта WiFi Share. Считай разделы
«Фактически реализовано», «Проверенное состояние» и «Известные риски» текущей
правдой, а старые ADR/roadmap — намерениями. Не запрашивай и не публикуй
config.local.json, password, data/*.db или содержимое общей папки.

Задача: <описать конкретный bug/этап>.

Сначала:
1) назови затрагиваемые компоненты и trust boundaries;
2) отдели подтверждённые факты от гипотез;
3) предложи минимальный план и тесты;
4) для сетевых/auth/file-write изменений обязательно проверь data loss,
   traversal/symlink, limits, partial failure, retry и LAN exposure;
5) не считай MCP/hooks/subagents активными только по наличию внешнего preset.
```

## 16. Ограничения слепка

- Это статический снимок, а не полная копия исходников; для точечного patch/debugging ChatGPT всё равно понадобятся затронутые файлы или repository access.
- Значения локального пароля и root path намеренно удалены.
- Runtime SQLite не анализировалась по содержимому; подтверждены только файлы и schema из кода.
- Браузерный E2E и тест на реальном втором Wi-Fi устройстве не выполнялись.
- Глобальная AI/MCP-конфигурация может измениться независимо от репозитория.
