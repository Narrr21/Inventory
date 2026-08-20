package handlers

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"my-backend/internal/models"
	"my-backend/internal/repository"
	"my-backend/internal/response"
)

// Analytics serves the three read-only dashboard endpoints. Everything is
// computed per request straight from the collections — no cached counters to
// fall out of sync with CRUD, at the cost of a handful of aggregations per
// call. At this data size (hundreds of items) that trade is obviously right;
// if the collection ever grows enough for it not to be, the fix is a cache
// here, not a different response shape.
//
// Deliberately absent: any notion of which status is "bad". Status is a
// free-text field with no master list, so the backend has no basis for
// declaring "Broken" more severe than "Under Maintenance" — it reports counts
// per value and lets the frontend own that mapping.
type AnalyticsHandler struct {
	items     *repository.ItemRepository
	projects  *repository.ProjectRepository
	itemTypes *repository.ItemTypeRepository
}

func NewAnalyticsHandler(items *repository.ItemRepository, projects *repository.ProjectRepository, itemTypes *repository.ItemTypeRepository) *AnalyticsHandler {
	return &AnalyticsHandler{items: items, projects: projects, itemTypes: itemTypes}
}

const (
	defaultStaleDays = 90
	maxStaleDays     = 3650
	defaultMonths    = 12
	maxMonths        = 60
	recentLimit      = 5
)

// clampedIntParam reads a positive integer query param, falling back to def
// for anything unparseable or out of range. Bad input is never an error here:
// these are display knobs, and a 422 on "?months=abc" would break a chart for
// no good reason.
func clampedIntParam(raw string, def, min, max int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		if raw == "" || err != nil {
			return def
		}
		return min
	}
	if value > max {
		return max
	}
	return value
}

// labeledBucket renames CountBy's generic `_id` to the field it actually
// counted, so a client reads `status`/`jenis`/`value` instead of `_id`.
func labeledBucket(label string, buckets []repository.CountBucket) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(buckets))
	for _, b := range buckets {
		out = append(out, map[string]interface{}{label: b.ID, "count": b.Count})
	}
	return out
}

func (h *AnalyticsHandler) publicItems(ctx context.Context, items []models.Item) []models.ItemPublic {
	namaProyekByID := buildNamaProyekIndex(ctx, h.projects)
	out := make([]models.ItemPublic, 0, len(items))
	for _, it := range items {
		it.EnsureMaps()
		it.NamaProyek = namaProyekByID[it.IdProyek]
		out = append(out, it.Public())
	}
	return out
}

// Summary: GET /api/v1/analytics/summary
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fail := func() {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load analytics summary", nil)
	}

	staleDays := clampedIntParam(r.URL.Query().Get("staleDays"), defaultStaleDays, 1, maxStaleDays)

	totalItems, err := h.items.Count(ctx)
	if err != nil {
		fail()
		return
	}
	projects, err := h.projects.List(ctx)
	if err != nil {
		fail()
		return
	}
	itemTypeCount, err := h.itemTypes.Count(ctx)
	if err != nil {
		fail()
		return
	}
	byStatus, err := h.items.CountBy(ctx, "status")
	if err != nil {
		fail()
		return
	}
	byJenis, err := h.items.CountBy(ctx, "jenis")
	if err != nil {
		fail()
		return
	}
	byProyekRaw, err := h.items.CountBy(ctx, "idProyek")
	if err != nil {
		fail()
		return
	}
	licenseWindows, err := h.items.CountBy(ctx, "licenseWindows")
	if err != nil {
		fail()
		return
	}
	licenseOffice, err := h.items.CountBy(ctx, "licenseOffice")
	if err != nil {
		fail()
		return
	}

	staleBefore := time.Now().UTC().AddDate(0, 0, -staleDays).Format("2006-01-02T15:04:05.000Z")
	staleItems, err := h.items.CountUpdatedBefore(ctx, staleBefore)
	if err != nil {
		fail()
		return
	}

	projectIDs := make([]string, 0, len(projects))
	for _, p := range projects {
		projectIDs = append(projectIDs, p.ID)
	}
	orphanItems, err := h.items.CountOrphans(ctx, projectIDs)
	if err != nil {
		fail()
		return
	}

	recentlyAdded, err := h.items.RecentlyAdded(ctx, recentLimit)
	if err != nil {
		fail()
		return
	}
	recentlyUpdated, err := h.items.RecentlyUpdated(ctx, recentLimit)
	if err != nil {
		fail()
		return
	}

	// byProyek is joined with the project list here rather than left as bare
	// IDs, since an ID alone is useless in a chart legend. Projects with zero
	// items are included so "which sites are empty" is answerable from the
	// same array; orphan groups keep their unresolvable ID with a blank name.
	countByProject := make(map[string]int, len(byProyekRaw))
	for _, b := range byProyekRaw {
		countByProject[b.ID] = b.Count
	}
	byProyek := make([]map[string]interface{}, 0, len(projects)+len(byProyekRaw))
	projectsWithoutItems := 0
	projectsMapped := 0
	for _, p := range projects {
		if p.Koordinat != nil {
			projectsMapped++
		}
		count := countByProject[p.ID]
		if count == 0 {
			projectsWithoutItems++
		}
		byProyek = append(byProyek, map[string]interface{}{
			"idProyek":   p.ID,
			"namaProyek": p.NamaProyek,
			"lokasi":     p.Lokasi,
			"count":      count,
		})
		delete(countByProject, p.ID)
	}
	orphanGroups := make([]string, 0, len(countByProject))
	for id := range countByProject {
		orphanGroups = append(orphanGroups, id)
	}
	sort.Strings(orphanGroups)
	for _, id := range orphanGroups {
		byProyek = append(byProyek, map[string]interface{}{
			"idProyek":   id,
			"namaProyek": "",
			"lokasi":     "",
			"count":      countByProject[id],
		})
	}
	sort.SliceStable(byProyek, func(i, j int) bool {
		return byProyek[i]["count"].(int) > byProyek[j]["count"].(int)
	})

	data := map[string]interface{}{
		"totals": map[string]interface{}{
			"items":          totalItems,
			"projects":       len(projects),
			"itemTypes":      itemTypeCount,
			"projectsMapped": projectsMapped,
		},
		"byStatus": labeledBucket("status", byStatus),
		"byJenis":  labeledBucket("jenis", byJenis),
		"byProyek": byProyek,
		"licenses": map[string]interface{}{
			"windows": labeledBucket("value", licenseWindows),
			"office":  labeledBucket("value", licenseOffice),
		},
		"needsAttention": map[string]interface{}{
			"staleDays":                staleDays,
			"staleItems":               staleItems,
			"orphanItems":              orphanItems,
			"projectsWithoutKoordinat": len(projects) - projectsMapped,
			"projectsWithoutItems":     projectsWithoutItems,
		},
		"recentlyAdded":   h.publicItems(ctx, recentlyAdded),
		"recentlyUpdated": h.publicItems(ctx, recentlyUpdated),
	}
	response.OK(w, http.StatusOK, data, nil)
}

// Map: GET /api/v1/analytics/map
//
// One request per map render: the points, how many items sit on each, and the
// status mix for popups/marker colors. Projects without coordinates are
// returned separately instead of being dropped — silently omitting them would
// make items disappear from the dashboard with no explanation, so they're
// surfaced as an explicit "not mapped yet" list.
func (h *AnalyticsHandler) Map(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fail := func() {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load analytics map", nil)
	}

	projects, err := h.projects.List(ctx)
	if err != nil {
		fail()
		return
	}
	buckets, err := h.items.CountByProjectAndStatus(ctx)
	if err != nil {
		fail()
		return
	}
	totalItems, err := h.items.Count(ctx)
	if err != nil {
		fail()
		return
	}

	statusByProject := make(map[string][]repository.ProjectStatusBucket)
	for _, b := range buckets {
		statusByProject[b.IdProyek] = append(statusByProject[b.IdProyek], b)
	}

	mapped := make([]map[string]interface{}, 0, len(projects))
	unmapped := make([]map[string]interface{}, 0)
	var itemsMapped int
	var north, south, east, west float64
	hasBounds := false

	for _, p := range projects {
		statuses := statusByProject[p.ID]
		// count desc, then name asc — a stable order matters here because
		// dominantStatus is read off the front of this slice.
		sort.SliceStable(statuses, func(i, j int) bool {
			if statuses[i].Count != statuses[j].Count {
				return statuses[i].Count > statuses[j].Count
			}
			return statuses[i].Status < statuses[j].Status
		})

		byStatus := make([]map[string]interface{}, 0, len(statuses))
		total := 0
		for _, s := range statuses {
			byStatus = append(byStatus, map[string]interface{}{"status": s.Status, "count": s.Count})
			total += s.Count
		}
		dominant := ""
		if len(statuses) > 0 {
			dominant = statuses[0].Status
		}

		if p.Koordinat == nil {
			unmapped = append(unmapped, map[string]interface{}{
				"_id":        p.ID,
				"namaProyek": p.NamaProyek,
				"lokasi":     p.Lokasi,
				"totalItems": total,
			})
			continue
		}

		mapped = append(mapped, map[string]interface{}{
			"_id":            p.ID,
			"namaProyek":     p.NamaProyek,
			"lokasi":         p.Lokasi,
			"koordinat":      p.Koordinat,
			"totalItems":     total,
			"byStatus":       byStatus,
			"dominantStatus": dominant,
		})
		itemsMapped += total

		if !hasBounds {
			north, south = p.Koordinat.Lat, p.Koordinat.Lat
			east, west = p.Koordinat.Lng, p.Koordinat.Lng
			hasBounds = true
			continue
		}
		north = maxFloat(north, p.Koordinat.Lat)
		south = minFloat(south, p.Koordinat.Lat)
		east = maxFloat(east, p.Koordinat.Lng)
		west = minFloat(west, p.Koordinat.Lng)
	}

	// bounds stays null rather than collapsing to 0,0 when nothing is mapped —
	// fitBounds on a zero box would center the map on the Atlantic.
	var bounds interface{}
	if hasBounds {
		bounds = map[string]float64{"north": north, "south": south, "east": east, "west": west}
	}

	data := map[string]interface{}{
		"projects": mapped,
		"unmapped": unmapped,
		"bounds":   bounds,
		"totals": map[string]interface{}{
			"projects":    len(projects),
			"mapped":      len(mapped),
			"unmapped":    len(unmapped),
			"items":       totalItems,
			"itemsMapped": itemsMapped,
		},
	}
	response.OK(w, http.StatusOK, data, nil)
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Timeline: GET /api/v1/analytics/timeline
//
// Items created per calendar month (UTC), zero-filled across the whole window
// so a chart never has to reconstruct missing months itself. `cumulative`
// counts everything created up to the end of each bucket, including items
// older than the window, so the line reads as inventory size rather than
// restarting from zero at the left edge.
func (h *AnalyticsHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fail := func() {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load analytics timeline", nil)
	}

	months := clampedIntParam(r.URL.Query().Get("months"), defaultMonths, 1, maxMonths)

	// Window is [first day of (this month - months + 1), now], so the current
	// month is always the last bucket.
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -(months - 1), 0)

	periods := make([]string, 0, months)
	for i := 0; i < months; i++ {
		periods = append(periods, start.AddDate(0, i, 0).Format("2006-01"))
	}

	buckets, err := h.items.CreatedPerMonth(ctx, periods[0])
	if err != nil {
		fail()
		return
	}
	countByPeriod := make(map[string]int, len(buckets))
	for _, b := range buckets {
		countByPeriod[b.Period] = b.Count
	}

	// Everything created before the window still counts towards the running
	// total. Comparing createdAt against the bare "YYYY-MM" prefix works
	// because ISO-8601 sorts lexicographically: "2026-03" < "2026-03-01T...".
	older, err := h.items.CountCreatedBefore(ctx, periods[0])
	if err != nil {
		fail()
		return
	}

	cumulative := int(older)
	out := make([]map[string]interface{}, 0, months)
	for _, period := range periods {
		created := countByPeriod[period]
		cumulative += created
		out = append(out, map[string]interface{}{
			"period":     period,
			"created":    created,
			"cumulative": cumulative,
		})
	}

	data := map[string]interface{}{
		"months":  months,
		"from":    periods[0],
		"to":      periods[len(periods)-1],
		"buckets": out,
	}
	response.OK(w, http.StatusOK, data, nil)
}
