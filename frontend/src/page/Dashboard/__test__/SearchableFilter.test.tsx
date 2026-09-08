import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import SearchableFilter from "../../../utils/Table/SearchableFilter";
import { describe, expect, it, vi } from "vitest";

describe("SearchableFilter Component", () => {
  const mockOptions: string[] = [
    "Kategori A",
    "Kategori B",
    "Kategori C",
    "Kategori D",
    "Kategori E",
    "Kategori F", // Opsi tambahan untuk menguji maxVisible
  ];

  it("harus merender tombol filter dengan label saat belum ada yang dipilih", () => {
    // Render komponen dengan kondisi awal selected kosong
    const { rerender } = render(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected=""
        onChange={() => {}}
      />,
    );

    // Pastikan tombol filter muncul dengan label yang benar
    expect(
      screen.getByRole("button", { name: /kategori/i }),
    ).toBeInTheDocument();

    // Rerender komponen dengan selected yang berbeda
    rerender(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected="Kategori A"
        onChange={() => {}}
      />,
    );

    // Pastikan tombol filter menampilkan opsi yang dipilih
    expect(
      screen.getByRole("button", { name: "Kategori A" }),
    ).toBeInTheDocument();
  });

  it("harus membuka menu dan merender opsi sampai batas maxVisible saat tombol diklik", async () => {
    const user = userEvent.setup();

    // Render komponen dengan maxVisible = 5
    render(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected=""
        onChange={() => {}}
        maxVisible={5}
      />,
    );

    // Ambil tombol filter berdasarkan label
    const button = screen.getByRole("button", { name: /kategori/i });

    // Klik tombol untuk membuka menu
    await act(async () => {
      await user.click(button);
    });

    // Pastikan input pencarian muncul di menu
    expect(
      screen.getByPlaceholderText(/cari kategori.../i),
    ).toBeInTheDocument();

    // Pastikan opsi yang terlihat sesuai dengan maxVisible
    expect(
      screen.getByRole("menuitem", { name: "Kategori A" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("menuitem", { name: "Kategori E" }),
    ).toBeInTheDocument();

    // Pastikan opsi yang melebihi maxVisible tidak muncul di menu
    expect(
      screen.queryByRole("menuitem", { name: "Kategori F" }),
    ).not.toBeInTheDocument();

    // Pastikan teks "+1 lainnya, persempit pencarian" muncul di menu karena ada opsi yang tersembunyi
    expect(
      screen.getByText(/\+1 lainnya, persempit pencarian/i),
    ).toBeInTheDocument();
  });

  it("harus menyaring opsi berdasarkan query pencarian", async () => {
    const user = userEvent.setup();

    // Render komponen dengan opsi yang sudah ditentukan
    render(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected=""
        onChange={() => {}}
      />,
    );

    // Ambil tombol filter berdasarkan label
    const button = screen.getByRole("button", { name: /kategori/i });

    // Klik tombol filter untuk membuka menu
    await act(async () => {
      await user.click(button);
    });

    // Ambil input pencarian di menu
    const searchInput = screen.getByPlaceholderText(/cari kategori.../i);

    // Ketik query pencarian "Kategori A" di input pencarian
    await act(async () => {
      await user.type(searchInput, "Kategori A");
    });

    // Pastikan opsi tetap muncul jika sesuai dengan query
    expect(
      screen.getByRole("menuitem", { name: "Kategori A" }),
    ).toBeInTheDocument();

    // Pastikan opsi yang tidak sesuai dengan query tidak muncul di menu
    expect(
      screen.queryByRole("menuitem", { name: "Kategori B" }),
    ).not.toBeInTheDocument();
  });

  it("harus menampilkan teks 'Tidak ada hasil' jika query tidak cocok dengan opsi manapun", async () => {
    const user = userEvent.setup();
    render(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected=""
        onChange={() => {}}
      />,
    );

    const button = screen.getByRole("button", { name: /kategori/i });

    await act(async () => {
      await user.click(button);
    });

    const searchInput = screen.getByPlaceholderText(/cari kategori.../i);

    await act(async () => {
      await user.type(searchInput, "xyz123");
    });

    expect(
      screen.getByRole("menuitem", { name: /tidak ada hasil/i }),
    ).toBeInTheDocument();
  });

  it("harus memanggil onChange dengan value baru saat sebuah opsi di-klik", async () => {
    const user = userEvent.setup();
    const mockOnChange = vi.fn();

    render(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected="Kategori A"
        onChange={mockOnChange}
      />,
    );

    const button = screen.getByRole("button", { name: "Kategori A" });

    await act(async () => {
      await user.click(button);
    });

    // Klik opsi yang sudah selected ("Kategori A") -> toggle off, jadi string kosong
    const optionA = screen.getByRole("menuitem", { name: "Kategori A" });
    await act(async () => {
      await user.click(optionA);
    });
    expect(mockOnChange).toHaveBeenCalledWith("");

    // Klik opsi lain ("Kategori B") -> jadi opsi yang baru dipilih
    const optionB = screen.getByRole("menuitem", { name: "Kategori B" });
    await act(async () => {
      await user.click(optionB);
    });
    expect(mockOnChange).toHaveBeenCalledWith("Kategori B");
  });
});
