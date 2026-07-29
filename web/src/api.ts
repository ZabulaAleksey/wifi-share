export type FileItem = {
  id: string;
  parentId?: string;
  name: string;
  kind: "file" | "folder";
  mime: string;
  size: number;
  modified: string;
};

export type FileListing = {
  parent: { id: string; name: string };
  files: FileItem[];
};

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string } | null;
    throw new Error(body?.error ?? `HTTP ${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export const api = {
  list(parent = "root") {
    return request<FileListing>(`/api/files?parent=${encodeURIComponent(parent)}`);
  },
  upload(parent: string, files: globalThis.FileList) {
    const body = new FormData();
    Array.from(files).forEach((file) => body.append("files", file));
    return request<FileItem[]>(`/api/files/${parent}/upload`, { method: "POST", body });
  },
  createFolder(parent: string, name: string) {
    return request<FileItem>(`/api/files/${parent}/folders`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
  },
  rename(id: string, name: string) {
    return request<{ id: string; name: string }>(`/api/files/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
  },
  remove(id: string) {
    return request<void>(`/api/files/${id}`, { method: "DELETE" });
  },
  content(id: string) {
    return `/api/files/${id}/content`;
  },
};

