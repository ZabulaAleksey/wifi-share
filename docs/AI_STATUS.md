# AI Project Status — WiFi Share

Обновлено: 2026-08-10.

## Текущий этап

Рабочий local MVP / alpha: базовый file browsing, password-gated mutations, media playback и Windows tray реализованы. Следующий разумный этап — security/data-integrity hardening до расширения функций.

## Уже реализовано

- Go REST/static server и React/Vite UI;
- safe relative IDs и root containment checks;
- read-only browsing/content и authenticated mutations;
- 24-hour in-memory sessions;
- upload progress, folders, rename, Windows recycle;
- SQLite audit;
- Vidstack playlist и cross-tab coordination;
- live ShareDir switch через Windows tray.

## В работе

Активная coding-задача не зафиксирована. Ветка `docs/project-snapshot` документирует состояние без изменения runtime-кода.

## Известные проблемы

- active HTML/content открывается inline на origin приложения;
- upload без total size cap, temp/atomic finalize и crash-safe retry;
- bind на всех interfaces без TLS/pairing/rate limit;
- root/data separation не enforced кодом;
- ESLint 9 config отсутствует, поэтому `npm run lint` падает;
- frontend tests и E2E отсутствуют;
- основной frontend chunk около 611 kB minified;
- non-Windows trash не реализован;
- mutation errors create/rename/delete не видны в UI.

Полный список рисков: `docs/PROJECT_SNAPSHOT.md`, раздел 12.

## Последние проверки

```text
go test ./cmd/... ./internal/...       PASS
npm run build --prefix web             PASS with chunk-size warning
npm run lint --prefix web              FAIL: eslint.config.* missing
```

Toolchain: Go 1.26.5; Node 22.23.1; npm 10.9.8.

## Следующая рекомендуемая задача

Один bounded security stage:

1. temp + atomic upload finalize;
2. `MaxBytesReader`/quota/concurrency limits;
3. безопасная policy для active content;
4. enforced DataDir outside ShareDir;
5. backend regression/security tests.

После этого восстановить frontend lint/tests и только затем проектировать pairing/resumable protocol.

## Контекст для следующей AI-сессии

- Начать с `docs/PROJECT_SNAPSHOT.md`, затем открыть только затронутые исходники.
- Не читать/публиковать `config.local.json`, `data/*.db` или ShareDir content.
- Не считать roadmap-функцию реализованной без подтверждения в source/tests.
- LAN peer, filename и payload считать недоверенными.
- External preset agents/hooks существуют отдельно от repo и могут быть не установлены.
