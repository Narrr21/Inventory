import React from "react";
import {
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  type SelectChangeEvent,
} from "@mui/material";

export interface StatusOption {
  value: string;
  label: string;
}

interface StatusFilterProps {
  value: string;
  options: StatusOption[];
  onChange: (value: string) => void;
  label?: string;
}

const StatusFilter: React.FC<StatusFilterProps> = ({
  value,
  options,
  onChange,
  label = "Status",
}) => {
  const handleChange = (e: SelectChangeEvent) => onChange(e.target.value);

  return (
    <FormControl size="small" sx={{ minWidth: 160 }}>
      <InputLabel id="status-filter-label">{label}</InputLabel>
      <Select
        labelId="status-filter-label"
        value={value}
        label={label}
        onChange={handleChange}
        inputProps={{ "aria-label": "status filter" }}
      >
        <MenuItem value="">Semua</MenuItem>
        {options.map((opt) => (
          <MenuItem key={opt.value} value={opt.value}>
            {opt.label}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
};

export default StatusFilter;