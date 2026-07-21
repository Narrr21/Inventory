// Tipe Data dari Backend API

import type { ItemFormData } from "../types/form";
import type { FieldInfo } from "../types/form";
import type { BackendItem } from "../types/dashboard";
import type { CreateItemRequest, UpdateItemRequest } from "./CRUDitems";

// Helper konversi objek { [key]: value } dari backend ke FieldInfo[]
const mapObjectToFields = (obj?: Record<string, string>): FieldInfo[] => {
  if (!obj) return [];
  return Object.entries(obj).map(([key, val], index) => ({
    id: `${key}-${index}`,
    label: key,
    placeholder: "Isikan data...",
    nullable: true,
    value: val,
  }));
};

// Helper konversi FieldInfo[] frontend kembali ke objek { [key]: value } untuk dikirim ke backend
const mapFieldsToObject = (fields: FieldInfo[]): Record<string, string> => {
  const result: Record<string, string> = {};
  fields.forEach((field) => {
    if (field.label.trim()) {
      result[field.label] = field.value;
    }
  });
  return result;
};

// 1. Transform Backend -> Frontend (Untuk Edit/Pre-fill)
export const mapBackendToItemForm = (item: BackendItem): ItemFormData => {
  return {
    id: item._id,
    name: item.nama || "",
    serial_number: item.serialNumber || "",
    licenseWindows: item.licenseWindows || "",
    licenseOffice: item.licenseOffice || "",
    status: item.status || "",
    jenis: item.jenis || "",
    proyek: item.namaProyek || "",
    credentials: mapObjectToFields(item.credentials),
    remoteInfo: mapObjectToFields(item.remoteInfo),
    customAttributes: mapObjectToFields(item.customAttributes),
    deskripsi: item.deskripsi || "",
    idProyek: item.idProyek || "",
  };
};

// 2. Transform Frontend -> Backend (Untuk Submit Edit/PUT)
export const mapItemFormToBackend = (formData: ItemFormData): BackendItem => {
  return {
    _id: formData.id || "",
    nama: formData.name,
    serialNumber: formData.serial_number,
    licenseWindows: formData.licenseWindows,
    licenseOffice: formData.licenseOffice,
    status: formData.status,
    jenis: formData.jenis,
    idProyek: formData.idProyek,
    namaProyek: formData.proyek,
    credentials: mapFieldsToObject(formData.credentials),
    remoteInfo: mapFieldsToObject(formData.remoteInfo),
    customAttributes: mapFieldsToObject(formData.customAttributes),
    createdAt: "", // Placeholder, backend will handle this
    updatedAt: "", // Placeholder, backend will handle this
  };
};

export const mapItemFormToCreateRequest = (formData: ItemFormData): CreateItemRequest => {
  return {
    jenis: formData.jenis,
    serialNumber: formData.serial_number,
    nama: formData.name,
    idProyek: formData.idProyek,
    status: formData.status,
    licenseWindows: formData.licenseWindows,
    licenseOffice: formData.licenseOffice,
    deskripsi: formData.deskripsi,
    credentials: mapFieldsToObject(formData.credentials),
    remoteInfo: mapFieldsToObject(formData.remoteInfo),
    customAttributes: mapFieldsToObject(formData.customAttributes),
  };
}

export const mapItemFormToUpdateRequest = (formData: ItemFormData): Partial<UpdateItemRequest> => {
  return {
    jenis: formData.jenis,
    serialNumber: formData.serial_number,
    nama: formData.name,
    idProyek: formData.idProyek,
    status: formData.status,
    licenseWindows: formData.licenseWindows,
    licenseOffice: formData.licenseOffice,
    deskripsi: formData.deskripsi,
    credentials: mapFieldsToObject(formData.credentials),
    remoteInfo: mapFieldsToObject(formData.remoteInfo),
    customAttributes: mapFieldsToObject(formData.customAttributes),
  };
}