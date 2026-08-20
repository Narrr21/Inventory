import React, { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Card,
  CardContent,
  Chip,
  Divider,
  Grid,
  LinearProgress,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import MainLayout from "../layouts/MainLayout";
import {
  fetchAnalyticsSummary,
  fetchAnalyticsTimeline,
  type AnalyticsSummaryData,
  type AnalyticsTimelineData,
} from "../api/AnalyticsAPI";

const STATUS_COLORS: Record<string, string> = {
  Healthy: "#2e7d32",
  "Under Maintenance": "#ed6c02",
  Broken: "#c62828",
};

const statusColor = (status: string): string =>
  STATUS_COLORS[status] ?? "#757575";

const formatNumber = (value: number): string =>
  new Intl.NumberFormat("id-ID").format(value);

const formatPeriod = (period: string): string => {
  const [year, month] = period.split("-");
  if (!year || !month) return period;
  const date = new Date(Number(year), Number(month) - 1, 1);
  return date.toLocaleString("id-ID", { month: "short", year: "2-digit" });
};

const AnalyticPage: React.FC = () => {
  const [summary, setSummary] = useState<AnalyticsSummaryData | null>(null);
  const [timeline, setTimeline] = useState<AnalyticsTimelineData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    setLoading(true);
    setError(null);

    Promise.all([
      fetchAnalyticsSummary(90, controller.signal),
      fetchAnalyticsTimeline(12, controller.signal),
    ])
      .then(([summaryResponse, timelineResponse]) => {
        setSummary(summaryResponse.data);
        setTimeline(timelineResponse.data);
      })
      .catch((err) => {
        if (err?.name !== "AbortError") {
          setError("Gagal memuat dashboard analitik.");
        }
      })
      .finally(() => setLoading(false));

    return () => controller.abort();
  }, []);

  const statusRows = useMemo(() => summary?.byStatus ?? [], [summary]);
  const jenisRows = useMemo(() => summary?.byJenis ?? [], [summary]);
  const proyekRows = useMemo(() => summary?.byProyek ?? [], [summary]);
  const needsAttention = useMemo(
    () => summary?.needsAttention ?? null,
    [summary],
  );

  if (loading) {
    return (
      <MainLayout>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <Typography variant="h5" sx={{ fontWeight: 600 }}>
            Dashboard Analitik
          </Typography>
          <LinearProgress />
        </Box>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 700 }}>
          Dashboard Analitik
        </Typography>

        {error && <Alert severity="error">{error}</Alert>}

        {summary && (
          <>
            <Grid container spacing={2}>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography variant="overline" color="text.secondary">
                      Total Barang
                    </Typography>
                    <Typography variant="h4">
                      {formatNumber(summary.totals.items)}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography variant="overline" color="text.secondary">
                      Total Project
                    </Typography>
                    <Typography variant="h4">
                      {formatNumber(summary.totals.projects)}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography variant="overline" color="text.secondary">
                      Jenis Barang
                    </Typography>
                    <Typography variant="h4">
                      {formatNumber(summary.totals.itemTypes)}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography variant="overline" color="text.secondary">
                      Project Terpetakan
                    </Typography>
                    <Typography variant="h4">
                      {formatNumber(summary.totals.projectsMapped)} /{" "}
                      {formatNumber(summary.totals.projects)}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>

            <Grid container spacing={2}>
              <Grid size={{ xs: 12, md: 6 }}>
                <Paper sx={{ p: 2, height: "100%" }}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Status Barang
                  </Typography>
                  <Stack spacing={1.5}>
                    {statusRows.map((row) => (
                      <Box key={row.status}>
                        <Box
                          sx={{
                            display: "flex",
                            justifyContent: "space-between",
                            mb: 0.5,
                          }}
                        >
                          <Typography variant="body2">{row.status}</Typography>
                          <Typography variant="body2" fontWeight={600}>
                            {row.count}
                          </Typography>
                        </Box>
                        <Box
                          sx={{
                            width: "100%",
                            height: 10,
                            borderRadius: 999,
                            bgcolor: "rgba(0,0,0,0.08)",
                            overflow: "hidden",
                          }}
                        >
                          <Box
                            sx={{
                              width: `${(row.count / Math.max(summary.totals.items, 1)) * 100}%`,
                              height: "100%",
                              borderRadius: 999,
                              background: statusColor(row.status),
                            }}
                          />
                        </Box>
                      </Box>
                    ))}
                  </Stack>
                </Paper>
              </Grid>

              <Grid size={{ xs: 12, md: 6 }}>
                <Paper sx={{ p: 2, height: "100%" }}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Jenis Barang
                  </Typography>
                  <Stack spacing={1.5}>
                    {jenisRows.slice(0, 8).map((row, index) => (
                      <Box key={row.jenis}>
                        <Box
                          sx={{
                            display: "flex",
                            justifyContent: "space-between",
                            mb: 0.5,
                          }}
                        >
                          <Typography variant="body2">{row.jenis}</Typography>
                          <Typography variant="body2" fontWeight={600}>
                            {row.count}
                          </Typography>
                        </Box>
                        <Box
                          sx={{
                            width: "100%",
                            height: 10,
                            borderRadius: 999,
                            bgcolor: "rgba(0,0,0,0.08)",
                            overflow: "hidden",
                          }}
                        >
                          <Box
                            sx={{
                              width: `${(row.count / Math.max(jenisRows[0]?.count ?? 1, 1)) * 100}%`,
                              height: "100%",
                              borderRadius: 999,
                              background: `hsl(${(index * 55) % 360} 60% 50%)`,
                            }}
                          />
                        </Box>
                      </Box>
                    ))}
                  </Stack>
                </Paper>
              </Grid>
            </Grid>

            <Grid container spacing={2}>
              <Grid size={{ xs: 12, md: 8 }}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Project dengan Jumlah Barang
                  </Typography>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Project</TableCell>
                        <TableCell align="right">Jumlah</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {proyekRows.map((row) => (
                        <TableRow key={row.idProyek}>
                          <TableCell>
                            <Typography variant="body2" fontWeight={600}>
                              {row.namaProyek}
                            </Typography>
                            <Typography
                              variant="caption"
                              color="text.secondary"
                            >
                              {row.lokasi}
                            </Typography>
                          </TableCell>
                          <TableCell align="right">{row.count}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </Paper>
              </Grid>

              <Grid size={{ xs: 12, md: 4 }}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Perhatian
                  </Typography>
                  {needsAttention && (
                    <Stack spacing={1}>
                      {needsAttention.staleItems > 0 && (
                        <Chip
                          label={`${needsAttention.staleItems} barang belum diupdate lebih dari ${needsAttention.staleDays} hari`}
                          color="warning"
                          variant="outlined"
                        />
                      )}
                      {needsAttention.projectsWithoutKoordinat > 0 && (
                        <Chip
                          label={`${needsAttention.projectsWithoutKoordinat} project belum punya titik di peta`}
                          color="info"
                          variant="outlined"
                        />
                      )}
                      {needsAttention.projectsWithoutItems > 0 && (
                        <Chip
                          label={`${needsAttention.projectsWithoutItems} project masih kosong`}
                          color="secondary"
                          variant="outlined"
                        />
                      )}
                      {needsAttention.orphanItems > 0 && (
                        <Chip
                          label={`${needsAttention.orphanItems} barang merujuk project yang tidak ada`}
                          color="error"
                          variant="outlined"
                        />
                      )}
                    </Stack>
                  )}
                </Paper>
              </Grid>
            </Grid>

            {timeline && (
              <Paper sx={{ p: 2 }}>
                <Typography variant="h6" sx={{ mb: 1 }}>
                  Barang yang didata per bulan
                </Typography>
                <Box
                  sx={{
                    display: "flex",
                    alignItems: "flex-end",
                    gap: 1,
                    minHeight: 220,
                  }}
                >
                  {timeline.buckets.map((bucket) => (
                    <Box
                      key={bucket.period}
                      sx={{
                        flex: 1,
                        display: "flex",
                        flexDirection: "column",
                        alignItems: "center",
                        gap: 1,
                      }}
                    >
                      <Box
                        sx={{
                          width: "100%",
                          maxWidth: 32,
                          height: `${Math.max((bucket.created / Math.max(...timeline.buckets.map((b) => b.created), 1)) * 180, 12)}px`,
                          minHeight: 12,
                          bgcolor: "primary.main",
                          borderRadius: 1,
                        }}
                      />
                      <Typography
                        variant="caption"
                        sx={{ textAlign: "center" }}
                      >
                        {formatPeriod(bucket.period)}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              </Paper>
            )}

            <Grid container spacing={2}>
              <Grid size={{ xs: 12, md: 6 }}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Baru Ditambahkan
                  </Typography>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Nama</TableCell>
                        <TableCell>Project</TableCell>
                        <TableCell>Jenis</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {(summary.recentlyAdded ?? [])
                        .slice(0, 5)
                        .map((item: any, index: number) => (
                          <TableRow key={index}>
                            <TableCell>{item.nama ?? "-"}</TableCell>
                            <TableCell>{item.namaProyek ?? "-"}</TableCell>
                            <TableCell>{item.jenis ?? "-"}</TableCell>
                          </TableRow>
                        ))}
                    </TableBody>
                  </Table>
                </Paper>
              </Grid>

              <Grid size={{ xs: 12, md: 6 }}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Baru Diperbarui
                  </Typography>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Nama</TableCell>
                        <TableCell>Project</TableCell>
                        <TableCell>Jenis</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {(summary.recentlyUpdated ?? [])
                        .slice(0, 5)
                        .map((item: any, index: number) => (
                          <TableRow key={index}>
                            <TableCell>{item.nama ?? "-"}</TableCell>
                            <TableCell>{item.namaProyek ?? "-"}</TableCell>
                            <TableCell>{item.jenis ?? "-"}</TableCell>
                          </TableRow>
                        ))}
                    </TableBody>
                  </Table>
                </Paper>
              </Grid>
            </Grid>
          </>
        )}
      </Box>
    </MainLayout>
  );
};

export default AnalyticPage;
