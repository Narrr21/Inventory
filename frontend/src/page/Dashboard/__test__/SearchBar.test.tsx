import { render, screen, fireEvent, act } from "@testing-library/react";
import SearchBar from "../../../utils/Table/SearchBar";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

describe("SearchBar Component", () => {
  beforeEach(() => {
    // Gunakan fake timers untuk mengontrol debounce
    vi.useFakeTimers();
  });

  afterEach(() => {
    // Kembalikan ke real timers setelah setiap test
    vi.useRealTimers();
  });

  it("harus merender input dengan placeholder default", () => {
    render(<SearchBar onSearch={() => {}} />);

    // Pastikan input dengan placeholder default "Cari..." muncul di layar
    expect(screen.getByPlaceholderText(/cari.../i)).toBeInTheDocument();
  });

  it("harus merender input dengan placeholder kustom jika diberikan", () => {
    render(<SearchBar onSearch={() => {}} placeholder="Cari produk..." />);

    // Pastikan input dengan placeholder kustom muncul di layar
    expect(screen.getByPlaceholderText(/cari produk.../i)).toBeInTheDocument();
  });

  it("harus memanggil onSearch dengan value yang di-trim setelah debounce selesai", () => {
    const mockOnSearch = vi.fn();

    render(<SearchBar onSearch={mockOnSearch} debounceMs={400} />);

    // Ambil input
    const input = screen.getByPlaceholderText(/cari.../i);

    act(() => {
      fireEvent.change(input, { target: { value: " sepatu baru " } });
    });

    // Majukan waktu 200ms, onSearch seharusnya belum dipanggil karena debounce belum selesai
    act(() => {
      vi.advanceTimersByTime(200);
    });

    // Pastikan onSearch belum dipanggil dengan value yang di-trim
    expect(mockOnSearch).not.toHaveBeenCalledWith("sepatu baru");

    // Majukan waktu hingga total 400ms, onSearch seharusnya dipanggil dengan value yang di-trim
    act(() => {
      vi.advanceTimersByTime(200);
    });

    // Pastikan onSearch dipanggil dengan value yang di-trim
    expect(mockOnSearch).toHaveBeenCalledWith("sepatu baru");
  });
});
