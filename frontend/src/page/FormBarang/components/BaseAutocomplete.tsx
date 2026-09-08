import React from "react";
import {
  Autocomplete,
  Box,
  IconButton,
  TextField,
  Typography,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import type { Jenis } from "../../../types/dashboard";

interface AddOption {
  inputValue: string;
  label: string;
}

type AutocompleteOption = Jenis | AddOption;

interface BaseAutocompleteProps {
  label: string;
  value: string;
  inputValue: string;
  options: Jenis[];
  onValueChange: (value: string) => void;
  onInputChange: (value: string) => void;
  onCreateOption?: (value: string) => void | Promise<void>;
  onDeleteOption?: (value: string) => void | Promise<void>;
  placeholder?: string;
  helperText?: string;
  loading?: boolean;
}

export const BaseAutocomplete: React.FC<BaseAutocompleteProps> = ({
  label,
  value,
  inputValue,
  options,
  onValueChange,
  onInputChange,
  onCreateOption,
  onDeleteOption,
  placeholder,
  helperText,
  loading = false,
}) => {
  const normalizedInput = inputValue.trim().toLowerCase();
  const hasExactMatch = options.some(
    (option) => option.jenis.toLowerCase() === normalizedInput,
  );

  const filteredOptions: AutocompleteOption[] = options
    .filter((option) => option.jenis.toLowerCase().includes(normalizedInput))
    .slice();

  if (inputValue.trim() && !hasExactMatch) {
    filteredOptions.push({
      inputValue: inputValue.trim(),
      label: `Tambahkan ke database: ${inputValue.trim()}`,
    });
  }

  return (
    <Autocomplete<AutocompleteOption, false, false, true>
      freeSolo
      size="small"
      loading={loading}
      options={filteredOptions}
      value={value}
      inputValue={inputValue}
      onChange={(_, newValue) => {
        if (!newValue) {
          onValueChange("");
          onInputChange("");
          return;
        }

        if (typeof newValue === "string") {
          onValueChange(newValue);
          onInputChange(newValue);
          return;
        }

        // newValue is object: either AddOption or Jenis
        if ("inputValue" in newValue) {
          const v = newValue.inputValue;
          if (onCreateOption) void onCreateOption(v);
          onValueChange(v);
          onInputChange(v);
          return;
        }

        // Jenis
        onValueChange(newValue.jenis);
        onInputChange(newValue.jenis);
      }}
      onInputChange={(_, newInputValue, reason) => {
        if (reason === "reset") {
          return;
        }

        onInputChange(newInputValue);
        onValueChange(newInputValue);
      }}
      getOptionLabel={(option) => {
        if (typeof option === "string") return option;
        if ("inputValue" in option) return option.inputValue;
        return option.jenis;
      }}
      isOptionEqualToValue={(option, selectedValue) => {
        const toStringValue = (
          v: AutocompleteOption | string | null | undefined,
        ) => {
          if (!v) return "";
          if (typeof v === "string") return v;
          if ("inputValue" in v) return v.inputValue;
          return v.jenis;
        };

        return toStringValue(option) === toStringValue(selectedValue as any);
      }}
      renderOption={(props, option) => {
        const label =
          typeof option === "string"
            ? option
            : "inputValue" in option
              ? option.label
              : option.jenis;

        const { key, ...restProps } = props as any;

        return (
          <li key={key} {...restProps}>
            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                width: "100%",
                gap: 1,
              }}
            >
              <Typography variant="body2" noWrap>
                {label}
              </Typography>
              {/* show delete only for existing Jenis items */}
              {!("inputValue" in option) && onDeleteOption && (
                <IconButton
                  size="small"
                  edge="end"
                  aria-label={`hapus-${"jenis" in option ? option.jenis : String(option)}`}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    if ("_id" in option) void onDeleteOption(option._id);
                  }}
                >
                  <CloseIcon fontSize="small" />
                </IconButton>
              )}
            </Box>
          </li>
        );
      }}
      renderInput={(params) => (
        <TextField
          {...params}
          label={label}
          placeholder={placeholder}
          helperText={helperText}
        />
      )}
    />
  );
};
