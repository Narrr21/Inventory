import React from "react";
import { Box, Typography, Button } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import { BaseRow } from "./BaseRow";
import type { FieldInfo } from "../../../types/form";

interface BaseSectionProps {
  label: string;
  fields: FieldInfo[];
  onChangeField: (id: string, updatedField: Partial<FieldInfo>) => void;
  onAddField: () => void;
}

export const BaseSection: React.FC<BaseSectionProps> = ({
  label,
  fields,
  onChangeField,
  onAddField,
}) => {
  return (
    <Box
      sx={{
        width: "100%",
        display: "flex",
        flexDirection: "column",
        gap: 1.5,
        mt: 1,
        mb: 1,
      }}
    >
      <Typography
        variant="subtitle1"
        sx={{
          fontWeight: 700,
          borderBottom: "1px solid",
          borderColor: "divider",
          pb: 0.5,
        }}
      >
        {label}
      </Typography>

      <Box sx={{ display: "flex", flexWrap: "wrap", width: "100%" }}>
        {fields.map((field) => (
          <BaseRow
            key={field.id}
            label={field.label}
            value={field.value}
            placeholder={field.placeholder}
            required={!field.nullable}
            editableLabel={true}
            onLabelChange={(newLabel) =>
              onChangeField(field.id, { label: newLabel })
            }
            onChange={(val) => onChangeField(field.id, { value: val })}
          />
        ))}
      </Box>

      {/* Dotted border button to add new row */}
      <Button
        variant="outlined"
        startIcon={<AddIcon />}
        onClick={onAddField}
        sx={{
          borderStyle: "dashed",
          borderWidth: "1.5px",
          width: "100%",
          mt: 1,
          justifyContent: "center",
          color: "text.secondary",
          borderColor: "text.disabled",
          "&:hover": {
            borderStyle: "dashed",
            borderWidth: "1.5px",
          },
        }}
      >
        Tambah Baris Baru
      </Button>
    </Box>
  );
};
