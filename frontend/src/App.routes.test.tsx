import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

describe("App routes", () => {
  beforeEach(() => {
    window.history.pushState({}, "", "/dashboard");

    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL) => {
        const url = String(input);

        if (url.includes("/api/v1/analytics/map")) {
          return Promise.resolve({
            ok: true,
            json: async () => ({
              success: true,
              code: 200,
              data: {
                projects: [],
                unmapped: [],
                bounds: null,
                totals: {
                  projects: 0,
                  mapped: 0,
                  unmapped: 0,
                  items: 0,
                  itemsMapped: 0,
                },
              },
            }),
          } as Response);
        }

        if (url.includes("/api/v1/analytics/summary")) {
          return Promise.resolve({
            ok: true,
            json: async () => ({
              success: true,
              code: 200,
              data: {
                totals: {
                  items: 126,
                  projects: 9,
                  itemTypes: 16,
                  projectsMapped: 7,
                },
                byStatus: [{ status: "Healthy", count: 94 }],
                byJenis: [{ jenis: "Laptop", count: 24 }],
                byProyek: [
                  {
                    idProyek: "1",
                    namaProyek: "ALPHA",
                    lokasi: "Jakarta HQ",
                    count: 31,
                  },
                ],
                licenses: {
                  windows: [{ value: "Pro", count: 9 }],
                  office: [{ value: "365", count: 4 }],
                },
                needsAttention: {
                  staleDays: 90,
                  staleItems: 0,
                  orphanItems: 0,
                  projectsWithoutKoordinat: 0,
                  projectsWithoutItems: 0,
                },
                recentlyAdded: [],
                recentlyUpdated: [],
              },
            }),
          } as Response);
        }

        if (url.includes("/api/v1/analytics/timeline")) {
          return Promise.resolve({
            ok: true,
            json: async () => ({
              success: true,
              code: 200,
              data: {
                months: 6,
                from: "2026-03",
                to: "2026-08",
                buckets: [
                  { period: "2026-03", created: 5, cumulative: 89 },
                  { period: "2026-04", created: 9, cumulative: 98 },
                ],
              },
            }),
          } as Response);
        }

        return Promise.resolve({
          ok: true,
          json: async () => ({}),
        } as Response);
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders the map page at /map", async () => {
    window.history.pushState({}, "", "/map");
    render(<App />);

    expect(await screen.findByText(/Peta Project/i)).toBeInTheDocument();
  });

  it("renders the analytics page at /analytic", async () => {
    window.history.pushState({}, "", "/analytic");
    render(<App />);

    expect(await screen.findByText(/Dashboard Analitik/i)).toBeInTheDocument();
  });
});
