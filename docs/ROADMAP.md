# Roadmap — WiFi Share

Статусы отражают код на 2026-08-10, а не только старый план.

## Выполнено

- local Go HTTP server + React SPA;
- directory listing и content/Range;
- password sessions и protected mutations;
- upload progress, folder creation, rename, Windows recycle;
- SQLite audit baseline;
- audio/video player, folder playlist, cross-tab pause;
- Windows tray и dynamic ShareDir;
- responsive dark UI.

## Сейчас: foundation hardening

- atomic/crash-safe uploads;
- size/quota/concurrency/time limits;
- active-content origin isolation;
- root/data validation;
- login throttling и security regression suite;
- исправление ESLint config, frontend unit/E2E baseline.

## Затем: надёжный transfer protocol

- chunked/resumable uploads;
- checksum verification;
- retry, backpressure и cleanup metadata;
- progress/recovery при reconnect и restart;
- performance benchmarks для больших файлов и нескольких peers.

## После protocol: connection UX

- pairing/session model;
- QR connection;
- explicit interface/firewall onboarding;
- device discovery и список подключений;
- WebSocket/file watching для live updates.

## Позднее

- thumbnails WebP/FFmpeg;
- folder transfer, ZIP streaming и restore UX;
- PWA, installer, autostart и platform hardening;
- non-Windows trash adapters;
- code splitting/media lazy load.

## Experimental / optional

- server lease для единственного player между разными devices;
- HTTPS на LAN;
- granular per-share permissions;
- multiple independent shares.
