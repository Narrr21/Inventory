import React from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  IconButton,
  Box,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";

interface BaseFormProps {
  open: boolean;
  title: string;
  onClose: () => void;
  onSubmit: (e: React.FormEvent) => void;
  children: React.ReactNode;
  submitLabel?: string;
}

export const BaseForm: React.FC<BaseFormProps> = ({
  open,
  title,
  onClose,
  onSubmit,
  children,
  submitLabel = "Simpan",
}) => {
  return (
    <Dialog
      open={open}
      onClose={onClose}
      fullWidth
      maxWidth="md"
      slotProps={{
        paper: {
          sx: {
            height: "80vh",
            maxHeight: "700px",
            display: "flex",
            flexDirection: "column",
            p: "4px",
            backgroundColor: "background.paper"
          },
        },
      }}
    >
      <DialogTitle
        sx={{
          m: 0,
          p: 2,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        {title}
        <IconButton onClick={onClose} size="small">
          <CloseIcon />
        </IconButton>
      </DialogTitle>

      <DialogContent
        dividers
        sx={{
          flexGrow: 1,
          overflowY: "auto",
          p: 2,
        }}
      >
        <form id="base-form" onSubmit={onSubmit}>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {children}
          </Box>
        </form>
      </DialogContent>

      <DialogActions sx={{ p: 2 }}>
        <Button onClick={onClose} color="inherit">
          Batal
        </Button>
        <Button
          type="submit"
          form="base-form"
          variant="contained"
          color="primary"
        >
          {submitLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
};