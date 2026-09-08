import type { BackendItem, Jenis, Project } from "./dashboard";

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

export interface InventoryQueryParams {
  jenis?: string;
  status?: string;
  proyek?: string;
  search?: string;
  sortBy?: string;
  sortOrder?: "asc" | "desc";
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

export interface ProjectCoordinateInput {
  lat: number | null;
  lng: number | null;
}

export interface CreateProjectRequest {
  namaProyek: string;
  lokasi: string;
  koordinat: ProjectCoordinateInput | null;
}

export interface CreateProjectResponse {
  success: boolean;
  data: Project;
}

export interface UpdateProjectRequest {
  namaProyek?: string;
  lokasi?: string;
  koordinat?: ProjectCoordinateInput | null;
}

export interface UpdateProjectResponse {
  success: boolean;
  data: Project;
}

export interface CreateItemRequest {
  jenis: string;
  serialNumber: string;
  nama?: string;
  idProyek: string;
  status: string;
  licenseWindows?: string;
  licenseOffice?: string;
  deskripsi?: string;
  credentials?: Record<string, string>;
  remoteInfo?: Record<string, string>;
  customAttributes?: Record<string, string>;
}

export interface CreateItemResponse {
  success: boolean;
  data: BackendItem;
}

export interface UpdateItemRequest {
  jenis?: string;
  serialNumber?: string;
  nama?: string;
  idProyek?: string;
  status?: string;
  licenseWindows?: string;
  licenseOffice?: string;
  deskripsi?: string;
  credentials?: Record<string, string>;
  remoteInfo?: Record<string, string>;
  customAttributes?: Record<string, string>;
}

export interface UpdateItemResponse {
  success: boolean;
  data: BackendItem;
}

export interface DeleteItemResponse {
  success: boolean;
  data: {
    _id: string;
  };
}
