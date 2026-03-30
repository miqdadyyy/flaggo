package flaggo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// memProvider is an in-memory provider for testing.
type memProvider struct {
	flags map[string]FlagConfig
}

func newMemProvider() *memProvider {
	return &memProvider{flags: make(map[string]FlagConfig)}
}

func (m *memProvider) Set(_ context.Context, key string, config FlagConfig) error {
	m.flags[key] = config
	return nil
}

func (m *memProvider) Get(_ context.Context, key string) (FlagConfig, bool) {
	cfg, ok := m.flags[key]
	return cfg, ok
}

func (m *memProvider) All(_ context.Context) (map[string]FlagConfig, error) {
	out := make(map[string]FlagConfig, len(m.flags))
	for k, v := range m.flags {
		out[k] = v
	}
	return out, nil
}

// must is a test helper that fails the test on error.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	ff := New(newMemProvider())
	if ff == nil {
		t.Fatal("expected non-nil Flaggo")
	}
}

func TestNewWithConfig(t *testing.T) {
	ff := New(newMemProvider(), Config{Username: "admin", Password: "secret"})
	if ff.config.Username != "admin" {
		t.Fatalf("expected username 'admin', got '%s'", ff.config.Username)
	}
}

func TestEnableDisableToggle(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())

	must(t, ff.Enable(ctx, "flag1"))
	cfg, ok := ff.GetConfig(ctx, "flag1")
	if !ok || !cfg.Enabled {
		t.Fatal("expected flag1 to be enabled")
	}

	must(t, ff.Disable(ctx, "flag1"))
	cfg, _ = ff.GetConfig(ctx, "flag1")
	if cfg.Enabled {
		t.Fatal("expected flag1 to be disabled")
	}

	must(t, ff.Toggle(ctx, "flag1"))
	cfg, _ = ff.GetConfig(ctx, "flag1")
	if !cfg.Enabled {
		t.Fatal("expected flag1 to be enabled after toggle")
	}
}

func TestEnablePreservesConfig(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())

	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{
		Enabled: false,
		Rollout: 50,
		Actors:  map[string][]string{"user_id": {"1"}},
	}))

	must(t, ff.Enable(ctx, "flag1"))
	cfg, _ := ff.GetConfig(ctx, "flag1")
	if cfg.Rollout != 50 {
		t.Fatalf("expected rollout 50, got %d", cfg.Rollout)
	}
	if len(cfg.Actors["user_id"]) != 1 {
		t.Fatal("expected actors to be preserved")
	}
}

func TestSetConfigAndGetConfig(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())

	want := FlagConfig{
		Enabled: true,
		Actors:  map[string][]string{"user_id": {"1", "2"}, "org_id": {"99"}},
		Rollout: 75,
	}
	must(t, ff.SetConfig(ctx, "flag1", want))

	got, ok := ff.GetConfig(ctx, "flag1")
	if !ok {
		t.Fatal("expected flag1 to exist")
	}
	if got.Enabled != want.Enabled || got.Rollout != want.Rollout {
		t.Fatalf("config mismatch: got %+v", got)
	}
	if len(got.Actors["user_id"]) != 2 {
		t.Fatalf("expected 2 user_id actors, got %d", len(got.Actors["user_id"]))
	}
}

func TestGetConfigNotFound(t *testing.T) {
	ff := New(newMemProvider())
	_, ok := ff.GetConfig(context.Background(), "nonexistent")
	if ok {
		t.Fatal("expected flag not to exist")
	}
}

func TestAll(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())

	must(t, ff.SetConfig(ctx, "a", FlagConfig{Enabled: true}))
	must(t, ff.SetConfig(ctx, "b", FlagConfig{Enabled: false}))

	all, err := ff.All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(all))
	}
}

func TestPopulate(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())

	must(t, ff.SetConfig(ctx, "existing", FlagConfig{Enabled: true, Rollout: 80}))

	must(t, ff.Populate(ctx, []string{"existing", "new1", "new2"}))

	cfg, _ := ff.GetConfig(ctx, "existing")
	if !cfg.Enabled || cfg.Rollout != 80 {
		t.Fatal("existing flag should not be modified")
	}

	cfg, ok := ff.GetConfig(ctx, "new1")
	if !ok {
		t.Fatal("new1 should exist")
	}
	if cfg.Enabled {
		t.Fatal("new1 should be disabled")
	}
}

// --- IsEnabled evaluation tests ---

func TestIsEnabled_NonExistentFlag(t *testing.T) {
	ff := New(newMemProvider())
	if ff.IsEnabled(context.Background(), "nope") {
		t.Fatal("non-existent flag should not be enabled")
	}
}

func TestIsEnabled_DisabledFlag(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: false, Rollout: 100}))

	if ff.IsEnabled(ctx, "flag1") {
		t.Fatal("disabled flag should not be enabled even with 100% rollout")
	}
}

func TestIsEnabled_FullRollout(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 100}))

	if !ff.IsEnabled(ctx, "flag1") {
		t.Fatal("flag with 100% rollout should be enabled")
	}
}

func TestIsEnabled_ZeroRolloutNoActors(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 0}))

	ctx = WithActor(ctx, Actor{Attributes: map[string]string{"session_id": "sess1"}})
	if ff.IsEnabled(ctx, "flag1") {
		t.Fatal("flag with 0% rollout and no actor match should be disabled")
	}
}

func TestIsEnabled_ActorTargeting(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{
		Enabled: true,
		Rollout: 0,
		Actors:  map[string][]string{"user_id": {"42", "99"}},
	}))

	ctx1 := WithActor(ctx, Actor{Attributes: map[string]string{"user_id": "42", "session_id": "s1"}})
	if !ff.IsEnabled(ctx1, "flag1") {
		t.Fatal("targeted actor should see the flag")
	}

	ctx2 := WithActor(ctx, Actor{Attributes: map[string]string{"user_id": "1", "session_id": "s2"}})
	if ff.IsEnabled(ctx2, "flag1") {
		t.Fatal("non-targeted actor should not see the flag with 0% rollout")
	}
}

func TestIsEnabled_ActorMultipleAttributes(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{
		Enabled: true,
		Rollout: 0,
		Actors:  map[string][]string{"user_id": {"1"}, "org_id": {"99"}},
	}))

	ctx1 := WithActor(ctx, Actor{Attributes: map[string]string{"user_id": "5", "org_id": "99"}})
	if !ff.IsEnabled(ctx1, "flag1") {
		t.Fatal("actor matching org_id should see the flag")
	}
}

func TestIsEnabled_RolloutDeterministic(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 50}))

	ctx1 := WithActor(ctx, Actor{Attributes: map[string]string{"session_id": "test-session-1"}})
	result1 := ff.IsEnabled(ctx1, "flag1")

	for i := 0; i < 100; i++ {
		if ff.IsEnabled(ctx1, "flag1") != result1 {
			t.Fatal("rollout should be deterministic for the same actor")
		}
	}
}

func TestIsEnabled_RolloutNoSessionID(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 50}))

	if ff.IsEnabled(ctx, "flag1") {
		t.Fatal("no session_id should return false for partial rollout")
	}

	ctx2 := WithActor(ctx, Actor{Attributes: map[string]string{"user_id": "1"}})
	if ff.IsEnabled(ctx2, "flag1") {
		t.Fatal("actor without session_id should return false for partial rollout")
	}
}

func TestIsDisabled(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 100}))

	if ff.IsDisabled(ctx, "flag1") {
		t.Fatal("enabled flag should not be disabled")
	}

	must(t, ff.Disable(ctx, "flag1"))
	if !ff.IsDisabled(ctx, "flag1") {
		t.Fatal("disabled flag should be disabled")
	}
}

// --- Actor context tests ---

func TestWithActorAndActorFrom(t *testing.T) {
	ctx := WithActor(context.Background(), Actor{Attributes: map[string]string{"user_id": "42"}})
	got := ActorFrom(ctx)
	if got.Attributes["user_id"] != "42" {
		t.Fatalf("expected user_id '42', got '%s'", got.Attributes["user_id"])
	}
}

func TestActorFromEmpty(t *testing.T) {
	got := ActorFrom(context.Background())
	if got.Attributes != nil {
		t.Fatal("expected nil attributes from empty context")
	}
}

// --- HTTP handler tests ---

func jsonReq(method, url, body string) *http.Request {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, url, strings.NewReader(body))
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	return r
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestHandler_GetAllFlags(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 100}))

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodGet, "/", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]FlagConfig
	decodeJSON(t, w, &body)
	if _, ok := body["flag1"]; !ok {
		t.Fatal("expected flag1 in response")
	}
}

func TestHandler_GetSingleFlag(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 75}))

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodGet, "/?key=flag1", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	decodeJSON(t, w, &body)
	if body["key"] != "flag1" {
		t.Fatal("expected key flag1")
	}
	if body["rollout"].(float64) != 75 {
		t.Fatalf("expected rollout 75, got %v", body["rollout"])
	}
}

func TestHandler_GetFlagNotFound(t *testing.T) {
	ff := New(newMemProvider())

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodGet, "/?key=nope", ""))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_PostCreateFlag(t *testing.T) {
	ff := New(newMemProvider())

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodPost, "/",
		`{"key":"new-flag","enabled":true,"rollout":50,"actors":{"user_id":["1","2"]}}`))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	cfg, ok := ff.GetConfig(context.Background(), "new-flag")
	if !ok {
		t.Fatal("expected flag to exist")
	}
	if !cfg.Enabled || cfg.Rollout != 50 || len(cfg.Actors["user_id"]) != 2 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestHandler_PostPartialUpdate(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 30, Actors: map[string][]string{"user_id": {"1"}}}))

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodPost, "/", `{"key":"flag1","rollout":80}`))

	cfg, _ := ff.GetConfig(ctx, "flag1")
	if cfg.Rollout != 80 {
		t.Fatalf("expected rollout 80, got %d", cfg.Rollout)
	}
	if !cfg.Enabled {
		t.Fatal("enabled should be preserved")
	}
	if len(cfg.Actors["user_id"]) != 1 {
		t.Fatal("actors should be preserved")
	}
}

func TestHandler_PostMissingKey(t *testing.T) {
	ff := New(newMemProvider())

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodPost, "/", `{"enabled":true}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_PostInvalidBody(t *testing.T) {
	ff := New(newMemProvider())

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodPost, "/", `not json`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: true, Rollout: 100}))

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodDelete, "/", `{"key":"flag1"}`))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	cfg, _ := ff.GetConfig(ctx, "flag1")
	if cfg.Enabled {
		t.Fatal("flag should be disabled after DELETE")
	}
}

func TestHandler_Patch(t *testing.T) {
	ctx := context.Background()
	ff := New(newMemProvider())
	must(t, ff.SetConfig(ctx, "flag1", FlagConfig{Enabled: false, Rollout: 100}))

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodPatch, "/", `{"key":"flag1"}`))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	cfg, _ := ff.GetConfig(ctx, "flag1")
	if !cfg.Enabled {
		t.Fatal("flag should be enabled after PATCH toggle")
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	ff := New(newMemProvider())

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodPut, "/", ""))

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandler_NonJSONReturns200(t *testing.T) {
	ff := New(newMemProvider())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Basic auth tests ---

func TestHandler_BasicAuth_Unauthorized(t *testing.T) {
	ff := New(newMemProvider(), Config{Username: "admin", Password: "secret"})

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodGet, "/", ""))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate header")
	}
}

func TestHandler_BasicAuth_WrongCredentials(t *testing.T) {
	ff := New(newMemProvider(), Config{Username: "admin", Password: "secret"})

	req := jsonReq(http.MethodGet, "/", "")
	req.SetBasicAuth("admin", "wrong")
	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandler_BasicAuth_Success(t *testing.T) {
	ff := New(newMemProvider(), Config{Username: "admin", Password: "secret"})

	req := jsonReq(http.MethodGet, "/", "")
	req.SetBasicAuth("admin", "secret")
	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandler_NoAuthWhenNotConfigured(t *testing.T) {
	ff := New(newMemProvider())

	w := httptest.NewRecorder()
	ff.GetHandler().ServeHTTP(w, jsonReq(http.MethodGet, "/", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 without auth, got %d", w.Code)
	}
}

// --- hashPercent tests ---

func TestHashPercent_Range(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := hashPercent("flag", string(rune(i)))
		if v < 0 || v >= 100 {
			t.Fatalf("hashPercent out of range: %d", v)
		}
	}
}

func TestHashPercent_Deterministic(t *testing.T) {
	a := hashPercent("flag", "session1")
	b := hashPercent("flag", "session1")
	if a != b {
		t.Fatal("hashPercent should be deterministic")
	}
}
