import type {
  AnalyticsSummaryResponse,
  AnalyticsTimelineResponse,
  ProjectMapResponse,
} from "../types/api";

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
