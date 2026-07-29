# WiFi Share

Локальная передача и просмотр файлов между устройствами в одной Wi‑Fi сети.

## Требования

- Go 1.24+
- Node.js 20.19+ или 22.12+
- npm

## Где указать папку с файлами

Путь к доступным по Wi‑Fi файлам передаётся серверу через параметр `-root`.
Если параметр не указан, используется папка `shared` в каталоге проекта.

Например, чтобы открыть доступ к папке `D:\Media`:

```powershell
.\wifi-share.exe -root "D:\Media"
```

Если в пути есть пробелы, обязательно заключите его в кавычки:

```powershell
.\wifi-share.exe -root "D:\Общие файлы\Фото и видео"
```

Во время разработки используется тот же параметр:

```powershell
go run ./cmd/wifi-share -root "D:\Media"
```

Можно одновременно изменить порт и каталоги служебных данных:

```powershell
.\wifi-share.exe `
  -addr ":8080" `
  -root "D:\Media" `
  -data "C:\WiFiShare\data" `
  -web ".\web\dist"
```

| Параметр | По умолчанию | Назначение |
|---|---|---|
| `-root` | `.\shared` | Папка, файлы которой видны другим устройствам |
| `-data` | `.\data` | SQLite, корзина, кэш и служебные данные |
| `-web` | `.\web\dist` | Собранный React-интерфейс |
| `-addr` | `:8080` | IP-адрес и порт HTTP-сервера |

> Указывайте в `-root` только ту папку, содержимое которой можно открыть другим
> устройствам в локальной сети. Остальные папки компьютера приложению недоступны.

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
# wifi-share
