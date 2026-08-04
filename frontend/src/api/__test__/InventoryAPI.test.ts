import { describe, expect, it } from "vitest";
import { fetchJenisSuggestions } from "../InventoryAPI";

describe("fetchJenisSuggestions", () => {
  it("mengembalikan suggestion yang cocok dari mock data", async () => {
    const response = await fetchJenisSuggestions("el");

    expect(response.success).toBe(true);
    expect(response.data).toContain("Elektronik");
  });

  it("mengembalikan hasil kosong jika query tidak cocok", async () => {
    const response = await fetchJenisSuggestions("jenis-yang-tidak-ada");

    expect(response.success).toBe(true);
    expect(response.data).toHaveLength(0);
  });
});
