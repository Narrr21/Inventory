// Tipe Data dari Backend API

import type { ItemFormData } from "../types/form";
import type { FieldInfo } from "../types/form";
import type { BackendItem } from "../types/dashboard";

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
    id: item.id,
    name: item.name || "",
    serial_number: item.serial_number || "",
    license_windows: item.license_windows || "",
    license_office: item.license_office || "",
    status: item.status || "",
    jenis: item.jenis || "",
    proyek: item.proyek || "",
    credentials: mapObjectToFields(item.credentials),
    remote_info: mapObjectToFields(item.remote_info),
    other: mapObjectToFields(item.other),
  };
};

// 2. Transform Frontend -> Backend (Untuk Submit Edit/PUT)
export const mapItemFormToBackend = (formData: ItemFormData): BackendItem => {
  return {
    id: formData.id || "",
    name: formData.name,
    serial_number: formData.serial_number,
    license_windows: formData.license_windows,
    license_office: formData.license_office,
    status: formData.status,
    jenis: formData.jenis,
    proyek: formData.proyek,
    credentials: mapFieldsToObject(formData.credentials),
    remote_info: mapFieldsToObject(formData.remote_info),
    other: mapFieldsToObject(formData.other),
  };
};