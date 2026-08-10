# Design — WiFi Share

Каноническое описание фактически реализованного UI на 2026-08-10. Источник реализации: `web/src/App.tsx`, `MediaViewer.tsx`, `styles.css`.

## Принципы

- Локальный файловый менеджер должен быстро сообщать: какой root открыт, доступен ли сервер и есть ли full access.
- Read-only режим остаётся полноценным для browsing/media, а dangerous actions скрыты до login.
- Desktop показывает табличную плотность; mobile сохраняет главное имя/иконку/action и убирает вторичные метаданные.

## Визуальные tokens

| Token | Значение |
|---|---|
| Page background | `#07100e` |
| Surface | `#0c1714` |
| Elevated surface | `#111f1b` |
| Text | `#e9f3ef` |
| Muted | `#84958f` |
| Accent | `#80e5bc` |
| Accent dark | `#153c30` |
| Font | Inter fallback на system UI sans |

Скругления обычно 10–20 px, границы полупрозрачные, accent применяется для primary action, connection status и media brand.

## Layout и компоненты

- Desktop: fixed sidebar 245 px, topbar, breadcrumbs, file panel.
- Sidebar: brand, единственный nav item и connection card.
- Topbar: search, login или upload/create/logout actions.
- File row: icon, name/type label, modified, size и contextual menu.
- Overlays: login modal, upload progress panel и fixed media dock.
- Иконки Lucide; file categories имеют разные цветовые пары.

## Состояния

- loading: «Загружаем файлы…»;
- empty folder и empty search;
- listing error;
- read-only notice;
- login pending/error;
- upload uploading/done/error с процентом;
- audio/video playing и close;
- mutation errors create/rename/delete пока визуально не представлены — это известный UX debt.

## Responsive

- `<=900px`: sidebar превращается в верхнюю строку, actions переносятся, search занимает строку, media dock растягивается между краями.
- `<=650px`: topbar вертикальный, подписи action buttons скрыты, table header/date/size убраны, row становится двухколоночным.
- Минимальная ширина body: 320 px.

## Не реализовано

- light theme;
- drag-and-drop;
- thumbnails/gallery;
- inline text editor;
- toast/error system для всех mutations;
- accessibility audit и browser E2E baseline.
