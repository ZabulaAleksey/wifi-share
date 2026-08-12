# Стратегия тестирования WiFi Share

## Базовые команды

```powershell
go test ./...
cd web
npm run lint
npm run build
```

## Обязательная security regression suite

- закодированные и декодированные варианты traversal, внешняя символическая ссылка и пересечение root/data;
- огромный request body, отдельный файл и множество частей multipart;
- заполненный диск, прерванная запись, stale temp file и отсутствие partial file под конечным именем;
- коллизия имён, странный Unicode и запрещённые path-сегменты;
- конкурирующие upload/rename/delete и соблюдение quota/concurrency/timeouts/cancellation;
- malicious HTML и SVG, MIME confusion, CSP, `nosniff` и `Content-Disposition`;
- истечение/отзыв сессии, brute force/login flooding, CSRF и Origin;
- malformed и overlapping Range;
- повреждённые изображения/аудио/видео и bounded FFmpeg;
- ZIP traversal/bomb, отмена и ограниченное потребление временного диска;
- просроченная resumable session, повреждённый chunk, checksum mismatch, replay финализации;
- истёкший или повторно использованный pairing token/QR и отзыв устройства.

## Надёжность и realtime

- retry/backpressure/reconnect после обрыва сети, сна, смены IP/VPN и рестарта сервера;
- snapshot + events без потери canonical state;
- heartbeat, bounded WebSocket buffer и совместимость версий envelope;
- потерянные/сдвоенные watcher events с rescan fallback;
- недоступность thumbnail не блокирует оригинал;
- потоковый ZIP и большие файлы отменяются без утечки ресурсов.

## Gate эксперимента

Любой performance/transport experiment проходит те же функциональные и security-тесты, feature flag, capability detection, fallback и end-to-end benchmark. Microbenchmark без пользовательского выигрыша не является основанием для принятия.

## Definition of Done этапа

- реализованы только требования текущей фазы;
- положительные, негативные и граничные тесты проходят;
- ограничения ресурсов и режим отказа проверены;
- документация не называет future-возможность реализованной;
- `docs/AI_STATUS.md` обновлён фактами и доказательствами.
