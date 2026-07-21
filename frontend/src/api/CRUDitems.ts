import type { BackendItem, Project } from "../types/dashboard";

export interface CreateItemRequest {}

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

async function createItemAPI(
  req: CreateItemRequest,
  signal?: AbortSignal,
): Promise<CreateItemResponse> {
  const res = await fetch("/api/v1/items", {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(req)
  });
  if (!res.ok) throw new Error("Gagal membuat item");
  return (await res.json()) as CreateItemResponse;
}

export async function createItem(
  req: CreateItemRequest,
  signal?: AbortSignal,
): Promise<CreateItemResponse> {
  return createItemAPI(req, signal);
}
