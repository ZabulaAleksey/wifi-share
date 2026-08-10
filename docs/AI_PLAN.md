# AI Plan — WiFi Share

Обновлено: 2026-08-10.

## Текущая цель

Активный implementation stage не выбран. Рекомендуемая следующая цель: закрыть один security/data-integrity stage для uploads и same-origin file serving.

## Предлагаемый scope следующего stage

### В scope

- crash-safe temp upload + atomic publish;
- server-side request/file limits и cleanup;
- policy для HTML/SVG/другого active content;
- DataDir/ShareDir separation validation;
- regression tests.

### Вне scope

- discovery, QR и pairing UI;
- resumable chunk protocol;
- thumbnails и redesign;
- deployment/publishing.

## Acceptance criteria

- partial file никогда не виден под финальным именем;
- failed/crashed upload можно безопасно повторить;
- oversized body отклоняется с предсказуемым status без unbounded disk use;
- shared active content не получает полномочия app origin;
- DataDir внутри ShareDir отклоняется при startup;
- Go tests проходят, новые threat cases покрыты.

## Рекомендуемые специалисты preset

Минимум: `wifi_security_engineer`; `network_protocol_engineer` только если stage включает retry/resume contract. Перед parallel writes заморозить API interface.
