import React, { useEffect, useState } from "react";
import { Box } from "@mui/material";
import { BaseForm } from "./BaseForm";
import { BaseRow } from "./BaseRow";

export interface ProjectFormData {
  id?: string;
  name: string;
  location: string;
}

interface ProjectFormModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: ProjectFormData) => void;
  initialData?: ProjectFormData | null;
}

export const ProjectFormModal: React.FC<ProjectFormModalProps> = ({
  open,
  onClose,
  onSubmit,
  initialData,
}) => {
  const [formData, setFormData] = useState<ProjectFormData>({ name: "", location: "" });
  const isEdit = Boolean(initialData?.id);

  useEffect(() => {
    if (initialData) {
      setFormData(initialData);
    } else {
      setFormData({ name: "", location: "" });
    }
  }, [initialData, open]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
    onClose();
  };

  return (
    <BaseForm
      open={open}
      title={isEdit ? "Edit Project" : "Create Project"}
      onClose={onClose}
      onSubmit={handleSubmit}
      submitLabel={isEdit ? "Update Project" : "Simpan Project"}
    >
      <Box sx={{ display: "flex", flexWrap: "wrap", width: "100%" }}>
        <BaseRow
          label="Nama Project"
          value={formData.name}
          placeholder="Masukkan nama project"
          onChange={(val) => setFormData((p) => ({ ...p, name: val }))}
          required
        />
        <BaseRow
          label="Lokasi"
          value={formData.location}
          placeholder="Masukkan lokasi project"
          onChange={(val) => setFormData((p) => ({ ...p, location: val }))}
        />
      </Box>
    </BaseForm>
  );
};