import type {
  SortDirection,
  BackendItem,
  Project,
  Jenis,
} from "../types/dashboard";

export interface InventoryQueryParams {
  jenis?: string;
  status?: string;
  proyek?: string;
  search?: string;
  sortBy?: string;
  sortOrder?: SortDirection;
  page: number;
  limit: number;
}

export interface InventoryResponse {
  success: boolean;
  data: BackendItem[];
  meta: {
    total: number;
    page: number;
    limit: number;
    totalPages: number;
  };
}

export interface FilterOptionsResponse {
  success: boolean;
  data: {
    project: Project[];
    jenis: Jenis[];
    status: string[];
  };
}

export interface ListOfProjectsResponse {
  success: boolean;
  data: Project[];
}

export interface JenisSuggestionsResponse {
  success: boolean;
  data: Jenis[];
}

export interface AddJenisResponse {
  success: boolean;
  data: Jenis[];
}
export interface DeleteJenisResponse {
  success: boolean;
  data: {
    _id: string;
    jenis: string;
    defaultJenis: string;
    reassignedItems: number;
  };
}

export interface DeleteProjectResponse {
  success: boolean;
  data: {
    _id: string;
  };
}

export interface CreateProjectRequest {
  namaProyek: string;
  lokasi: string;
}

export interface CreateProjectResponse {
  success: boolean;
  data: Project;
}

export interface UpdateProjectRequest {
  namaProyek?: string;
  lokasi?: string;
}

export interface UpdateProjectResponse {
  success: boolean;
  data: Project;
}

// API

async function fetchInventoryApi(
  params: InventoryQueryParams,
  signal?: AbortSignal,
): Promise<InventoryResponse> {
  const qs = new URLSearchParams();
  if (params.jenis) qs.append("jenis", params.jenis);
  if (params.status) qs.append("status", params.status);
  if (params.proyek) qs.append("namaProyek", params.proyek);
  if (params.search) qs.append("q", params.search);
  if (params.sortBy) qs.append("sortBy", params.sortBy);
  if (params.sortOrder) qs.append("sortOrder", params.sortOrder);
  qs.append("page", String(params.page));
  qs.append("limit", String(params.limit));
  const res = await fetch(`/api/v1/items?${qs.toString()}`, { signal });
  if (!res.ok) throw new Error("Gagal mengambil data inventory");
  return (await res.json()) as InventoryResponse;
}

async function fetchFilterOptionsApi(
  signal?: AbortSignal,
): Promise<FilterOptionsResponse> {
  const res = await fetch("/api/v1/items/filter-options", { signal });
  if (!res.ok) throw new Error("Gagal mengambil opsi filter");
  const jenisData = await fetch("/api/v1/item-types", { signal });
  const jenisOptions = (await jenisData.json()).data;
  return {
    success: true,
    data: {
      ...((await res.json()).data as any),
      jenis: jenisOptions,
    },
  } as FilterOptionsResponse;
}

export async function fetchListOfProjects(
  signal?: AbortSignal,
): Promise<ListOfProjectsResponse> {
  const res = await fetch("/api/v1/projects", { signal });
  if (!res.ok) throw new Error("Gagal mengambil opsi proyek");
  return (await res.json()) as ListOfProjectsResponse;
}

export async function createProject(
  payload: CreateProjectRequest,
  signal?: AbortSignal,
): Promise<CreateProjectResponse> {
  const res = await fetch("/api/v1/projects", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    signal,
  });

  if (!res.ok) throw new Error("Gagal membuat project baru");
  return (await res.json()) as CreateProjectResponse;
}

export async function updateProject(
  id: string,
  payload: UpdateProjectRequest,
  signal?: AbortSignal,
): Promise<UpdateProjectResponse> {
  const res = await fetch(`/api/v1/projects/${id}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    signal,
  });

  if (!res.ok) throw new Error("Gagal meng-update project");
  return (await res.json()) as UpdateProjectResponse;
}

export async function deleteProject(
  id: string,
  signal?: AbortSignal,
): Promise<DeleteProjectResponse> {
  const res = await fetch(`/api/v1/projects/${id}`, {
    method: "DELETE",
    signal,
  });

  if (!res.ok) throw new Error("Gagal menghapus project");
  return (await res.json()) as DeleteProjectResponse;
}

async function fetchJenisSuggestionsApi(
  query: string,
  signal?: AbortSignal,
): Promise<JenisSuggestionsResponse> {
  const qs = new URLSearchParams();
  if (query.trim()) qs.append("q", query.trim());

  const res = await fetch(`/api/v1/item-types?${qs.toString()}`, {
    signal,
  });

  if (!res.ok) throw new Error("Gagal mengambil suggestion jenis");
  return (await res.json()) as JenisSuggestionsResponse;
}

export async function addJenisSuggestion(
  jenis: string,
  signal?: AbortSignal,
): Promise<AddJenisResponse> {
  const res = await fetch("/api/v1/item-types", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ jenis }),
    signal,
  });
  if (!res.ok) throw new Error("Gagal menambahkan suggestion jenis");
  return (await res.json()) as AddJenisResponse;
}

export async function deleteJenisSuggestion(
  id: string,
  signal?: AbortSignal,
): Promise<DeleteJenisResponse> {
  const res = await fetch(`/api/v1/item-types/${id}`, {
    method: "DELETE",
    signal,
  });

  if (!res.ok) throw new Error("Gagal menghapus suggestion jenis");
  return (await res.json()) as DeleteJenisResponse;
}

// EXPORT

export async function fetchInventory(
  params: InventoryQueryParams,
  signal?: AbortSignal,
): Promise<InventoryResponse> {
  return fetchInventoryApi(params, signal);
}

export async function fetchFilterOptions(
  signal?: AbortSignal,
): Promise<FilterOptionsResponse> {
  return fetchFilterOptionsApi(signal);
}

export async function fetchJenisSuggestions(
  query: string,
  signal?: AbortSignal,
): Promise<JenisSuggestionsResponse> {
  return fetchJenisSuggestionsApi(query, signal);
}
