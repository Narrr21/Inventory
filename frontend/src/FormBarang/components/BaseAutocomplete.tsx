import React from "react";
import {
  Autocomplete,
  Box,
  IconButton,
  TextField,
  Typography,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";

interface AddOption {
  inputValue: string;
  label: string;
}

type AutocompleteOption = string | AddOption;

interface BaseAutocompleteProps {
  label: string;
  value: string;
  inputValue: string;
  options: string[];
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
    (option) => option.toLowerCase() === normalizedInput,
  );

  const filteredOptions: AutocompleteOption[] = options.filter((option) =>
    option.toLowerCase().includes(normalizedInput),
  );

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
        if (typeof newValue === "string") {
          onValueChange(newValue);
          onInputChange(newValue);
          return;
        }

        if (newValue && typeof newValue !== "string") {
          if (onCreateOption) {
            void onCreateOption(newValue.inputValue);
          }
          onValueChange(newValue.inputValue);
          onInputChange(newValue.inputValue);
          return;
        }

        onValueChange("");
        onInputChange("");
      }}
      onInputChange={(_, newInputValue, reason) => {
        if (reason === "reset") {
          return;
        }

        onInputChange(newInputValue);
        onValueChange(newInputValue);
      }}
      getOptionLabel={(option) => {
        if (typeof option === "string") {
          return option;
        }

        return option.inputValue;
      }}
      isOptionEqualToValue={(option, selectedValue) => {
        if (typeof option === "string") {
          return option === selectedValue;
        }

        return option.inputValue === selectedValue;
      }}
      renderOption={(props, option) => {
        if (typeof option === "string") {
          return (
            <li {...props}>
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
                  {option}
                </Typography>
                {onDeleteOption && (
                  <IconButton
                    size="small"
                    edge="end"
                    aria-label={`hapus-${option}`}
                    onMouseDown={(event) => event.preventDefault()}
                    onClick={(event) => {
                      event.preventDefault();
                      event.stopPropagation();
                      void onDeleteOption(option);
                    }}
                  >
                    <CloseIcon fontSize="small" />
                  </IconButton>
                )}
              </Box>
            </li>
          );
        }

        return (
          <li {...props}>
            <Typography variant="body2">
              {option.label}
            </Typography>
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
