import React, { useMemo, useState } from "react";
import {
  Box,
  Button,
  Checkbox,
  ListItemText,
  Menu,
  MenuItem,
  TextField,
  Typography,
} from "@mui/material";
import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";
import type { FilterOption } from "../../types/dashboard";

interface SearchableFilterProps {
  label: string;
  options: FilterOption[];
  selected: string[];
  onChange: (values: string[]) => void;
  maxVisible?: number;
}

const SearchableFilter: React.FC<SearchableFilterProps> = ({
  label,
  options,
  selected,
  onChange,
  maxVisible = 5,
}) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [query, setQuery] = useState("");
  const open = Boolean(anchorEl);

  const filtered = useMemo(
    () =>
      options.filter((opt) =>
        opt.label.toLowerCase().includes(query.toLowerCase()),
      ),
    [options, query],
  );

  const visible = filtered.slice(0, maxVisible);

  const closeMenu = () => {
    setAnchorEl(null);
    setQuery("");
  };

  const toggleValue = (value: string) => {
    const next = selected.includes(value)
      ? selected.filter((v) => v !== value)
      : [...selected, value];
    onChange(next);
  };

  return (
    <Box>
      <Button
        variant="outlined"
        color="inherit"
        endIcon={<ArrowDropDownIcon />}
        onClick={(e) => setAnchorEl(e.currentTarget)}
        sx={{ minWidth: 160, justifyContent: "space-between" }}
      >
        {label}
        {selected.length > 0 ? ` (${selected.length})` : ""}
      </Button>
      <Menu anchorEl={anchorEl} open={open} onClose={closeMenu}>
        <Box sx={{ px: 1.5, py: 1 }}>
          <TextField
            size="small"
            autoFocus
            fullWidth
            placeholder={`Cari ${label.toLowerCase()}...`}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.stopPropagation()}
            slotProps={{
              input: { "aria-label": `search ${label}` },
            }}
          />
        </Box>
        {/* max 5 opsi terlihat sekaligus, sisanya di-scroll */}
        <Box sx={{ maxHeight: maxVisible * 40, overflowY: "auto" }}>
          {visible.length === 0 ? (
            <MenuItem disabled>Tidak ada hasil</MenuItem>
          ) : (
            visible.map((opt) => (
              <MenuItem
                key={opt.value}
                onClick={() => toggleValue(opt.value)}
                dense
              >
                <Checkbox
                  size="small"
                  checked={selected.includes(opt.value)}
                  sx={{ mr: 1 }}
                />
                <ListItemText primary={opt.label} />
              </MenuItem>
            ))
          )}
        </Box>
        {filtered.length > maxVisible && (
          <Typography
            variant="caption"
            sx={{ display: "block", px: 1.5, py: 0.5, color: "text.secondary" }}
          >
            +{filtered.length - maxVisible} lainnya, persempit pencarian
          </Typography>
        )}
      </Menu>
    </Box>
  );
};

export default SearchableFilter;
