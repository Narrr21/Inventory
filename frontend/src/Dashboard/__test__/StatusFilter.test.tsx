import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import StatusFilter, { type StatusOption } from "../components/StatusFilter";
import { describe, expect, it, vi } from "vitest";

describe("StatusFilter Component", () => {
  const mockOptions: StatusOption[] = [
    { value: "active", label: "Aktif" },
    { value: "inactive", label: "Non-Aktif" },
  ];

  it("harus merender label dan nilai default dengan benar", () => {
    render(
      <StatusFilter
        value=""
        options={mockOptions}
        onChange={() => {}}
        label="Filter Status"
      />,
    );

    // Pastikan label "Filter Status" muncul di layar
    expect(screen.getByLabelText("Filter Status")).toBeInTheDocument();

    // Pastikan nilai default adalah string kosong
    const muiSelectInput = screen.getByRole("combobox")
      .nextElementSibling as HTMLInputElement;
    expect(muiSelectInput.value).toBe("");
  });

  it("harus menampilkan semua opsi ketika dropdown diklik", async () => {
    const user = userEvent.setup();
    render(<StatusFilter value="" options={mockOptions} onChange={() => {}} />);

    // Ambil elemen select MUI (combobox)
    const selectTrigger = screen.getByRole("combobox");

    // Klik untuk membuka dropdown
    await act(async () => {
      await user.click(selectTrigger);
    });

    // Pastikan semua opsi muncul di layar
    expect(screen.getByRole("option", { name: "Semua" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Aktif" })).toBeInTheDocument();
    expect(
      screen.getByRole("option", { name: "Non-Aktif" }),
    ).toBeInTheDocument();
  });

  it("harus memanggil fungsi onChange ketika opsi dipilih", async () => {
    const user = userEvent.setup();
    const mockOnChange = vi.fn();

    render(
      <StatusFilter value="" options={mockOptions} onChange={mockOnChange} />,
    );

    // Ambil elemen select MUI (combobox)
    const selectTrigger = screen.getByRole("combobox");

    // Klik untuk membuka dropdown
    await act(async () => {
      await user.click(selectTrigger);
    });

    // Ambil opsi "Aktif"
    const optionActive = screen.getByRole("option", { name: "Aktif" });

    // Klik opsi "Aktif"
    await act(async () => {
      await user.click(optionActive);
    });

    // Pastikan fungsi onChange dipanggil sekali
    expect(mockOnChange).toHaveBeenCalledTimes(1);

    // Pastikan fungsi onChange dipanggil dengan nilai yang benar
    expect(mockOnChange).toHaveBeenCalledWith("active");
  });

  it("harus menampilkan teks opsi yang dipilih di layar setelah status berubah", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <StatusFilter value="" options={mockOptions} onChange={() => {}} />,
    );

    // Ambil elemen select MUI (combobox)
    const selectTrigger = screen.getByRole("combobox");

    // Klik untuk membuka dropdown
    await act(async () => {
      await user.click(selectTrigger);
    });

    // Ambil opsi "Aktif"
    const optionActive = screen.getByRole("option", { name: "Aktif" });

    // Klik opsi "Aktif"
    await act(async () => {
      await user.click(optionActive);
    });

    // Rerender komponen dengan nilai baru
    await act(async () => {
      rerender(
        <StatusFilter
          value="active"
          options={mockOptions}
          onChange={() => {}}
        />,
      );
    });

    // Pastikan teks opsi yang dipilih muncul di layar
    expect(screen.getByText("Aktif")).toBeInTheDocument();
  });
});
