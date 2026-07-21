import type { BackendItem, Project } from "../types/dashboard";

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

async function createItemAPI(
  req: CreateItemRequest,
  signal?: AbortSignal,
): Promise<CreateItemResponse> {
  const { customAttributes, ...rest } = req;

  const cleanedPayload = Object.fromEntries(
    Object.entries({
      ...rest,
      ...customAttributes,
    }).filter(([_, value]) => value !== undefined && value !== ""),
  );

  const res = await fetch("/api/v1/items", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(cleanedPayload),
    signal,
  });

  if (!res.ok) throw new Error("Gagal membuat item baru");
  return (await res.json()) as CreateItemResponse;
}

// Update Item API (PUT)
export async function updateItem(
  id: string,
  payload: UpdateItemRequest,
  signal?: AbortSignal,
): Promise<UpdateItemResponse> {
  const { customAttributes, ...rest } = payload;

  const cleanedPayload = Object.fromEntries(
    Object.entries({
      ...rest,
      ...customAttributes,
    }).filter(([_, value]) => value !== undefined && value !== ""),
  );

  const res = await fetch(`/api/v1/items/${id}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(cleanedPayload),
    signal,
  });

  if (!res.ok) throw new Error("Gagal meng-update item");
  return (await res.json()) as UpdateItemResponse;
}

// Delete Item API (DELETE)
export async function deleteItemAPI(
  id: string,
  signal?: AbortSignal,
): Promise<DeleteItemResponse> {
  const res = await fetch(`/api/v1/items/${id}`, {
    method: "DELETE",
    signal,
  });

  if (!res.ok) throw new Error("Gagal menghapus item");
  return (await res.json()) as DeleteItemResponse;
}

export async function createItem(
  req: CreateItemRequest,
  signal?: AbortSignal,
): Promise<CreateItemResponse> {
  return createItemAPI(req, signal);
}
