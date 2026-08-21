# Дополнительные архитектурные решения

## ADR-019 — fail-closed LAN bind и изолированные temporary uploads

**Статус:** принято, этап A.

Сервер принимает только `localhost` или конкретный private/loopback IP и запрещает wildcard/динамический порт. Временные загрузки хранятся в зарезервированном `.wifi-share-tmp` с правами `0700`, не видны API и атомарно переименовываются только после `Sync` и закрытия. Это уменьшает случайную сетевую экспозицию и исключает выдачу partial files из share root.

ADR-001–ADR-012, описывающие baseline, остаются в `ARCHITECTURE.mdx`. Здесь зафиксированы новые решения, полученные при интеграции будущих требований.

## ADR-013 — Security-first последовательность

**Статус:** принято.

До resumability, pairing, discovery, realtime и performance-экспериментов закрываются crash-safe uploads, ресурсные лимиты, active-content isolation, root/data separation, усиление входа и security regression suite.

## ADR-014 — LAN не является границей аутентификации

**Статус:** принято.

Обнаружение и частный адрес помогают подключению, но не создают identity или права. Фактическая экспозиция показывается пользователю, а wildcard/public bind получает предупреждение.

## ADR-015 — Canonical state отделён от realtime

**Статус:** принято.

SQLite/файловое состояние остаётся источником истины. WebSocket и watcher доставляют версионированные события поверх snapshot и всегда допускают rescan/recovery.

## ADR-016 — Единый прикладной core

**Статус:** принято как целевое направление.

REST, UI и будущие CLI/MCP вызывают общие Application Services. Нельзя копировать файловые, transfer- и auth-правила в каждый adapter.

## ADR-017 — Gate экспериментальной технологии

**Статус:** принято.

Перед eBPF, `io_uring`, zero-copy, QUIC/WebTransport, E2EE, WebRTC или Rust-модулем фиксируются проблема, цель, альтернативы, feature flag/isolation, fallback, тесты, security review и end-to-end benchmark. DPDK/RDMA/AF_XDP остаются research-only.

## ADR-018 — LAN compute вне file-transfer core

**Статус:** принято.

Распределённые вычисления могут появиться только отдельным module/service adapter над явным trust/consent layer; общий оркестратор предпочтительно живёт в отдельном проекте.
