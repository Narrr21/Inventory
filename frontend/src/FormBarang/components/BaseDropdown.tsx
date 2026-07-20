import React from "react";
import { FormControl, InputLabel, Select, MenuItem, FormHelperText, type SelectChangeEvent } from "@mui/material";
import type { FilterOption } from "../../types/dashboard";

interface BaseDropdownProps {
  label: string;
  value: string;
  options: FilterOption[];
  onChange: (value: string) => void;
  error?: boolean;
  helperText?: string;
  fullWidth?: boolean;
}

export const BaseDropdown: React.FC<BaseDropdownProps> = ({
  label,
  value,
  options,
  onChange,
  error = false,
  helperText = "",
  fullWidth = true,
}) => {
  const handleChange = (e: SelectChangeEvent) => {
    onChange(e.target.value);
  };

  const labelId = `base-dropdown-${label.toLowerCase().replace(/\s+/g, "-")}`;

  return (
    <FormControl size="small" fullWidth={fullWidth} error={error}>
      <InputLabel id={labelId}>{label}</InputLabel>
      <Select
        labelId={labelId}
        value={value}
        label={label}
        onChange={handleChange}
      >
        {options.map((opt) => (
          <MenuItem key={opt.value} value={opt.value}>
            {opt.label}
          </MenuItem>
        ))}
      </Select>
      {helperText && <FormHelperText>{helperText}</FormHelperText>}
    </FormControl>
  );
};