import React, { useState, useEffect } from "react";
import { Box, TextField, Typography, IconButton } from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import CheckIcon from "@mui/icons-material/Check";

interface BaseRowProps {
  label: string;
  value: string;
  placeholder?: string;
  onChange: (val: string) => void;
  editableLabel?: boolean;
  onLabelChange?: (newLabel: string) => void;
  onValidate?: (val: string) => string | null;
  required?: boolean;
}

export const BaseRow: React.FC<BaseRowProps> = ({
  label,
  value,
  placeholder = "",
  onChange,
  editableLabel = false,
  onLabelChange,
  onValidate,
  required = false,
}) => {
  const [isEditingLabel, setIsEditingLabel] = useState(false);
  const [tempLabel, setTempLabel] = useState(label);
  const [errorText, setErrorText] = useState<string | null>(null);

  useEffect(() => {
    setTempLabel(label);
  }, [label]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newVal = e.target.value;
    onChange(newVal);

    if (onValidate) {
      const err = onValidate(newVal);
      setErrorText(err);
    }
  };

  const handleSaveLabel = () => {
    setIsEditingLabel(false);
    if (onLabelChange && tempLabel.trim()) {
      onLabelChange(tempLabel.trim());
    } else {
      setTempLabel(label);
    }
  };

  const handleKeyDownLabel = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSaveLabel();
    }
  };

  return (
    <Box
      sx={{
        width: { xs: "100%", lg: "50%" },
        display: "flex",
        flexDirection: "column",
        gap: 0.5,
        p: 0.5,
      }}
    >
      <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
        {isEditingLabel ? (
          <>
            <TextField
              size="small"
              variant="standard"
              value={tempLabel}
              onChange={(e) => setTempLabel(e.target.value)}
              onBlur={handleSaveLabel}
              onKeyDown={handleKeyDownLabel}
              autoFocus
            />
            <IconButton
              size="small"
              onClick={handleSaveLabel}
              onMouseDown={(e) => e.preventDefault()}
            >
              <CheckIcon fontSize="small" />
            </IconButton>
          </>
        ) : (
          <>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {label} {required && "*"}
            </Typography>
            {editableLabel && (
              <IconButton
                size="small"
                onClick={() => setIsEditingLabel(true)}
              >
                <EditIcon fontSize="small" sx={{ fontSize: 14 }} />
              </IconButton>
            )}
          </>
        )}
      </Box>

      <TextField
        size="small"
        fullWidth
        value={value}
        placeholder={placeholder}
        onChange={handleChange}
        error={Boolean(errorText)}
        helperText={errorText}
      />
    </Box>
  );
};