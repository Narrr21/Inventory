import React, { useEffect, useState } from "react";
import { Box, TextField } from "@mui/material";
import { BaseForm } from "./BaseForm";
import { BaseRow } from "./BaseRow";

export interface ProjectFormData {
  id?: string;
  name: string;
  location: string;
  lat?: string;
  lng?: string;
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
  const [formData, setFormData] = useState<ProjectFormData>({
    name: "",
    location: "",
    lat: "",
    lng: "",
  });
  const [coordError, setCoordError] = useState<string | null>(null);
  const isEdit = Boolean(initialData?.id);

  useEffect(() => {
    if (initialData) {
      setFormData({
        ...initialData,
        lat: initialData.lat ?? "",
        lng: initialData.lng ?? "",
      });
    } else {
      setFormData({ name: "", location: "", lat: "", lng: "" });
    }
    setCoordError(null);
  }, [initialData, open]);

  const handleFieldChange = (
    field: "name" | "location" | "lat" | "lng",
    value: string,
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (field === "lat" || field === "lng") {
      const lat = field === "lat" ? value : (formData.lat ?? "");
      const lng = field === "lng" ? value : (formData.lng ?? "");
      const hasLat = lat.trim() !== "";
      const hasLng = lng.trim() !== "";
      if ((hasLat && !hasLng) || (!hasLat && hasLng)) {
        setCoordError("Latitude dan longitude harus diisi bersamaan.");
      } else {
        setCoordError(null);
      }
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const hasLat = (formData.lat ?? "").trim() !== "";
    const hasLng = (formData.lng ?? "").trim() !== "";
    const bothEmpty = !hasLat && !hasLng;
    const partial = (hasLat && !hasLng) || (!hasLat && hasLng);

    if (partial) {
      setCoordError("Latitude dan longitude harus diisi bersamaan.");
      return;
    }

    setCoordError(null);

    if (bothEmpty) {
      onSubmit({ ...formData, lat: "", lng: "" });
      onClose();
      return;
    }

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
          onChange={(val) => handleFieldChange("name", val)}
          required
        />
        <BaseRow
          label="Lokasi"
          value={formData.location}
          placeholder="Masukkan lokasi project"
          onChange={(val) => handleFieldChange("location", val)}
        />
        <Box sx={{ width: { xs: "100%", lg: "50%" }, p: 0.5 }}>
          <TextField
            size="small"
            fullWidth
            label="Latitude"
            value={formData.lat ?? ""}
            placeholder="-6.2088"
            onChange={(e) => handleFieldChange("lat", e.target.value)}
            error={Boolean(coordError)}
            helperText={coordError}
          />
        </Box>
        <Box sx={{ width: { xs: "100%", lg: "50%" }, p: 0.5 }}>
          <TextField
            size="small"
            fullWidth
            label="Longitude"
            value={formData.lng ?? ""}
            placeholder="106.8456"
            onChange={(e) => handleFieldChange("lng", e.target.value)}
            error={Boolean(coordError)}
            helperText={coordError}
          />
        </Box>
      </Box>
    </BaseForm>
  );
};
