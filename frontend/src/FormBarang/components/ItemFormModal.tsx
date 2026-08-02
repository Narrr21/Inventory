import React, { useEffect, useState } from "react";
import { Box } from "@mui/material";
import { BaseForm } from "./BaseForm";
import { BaseRow } from "./BaseRow";
import { BaseAutocomplete } from "./BaseAutocomplete";
import { BaseDropdown } from "./BaseDropdown";
import { BaseSection } from "./BaseSection";
import type { FieldInfo, ItemFormData } from "../../types/form";
import type { Project } from "../../types/dashboard";
import { useDebounce } from "../../Dashboard/hooks/useDebounce";
import {
  addJenisSuggestion,
  deleteJenisSuggestion,
  fetchJenisSuggestions,
} from "../../api/InventoryAPI";

interface ItemFormModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: ItemFormData) => void;
  initialData?: ItemFormData | null; // Untuk Edit Item
  options: {
    status: string[];
    jenis: string[];
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
  const [jenisSuggestions, setJenisSuggestions] = useState<string[]>(
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
        setJenisSuggestions(response.data);
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
    setJenisSuggestions(response.data);
    setJenisInput(nextValue);
    setFormData((prev) => ({ ...prev, jenis: nextValue }));
  };

  const handleDeleteJenis = async (value: string) => {
    const confirmed = window.confirm(`Hapus jenis "${value}" dari suggestion?`);

    if (!confirmed) {
      return;
    }

    const response = await deleteJenisSuggestion(value);
    setJenisSuggestions(response.data);

    setFormData((prev) =>
      prev.jenis === value ? { ...prev, jenis: "" } : prev,
    );

    setJenisInput((prev) => (prev === value ? "" : prev));
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
