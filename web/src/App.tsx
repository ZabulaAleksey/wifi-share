import { type FormEvent, useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Archive, AudioLines, ChevronRight, Code2, File, FileImage, FileText,
  Film, Folder, FolderPlus, HardDrive, LockKeyhole, LogIn, LogOut,
  MoreHorizontal, Pencil, Search, Trash2, Upload, Wifi, X,
} from "lucide-react";
import { api, type FileItem } from "./api";
import { MediaViewer } from "./MediaViewer";

type Crumb = { id: string; name: string };
type UploadItem = {
  id: string;
  name: string;
  percent: number;
  status: "uploading" | "done" | "error";
  error?: string;
};

const rootCrumb: Crumb = { id: "root", name: "Общие файлы" };

export function App() {
  const [crumbs, setCrumbs] = useState<Crumb[]>([rootCrumb]);
  const [search, setSearch] = useState("");
  const [activeMedia, setActiveMedia] = useState<FileItem | null>(null);
  const [loginOpen, setLoginOpen] = useState(false);
  const [uploads, setUploads] = useState<UploadItem[]>([]);
  const fileInput = useRef<HTMLInputElement>(null);
  const queryClient = useQueryClient();
  const current = crumbs[crumbs.length - 1];

  const listing = useQuery({
    queryKey: ["files", current.id],
    queryFn: () => api.list(current.id),
  });
  const auth = useQuery({ queryKey: ["auth"], queryFn: api.authStatus });
  const authenticated = auth.data?.authenticated ?? false;
  const refresh = () => queryClient.invalidateQueries({ queryKey: ["files", current.id] });

  const createFolder = useMutation({
    mutationFn: (name: string) => api.createFolder(current.id, name),
    onSuccess: refresh,
  });
  const rename = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => api.rename(id, name),
    onSuccess: refresh,
  });
  const remove = useMutation({ mutationFn: api.remove, onSuccess: refresh });
  const logout = useMutation({
    mutationFn: api.logout,
    onSuccess: () => queryClient.setQueryData(["auth"], { authenticated: false }),
  });

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    return (listing.data?.files ?? []).filter((item) =>
      item.name.toLowerCase().includes(query),
    );
  }, [listing.data, search]);

  function goRoot() {
    setCrumbs([rootCrumb]);
    setSearch("");
  }

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

  async function uploadFiles(fileList: FileList) {
    const files = Array.from(fileList);
    const pending = files.map((file, index) => ({
      id: globalThis.crypto?.randomUUID?.()
        ?? `upload-${Date.now()}-${index}-${Math.random().toString(36).slice(2)}`,
      name: file.name, percent: 0,
      status: "uploading" as const,
    }));
    setUploads((currentItems) => [...currentItems, ...pending]);
    await Promise.all(files.map(async (file, index) => {
      const id = pending[index].id;
      try {
        await api.upload(current.id, file, (percent) => {
          setUploads((items) => items.map((item) => item.id === id ? { ...item, percent } : item));
        });
        setUploads((items) => items.map((item) => item.id === id ? { ...item, percent: 100, status: "done" } : item));
      } catch (error) {
        setUploads((items) => items.map((item) => item.id === id ? {
          ...item, status: "error", error: error instanceof Error ? error.message : "Ошибка загрузки",
        } : item));
      }
    }));
    await refresh();
  }

  function askCreateFolder() {
    const name = window.prompt("Название новой папки");
    if (name?.trim()) createFolder.mutate(name.trim());
  }

  function askRename(item: FileItem) {
    const name = window.prompt("Новое название", item.name);
    if (name?.trim() && name.trim() !== item.name) rename.mutate({ id: item.id, name: name.trim() });
  }

  function askRemove(item: FileItem) {
    if (window.confirm(`Переместить «${item.name}» в корзину компьютера?`)) remove.mutate(item.id);
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <button className="brand brand--button" onClick={goRoot} aria-label="Вернуться в корень">
          <span className="brand__mark"><Wifi size={20} /></span>
          <div><strong>WiFi Share</strong><span>локальное пространство</span></div>
        </button>
        <nav>
          <button className="nav-item nav-item--active" onClick={goRoot}>
            <HardDrive size={19} /> Мои файлы
          </button>
        </nav>
        <div className="connection-card">
          <span className="connection-card__pulse" />
          <div><strong>Сервер доступен</strong><span>{window.location.host}</span></div>
        </div>
      </aside>

      <main>
        <header className="topbar">
          <div><span className="eyebrow">Локальное хранилище</span><h1>Ваши файлы</h1></div>
          <div className="topbar__actions">
            <label className="search"><Search size={18} /><input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Найти файл" /></label>
            {authenticated ? (
              <>
                <input ref={fileInput} hidden type="file" multiple onChange={(event) => {
                  if (event.target.files) void uploadFiles(event.target.files);
                  event.target.value = "";
                }} />
                <button className="button button--secondary" onClick={askCreateFolder}><FolderPlus size={18} /><span>Новая папка</span></button>
                <button className="button button--primary" onClick={() => fileInput.current?.click()}><Upload size={18} /><span>Загрузить</span></button>
                <button className="icon-button topbar__logout" onClick={() => logout.mutate()} aria-label="Выйти"><LogOut size={18} /></button>
              </>
            ) : (
              <button className="button button--primary" onClick={() => setLoginOpen(true)}><LogIn size={18} /><span>Полный доступ</span></button>
            )}
          </div>
        </header>

        {!authenticated && (
          <div className="readonly-notice"><LockKeyhole size={16} /> Режим просмотра. Войдите, чтобы загружать, переименовывать и удалять файлы.</div>
        )}

        <div className="breadcrumbs">
          {crumbs.map((crumb, index) => (
            <span key={crumb.id}>{index > 0 && <ChevronRight size={15} />}<button onClick={() => setCrumbs((value) => value.slice(0, index + 1))}>{crumb.name}</button></span>
          ))}
        </div>

        <section className="file-panel">
          <div className="file-panel__heading"><span>Название</span><span>Изменён</span><span>Размер</span><span /></div>
          {listing.isLoading && <div className="empty-state">Загружаем файлы…</div>}
          {listing.error && <div className="empty-state empty-state--error">{listing.error.message}</div>}
          {!listing.isLoading && !listing.error && filtered.length === 0 && (
            <div className="empty-state"><Folder size={42} /><strong>{search ? "Ничего не найдено" : "Папка пока пуста"}</strong><span>{search ? "Попробуйте другой запрос" : "Здесь появятся доступные файлы"}</span></div>
          )}
          {filtered.map((item) => (
            <FileRow key={item.id} item={item} canEdit={authenticated} onOpen={() => open(item)} onRename={() => askRename(item)} onRemove={() => askRemove(item)} />
          ))}
        </section>
      </main>

      {activeMedia && <MediaViewer file={activeMedia} source={api.content(activeMedia.id)} onClose={() => setActiveMedia(null)} />}
      {loginOpen && <LoginDialog onClose={() => setLoginOpen(false)} onSuccess={() => {
        queryClient.setQueryData(["auth"], { authenticated: true });
        setLoginOpen(false);
      }} />}
      {uploads.length > 0 && <UploadProgress items={uploads} onClose={() => setUploads([])} />}
    </div>
  );
}

function LoginDialog({ onClose, onSuccess }: { onClose: () => void; onSuccess: () => void }) {
  const [password, setPassword] = useState("");
  const login = useMutation({ mutationFn: api.login, onSuccess });
  function submit(event: FormEvent) { event.preventDefault(); login.mutate(password); }
  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <form className="login-card" onSubmit={submit}>
        <button type="button" className="icon-button login-card__close" onClick={onClose}><X size={18} /></button>
        <span className="login-card__icon"><LockKeyhole /></span><span className="eyebrow">Полный доступ</span><h2>Введите пароль</h2>
        <p>После входа можно загружать, переименовывать и удалять файлы.</p>
        <input autoFocus type="password" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="Пароль" />
        {login.error && <span className="form-error">Неверный пароль</span>}
        <button className="button button--primary" disabled={!password || login.isPending}>{login.isPending ? "Проверяем…" : "Войти"}</button>
      </form>
    </div>
  );
}

function UploadProgress({ items, onClose }: { items: UploadItem[]; onClose: () => void }) {
  const active = items.some((item) => item.status === "uploading");
  return (
    <section className="upload-panel">
      <header><div><strong>{active ? "Загрузка файлов" : "Загрузка завершена"}</strong><span>{items.length} файл(а)</span></div><button className="icon-button" onClick={onClose} disabled={active}><X size={18} /></button></header>
      <div className="upload-panel__list">{items.map((item) => (
        <div className="upload-item" key={item.id}><div><span title={item.name}>{item.name}</span><strong>{item.status === "error" ? "Ошибка" : `${item.percent}%`}</strong></div><div className={`progress ${item.status === "error" ? "progress--error" : ""}`}><i style={{ width: `${item.percent}%` }} /></div>{item.error && <small>{item.error}</small>}</div>
      ))}</div>
    </section>
  );
}

function FileRow({ item, canEdit, onOpen, onRename, onRemove }: { item: FileItem; canEdit: boolean; onOpen: () => void; onRename: () => void; onRemove: () => void }) {
  return <article className="file-row"><button className="file-row__main" onClick={onOpen}><FileIcon item={item} /><span><strong>{item.name}</strong><small>{labelFor(item)}</small></span></button><time>{formatDate(item.modified)}</time><span className="file-size">{item.kind === "folder" ? "—" : formatBytes(item.size)}</span>{canEdit ? <details className="menu"><summary aria-label="Действия"><MoreHorizontal size={20} /></summary><div><button onClick={onRename}><Pencil size={15} /> Переименовать</button><button className="danger" onClick={onRemove}><Trash2 size={15} /> В корзину</button></div></details> : <span />}</article>;
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
  const parts = item.name.split(".");
  return item.kind === "folder" ? "Папка" : parts[parts.length - 1]?.toUpperCase() || "Файл";
}
function formatDate(value: string) { return new Intl.DateTimeFormat("ru", { day: "2-digit", month: "short", year: "numeric" }).format(new Date(value)); }
function formatBytes(value: number) { if (!value) return "0 Б"; const units = ["Б", "КБ", "МБ", "ГБ", "ТБ"]; const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1); return `${(value / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`; }