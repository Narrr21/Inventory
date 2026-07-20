import React from "react";
import { Box, LinearProgress, Tooltip, Typography } from "@mui/material";
import ArrowUpwardIcon from "@mui/icons-material/ArrowUpward";
import ArrowDownwardIcon from "@mui/icons-material/ArrowDownward";
import type { ColumnDef, SortDirection } from "../../types/dashboard";
import type { InventoryRow } from "../../api/InventoryAPI";

interface DataTableProps {
  columns: ColumnDef[];
  rows: InventoryRow[];
  sortKey?: string;
  sortDirection: SortDirection;
  onSortChange: (col: ColumnDef) => void;
  loading?: boolean;
}

const DEFAULT_COLUMN_WIDTH = 160;
const TOOLTIP_THRESHOLD = 24;

const DataTable: React.FC<DataTableProps> = ({
  columns,
  rows,
  sortKey,
  sortDirection,
  onSortChange,
  loading = false,
}) => {
  return (
    <Box
      role="region"
      aria-label="data-table"
      sx={{
        display: "flex",
        flexDirection: "column",
        width: "100%",
        overflowX: "auto",
        border: "1px solid",
        borderColor: "divider",
        borderRadius: 2,
        position: "relative",
      }}
    >
      <Box sx={{ height: 2 }}>
        {loading && <LinearProgress sx={{ height: 2 }} />}
      </Box>
      <Box sx={{ minWidth: "max-content", width: "100%" }}>
        {/* AttributeRow */}
        <Box
          role="row"
          sx={{
            display: "flex",
            bgcolor: "background.paper",
            borderBottom: "1px solid",
            borderColor: "divider",
            position: "sticky",
            top: 0,
            zIndex: 1,
          }}
        >
          {columns.map((col) => {
            const isActive = sortKey === col.key;
            const isSortable = col.sortable !== false;
            return (
              <Box
                key={col.key}
                role="columnheader"
                onClick={() => isSortable && onSortChange(col)}
                sx={{
                  width: col.width ?? DEFAULT_COLUMN_WIDTH,
                  flex: `0 0 ${col.width ?? DEFAULT_COLUMN_WIDTH}px`,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  gap: 0.5,
                  py: 1.25,
                  px: 1,
                  cursor: isSortable ? "pointer" : "default",
                  userSelect: "none",
                }}
              >
                <Typography
                  variant="subtitle2"
                  noWrap
                  sx={{ fontWeight: 600, color: "text.primary" }}
                >
                  {col.label}
                </Typography>
                {isSortable && (
                  <Box
                    component="span"
                    aria-label={
                      isActive ? `sorted ${sortDirection}` : "not sorted"
                    }
                    sx={{
                      display: "inline-flex",
                      color: isActive ? "primary.main" : "text.disabled",
                    }}
                  >
                    {isActive && sortDirection === "desc" ? (
                      <ArrowDownwardIcon sx={{ fontSize: 16 }} />
                    ) : (
                      <ArrowUpwardIcon sx={{ fontSize: 16 }} />
                    )}
                  </Box>
                )}
              </Box>
            );
          })}
        </Box>

        {/* TableRow */}
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            opacity: loading ? 0.6 : 1,
            transition: "opacity 150ms",
          }}
        >
          {rows.map((row, idx) => (
            <Box
              key={idx}
              role="row"
              sx={{
                display: "flex",
                borderBottom: "1px solid",
                borderColor: "divider",
                "&:last-of-type": { borderBottom: "none" },
                "&:hover": { bgcolor: "action.hover" },
              }}
            >
              {columns.map((col) => {
                const value = row[col.key];
                const text =
                  value === undefined || value === null ? "" : String(value);
                return (
                  <Box
                    key={col.key}
                    role="cell"
                    sx={{
                      width: col.width ?? DEFAULT_COLUMN_WIDTH,
                      flex: `0 0 ${col.width ?? DEFAULT_COLUMN_WIDTH}px`,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      py: 1,
                      px: 1,
                    }}
                  >
                    <Tooltip
                      title={text}
                      disableHoverListener={text.length <= TOOLTIP_THRESHOLD}
                    >
                      <Typography
                        variant="body2"
                        noWrap
                        align="center"
                        sx={{ maxWidth: "100%" }}
                      >
                        {text}
                      </Typography>
                    </Tooltip>
                  </Box>
                );
              })}
            </Box>
          ))}
          {!loading && rows.length === 0 && (
            <Box sx={{ py: 4, textAlign: "center" }}>
              <Typography variant="body2" color="text.secondary">
                Tidak ada data untuk ditampilkan.
              </Typography>
            </Box>
          )}
        </Box>
      </Box>
    </Box>
  );
};

export default DataTable;
