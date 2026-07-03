import {
  render,
  screen,
  act,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import SearchableFilter from "../components/SearchableFilter";
import { describe, expect, it, vi } from "vitest";

interface FilterOption {
  value: string;
  label: string;
}

describe("SearchableFilter Component", () => {
  const mockOptions: FilterOption[] = [
    { value: "1", label: "Kategori A" },
    { value: "2", label: "Kategori B" },
    { value: "3", label: "Kategori C" },
    { value: "4", label: "Kategori D" },
    { value: "5", label: "Kategori E" },
    { value: "6", label: "Kategori F" }, // Opsi tambahan untuk menguji maxVisible
  ];

  it("harus merender tombol filter dengan jumlah selected yang sesuai", () => {
    // Render komponen dengan kondisi awal selected kosong
    const { rerender } = render(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected={[]}
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
        selected={["1", "2"]}
        onChange={() => {}}
      />,
    );

    // Pastikan tombol filter menampilkan jumlah selected yang benar
    expect(
      screen.getByRole("button", { name: /kategori \(2\)/i }),
    ).toBeInTheDocument();
  });

  it("harus membuka menu dan merender opsi sampai batas maxVisible saat tombol diklik", async () => {
    const user = userEvent.setup();

    // Render komponen dengan maxVisible = 5
    render(
      <SearchableFilter
        label="Kategori"
        options={mockOptions}
        selected={[]}
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
        selected={[]}
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
    // act(() => {
    //   fireEvent.change(searchInput, { target: { value: "Kategori A" } });
    // });
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
        selected={[]}
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
        selected={["1"]}
        onChange={mockOnChange}
      />,
    );

    const button = screen.getByRole("button", { name: /kategori \(1\)/i });

    await act(async () => {
      await user.click(button);
    });

    // PERBAIKAN: Mengubah "option" menjadi "menuitem" untuk uncheck Kategori A
    const optionA = screen.getByRole("menuitem", { name: "Kategori A" });
    await act(async () => {
      await user.click(optionA);
    });
    expect(mockOnChange).toHaveBeenCalledWith([]);

    // PERBAIKAN: Mengubah "option" menjadi "menuitem" untuk check Kategori B
    const optionB = screen.getByRole("menuitem", { name: "Kategori B" });
    await act(async () => {
      await user.click(optionB);
    });
    expect(mockOnChange).toHaveBeenCalledWith(["1", "2"]);
  });
});
