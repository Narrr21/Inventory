import type { SortDirection, BackendItem, Project } from "../types/dashboard";

const USE_MOCK = false;
// Note: Dapat menghapus USE_MOCK dan logic MOCK dibawah jika backend selesai

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
    project: string[];
    jenis: string[];
    status: string[];
  };
}

export interface ListOfProjectsResponse {
  success: boolean;
  data: Project[];
}

// MOCK

const PROJECT_POOL = ["Migrasi Sistem", "Renovasi Kantor", "Ruang Meeting", "Ekspansi Cabang"];
const JENIS_POOL = ["Elektronik", "Furnitur", "Aksesoris", "Alat Tulis"];
const STATUS_POOL = ["Healthy", "Under Maintenance", "Broken"];
const NAMA_POOL = [
  "Laptop Dell XPS 13",
  "Meja Kerja Standing Desk Elektrik",
  "Proyektor Epson EB-X05",
  "Kabel HDMI 10 Meter",
  "Kursi Ergonomis Herman Miller",
  'Monitor LG UltraWide 34"',
  "Printer HP LaserJet Pro",
  "Rak Server 42U",
  "Whiteboard Magnetik 120x90",
  "Router Mikrotik CCR2004",
  "Switch 24 Port Gigabit",
  "Lemari Arsip Besi",
];

const MOCK_DB: BackendItem[] = Array.from({ length: 200 }, (_, i) => ({
  _id: `item-${i + 1}`,
  nama: NAMA_POOL[i % NAMA_POOL.length],
  serialNumber: `SN-${1000 + i}`,
  license_windows: `WIN-${2000 + i}`,
  license_office: `OFFICE-${3000 + i}`,
  status: STATUS_POOL[i % STATUS_POOL.length],
  jenis: JENIS_POOL[i % JENIS_POOL.length],
  idProyek: PROJECT_POOL[i % PROJECT_POOL.length],
  created_at: new Date(Date.now() - i * 1000 * 60 * 60 * 24).toISOString(),
  updated_at: new Date(Date.now() - i * 1000 * 60 * 60 * 24).toISOString(),
  credentials: {
    username: `user${i + 1}`,
    password: `pass${i + 1}`,
  },
  remote_info: {
    ip_address: `192.168.1.${i + 1}`,
    mac_address: `00:1A:2B:3C:4D:${(i + 1).toString(16).padStart(2, "0")}`,
  },
  customAttributes: {
    notes: `Catatan untuk item ${i + 1}`,
  },
}));

const NETWORK_DELAY_MS = 350;

const simulateNetwork = <T,>(value: T, signal?: AbortSignal): Promise<T> =>
  new Promise((resolve, reject) => {
    const timer = setTimeout(() => resolve(value), NETWORK_DELAY_MS);
    signal?.addEventListener("abort", () => {
      clearTimeout(timer);
      reject(new DOMException("Aborted", "AbortError"));
    });
  });

async function fetchInventoryMock(
  params: InventoryQueryParams,
  signal?: AbortSignal
): Promise<InventoryResponse> {
  let rows = MOCK_DB.filter((row) => {
    const matchesSearch =
      params.search === "" || row.nama.toLowerCase().includes(params.search?.toLowerCase() ?? "") ||
      row.serialNumber.toLowerCase().includes(params.search?.toLowerCase() ?? "");
    const matchesStatus = params.status === "" || row.status === params.status;
    const matchesProject = params.proyek === "" || row.idProyek === params.proyek;
    const matchesJenis =
      params.jenis === "" || row.jenis === params.jenis;
    return matchesSearch && matchesStatus && matchesProject && matchesJenis;
  });

  if (params.sortBy) {
    const key = params.sortBy;
    rows = [...rows].sort((a, b) => {
      const av = String(a[key as keyof BackendItem] ?? "").toLowerCase();
      const bv = String(b[key as keyof BackendItem] ?? "").toLowerCase();
      if (av === "" && bv !== "") return params.sortOrder === "asc" ? 1 : -1;
      if (av === bv) return 0;
      const result = av > bv ? 1 : -1;
      return params.sortOrder === "asc" ? result : -result;
    });
  }

  const total = rows.length;
  const start = (params.page - 1) * params.limit;
  const paged = rows.slice(start, start + params.limit);

  return simulateNetwork(
    {
      success: true,
      data: paged,
      meta: {
        total,
        page: params.page,
        limit: params.limit,
        totalPages: Math.ceil(total / params.limit),
      },
    },
    signal
  );
}

async function fetchFilterOptionsMock(signal?: AbortSignal): Promise<FilterOptionsResponse> {
  return simulateNetwork(
    {
      success: true,
      data: {
        project: PROJECT_POOL,
        jenis: JENIS_POOL,
        status: STATUS_POOL,
      },
    },
    signal
  );
}

// API

async function fetchInventoryApi(
  params: InventoryQueryParams,
  signal?: AbortSignal
): Promise<InventoryResponse> {
  const qs = new URLSearchParams();
  if (params.jenis) qs.append("jenis", params.jenis);
  if (params.status) qs.append("status", params.status);
  if (params.proyek) qs.append("proyek", params.proyek);
  if (params.search) qs.append("search", params.search);
  if (params.sortBy) qs.append("sortBy", params.sortBy);
  if (params.sortOrder) qs.append("sortOrder", params.sortOrder);
  qs.append("page", String(params.page));
  qs.append("limit", String(params.limit));
  const res = await fetch(`/api/v1/items?${qs.toString()}`, { signal });
  if (!res.ok) throw new Error("Gagal mengambil data inventory");
  return (await res.json()) as InventoryResponse;
}

async function fetchFilterOptionsApi(signal?: AbortSignal): Promise<FilterOptionsResponse> {
  const res = await fetch("/api/v1/items/filter-options", { signal });
  if (!res.ok) throw new Error("Gagal mengambil opsi filter");
  return (await res.json()) as FilterOptionsResponse;
}

async function fetchListOfProjects(signal?: AbortSignal): Promise<ListOfProjectsResponse> {
  const res = await fetch("/api/v1/projects", { signal });
  if (!res.ok) throw new Error("Gagal mengambil opsi proyek");
  return (await res.json()) as ListOfProjectsResponse;
}

// EXPORT

export async function fetchInventory(
  params: InventoryQueryParams,
  signal?: AbortSignal
): Promise<InventoryResponse> {
  if (USE_MOCK) {
    return fetchInventoryMock(params, signal);
  }
  return fetchInventoryApi(params, signal);
}

export async function fetchFilterOptions(signal?: AbortSignal): Promise<FilterOptionsResponse> {
  if (USE_MOCK) {
    return fetchFilterOptionsMock(signal);
  }
  return fetchFilterOptionsApi(signal);
}