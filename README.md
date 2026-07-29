# WiFi Share

Локальная передача и просмотр файлов между устройствами в одной Wi‑Fi сети.

## Требования

- Go 1.24+
- Node.js 20.19+ или 22.12+
- npm

## Запуск для разработки

```powershell
New-Item -ItemType Directory -Force shared
go run ./cmd/wifi-share
```

Во втором терминале:

```powershell
Set-Location web
npm install
npm run dev
```

Откройте `http://localhost:5173`. Для другого устройства используйте LAN-адрес компьютера, например `http://192.168.1.15:5173`.

## Production-сборка

```powershell
Set-Location web
npm ci
npm run build
Set-Location ..
go build -o wifi-share.exe ./cmd/wifi-share
.\wifi-share.exe -root .\shared
```

После сборки интерфейс и API доступны на `http://localhost:8080`.

Архитектурные решения находятся в [ARCHITECTURE.mdx](./ARCHITECTURE.mdx).
