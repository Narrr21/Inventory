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
import type {
  ColumnDef,
  FilterOption,
  SortDirection,
} from "../types/dashboard";
import type { ItemFormData } from "../types/form";
import MainLayout from "../layouts/MainLayout";
import { ItemFormModal } from "../FormBarang/components/ItemFormModal";
import {
  mapBackendToItemForm,
  mapItemFormToBackend,
} from "../api/ItemMapper";

const PAGE_SIZE = 10;

const EMPTY_FILTER_OPTIONS: FilterOptionsResponse = {
  project: [],
  jenis: [],
};

const COLUMNS: ColumnDef[] = [
  {
    label: "ID",
    key: "id",
    width: 10,
  },
  {
    label: "Nama Barang",
    key: "name",
    width: 30,
  },
  {
    label: "Project",
    key: "proyek",
    width: 20,
  },
  {
    label: "Jenis",
    key: "jenis",
    width: 15,
  },
  {
    label: "Status",
    key: "status",
    width: 15,
  },
  {
    label: "Aksi",
    key: "actions",
    width: 10,
    sortable: false,
  }
];

const STATUS_OPTIONS: FilterOption[] = [
  { label: "Healthy", value: "Healthy" },
  { label: "Under Maintenance", value: "Under Maintenance" },
  { label: "Broken", value: "Broken" },
];

const Dashboard: React.FC = () => {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [projectSelected, setProjectSelected] = useState<string[]>([]);
  const [jenisBarangSelected, setJenisBarangSelected] = useState<string[]>([]);
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
      .then(setFilterOptions)
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
  }, [
    search,
    status,
    projectSelected,
    jenisBarangSelected,
    sortKey,
    sortDirection,
  ]);

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
        project: projectSelected,
        jenis: jenisBarangSelected,
        sortKey,
        sortDirection,
        page,
        pageSize: PAGE_SIZE,
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
    projectSelected,
    jenisBarangSelected,
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
    const rawBackendData = data?.rows.find((item) => item.id === index);
    if (!rawBackendData) return;

    // Transformasi data backend lewat Inventory API & Mapper contoh:
    const formattedData = mapBackendToItemForm(rawBackendData);
    setSelectedItem(formattedData);

    setIsModalOpen(true);
  };

  const handleDeleteClick = (index: string) => {
    if (confirm("Are you sure you want to delete item with id " + index + "?")) {
      handleDelete(index);
    }
  };

  const handleDelete = async (index: string) => {

    // PROSES DELETE API
    console.log("Item with ID:", index, "Deleted")
  }

  const handleSubmit = async (formData: ItemFormData) => {
    const payload = mapItemFormToBackend(formData);

    if (formData.id) {
      // PROSES EDIT (PUT / PATCH API)
      console.log("Update Item ID:", formData.id, payload);
      // await updateItemApi(formData.id, payload);
    } else {
      // PROSES CREATE (POST API)
      console.log("Create Item Baru:", payload);
      // await createItemApi(payload);
    }

    setIsModalOpen(false);
  };

  const rows = data?.rows ?? [];
  const total = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

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
          statusOptions={STATUS_OPTIONS}
          onStatusChange={setStatus}
          projectOptions={filterOptions.project}
          projectSelected={projectSelected}
          onProjectChange={setProjectSelected}
          jenisOptions={filterOptions.jenis}
          jenisSelected={jenisBarangSelected}
          onJenisChange={setJenisBarangSelected}
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
          status: STATUS_OPTIONS,
          proyek: filterOptions.project,
          jenis: filterOptions.jenis,
        }}
      />
    </MainLayout>
  );
};

export default Dashboard;
