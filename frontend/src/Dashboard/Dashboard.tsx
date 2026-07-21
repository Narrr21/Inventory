import React, { useEffect, useRef, useState } from "react";
import { Alert, Box, Button, Pagination, Typography } from "@mui/material";
import ControlRow from "./components/ControlRow";
import DataTable from "./components/DataTable";
import {
  fetchFilterOptions,
  fetchInventory,
  type FilterOptionsResponse,
  type InventoryResponse,
} from "../api/InventoryAPI";
import { createItem } from "../api/CRUDitems";
import type { ColumnDef, SortDirection } from "../types/dashboard";
import type { ItemFormData } from "../types/form";
import MainLayout from "../layouts/MainLayout";
import { ItemFormModal } from "../FormBarang/components/ItemFormModal";
import {
  mapBackendToItemForm,
  mapItemFormToCreateRequest,
} from "../api/ItemMapper";

const PAGE_SIZE = 10;

const EMPTY_FILTER_OPTIONS: FilterOptionsResponse = {
  success: true,
  data: {
    project: [],
    jenis: [],
    status: [],
  },
};

const COLUMNS: ColumnDef[] = [
  {
    label: "ID",
    key: "_id",
    width: 100,
  },
  {
    label: "Nama Barang",
    key: "nama",
    width: 300,
  },
  {
    label: "Serial Number",
    key: "serialNumber",
    width: 150,
  },
  {
    label: "Project",
    key: "namaProyek",
    width: 150,
  },
  {
    label: "Jenis",
    key: "jenis",
    width: 200,
  },
  {
    label: "Status",
    key: "status",
    width: 150,
  },
  {
    label: "Aksi",
    key: "actions",
    width: 100,
    sortable: false,
  },
];

const Dashboard: React.FC = () => {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [project, setProject] = useState("");
  const [jenis, setJenis] = useState("");
  const [sortKey, setSortKey] = useState<string | undefined>(undefined);
  const [sortDirection, setSortDirection] = useState<SortDirection>("asc");
  const [page, setPage] = useState(1);
  const [reloadToken, setReloadToken] = useState(0);

  const [data, setData] = useState<InventoryResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [filterOptions, setFilterOptions] =
    useState<FilterOptionsResponse>(EMPTY_FILTER_OPTIONS);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState<ItemFormData | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    fetchFilterOptions(controller.signal)
      .then((response) => setFilterOptions(response))
      .catch((err) => {
        // gagal memuat opsi filter bukan error fatal untuk seluruh halaman,
        // dropdown filter akan tampil kosong tapi tabel tetap bisa dipakai
        if (err?.name !== "AbortError") console.error(err);
      });
    return () => controller.abort();
  }, []);

  useEffect(() => {
    setPage(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search, status, project, jenis, sortKey, sortDirection]);

  const requestIdRef = useRef(0);

  useEffect(() => {
    const controller = new AbortController();
    const requestId = ++requestIdRef.current;

    setLoading(true);
    setError(null);

    fetchInventory(
      {
        search,
        status,
        proyek: project,
        jenis: jenis,
        sortBy: sortKey,
        sortOrder: sortDirection,
        page,
        limit: PAGE_SIZE,
      },
      controller.signal,
    )
      .then((res) => {
        if (requestIdRef.current !== requestId) return; // response basi, abaikan
        setData(res);
        setSortKey((prev) => prev ?? COLUMNS[0]?.key);
      })
      .catch((err) => {
        if (err?.name === "AbortError") return;
        if (requestIdRef.current !== requestId) return;
        setError("Gagal memuat data inventory. Silakan coba lagi.");
      })
      .finally(() => {
        if (requestIdRef.current === requestId) setLoading(false);
      });

    return () => controller.abort();
  }, [
    search,
    status,
    project,
    jenis,
    sortKey,
    sortDirection,
    page,
    reloadToken,
  ]);

  const handleSort = (col: ColumnDef) => {
    if (col.sortable === false) return;
    if (sortKey === col.key) {
      setSortDirection((prev) => (prev === "asc" ? "desc" : "asc"));
    } else {
      setSortKey(col.key);
      setSortDirection("asc");
    }
  };

  const handleAddClick = () => {
    setSelectedItem(null);
    setIsModalOpen(true);
  };

  const handleEditClick = (index: string) => {
    const rawBackendData = data?.data.find((item) => item._id === index);
    if (!rawBackendData) return;

    // Transformasi data backend lewat Inventory API & Mapper contoh:
    const formattedData = mapBackendToItemForm(rawBackendData);
    setSelectedItem(formattedData);

    setIsModalOpen(true);
  };

  const handleDeleteClick = (index: string) => {
    if (
      confirm("Are you sure you want to delete item with id " + index + "?")
    ) {
      handleDelete(index);
    }
  };

  const handleDelete = async (index: string) => {
    // PROSES DELETE API
    console.log("Item with ID:", index, "Deleted");
  };

  const handleSubmit = async (formData: ItemFormData) => {
    if (formData.id) {
      const payload = mapItemFormToCreateRequest(formData);
      console.log("Update Item ID:", formData.id, payload);
    } else {
      const payload = mapItemFormToCreateRequest(formData);
      createItem(payload)
        .then((res) => {
          console.log("Item Created:", res);
          setReloadToken((t) => t + 1);
        })
        .catch((err) => {
          console.error("Failed to create item:", err);
          alert("Gagal membuat item baru. Silakan coba lagi.");
        });
    }
    setIsModalOpen(false);
  };

  const rows = data?.data ?? [];
  const total = data?.meta.total ?? 0;
  const pageCount = data?.meta.totalPages ?? 0;

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
          Dashboard Inventory
        </Typography>

        <ControlRow
          onSearch={setSearch}
          status={status}
          statusOptions={filterOptions.data.status}
          onStatusChange={setStatus}
          projectOptions={filterOptions.data.project}
          projectSelected={project}
          onProjectChange={setProject}
          jenisOptions={filterOptions.data.jenis}
          jenisSelected={jenis}
          onJenisChange={setJenis}
          onAddItem={handleAddClick}
        />

        {error ? (
          <Alert
            severity="error"
            action={
              <Button
                color="inherit"
                size="small"
                onClick={() => setReloadToken((t) => t + 1)}
              >
                Coba Lagi
              </Button>
            }
            sx={{ mb: 2 }}
          >
            {error}
          </Alert>
        ) : (
          <>
            <DataTable
              columns={COLUMNS}
              rows={rows}
              sortKey={sortKey}
              sortDirection={sortDirection}
              onSortChange={handleSort}
              loading={loading}
              onEditClick={handleEditClick}
              onDeleteClick={handleDeleteClick}
            />

            {total > 0 && (
              <Box sx={{ display: "flex", justifyContent: "center", mt: 2 }}>
                <Pagination
                  count={pageCount}
                  page={page}
                  onChange={(_, value) => setPage(value)}
                  disabled={loading}
                  color="primary"
                />
              </Box>
            )}
          </>
        )}
      </Box>
      <ItemFormModal
        open={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleSubmit}
        initialData={selectedItem || undefined}
        options={{
          status: filterOptions.data.status,
          proyek: filterOptions.data.project,
          jenis: filterOptions.data.jenis,
        }}
      />
    </MainLayout>
  );
};

export default Dashboard;
