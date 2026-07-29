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

function uploadFile(
  parent: string,
  file: File,
  onProgress: (percent: number) => void,
): Promise<FileItem[]> {
  return new Promise((resolve, reject) => {
    const body = new FormData();
    body.append("files", file);
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `/api/files/${parent}/upload`);
    xhr.responseType = "json";
    xhr.upload.addEventListener("progress", (event) => {
      if (event.lengthComputable) onProgress(Math.round((event.loaded / event.total) * 100));
    });
    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        onProgress(100);
        resolve(xhr.response as FileItem[]);
      } else {
        reject(new Error(xhr.response?.error ?? `HTTP ${xhr.status}`));
      }
    });
    xhr.addEventListener("error", () => reject(new Error("Ошибка сети при загрузке")));
    xhr.send(body);
  });
}

export const api = {
  list(parent = "root") {
    return request<FileListing>(`/api/files?parent=${encodeURIComponent(parent)}`);
  },
  authStatus() {
    return request<{ authenticated: boolean }>("/api/auth/status");
  },
  login(password: string) {
    return request<{ authenticated: boolean }>("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password }),
    });
  },
  logout() {
    return request<void>("/api/auth/session", { method: "DELETE" });
  },
  upload: uploadFile,
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