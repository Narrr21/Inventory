import React, { useEffect, useState } from "react";
import { Box } from "@mui/material";
import { BaseForm } from "./BaseForm";
import { BaseRow } from "./BaseRow";
import { BaseDropdown } from "./BaseDropdown";
import { BaseSection } from "./BaseSection";
import type { FieldInfo } from "../../types/form";
import type { FilterOption } from "../../types/dashboard";

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
  other: FieldInfo[];
}

interface ItemFormModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: ItemFormData) => void;
  initialData?: ItemFormData | null; // Untuk Edit Item
  options: {
    status: FilterOption[];
    jenis: FilterOption[];
    proyek: FilterOption[];
  };
}

const DEFAULT_FORM: ItemFormData = {
  name: "",
  serial_number: "",
  license_windows: "",
  license_office: "",
  status: "",
  jenis: "",
  proyek: "",
  credentials: [],
  remote_info: [],
  other: [],
};

export const ItemFormModal: React.FC<ItemFormModalProps> = ({
  open,
  onClose,
  onSubmit,
  initialData,
  options,
}) => {
  const [formData, setFormData] = useState<ItemFormData>(DEFAULT_FORM);
  const isEdit = Boolean(initialData?.id);

  useEffect(() => {
    if (initialData) {
      // Backend mapping transform -> Frontend format
      setFormData(initialData);
    } else {
      setFormData(DEFAULT_FORM);
    }
  }, [initialData, open]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
    onClose();
  };

  const handleSectionFieldChange = (
    sectionKey: "credentials" | "remote_info" | "other",
    fieldId: string,
    updated: Partial<FieldInfo>
  ) => {
    setFormData((prev) => ({
      ...prev,
      [sectionKey]: prev[sectionKey].map((item) =>
        item.id === fieldId ? { ...item, ...updated } : item
      ),
    }));
  };

  const handleAddSectionField = (sectionKey: "credentials" | "remote_info" | "other") => {
    const newField: FieldInfo = {
      id: Date.now().toString(),
      label: "Label Baru",
      placeholder: "Isikan data...",
      nullable: true,
      value: "",
    };
    setFormData((prev) => ({
      ...prev,
      [sectionKey]: [...prev[sectionKey], newField],
    }));
  };

  return (
    <BaseForm
      open={open}
      title={isEdit ? "Edit Item" : "Create Item"}
      onClose={onClose}
      onSubmit={handleSubmit}
      submitLabel={isEdit ? "Update" : "Simpan"}
    >
      {/* Primary Rows */}
      <Box sx={{ display: "flex", flexWrap: "wrap", width: "100%" }}>
        <BaseRow
          label="Nama Barang"
          value={formData.name}
          onChange={(val) => setFormData((p) => ({ ...p, name: val }))}
          required
        />
        <BaseRow
          label="Serial Number"
          value={formData.serial_number}
          onChange={(val) => setFormData((p) => ({ ...p, serial_number: val }))}
        />
        <BaseRow
          label="License Windows"
          value={formData.license_windows}
          onChange={(val) => setFormData((p) => ({ ...p, license_windows: val }))}
        />
        <BaseRow
          label="License Office"
          value={formData.license_office}
          onChange={(val) => setFormData((p) => ({ ...p, license_office: val }))}
        />
      </Box>

      {/* Dropdowns */}
      <Box sx={{ display: "flex", flexWrap: "wrap", gap: 2, p: 0.5 }}>
        <Box sx={{ width: { xs: "100%", lg: "48%" } }}>
          <BaseDropdown
            label="Status"
            value={formData.status}
            options={options.status}
            onChange={(val) => setFormData((p) => ({ ...p, status: val }))}
          />
        </Box>
        <Box sx={{ width: { xs: "100%", lg: "48%" } }}>
          <BaseDropdown
            label="Jenis"
            value={formData.jenis}
            options={options.jenis}
            onChange={(val) => setFormData((p) => ({ ...p, jenis: val }))}
          />
        </Box>
        <Box sx={{ width: { xs: "100%", lg: "48%" } }}>
          <BaseDropdown
            label="Proyek"
            value={formData.proyek}
            options={options.proyek}
            onChange={(val) => setFormData((p) => ({ ...p, proyek: val }))}
          />
        </Box>
      </Box>

      {/* Sections */}
      <BaseSection
        label="Credentials"
        fields={formData.credentials}
        onChangeField={(id, updated) => handleSectionFieldChange("credentials", id, updated)}
        onAddField={() => handleAddSectionField("credentials")}
      />

      <BaseSection
        label="Remote Info"
        fields={formData.remote_info}
        onChangeField={(id, updated) => handleSectionFieldChange("remote_info", id, updated)}
        onAddField={() => handleAddSectionField("remote_info")}
      />

      <BaseSection
        label="Other"
        fields={formData.other}
        onChangeField={(id, updated) => handleSectionFieldChange("other", id, updated)}
        onAddField={() => handleAddSectionField("other")}
      />
    </BaseForm>
  );
};