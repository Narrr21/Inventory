export interface ColumnDef {
  /** key yang match dengan field pada setiap row */
  key: string;
  /** label yang ditampilkan pada AttributeRow (header) */
  label: string;
  /** lebar kolom tetap dalam px */
  width?: number;
  /** default true. set false untuk kolom yang tidak bisa di-sort */
  sortable?: boolean;
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
  id: string;
  name: string;
  serial_number: string;
  license_windows: string;
  license_office: string;
  status: string;
  jenis: string;
  proyek: string;
  created_at?: string;
  credentials?: Record<string, string>;
  remote_info?: Record<string, string>;
  other?: Record<string, string>;
}