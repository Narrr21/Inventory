import type { FilterOption, SortDirection } from "../types/dashboard";

const USE_MOCK = true;
// Note: Dapat menghapus USE_MOCK dan logic MOCK dibawah jika backend selesai

export interface InventoryQueryParams {
  search: string;
  status: string;
  project: string[];
  jenisBarang: string[];
  sortKey?: string;
  sortDirection: SortDirection;
  page: number;
  pageSize: number;
}

export type InventoryRow = Record<string, string | number>;

export interface InventoryResponse {
  rows: InventoryRow[];
  total: number;
}

export interface FilterOptionsResponse {
  project: FilterOption[];
  jenisBarang: FilterOption[];
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

const MOCK_DB: InventoryRow[] = Array.from({ length: 200 }, (_, i) => ({
  namaBarang: `${NAMA_POOL[i % NAMA_POOL.length]} #${i + 1}`,
  project: PROJECT_POOL[i % PROJECT_POOL.length],
  jenisBarang: JENIS_POOL[i % JENIS_POOL.length],
  status: STATUS_POOL[i % STATUS_POOL.length],
  qty: (i * 7) % 40,
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
      params.search === "" ||
      String(row.namaBarang).toLowerCase().includes(params.search.toLowerCase());
    const matchesStatus = params.status === "" || row.status === params.status;
    const matchesProject = params.project.length === 0 || params.project.includes(String(row.project));
    const matchesJenis =
      params.jenisBarang.length === 0 || params.jenisBarang.includes(String(row.jenisBarang));
    return matchesSearch && matchesStatus && matchesProject && matchesJenis;
  });

  if (params.sortKey) {
    const key = params.sortKey;
    rows = [...rows].sort((a, b) => {
      const av = a[key];
      const bv = b[key];
      if (av === bv) return 0;
      const result = av > bv ? 1 : -1;
      return params.sortDirection === "asc" ? result : -result;
    });
  }

  const total = rows.length;
  const start = (params.page - 1) * params.pageSize;
  const paged = rows.slice(start, start + params.pageSize);

  return simulateNetwork({ rows: paged, total }, signal);
}

async function fetchFilterOptionsMock(signal?: AbortSignal): Promise<FilterOptionsResponse> {
  const toOptions = (values: string[]): FilterOption[] => values.map((v) => ({ value: v, label: v }));

  return simulateNetwork(
    {
      project: toOptions(PROJECT_POOL),
      jenisBarang: toOptions(JENIS_POOL),
    },
    signal
  );
}

// API

async function fetchInventoryApi(
  params: InventoryQueryParams,
  signal?: AbortSignal
): Promise<InventoryResponse> {
  const qs = new URLSearchParams({
    search: params.search,
    status: params.status,
    project: params.project.join(","),
    jenisBarang: params.jenisBarang.join(","),
    sortKey: params.sortKey ?? "",
    sortDirection: params.sortDirection,
    page: String(params.page),
    pageSize: String(params.pageSize),
  });
  const res = await fetch(`/api/inventory?${qs.toString()}`, { signal });
  if (!res.ok) throw new Error("Gagal mengambil data inventory");
  return (await res.json()) as InventoryResponse;
}

async function fetchFilterOptionsApi(signal?: AbortSignal): Promise<FilterOptionsResponse> {
  const res = await fetch("/api/inventory/filter-options", { signal });
  if (!res.ok) throw new Error("Gagal mengambil opsi filter");
  return (await res.json()) as FilterOptionsResponse;
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