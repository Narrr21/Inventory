export interface ProjectMapCoordinate {
  lat: number;
  lng: number;
}

export interface ProjectMapItem {
  _id: string;
  namaProyek: string;
  lokasi: string;
  koordinat: ProjectMapCoordinate;
  totalItems: number;
  dominantStatus?: string;
  byStatus?: Array<{ status: string; count: number }>;
}

export interface UnmappedProjectItem {
  _id: string;
  namaProyek: string;
  lokasi: string;
  totalItems: number;
}

export interface ProjectMapBounds {
  north: number;
  south: number;
  east: number;
  west: number;
}

export interface ProjectMapResponse {
  success: boolean;
  code: number;
  data: {
    projects: ProjectMapItem[];
    unmapped: UnmappedProjectItem[];
    bounds: ProjectMapBounds | null;
    totals: {
      projects: number;
      mapped: number;
      unmapped: number;
      items: number;
      itemsMapped: number;
    };
  };
}

export interface AnalyticsStatusCount {
  status: string;
  count: number;
}

export interface AnalyticsCategoryCount {
  jenis: string;
  count: number;
}

export interface AnalyticsProjectCount {
  idProyek: string;
  namaProyek: string;
  lokasi: string;
  count: number;
}

export interface AnalyticsLicenseCount {
  value: string;
  count: number;
}

export interface AnalyticsNeedsAttention {
  staleDays: number;
  staleItems: number;
  orphanItems: number;
  projectsWithoutKoordinat: number;
  projectsWithoutItems: number;
}

export interface AnalyticsSummaryData {
  totals: {
    items: number;
    projects: number;
    itemTypes: number;
    projectsMapped: number;
  };
  byStatus: AnalyticsStatusCount[];
  byJenis: AnalyticsCategoryCount[];
  byProyek: AnalyticsProjectCount[];
  licenses: {
    windows: AnalyticsLicenseCount[];
    office: AnalyticsLicenseCount[];
  };
  needsAttention: AnalyticsNeedsAttention;
  recentlyAdded: Array<Record<string, unknown>>;
  recentlyUpdated: Array<Record<string, unknown>>;
}

export interface AnalyticsSummaryResponse {
  success: boolean;
  code: number;
  data: AnalyticsSummaryData;
}

export interface AnalyticsTimelineBucket {
  period: string;
  created: number;
  cumulative: number;
}

export interface AnalyticsTimelineData {
  months: number;
  from: string;
  to: string;
  buckets: AnalyticsTimelineBucket[];
}

export interface AnalyticsTimelineResponse {
  success: boolean;
  code: number;
  data: AnalyticsTimelineData;
}

const toApiUrl = (path: string): string => {
  if (typeof window !== "undefined" && window.location?.origin) {
    return new URL(path, window.location.origin).toString();
  }

  return new URL(path, "http://localhost").toString();
};

export async function fetchProjectMap(
  signal?: AbortSignal,
): Promise<ProjectMapResponse> {
  const res = await fetch(toApiUrl("/api/v1/analytics/map"), { signal });

  if (!res.ok) {
    throw new Error("Gagal mengambil data peta project");
  }

  return (await res.json()) as ProjectMapResponse;
}

export async function fetchAnalyticsSummary(
  staleDays = 90,
  signal?: AbortSignal,
): Promise<AnalyticsSummaryResponse> {
  const params = new URLSearchParams({ staleDays: String(staleDays) });
  const res = await fetch(toApiUrl(`/api/v1/analytics/summary?${params}`), {
    signal,
  });

  if (!res.ok) {
    throw new Error("Gagal mengambil data analitik");
  }

  return (await res.json()) as AnalyticsSummaryResponse;
}

export async function fetchAnalyticsTimeline(
  months = 12,
  signal?: AbortSignal,
): Promise<AnalyticsTimelineResponse> {
  const params = new URLSearchParams({ months: String(months) });
  const res = await fetch(toApiUrl(`/api/v1/analytics/timeline?${params}`), {
    signal,
  });

  if (!res.ok) {
    throw new Error("Gagal mengambil timeline analitik");
  }

  return (await res.json()) as AnalyticsTimelineResponse;
}
