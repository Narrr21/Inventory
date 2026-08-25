import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  buildProjectRequest,
  fetchJenisSuggestions,
  normalizeProjectCoordinateInput,
} from "../inventoryAPI";

describe("fetchJenisSuggestions", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        const isMatch = url.includes("q=el");

        return {
          ok: true,
          json: async () => ({
            success: true,
            data: isMatch
              ? [
                  { _id: "1", jenis: "Elektronik" },
                  { _id: "2", jenis: "Laptop" },
                ]
              : [],
          }),
        } as Response;
      }),
    );
  });

  it("mengembalikan suggestion yang cocok dari mock data", async () => {
    const response = await fetchJenisSuggestions("el");

    expect(response.success).toBe(true);
    expect(response.data.map((item) => item.jenis)).toContain("Elektronik");
  });

  it("mengembalikan hasil kosong jika query tidak cocok", async () => {
    const response = await fetchJenisSuggestions("jenis-yang-tidak-ada");

    expect(response.success).toBe(true);
    expect(response.data).toHaveLength(0);
  });
});

describe("project coordinate handling", () => {
  it("menghasilkan payload dengan koordinat null bila kosong", () => {
    const payload = buildProjectRequest({
      name: "ALPHA",
      location: "Jakarta",
      lat: "",
      lng: "",
    });

    expect(payload.koordinat).toBeNull();
    expect(payload.namaProyek).toBe("ALPHA");
    expect(payload.lokasi).toBe("Jakarta");
  });

  it("menolak koordinat parsial sebelum request dikirim", () => {
    const result = normalizeProjectCoordinateInput("-6.2088", "");

    expect(result.isValid).toBe(false);
    expect(result.errors["koordinat.lng"]).toBe(
      "Latitude dan longitude harus diisi bersamaan.",
    );
  });
});
