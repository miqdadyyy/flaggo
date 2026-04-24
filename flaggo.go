package flaggo

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"hash/fnv"
	"net/http"
)

// FlagConfig holds the full configuration for a feature flag.
type FlagConfig struct {
	Enabled bool                `json:"enabled"`
	Actors  map[string][]string `json:"actors,omitempty"`
	Rollout int                 `json:"rollout"`
}

// Actor represents the identity making a request.
type Actor struct {
	Attributes map[string]string
}

type actorKey struct{}

// WithActor returns a new context with the given actor attached.
func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

// ActorFrom extracts the actor from the context. Returns an empty Actor if none is set.
func ActorFrom(ctx context.Context) Actor {
	a, _ := ctx.Value(actorKey{}).(Actor)
	return a
}

// Provider is the interface that storage backends must implement.
type Provider interface {
	Set(ctx context.Context, key string, config FlagConfig) error
	Get(ctx context.Context, key string) (FlagConfig, bool)
	All(ctx context.Context) (map[string]FlagConfig, error)
}

// Config holds optional configuration for a Flaggo instance.
type Config struct {
	// Username for HTTP Basic Authentication on the handler.
	// Leave empty to disable auth.
	Username string
	// Password for HTTP Basic Authentication on the handler.
	Password string
}

// Flaggo is the main feature flag manager.
type Flaggo struct {
	provider Provider
	config   Config
}

// New creates a new Flaggo instance with the given provider and optional config.
func New(provider Provider, cfgs ...Config) *Flaggo {
	var cfg Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	return &Flaggo{provider: provider, config: cfg}
}

// IsEnabled evaluates whether the flag is enabled for the current actor.
//
// Evaluation order:
//  1. If the flag does not exist or Enabled is false → false
//  2. If the actor matches any entry in Actors → true
//  3. If hash(flagKey + sessionID) % 100 < Rollout → true
//  4. Otherwise → false
func (f *Flaggo) IsEnabled(ctx context.Context, key string) bool {
	cfg, exists := f.provider.Get(ctx, key)
	if !exists || !cfg.Enabled {
		return false
	}

	actor := ActorFrom(ctx)

	// Check actor targeting
	for attr, allowed := range cfg.Actors {
		val, ok := actor.Attributes[attr]
		if !ok {
			continue
		}
		for _, a := range allowed {
			if val == a {
				return true
			}
		}
	}

	// Check rollout percentage
	if cfg.Rollout >= 100 {
		return true
	}
	if cfg.Rollout <= 0 {
		return false
	}

	sessionID := actor.Attributes["session_id"]
	if sessionID == "" {
		return false
	}

	return hashPercent(key, sessionID) < cfg.Rollout
}

// IsDisabled returns whether the given feature flag is disabled for the current actor.
func (f *Flaggo) IsDisabled(ctx context.Context, key string) bool {
	return !f.IsEnabled(ctx, key)
}

// Enable sets the flag's Enabled field to true, preserving existing actors and rollout.
func (f *Flaggo) Enable(ctx context.Context, key string) error {
	cfg, _ := f.provider.Get(ctx, key)
	cfg.Enabled = true
	return f.provider.Set(ctx, key, cfg)
}

// Disable sets the flag's Enabled field to false, preserving existing actors and rollout.
func (f *Flaggo) Disable(ctx context.Context, key string) error {
	cfg, _ := f.provider.Get(ctx, key)
	cfg.Enabled = false
	return f.provider.Set(ctx, key, cfg)
}

// Toggle flips the flag's Enabled field, preserving existing actors and rollout.
func (f *Flaggo) Toggle(ctx context.Context, key string) error {
	cfg, _ := f.provider.Get(ctx, key)
	cfg.Enabled = !cfg.Enabled
	return f.provider.Set(ctx, key, cfg)
}

// SetConfig stores the full configuration for a flag.
func (f *Flaggo) SetConfig(ctx context.Context, key string, config FlagConfig) error {
	return f.provider.Set(ctx, key, config)
}

// GetConfig returns the full configuration for a flag.
func (f *Flaggo) GetConfig(ctx context.Context, key string) (FlagConfig, bool) {
	return f.provider.Get(ctx, key)
}

// All returns all feature flags and their configurations.
func (f *Flaggo) All(ctx context.Context) (map[string]FlagConfig, error) {
	return f.provider.All(ctx)
}

// Populate ensures the given keys exist as feature flags.
// Keys that already exist are left unchanged; new keys are created as disabled.
func (f *Flaggo) Populate(ctx context.Context, keys []string) error {
	existing, err := f.provider.All(ctx)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if _, ok := existing[key]; !ok {
			if err = f.provider.Set(ctx, key, FlagConfig{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetHandler returns an http.Handler that serves the JSON API for managing flags.
// If Config.Username and Config.Password are set, Basic Authentication is required.
//
// Routes (by HTTP method and Accept header):
//   - GET  (Accept: application/json)            — list all flags with full config
//   - GET  (Accept: application/json, ?key=name) — get specific flag config
//   - GET  (no JSON accept)                      — serves empty 200 (use web.Handler for dashboard UI)
//   - POST   — create/update flag config (body: {"key": "name", "enabled": true, "actors": {...}, "rollout": 50})
//   - DELETE  — disable flag (body: {"key": "name"})
//   - PATCH   — toggle flag enabled state (body: {"key": "name"})
func (f *Flaggo) GetHandler() http.Handler {
	var h http.Handler = &httpHandler{ff: f}
	return f.WrapAuth(h)
}

// WrapAuth wraps the given handler with Basic Authentication if configured.
func (f *Flaggo) WrapAuth(h http.Handler) http.Handler {
	if f.config.Username != "" || f.config.Password != "" {
		return basicAuthMiddleware(h, f.config.Username, f.config.Password)
	}
	return h
}

func basicAuthMiddleware(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || !secureCompare(u, username) || !secureCompare(p, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="flaggo"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func secureCompare(a, b string) bool {
	ha := sha256.Sum256([]byte(a))
	hb := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ha[:], hb[:]) == 1
}

type httpHandler struct {
	ff *Flaggo
}

type configRequest struct {
	Key     string              `json:"key"`
	Enabled *bool               `json:"enabled,omitempty"`
	Actors  map[string][]string `json:"actors,omitempty"`
	Rollout *int                `json:"rollout,omitempty"`
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if accept == "application/json" {
		h.handleJSON(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *httpHandler) handleJSON(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	case http.MethodPatch:
		h.handlePatch(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key != "" {
		cfg, exists := h.ff.GetConfig(r.Context(), key)
		if !exists {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "flag not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"key":     key,
			"enabled": cfg.Enabled,
			"actors":  cfg.Actors,
			"rollout": cfg.Rollout,
		})
		return
	}

	flags, err := h.ff.All(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, flags)
}

func (h *httpHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if req.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Key is required"})
		return
	}

	cfg, _ := h.ff.GetConfig(r.Context(), req.Key)

	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.Actors != nil {
		cfg.Actors = req.Actors
	}
	if req.Rollout != nil {
		cfg.Rollout = *req.Rollout
	}

	if err := h.ff.SetConfig(r.Context(), req.Key, cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Feature flag updated",
		"key":     req.Key,
		"enabled": cfg.Enabled,
		"actors":  cfg.Actors,
		"rollout": cfg.Rollout,
	})
}

func (h *httpHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if req.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Key is required"})
		return
	}

	if err := h.ff.Disable(r.Context(), req.Key); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Feature flag disabled", "key": req.Key})
}

func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if req.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Key is required"})
		return
	}

	if err := h.ff.Toggle(r.Context(), req.Key); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	cfg, _ := h.ff.GetConfig(r.Context(), req.Key)
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Feature flag toggled",
		"key":     req.Key,
		"enabled": cfg.Enabled,
		"actors":  cfg.Actors,
		"rollout": cfg.Rollout,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// hashPercent returns a deterministic value 0-99 for a given flag key and session ID.
func hashPercent(flagKey, sessionID string) int {
	h := fnv.New32a()
	h.Write([]byte(flagKey + ":" + sessionID))
	return int(h.Sum32() % 100)
}
