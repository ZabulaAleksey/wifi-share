# Technical Decisions — WiFi Share

Исторические развёрнутые ADR находятся в корневом `ARCHITECTURE.mdx`. Этот журнал фиксирует их фактический статус по отношению к текущему коду; новые решения добавляются отдельными записями и не переписывают историю молча.

| ID | Решение | Причина | Текущий статус / последствия |
|---|---|---|---|
| ADR-001 | Go + `net/http` | один локальный executable, streaming и Range | реализовано |
| ADR-002 | React + TypeScript + Vite | сложный интерактивный файловый UI | реализовано |
| ADR-003 | REST + opaque relative IDs | не принимать абсолютные пути клиента | реализовано; resumability отсутствует |
| ADR-004 | SQLite metadata | простой локальный WAL store | частично: только audit operations |
| ADR-005 | Vidstack | единая audio/video platform | реализовано |
| ADR-006 | `BroadcastChannel` для одного player | работает на LAN HTTP без secure-context requirement | реализовано только между вкладками одного browser profile |
| ADR-007 | MIME icons + thumbnail cache | быстрый media browsing | icons реализованы, thumbnails/FFmpeg нет |
| ADR-008 | local-first security | ограниченный root и отсутствие cloud telemetry | частично; pairing/limits/active-content isolation не готовы |
| ADR-009 | plaintext local password + in-memory sessions | простой full-access gate, logout при restart | реализовано |
| ADR-010 | отдельный XHR на файл | наблюдаемый upload progress | реализовано; concurrency/backpressure не ограничены |
| ADR-011 | Windows systray и динамический root | desktop UX без restart | реализовано |
| ADR-012 | playlist текущей папки | последовательное воспроизведение media | реализовано |

## 2026-08-10 — Разделить текущую реализацию и целевую архитектуру

**Контекст:** корневой `ARCHITECTURE.mdx` смешивал работающие функции с будущими этапами, из-за чего AI-сессия могла принять thumbnails, pairing, WebSocket или resumable transfer за готовый код.

**Решение:** `docs/ARCHITECTURE.md`, `docs/AI_STATUS.md` и `docs/PROJECT_SNAPSHOT.md` описывают current state; `docs/ROADMAP.md` хранит планы.

**Альтернатива:** переписать исторический MDX. Не выбрана, чтобы сохранить исходные ADR и не создавать большой несвязанный diff.

**Последствие:** при конфликте по статусу функции current-state docs имеют приоритет, пока код или документация не обновлены вместе.
