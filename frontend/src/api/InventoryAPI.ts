import type {
  AddJenisResponse,
  CreateProjectRequest,
  CreateProjectResponse,
  DeleteJenisResponse,
  DeleteProjectResponse,
  FilterOptionsResponse,
  InventoryQueryParams,
  InventoryResponse,
  JenisSuggestionsResponse,
  ListOfProjectsResponse,
  ProjectCoordinateInput,
  UpdateProjectRequest,
  UpdateProjectResponse,
} from "../types/api";

const toApiUrl = (path: string): string => {
  if (typeof window !== "undefined" && window.location?.origin) {
    return new URL(path, window.location.origin).toString();
  }

  return new URL(path, "http://localhost").toString();
};

export function normalizeProjectCoordinateInput(
  latInput: string,
  lngInput: string,
): {
  isValid: boolean;
  value: ProjectCoordinateInput | null;
  errors: Record<string, string>;
} {
  const errors: Record<string, string> = {};
  const hasLat = latInput.trim() !== "";
  const hasLng = lngInput.trim() !== "";

  if (!hasLat && !hasLng) {
    return { isValid: true, value: null, errors: {} };
  }

  if (!hasLat || !hasLng) {
    errors["koordinat.lat"] = "Latitude dan longitude harus diisi bersamaan.";
    errors["koordinat.lng"] = "Latitude dan longitude harus diisi bersamaan.";
    return { isValid: false, value: null, errors };
  }

  const lat = Number(latInput);
  const lng = Number(lngInput);

  if (!Number.isFinite(lat) || lat < -90 || lat > 90) {
    errors["koordinat.lat"] = "Latitude harus berupa angka antara -90 dan 90.";
  }

  if (!Number.isFinite(lng) || lng < -180 || lng > 180) {
    errors["koordinat.lng"] =
      "Longitude harus berupa angka antara -180 dan 180.";
  }

  if (Object.keys(errors).length > 0) {
    return { isValid: false, value: null, errors };
  }

  return {
    isValid: true,
    value: { lat, lng },
    errors: {},
  };
}

export function buildProjectRequest(input: {
  name: string;
  location: string;
  lat?: string;
  lng?: string;
}): CreateProjectRequest {
  const normalized = normalizeProjectCoordinateInput(
    input.lat ?? "",
    input.lng ?? "",
  );

  return {
    namaProyek: input.name.trim(),
    lokasi: input.location.trim(),
    koordinat: normalized.value,
  };
}

// API

export async function fetchInventory(
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
  const res = await fetch(toApiUrl(`/api/v1/items?${qs.toString()}`), {
    signal,
  });
  if (!res.ok) throw new Error("Gagal mengambil data inventory");
  return (await res.json()) as InventoryResponse;
}

export async function fetchFilterOptions(
  signal?: AbortSignal,
): Promise<FilterOptionsResponse> {
  const res = await fetch(toApiUrl("/api/v1/items/filter-options"), { signal });
  if (!res.ok) throw new Error("Gagal mengambil opsi filter");
  const jenisData = await fetch(toApiUrl("/api/v1/item-types"), { signal });
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
  const res = await fetch(toApiUrl("/api/v1/projects"), { signal });
  if (!res.ok) throw new Error("Gagal mengambil opsi proyek");
  return (await res.json()) as ListOfProjectsResponse;
}

export async function createProject(
  payload: CreateProjectRequest,
  signal?: AbortSignal,
): Promise<CreateProjectResponse> {
  const res = await fetch(toApiUrl("/api/v1/projects"), {
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
  const res = await fetch(toApiUrl(`/api/v1/projects/${id}`), {
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
  const res = await fetch(toApiUrl(`/api/v1/projects/${id}`), {
    method: "DELETE",
    signal,
  });

  if (!res.ok) throw new Error("Gagal menghapus project");
  return (await res.json()) as DeleteProjectResponse;
}

export async function fetchJenisSuggestions(
  query: string,
  signal?: AbortSignal,
): Promise<JenisSuggestionsResponse> {
  const qs = new URLSearchParams();
  if (query.trim()) qs.append("q", query.trim());

  const res = await fetch(toApiUrl(`/api/v1/item-types?${qs.toString()}`), {
    signal,
  });

  if (!res.ok) throw new Error("Gagal mengambil suggestion jenis");
  return (await res.json()) as JenisSuggestionsResponse;
}

export async function addJenisSuggestion(
  jenis: string,
  signal?: AbortSignal,
): Promise<AddJenisResponse> {
  const res = await fetch(toApiUrl("/api/v1/item-types"), {
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
  const res = await fetch(toApiUrl(`/api/v1/item-types/${id}`), {
    method: "DELETE",
    signal,
  });

  if (!res.ok) throw new Error("Gagal menghapus suggestion jenis");
  return (await res.json()) as DeleteJenisResponse;
}
