export interface ColumnDef {
  /** key yang match dengan field pada setiap row */
  key: string;
  /** label yang ditampilkan pada AttributeRow (header) */
  label: string;
  /** lebar kolom tetap dalam px */
  width?: number;
  /** default true. set false untuk kolom yang tidak bisa di-sort */
  sortable?: boolean;
  /** default false. jika true, tampilkan tombol salin per-cell */
  copy?: boolean;
}

export type SortDirection = "asc" | "desc";

export interface TableData {
  columns: ColumnDef[];
  rows: Record<string, string | number>[];
}

export interface FilterOption {
  /** value yang dikirim ke API */
  value: string;
  /** label yang ditampilkan pada UI */
  label: string;
}

export interface BackendItem {
  _id: string;
  jenis: string;
  serialNumber: string;
  nama: string;
  idProyek: string;
  namaProyek: string;
  credentials?: Record<string, string>;
  remoteInfo?: Record<string, string>;
  licenseWindows?: string;
  licenseOffice?: string;
  status: string;
  deskripsi?: string;
  customAttributes?: Record<string, string>;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectCoordinate {
  lat: number;
  lng: number;
}

export interface Project {
  _id: string;
  namaProyek: string;
  lokasi: string;
  koordinat?: ProjectCoordinate | null;
}

export interface Jenis {
  _id: string;
  jenis: string;
}
