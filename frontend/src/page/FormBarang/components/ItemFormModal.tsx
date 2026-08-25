import React, { useEffect, useState } from "react";
import { Box } from "@mui/material";
import { BaseForm } from "./BaseForm";
import { BaseRow } from "./BaseRow";
import { BaseAutocomplete } from "./BaseAutocomplete";
import { BaseDropdown } from "./BaseDropdown";
import { BaseSection } from "./BaseSection";
import type { FieldInfo, ItemFormData } from "../../../types/form";
import type { Project, Jenis } from "../../../types/dashboard";
import { useDebounce } from "../../../utils/useDebounce";
import {
  addJenisSuggestion,
  deleteJenisSuggestion,
  fetchJenisSuggestions,
} from "../../../api/inventoryAPI";

interface ItemFormModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: ItemFormData) => void;
  initialData?: ItemFormData | null; // Untuk Edit Item
  options: {
    status: string[];
    jenis: Jenis[];
    proyek: Project[];
  };
}

const projectNames = (projects: Project[]): string[] => {
  return projects.map((project) => project.namaProyek);
};

const DEFAULT_FORM: ItemFormData = {
  name: "",
  serial_number: "",
  licenseWindows: "",
  licenseOffice: "",
  status: "",
  jenis: "",
  proyek: "",
  credentials: [],
  remoteInfo: [],
  customAttributes: [],
  deskripsi: "",
  idProyek: "",
};

export const ItemFormModal: React.FC<ItemFormModalProps> = ({
  open,
  onClose,
  onSubmit,
  initialData,
  options,
}) => {
  const [formData, setFormData] = useState<ItemFormData>(DEFAULT_FORM);
  const [jenisInput, setJenisInput] = useState("");
  const [jenisSuggestions, setJenisSuggestions] = useState<Jenis[]>(
    options.jenis,
  );
  const [isJenisLoading, setIsJenisLoading] = useState(false);
  const isEdit = Boolean(initialData?.id);
  const debouncedJenisInput = useDebounce(jenisInput, 250);

  useEffect(() => {
    if (initialData) {
      // Backend mapping transform -> Frontend format
      setFormData(initialData);
      setJenisInput(initialData.jenis || "");
      console.log("Initial Data:", initialData);
    } else {
      setFormData(DEFAULT_FORM);
      setJenisInput("");
    }
  }, [initialData, open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const controller = new AbortController();
    const query = debouncedJenisInput.trim();

    setIsJenisLoading(true);

    fetchJenisSuggestions(query, controller.signal)
      .then((response) => {
        const data = response.data;
        const normalized = Array.isArray(data) ? data : data ? [data] : [];
        setJenisSuggestions(normalized);
      })
      .catch((error) => {
        if (error?.name !== "AbortError") {
          setJenisSuggestions(options.jenis);
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsJenisLoading(false);
        }
      });

    return () => controller.abort();
  }, [debouncedJenisInput, open, options.jenis]);

  useEffect(() => {
    if (formData.proyek) {
      const matchedProject = options.proyek.find(
        (p) => p.namaProyek === formData.proyek,
      );
      const matchedId = matchedProject ? matchedProject._id : "";

      // Mencegah re-render loop
      if (formData.idProyek !== matchedId) {
        setFormData((prev) => ({ ...prev, idProyek: matchedId }));
      }
    }
  }, [formData.proyek, options.proyek]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
    onClose();
  };

  const handleSectionFieldChange = (
    sectionKey: "credentials" | "remoteInfo" | "customAttributes",
    fieldId: string,
    updated: Partial<FieldInfo>,
  ) => {
    setFormData((prev) => ({
      ...prev,
      [sectionKey]: prev[sectionKey].map((item) =>
        item.id === fieldId ? { ...item, ...updated } : item,
      ),
    }));
  };

  const handleAddSectionField = (
    sectionKey: "credentials" | "remoteInfo" | "customAttributes",
  ) => {
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

  const handleCreateJenis = async (value: string) => {
    const nextValue = value.trim();

    if (!nextValue) {
      return;
    }

    const response = await addJenisSuggestion(nextValue);
    const data = response.data;
    const normalized = Array.isArray(data) ? data : data ? [data] : [];
    // merge new entries with existing suggestions, avoiding duplicates by jenis
    setJenisSuggestions((prev) => {
      const map = new Map<string, Jenis>();
      prev.forEach((j) => map.set(j.jenis, j));
      normalized.forEach((j) => map.set(j.jenis, j));
      return Array.from(map.values());
    });
    setJenisInput(nextValue);
    setFormData((prev) => ({ ...prev, jenis: nextValue }));
  };

  const handleDeleteJenis = async (id: string) => {
    const target = jenisSuggestions.find((j) => j._id === id);
    const name = target ? target.jenis : id;
    const confirmed = window.confirm(`Hapus jenis "${name}" dari suggestion?`);

    if (!confirmed) return;

    try {
      const response = await deleteJenisSuggestion(id);
      const data = response.data;
      const normalized = Array.isArray(data) ? data : data ? [data] : [];
      // if backend returns remaining list, use it; else remove deleted from current
      if (normalized.length > 0) {
        setJenisSuggestions(normalized);
      } else {
        setJenisSuggestions((prev) => prev.filter((j) => j._id !== id));
      }

      setFormData((prev) =>
        prev.jenis === name ? { ...prev, jenis: "" } : prev,
      );
      setJenisInput((prev) => (prev === name ? "" : prev));
    } catch {
      window.alert("Gagal menghapus jenis. Silakan coba lagi.");
    }
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
          value={formData.licenseWindows}
          onChange={(val) =>
            setFormData((p) => ({ ...p, licenseWindows: val }))
          }
        />
        <BaseRow
          label="License Office"
          value={formData.licenseOffice}
          onChange={(val) => setFormData((p) => ({ ...p, licenseOffice: val }))}
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
          <BaseAutocomplete
            label="Jenis"
            value={formData.jenis}
            inputValue={jenisInput}
            options={jenisSuggestions}
            onValueChange={(val) => setFormData((p) => ({ ...p, jenis: val }))}
            onInputChange={(val) => {
              setJenisInput(val);
              setFormData((p) => ({ ...p, jenis: val }));
            }}
            onCreateOption={handleCreateJenis}
            onDeleteOption={handleDeleteJenis}
            placeholder="Ketik atau pilih jenis"
            helperText="Pilih suggestion yang tersedia atau ketik jenis baru. Jika tidak ada yang cocok, pilih Tambahkan ke database."
            loading={isJenisLoading}
          />
        </Box>
        <Box sx={{ width: { xs: "100%", lg: "48%" } }}>
          <BaseDropdown
            label="Proyek"
            value={formData.proyek}
            options={projectNames(options.proyek)}
            onChange={(val) => setFormData((p) => ({ ...p, proyek: val }))}
          />
        </Box>
      </Box>

      {/* Sections */}
      <BaseSection
        label="Credentials"
        fields={formData.credentials}
        onChangeField={(id, updated) =>
          handleSectionFieldChange("credentials", id, updated)
        }
        onAddField={() => handleAddSectionField("credentials")}
      />

      <BaseSection
        label="Remote Info"
        fields={formData.remoteInfo}
        onChangeField={(id, updated) =>
          handleSectionFieldChange("remoteInfo", id, updated)
        }
        onAddField={() => handleAddSectionField("remoteInfo")}
      />

      <BaseSection
        label="Custom Attributes"
        fields={formData.customAttributes}
        onChangeField={(id, updated) =>
          handleSectionFieldChange("customAttributes", id, updated)
        }
        onAddField={() => handleAddSectionField("customAttributes")}
      />
    </BaseForm>
  );
};
