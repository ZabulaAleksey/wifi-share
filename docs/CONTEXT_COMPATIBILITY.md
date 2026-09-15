# Совместимость контекста WiFi Share

Read-only brownfield reconciliation выполнена на GitHub `main` `ccb6e64`
и в clean isolated clone до framework change. Code, tests, lockfiles,
runtime config/data и `ARCHITECTURE.mdx` защищены как
`FORBIDDEN_TO_OVERWRITE`; `AGENTS.md` и state-bearing docs требуют `MERGE`.

| Source | Classification | Resolution |
|---|---|---|
| `prompts/STAGES.md` | CONFLICT: нет selector/status/NEXT; Stage 1/2 и встроенный garbled Stage A prompt расходятся с ROADMAP | Сохранён SHA/Git parent; Stage A выбран по ROADMAP/SPEC/кодовым тестам, Stage B deferred. |
| `docs/AI_PLAN.md` | MERGE: dependency migration DONE, governance slice и future smoke | Исторические факты сохранены; текущий verification slice в `docs/STAGES.md`. |
| `docs/AI_STATUS.md` | CONFLICT: Stage A implementation и более ранний список ещё открытых security gaps одновременно | Подтверждённые тестами факты и незакрытые gates перенесены в `docs/STAGES.md`; остальное сохранено в historical note и Git parent. |
| `prompts/README.md` | MERGE: launcher указывает на отсутствующий `01-security-baseline.md` | Исторический запуск сохранён в note, активный route — `docs/STAGES.md`. |

Baseline до изменения: `go test ./...` PASS на изолированном Go 1.25.14;
frontend TypeScript и Vite build PASS через local binaries; ESLint 9
FAIL из-за отсутствия `eslint.config.*`; frozen install остановлен
`ERR_PNPM_IGNORED_BUILDS` для esbuild. Сгенерированное pnpm workspace
изменение от restore удалено до edit; Git branch чиста.

После migration следует повторить Go/frontend baseline и canonical
stage adapter. Formal `.codex/dev-project.toml` отсутствует, поэтому
полное DEV policy inheritance остаётся выключенным; перемещение STAGES
само по себе его не активирует.
