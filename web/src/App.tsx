import { useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Archive,
  AudioLines,
  ChevronRight,
  Code2,
  File,
  FileImage,
  FileText,
  Film,
  Folder,
  FolderPlus,
  HardDrive,
  MoreHorizontal,
  Pencil,
  Search,
  Trash2,
  Upload,
  Wifi,
} from "lucide-react";
import { api, type FileItem } from "./api";
import { MediaViewer } from "./MediaViewer";

type Crumb = { id: string; name: string };

export function App() {
  const [crumbs, setCrumbs] = useState<Crumb[]>([{ id: "root", name: "Общие файлы" }]);
  const [search, setSearch] = useState("");
  const [activeMedia, setActiveMedia] = useState<FileItem | null>(null);
  const fileInput = useRef<HTMLInputElement>(null);
  const queryClient = useQueryClient();
  const current = crumbs.at(-1)!;

  const listing = useQuery({
    queryKey: ["files", current.id],
    queryFn: () => api.list(current.id),
  });

  const refresh = () => queryClient.invalidateQueries({ queryKey: ["files", current.id] });
  const upload = useMutation({
    mutationFn: (files: FileList) => api.upload(current.id, files),
    onSuccess: refresh,
  });
  const createFolder = useMutation({
    mutationFn: (name: string) => api.createFolder(current.id, name),
    onSuccess: refresh,
  });
  const rename = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => api.rename(id, name),
    onSuccess: refresh,
  });
  const remove = useMutation({
    mutationFn: api.remove,
    onSuccess: refresh,
  });

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    return (listing.data?.files ?? []).filter((item) =>
      item.name.toLowerCase().includes(query),
    );
  }, [listing.data, search]);

  function open(item: FileItem) {
    if (item.kind === "folder") {
      setCrumbs((value) => [...value, { id: item.id, name: item.name }]);
      return;
    }
    if (item.mime.startsWith("audio/") || item.mime.startsWith("video/")) {
      setActiveMedia(item);
      return;
    }
    window.open(api.content(item.id), "_blank", "noopener,noreferrer");
  }

  function askCreateFolder() {
    const name = window.prompt("Название новой папки");
    if (name?.trim()) createFolder.mutate(name.trim());
  }

  function askRename(item: FileItem) {
    const name = window.prompt("Новое название", item.name);
    if (name?.trim() && name.trim() !== item.name) {
      rename.mutate({ id: item.id, name: name.trim() });
    }
  }

  function askRemove(item: FileItem) {
    if (window.confirm(`Переместить «${item.name}» в корзину?`)) {
      remove.mutate(item.id);
    }
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand__mark"><Wifi size={20} /></span>
          <div><strong>WiFi Share</strong><span>локальное пространство</span></div>
        </div>
        <nav>
          <button className="nav-item nav-item--active"><HardDrive size={19} /> Мои файлы</button>
          <button className="nav-item"><AudioLines size={19} /> Аудио</button>
          <button className="nav-item"><Film size={19} /> Видео</button>
          <button className="nav-item"><Trash2 size={19} /> Корзина</button>
        </nav>
        <div className="connection-card">
          <span className="connection-card__pulse" />
          <div><strong>Сервер доступен</strong><span>{window.location.host}</span></div>
        </div>
      </aside>

      <main>
        <header className="topbar">
          <div>
            <span className="eyebrow">Локальное хранилище</span>
            <h1>Ваши файлы</h1>
          </div>
          <div className="topbar__actions">
            <label className="search">
              <Search size={18} />
              <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Найти файл" />
            </label>
            <input
              ref={fileInput}
              hidden
              type="file"
              multiple
              onChange={(event) => event.target.files && upload.mutate(event.target.files)}
            />
            <button className="button button--secondary" onClick={askCreateFolder}>
              <FolderPlus size={18} /> <span>Новая папка</span>
            </button>
            <button className="button button--primary" onClick={() => fileInput.current?.click()}>
              <Upload size={18} /> <span>Загрузить</span>
            </button>
          </div>
        </header>

        <div className="breadcrumbs">
          {crumbs.map((crumb, index) => (
            <span key={crumb.id}>
              {index > 0 && <ChevronRight size={15} />}
              <button onClick={() => setCrumbs((value) => value.slice(0, index + 1))}>
                {crumb.name}
              </button>
            </span>
          ))}
        </div>

        <section className="file-panel">
          <div className="file-panel__heading">
            <span>Название</span><span>Изменён</span><span>Размер</span><span />
          </div>
          {listing.isLoading && <div className="empty-state">Загружаем файлы…</div>}
          {listing.error && <div className="empty-state empty-state--error">{listing.error.message}</div>}
          {!listing.isLoading && !listing.error && filtered.length === 0 && (
            <div className="empty-state">
              <Folder size={42} />
              <strong>{search ? "Ничего не найдено" : "Папка пока пуста"}</strong>
              <span>{search ? "Попробуйте другой запрос" : "Загрузите файлы с компьютера или телефона"}</span>
            </div>
          )}
          {filtered.map((item) => (
            <FileRow
              key={item.id}
              item={item}
              onOpen={() => open(item)}
              onRename={() => askRename(item)}
              onRemove={() => askRemove(item)}
            />
          ))}
        </section>
      </main>

      {activeMedia && (
        <MediaViewer
          file={activeMedia}
          source={api.content(activeMedia.id)}
          onClose={() => setActiveMedia(null)}
        />
      )}
    </div>
  );
}

function FileRow({
  item,
  onOpen,
  onRename,
  onRemove,
}: {
  item: FileItem;
  onOpen: () => void;
  onRename: () => void;
  onRemove: () => void;
}) {
  return (
    <article className="file-row">
      <button className="file-row__main" onClick={onOpen}>
        <FileIcon item={item} />
        <span><strong>{item.name}</strong><small>{labelFor(item)}</small></span>
      </button>
      <time>{formatDate(item.modified)}</time>
      <span className="file-size">{item.kind === "folder" ? "—" : formatBytes(item.size)}</span>
      <details className="menu">
        <summary aria-label="Действия"><MoreHorizontal size={20} /></summary>
        <div>
          <button onClick={onRename}><Pencil size={15} /> Переименовать</button>
          <button className="danger" onClick={onRemove}><Trash2 size={15} /> В корзину</button>
        </div>
      </details>
    </article>
  );
}

function FileIcon({ item }: { item: FileItem }) {
  if (item.kind === "folder") return <span className="file-icon folder"><Folder /></span>;
  if (item.mime.startsWith("image/")) return <span className="file-icon image"><FileImage /></span>;
  if (item.mime.startsWith("audio/")) return <span className="file-icon audio"><AudioLines /></span>;
  if (item.mime.startsWith("video/")) return <span className="file-icon video"><Film /></span>;
  if (item.mime.includes("zip") || item.mime.includes("archive")) return <span className="file-icon archive"><Archive /></span>;
  if (/\.(ts|tsx|js|jsx|go|py|rs|html|css)$/i.test(item.name)) return <span className="file-icon code"><Code2 /></span>;
  if (item.mime.startsWith("text/") || item.mime.includes("pdf")) return <span className="file-icon text"><FileText /></span>;
  return <span className="file-icon"><File /></span>;
}

function labelFor(item: FileItem) {
  if (item.kind === "folder") return "Папка";
  return item.name.split(".").at(-1)?.toUpperCase() || "Файл";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("ru", { day: "2-digit", month: "short", year: "numeric" }).format(new Date(value));
}

function formatBytes(value: number) {
  if (!value) return "0 Б";
  const units = ["Б", "КБ", "МБ", "ГБ", "ТБ"];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`;
}

