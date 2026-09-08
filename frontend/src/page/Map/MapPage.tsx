import React, { useEffect, useMemo, useState } from "react";
import { Alert, Box, Chip, Divider, Paper, Typography } from "@mui/material";
import { MapContainer, Marker, Popup, TileLayer, useMap } from "react-leaflet";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import type { ProjectMapBounds, ProjectMapResponse } from "../../types/api";
import { fetchProjectMap } from "../../api/analyticsAPI";
import MainLayout from "../layouts/MainLayout";

const DEFAULT_BOUNDS: ProjectMapBounds = {
  north: 6,
  south: -10,
  east: 141,
  west: 90,
};

const createMarkerIcon = () =>
  L.divIcon({
    className: "custom-map-pin",
    html: `
      <span style="
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 18px;
        height: 18px;
        border-radius: 50%;
        background: #1976d2;
        border: 2px solid white;
        box-shadow: 0 2px 8px rgba(0,0,0,0.25);
        color: white;
        font-size: 10px;
        font-weight: 700;
      ">P</span>
    `,
    iconSize: [18, 18],
    iconAnchor: [9, 9],
    popupAnchor: [0, -10],
  });

const FitMapToBounds: React.FC<{ bounds: ProjectMapBounds | null }> = ({
  bounds,
}) => {
  const map = useMap();

  useEffect(() => {
    if (!bounds) {
      return;
    }

    const corner1 = L.latLng(bounds.south, bounds.west);
    const corner2 = L.latLng(bounds.north, bounds.east);
    map.fitBounds(L.latLngBounds(corner1, corner2), {
      padding: [40, 40],
      maxZoom: 12,
    });
  }, [bounds, map]);

  return null;
};

const MapPage: React.FC = () => {
  const [data, setData] = useState<ProjectMapResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    setLoading(true);
    setError(null);

    fetchProjectMap(controller.signal)
      .then((response) => setData(response))
      .catch((err) => {
        if (err?.name !== "AbortError") {
          setError("Gagal memuat data peta project.");
        }
      })
      .finally(() => setLoading(false));

    return () => controller.abort();
  }, []);

  const mappedProjects = useMemo(() => data?.data.projects ?? [], [data]);
  const unmappedProjects = useMemo(() => data?.data.unmapped ?? [], [data]);
  const bounds = data?.data.bounds ?? null;

  return (
    <MainLayout>
      <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>
          Peta Project
        </Typography>

        {error && <Alert severity="error">{error}</Alert>}

        <Paper
          elevation={0}
          sx={{
            border: "1px solid",
            borderColor: "divider",
            borderRadius: 2,
            p: 2,
          }}
        >
          <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1, mb: 2 }}>
            <Chip
              label={`${mappedProjects.length} pin aktif`}
              color="primary"
              variant="outlined"
            />
            <Chip
              label={`${unmappedProjects.length} belum dipetakan`}
              color="secondary"
              variant="outlined"
            />
          </Box>

          <Box sx={{ height: "70vh", width: "100%" }}>
            {!loading && !error ? (
              <MapContainer
                center={[-2.5, 118]}
                zoom={5}
                scrollWheelZoom
                style={{ height: "100%", width: "100%", borderRadius: 12 }}
              >
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />

                <FitMapToBounds bounds={bounds ?? DEFAULT_BOUNDS} />

                {mappedProjects.map((project) => (
                  <Marker
                    key={project._id}
                    position={[project.koordinat.lat, project.koordinat.lng]}
                    icon={createMarkerIcon()}
                  >
                    <Popup>
                      <Box sx={{ minWidth: 180 }}>
                        <Typography
                          variant="subtitle2"
                          sx={{ fontWeight: 700 }}
                        >
                          {project.namaProyek}
                        </Typography>
                        <Typography variant="body2">
                          {project.lokasi}
                        </Typography>
                        <Divider sx={{ my: 1 }} />
                        <Typography variant="caption" color="text.secondary">
                          {project.totalItems} barang
                        </Typography>
                      </Box>
                    </Popup>
                  </Marker>
                ))}
              </MapContainer>
            ) : (
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  height: "100%",
                  color: "text.secondary",
                }}
              >
                {loading ? "Memuat peta..." : "Tidak ada data peta."}
              </Box>
            )}
          </Box>
        </Paper>

        {unmappedProjects.length > 0 && (
          <Paper
            elevation={0}
            sx={{
              borderRadius: 2,
              p: 2,
              border: "1px solid",
              borderColor: "divider",
            }}
          >
            <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
              Belum dipetakan
            </Typography>
            <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
              {unmappedProjects.map((project) => (
                <Chip
                  key={project._id}
                  label={`${project.namaProyek} (${project.totalItems} barang)`}
                  variant="outlined"
                />
              ))}
            </Box>
          </Paper>
        )}
      </Box>
    </MainLayout>
  );
};

export default MapPage;
