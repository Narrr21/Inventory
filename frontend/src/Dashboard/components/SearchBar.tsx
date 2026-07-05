import React, { useEffect, useState } from "react";
import { InputAdornment, TextField } from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import { useDebounce } from "../hooks/useDebounce";

interface SearchBarProps {
  placeholder?: string;
  onSearch: (query: string) => void;
  debounceMs?: number;
}

const SearchBar: React.FC<SearchBarProps> = ({
  placeholder = "Cari...",
  onSearch,
  debounceMs = 400,
}) => {
  const [value, setValue] = useState("");
  const debouncedValue = useDebounce(value, debounceMs);

  useEffect(() => {
    onSearch(debouncedValue.trim());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedValue]);

  return (
    <TextField
      size="small"
      value={value}
      onChange={(e) => setValue(e.target.value)}
      placeholder={placeholder}
      sx={{
        minWidth: 240,
        bgcolor: value ? "background.paper" : "background.default",
      }}
      slotProps={{
        htmlInput: {
          "aria-label": "search",
        },
        input: {
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon fontSize="small" />
            </InputAdornment>
          ),
        },
      }}
    />
  );
};

export default SearchBar;
