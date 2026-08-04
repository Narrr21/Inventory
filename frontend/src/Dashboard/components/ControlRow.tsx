import React from "react";
import { Box, IconButton } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import SearchBar from "./SearchBar";
import StatusFilter from "./StatusFilter";
import SearchableFilter from "./SearchableFilter";
import type { Jenis, Project } from "../../types/dashboard"

interface ControlRowProps {
  onSearch: (query: string) => void;
  status: string;
  statusOptions: string[];
  onStatusChange: (value: string) => void;
  projectOptions: Project[];
  projectSelected: string;
  onProjectChange: (value: string) => void;
  jenisOptions: Jenis[];
  jenisSelected: string;
  onJenisChange: (value: string) => void;
  onAddItem: () => void;
}

const projectNames = (projects: Project[]): string[] => {
  return projects.map((project) => project.namaProyek);
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
        <IconButton
          onClick={props.onAddItem}
          sx={{
            border: "1px solid",
            borderColor: "primary.main",
            color: "primary.main",
            transition: "all 0.2s ease-in-out",
            "&:hover": {
              bgcolor: "primary.light",
              color: "primary.contrastText",
            },
          }}
        >
          <AddIcon />
        </IconButton>
        <StatusFilter
          value={props.status}
          options={props.statusOptions}
          onChange={props.onStatusChange}
        />
        <SearchableFilter
          label="Project"
          options={projectNames(props.projectOptions)}
          selected={props.projectSelected}
          onChange={props.onProjectChange}
        />
        <SearchableFilter
          label="Jenis Barang"
          options={props.jenisOptions.map((j) => j.jenis)}
          selected={props.jenisSelected}
          onChange={props.onJenisChange}
        />
      </Box>
    </Box>
  );
};

export default ControlRow;
