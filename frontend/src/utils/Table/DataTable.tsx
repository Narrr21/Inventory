import {
  Box,
  LinearProgress,
  Tooltip,
  Typography,
  IconButton,
} from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import ArrowUpwardIcon from "@mui/icons-material/ArrowUpward";
import ArrowDownwardIcon from "@mui/icons-material/ArrowDownward";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import CheckIcon from "@mui/icons-material/Check";
import type { ColumnDef, SortDirection } from "../../types/dashboard";

type RowWithId = {
  _id: string;
};

interface DataTableProps<T extends RowWithId = RowWithId> {
  columns: ColumnDef[];
  rows: T[];
  sortKey?: string;
  sortDirection: SortDirection;
  onSortChange: (col: ColumnDef) => void;
  loading?: boolean;
  onEditClick?: (id: string) => void;
  onDeleteClick?: (id: string) => void;
}

const DEFAULT_COLUMN_WIDTH = 160;
const TOOLTIP_THRESHOLD = 24;

const DataTable = <T extends RowWithId>({
  columns,
  rows,
  sortKey,
  sortDirection,
  onSortChange,
  loading = false,
  onEditClick,
  onDeleteClick,
}: DataTableProps<T>) => {
  const hasRowActions = Boolean(onEditClick || onDeleteClick);

  const handleCopy = async (text: string) => {
    if (!text) return;

    try {
      await navigator.clipboard.writeText(text);
    } catch {
      // noop: clipboard failures should not block the table UI
    }
  };

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
        {/* AttributeRow / Table Header */}
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

        {/* Table Body / Rows */}
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            opacity: loading ? 0.6 : 1,
            transition: "opacity 150ms",
          }}
        >
          {rows.map((row) => (
            <Box
              key={row._id}
              role="row"
              sx={{
                display: "flex",
                borderBottom: "1px solid",
                borderColor: "divider",
                cursor: onEditClick ? "pointer" : "default",
                "&:last-of-type": { borderBottom: "none" },
                ...(onEditClick
                  ? { "&:hover": { bgcolor: "action.hover" } }
                  : {}),
              }}
              onClick={onEditClick ? () => onEditClick(row._id) : undefined}
            >
              {columns.map((col) => {
                if (col.key === "actions" && hasRowActions) {
                  return (
                    <Box
                      key={col.key}
                      role="cell"
                      onClick={(e) => e.stopPropagation()}
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
                      <IconButton
                        size="small"
                        aria-label="edit item"
                        onClick={(e) => {
                          e.stopPropagation();
                          e.currentTarget.blur();
                          onEditClick?.(row._id);
                        }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton
                        size="small"
                        aria-label="delete item"
                        onClick={(e) => {
                          e.stopPropagation();
                          e.currentTarget.blur();
                          onDeleteClick?.(row._id);
                        }}
                        sx={{
                          color: "error.main",
                          transition: "all 0.2s ease-in-out",
                          "&:hover": {
                            bgcolor: "error.light",
                            color: "error.contrastText",
                          },
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Box>
                  );
                }

                if (col.key === "actions") {
                  return null;
                }

                const value = (row as Record<string, unknown>)[col.key];
                const text =
                  value === undefined || value === null ? "" : String(value);
                const canCopy = Boolean(col.copy);

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
                      gap: 0.5,
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
                        sx={{
                          maxWidth: canCopy ? "calc(100% - 24px)" : "100%",
                        }}
                      >
                        {text}
                      </Typography>
                    </Tooltip>

                    {canCopy && text && (
                      <Tooltip title="Salin">
                        <IconButton
                          size="small"
                          aria-label={`copy ${col.key}`}
                          onClick={(e) => {
                            e.stopPropagation();
                            void handleCopy(text);
                          }}
                          sx={{
                            width: 20,
                            height: 20,
                            color: "text.secondary",
                            "&:hover": {
                              bgcolor: "action.hover",
                            },
                          }}
                        >
                          <ContentCopyIcon sx={{ fontSize: 14 }} />
                        </IconButton>
                      </Tooltip>
                    )}
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
