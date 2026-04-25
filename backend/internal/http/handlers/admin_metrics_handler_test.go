package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

func TestDashboardReturnsSceneConversionMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB(t)
	defer db.Close()

	r := gin.New()
	r.GET("/api/admin/metrics", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		DashboardMetricsHandler(db)(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp DashboardMetrics
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.NotNil(t, resp.SceneClicks)
	require.NotNil(t, resp.GenerationSuccess)
	require.NotNil(t, resp.Payments)
}

func TestDashboardMetricsWithData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB(t)
	defer db.Close()

	// Seed tracking events
	_, err := db.Exec(`
		INSERT INTO tracking_events (user_id, event, payload, created_at)
		VALUES
			(1, 'scene_clicked', '{"scene_key": "portrait"}', datetime('now')),
			(1, 'scene_clicked', '{"scene_key": "portrait"}', datetime('now')),
			(2, 'scene_clicked', '{"scene_key": "invitation"}', datetime('now'))
	`)
	require.NoError(t, err)

	// Seed generations
	_, err = db.Exec(`
		INSERT INTO generations (user_id, client_request_id, scene_key, template_key, fields, status, created_at, updated_at)
		VALUES
			(1, 'req-1', 'portrait', 'office-pro', '{}', 'completed', datetime('now'), datetime('now')),
			(1, 'req-2', 'portrait', 'office-pro', '{}', 'completed', datetime('now'), datetime('now')),
			(2, 'req-3', 'invitation', 'wedding-classic', '{}', 'failed', datetime('now'), datetime('now'))
	`)
	require.NoError(t, err)

	// Seed orders across two different dates
	_, err = db.Exec(`
		INSERT INTO orders (user_id, order_no, package_code, amount, status, created_at, updated_at)
		VALUES
			(1, 'ORD-001', 'basic', 9.9, 'paid', datetime('now', '-1 day'), datetime('now')),
			(2, 'ORD-002', 'pro', 29.9, 'paid', datetime('now'), datetime('now')),
			(1, 'ORD-003', 'basic', 9.9, 'pending', datetime('now'), datetime('now'))
	`)
	require.NoError(t, err)

	r := gin.New()
	r.GET("/api/admin/metrics", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		DashboardMetricsHandler(db)(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp DashboardMetrics
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Equal(t, 2, resp.SceneClicks["portrait"])
	require.Equal(t, 1, resp.SceneClicks["invitation"])
	require.Equal(t, 2, resp.GenerationSuccess["portrait"])
	require.Equal(t, 0, resp.GenerationSuccess["invitation"])
	require.Equal(t, 2, len(resp.Payments))
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	schema := `
		CREATE TABLE tracking_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			event TEXT NOT NULL,
			payload TEXT,
			created_at TIMESTAMP
		);
		CREATE TABLE generations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			client_request_id TEXT NOT NULL,
			scene_key TEXT NOT NULL,
			template_key TEXT NOT NULL,
			fields TEXT,
			source_asset_id INTEGER,
			status TEXT NOT NULL,
			result_url TEXT,
			prompt TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			order_no TEXT NOT NULL,
			package_code TEXT NOT NULL,
			amount REAL NOT NULL,
			status TEXT NOT NULL,
			wx_prepay_id TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}
