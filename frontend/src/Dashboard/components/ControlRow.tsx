import React from "react";
import { Box, Button, Typography } from "@mui/material";
import SearchBar from "./SearchBar";
import StatusFilter, { type StatusOption } from "./StatusFilter";
import SearchableFilter from "./SearchableFilter";
import type { FilterOption } from "../../types/dashboard";

interface ControlRowProps {
  onSearch: (query: string) => void;
  status: string;
  statusOptions: StatusOption[];
  onStatusChange: (value: string) => void;
  projectOptions: FilterOption[];
  projectSelected: string[];
  onProjectChange: (values: string[]) => void;
  jenisBarangOptions: FilterOption[];
  jenisBarangSelected: string[];
  onJenisBarangChange: (values: string[]) => void;
  onAddItem: () => void;
}

const ControlRow: React.FC<ControlRowProps> = (props) => {
  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: { xs: "column", md: "row" },
        justifyContent: { md: "space-between" },
        alignItems: { xs: "stretch", md: "center" },
        gap: 1.5,
        mb: 2,
      }}
    >
      <Box sx={{ order: 0 }}>
        <SearchBar onSearch={props.onSearch} />
      </Box>

      <Box sx={{ order: 1, display: "flex", flexWrap: "wrap", gap: 1.5 }}>
        <Button
          variant="outlined"
          onClick={props.onAddItem}
        >
          <Typography variant="button" sx={{ fontWeight: 600 }}>
            Tambah Item
          </Typography>
        </Button>
        <StatusFilter
          value={props.status}
          options={props.statusOptions}
          onChange={props.onStatusChange}
        />
        <SearchableFilter
          label="Project"
          options={props.projectOptions}
          selected={props.projectSelected}
          onChange={props.onProjectChange}
        />
        <SearchableFilter
          label="Jenis Barang"
          options={props.jenisBarangOptions}
          selected={props.jenisBarangSelected}
          onChange={props.onJenisBarangChange}
        />
      </Box>
    </Box>
  );
};

export default ControlRow;
