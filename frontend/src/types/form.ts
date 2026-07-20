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
  license_windows: string;
  license_office: string;
  status: string;
  jenis: string;
  proyek: string;
  credentials: FieldInfo[];
  remote_info: FieldInfo[];
  customAttributes: FieldInfo[];
}