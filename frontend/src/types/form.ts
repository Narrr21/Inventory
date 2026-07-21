export interface FieldInfo {
  id: string;
  label: string;
  placeholder?: string;
  nullable: boolean;
  value: string;
}

export interface SectionProps {
  label: string;
  fields: FieldInfo[];
}

export interface ItemFormData {
  id?: string;
  name: string;
  serial_number: string;
  licenseWindows: string;
  licenseOffice: string;
  status: string;
  jenis: string;
  proyek: string;
  idProyek: string;
  credentials: FieldInfo[];
  remoteInfo: FieldInfo[];
  customAttributes: FieldInfo[];
  deskripsi: string;
}