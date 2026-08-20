import React, { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  InputAdornment,
  TextField,
  Typography,
} from "@mui/material";
import DataTable from "./components/DataTable";
import MainLayout from "../layouts/MainLayout";
import { ProjectFormModal } from "../FormBarang/components/ProjectFormModal";
import {
  buildProjectRequest,
  createProject,
  deleteProject,
  fetchListOfProjects,
  updateProject,
  type ListOfProjectsResponse,
} from "../api/InventoryAPI";
import type { ColumnDef, SortDirection, Project } from "../types/dashboard";
import type { ProjectFormData } from "../FormBarang/components/ProjectFormModal";
import SearchIcon from "@mui/icons-material/Search";

const COLUMNS: ColumnDef[] = [
  {
    label: "ID",
    key: "_id",
    width: 220,
    copy: true,
  },
  {
    label: "Nama Project",
    key: "namaProyek",
    width: 310,
    copy: true,
  },
  {
    label: "Lokasi",
    key: "lokasi",
    width: 300,
  },
  {
    label: "Koordinat",
    key: "koordinat",
    width: 200,
    sortable: false,
  },
  {
    label: "Aksi",
    key: "actions",
    width: 120,
    sortable: false,
  },
];

const EMPTY_PROJECT_RESPONSE: ListOfProjectsResponse = {
  success: true,
  data: [],
};

const ProjectPage: React.FC = () => {
  const [data, setData] = useState<ListOfProjectsResponse>(
    EMPTY_PROJECT_RESPONSE,
  );
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sortKey, setSortKey] = useState<string | undefined>("namaProyek");
  const [sortDirection, setSortDirection] = useState<SortDirection>("asc");
  const [search, setSearch] = useState("");
  const [reloadToken, setReloadToken] = useState(0);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedProject, setSelectedProject] =
    useState<ProjectFormData | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    setLoading(true);
    setError(null);

    fetchListOfProjects(controller.signal)
      .then((response) => setData(response))
      .catch((err) => {
        if (err?.name !== "AbortError") {
          setError("Gagal memuat list project. Silakan coba lagi.");
        }
      })
      .finally(() => setLoading(false));

    return () => controller.abort();
  }, [reloadToken]);

  const filteredRows = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();

    if (!normalizedSearch) {
      return [...data.data];
    }

    return data.data.filter((project) => {
      const id = project._id.toLowerCase();
      const name = project.namaProyek.toLowerCase();
      const location = project.lokasi.toLowerCase();
      const coordText = project.koordinat
        ? `${project.koordinat.lat}, ${project.koordinat.lng}`
        : "";

      return (
        id.includes(normalizedSearch) ||
        name.includes(normalizedSearch) ||
        location.includes(normalizedSearch) ||
        coordText.includes(normalizedSearch)
      );
    });
  }, [data.data, search]);

  const sortedRows = useMemo(() => {
    const rows = [...filteredRows];

    if (!sortKey) {
      return rows;
    }

    return rows.sort((left, right) => {
      const leftText =
        sortKey === "koordinat"
          ? left.koordinat
            ? `${left.koordinat.lat}, ${left.koordinat.lng}`
            : ""
          : String(left[sortKey as keyof Project] ?? "").toLowerCase();

      const rightText =
        sortKey === "koordinat"
          ? right.koordinat
            ? `${right.koordinat.lat}, ${right.koordinat.lng}`
            : ""
          : String(right[sortKey as keyof Project] ?? "").toLowerCase();

      if (leftText === rightText) {
        return 0;
      }

      const result = leftText > rightText ? 1 : -1;
      return sortDirection === "asc" ? result : -result;
    });
  }, [filteredRows, sortDirection, sortKey]);

  const handleSortChange = (column: ColumnDef) => {
    if (column.sortable === false) {
      return;
    }

    if (sortKey === column.key) {
      setSortDirection((current) => (current === "asc" ? "desc" : "asc"));
      return;
    }

    setSortKey(column.key);
    setSortDirection("asc");
  };

  const handleAddClick = () => {
    setSelectedProject(null);
    setIsModalOpen(true);
  };

  const handleEditClick = (id: string) => {
    const targetProject = data.data.find((project) => project._id === id);

    if (!targetProject) {
      return;
    }

    setSelectedProject({
      id: targetProject._id,
      name: targetProject.namaProyek,
      location: targetProject.lokasi,
      lat: targetProject.koordinat ? String(targetProject.koordinat.lat) : "",
      lng: targetProject.koordinat ? String(targetProject.koordinat.lng) : "",
    });
    setIsModalOpen(true);
  };

  const handleDeleteClick = (id: string) => {
    const targetProject = data.data.find((project) => project._id === id);

    if (!targetProject) {
      return;
    }

    if (!window.confirm(`Hapus project "${targetProject.namaProyek}"?`)) {
      return;
    }

    deleteProject(id)
      .then(() => fetchListOfProjects().then((resp) => setData(resp)))
      .catch((err) => {
        console.error("Failed to delete project:", err);
        alert("Gagal menghapus project. Silakan coba lagi.");
      });
  };

  const handleSubmit = (formData: ProjectFormData) => {
    const request = buildProjectRequest({
      name: formData.name,
      location: formData.location,
      lat: formData.lat ?? "",
      lng: formData.lng ?? "",
    });

    if (formData.id) {
      updateProject(formData.id, request)
        .then(() => fetchListOfProjects().then((resp) => setData(resp)))
        .catch((err) => {
          console.error("Failed to update project:", err);
          alert("Gagal memperbarui project. Silakan coba lagi.");
        });
    } else {
      createProject(request)
        .then(() => fetchListOfProjects().then((resp) => setData(resp)))
        .catch((err) => {
          console.error("Failed to create project:", err);
          alert("Gagal membuat project baru. Silakan coba lagi.");
        });
    }
  };

  const rows = sortedRows.map((project) => ({
    ...project,
    koordinat: project.koordinat
      ? `${project.koordinat.lat}, ${project.koordinat.lng}`
      : "— belum dipetakan",
  }));

  return (
    <MainLayout>
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          width: "100%",
          height: "100%",
          overflow: "auto",
        }}
      >
        <Typography variant="h5" sx={{ fontWeight: 600, mb: 2 }}>
          Project
        </Typography>

        <Box
          sx={{
            display: "flex",
            flexWrap: "wrap",
            gap: 2,
            justifyContent: "space-between",
            alignItems: "center",
            mb: 2,
          }}
        >
          <TextField
            size="small"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari project by ID, nama, atau lokasi"
            sx={{ width: { xs: "100%", sm: 420 } }}
            slotProps={{
              htmlInput: {
                "aria-label": "search-project",
              },
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon fontSize="small" />
                  </InputAdornment>
                ),
              },
            }}
          />

          <Button variant="contained" onClick={handleAddClick}>
            Tambah Project
          </Button>
        </Box>

        {error ? (
          <Alert
            severity="error"
            action={
              <Button
                color="inherit"
                size="small"
                onClick={() => setReloadToken((value) => value + 1)}
              >
                Coba Lagi
              </Button>
            }
            sx={{ mb: 2 }}
          >
            {error}
          </Alert>
        ) : (
          <DataTable
            columns={COLUMNS}
            rows={rows}
            sortKey={sortKey}
            sortDirection={sortDirection}
            onSortChange={handleSortChange}
            loading={loading}
            onEditClick={handleEditClick}
            onDeleteClick={handleDeleteClick}
          />
        )}
      </Box>
      <ProjectFormModal
        open={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleSubmit}
        initialData={selectedProject || undefined}
      />
    </MainLayout>
  );
};

export default ProjectPage;
