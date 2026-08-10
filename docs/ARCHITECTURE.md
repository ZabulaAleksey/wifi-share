# Architecture — WiFi Share

Статус: current-state baseline на 2026-08-10. Подробная переносимая карта находится в [PROJECT_SNAPSHOT.md](./PROJECT_SNAPSHOT.md). Исторические ADR сохранены в корневом `ARCHITECTURE.mdx`; его будущие функции не следует считать реализованными.

## Контекст

WiFi Share публикует одну локальную директорию через HTTP для устройств в той же сети. Незалогиненные клиенты могут читать список и содержимое; пароль открывает файловые мутации.

```text
Browser UI
   |
   | HTTP REST + static assets
   v
Go net/http App
   |-- auth/session map
   |-- safe relative-ID filesystem adapter --> ShareDir
   |-- audit -------------------------------> SQLite WAL
   |-- delete ------------------------------> Windows Recycle Bin
   `-- SPA ---------------------------------> web/dist

Windows tray --> App.SetShareDir / config.local.json / shutdown
Browser tabs  --> BroadcastChannel media coordination
```

## Компоненты

- `cmd/wifi-share`: загрузка локального config, CLI overrides, LAN URL discovery и platform runtime.
- `internal/app`: HTTP routes, auth middleware, path validation, file operations, SQLite audit и static SPA.
- `web/src`: React UI, REST client, query cache, upload progress и Vidstack player.
- `config.local.json`: plaintext local password и runtime paths; ignored.
- `data/wifi-share.db`: audit log; ignored и должен находиться вне ShareDir.

## Интерфейсы

- REST `/api/auth/*` и `/api/files*`; контракт подробно перечислен в `PROJECT_SNAPSHOT.md`.
- File ID — `root` или Base64URL от относительного пути.
- `App.SetShareDir(path)` атомарно меняет значение root под `RWMutex`; уже начавшаяся операция продолжает работать со старым snapshot пути.
- Frontend получает server state через TanStack Query и invalidates текущий listing после мутаций.

## Инварианты

- Абсолютный путь клиента не используется напрямую.
- Итоговый путь должен оставаться внутри ShareDir; существующие symlink также проверяются.
- Мутации требуют in-memory session cookie.
- Root нельзя переименовать или удалить.
- Upload не перезаписывает существующий файл.
- Секреты, база и build/runtime artifacts не коммитятся.

## Текущие ограничения

- HTTP без TLS, pairing и rate limiting; bind может охватывать все interfaces.
- Upload не resumable и пока пишет в финальное имя без server-side total body cap.
- Active content обслуживается inline на origin приложения.
- SQLite используется только для audit.
- WebSocket/discovery/thumbnails/checksums/PWA отсутствуют.
- System trash реализован только на Windows.
