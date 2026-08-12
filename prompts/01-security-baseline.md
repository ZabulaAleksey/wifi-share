# Этап A — Базовая безопасность

## Требования

`WFS-NET-001..003`, `WFS-UP-001..002`, `WFS-CONTENT-001`, `WFS-AUTH-004..005` из `docs/SPECIFICATION.md`.

## Scope

Реализуй серверные лимиты, crash-safe uploads, root/data validation, изоляцию активного содержимого, усиление входа и соответствующие негативные тесты. Обнови отображение bind/interface и инструкции firewall только настолько, насколько требуется для безопасной экспозиции baseline.

## Non-goals

Не реализуй resumable chunks, pairing, WebSocket, discovery, PWA, MCP и performance experiments.

## Проверки

- `go test ./...`;
- `npm run lint` и `npm run build` в `web/`;
- сценарии этапа A из `docs/TESTING.md`.

## Definition of Done

- partial upload никогда не виден под конечным именем;
- лимиты проверяются до исчерпания ресурсов;
- root/data overlap и active content отклоняются безопасно;
- login/CSRF/Origin policy имеет негативные тесты;
- `docs/AI_STATUS.md` описывает только проверенную реализацию.

Остановись после этапа A и не начинай следующий этап без отдельной задачи.
