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