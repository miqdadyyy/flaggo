package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/miqdadyyy/flaggo"
	"github.com/miqdadyyy/flaggo/providers/fileprovider"
	"github.com/miqdadyyy/flaggo/web"
)

type userResult struct {
	ID      int
	Enabled bool
}

func main() {
	fp, err := fileprovider.New(fileprovider.Options{
		Path: "storage/flags.json",
	})
	if err != nil {
		log.Fatal(err)
	}

	ff := flaggo.New(fp)

	ctx := context.Background()
	if _, exists := ff.GetConfig(ctx, "new-dashboard"); !exists {
		_ = ff.SetConfig(ctx, "new-dashboard", flaggo.FlagConfig{
			Enabled: true,
			Rollout: 30,
		})
	}

	http.Handle("/flags", web.Handler(ff))
	http.HandleFunc("/simulation", simulationHandler(ff))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/simulation", http.StatusFound)
	})

	log.Println("Server running on http://localhost:3000")
	log.Println("  /simulation  - 10x10 user grid")
	log.Println("  /flags       - flaggo dashboard")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func simulationHandler(ff *flaggo.Flaggo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flagKey := r.URL.Query().Get("key")
		if flagKey == "" {
			flagKey = "new-dashboard"
		}

		results := make([]userResult, 100)
		enabledCount := 0

		for i := 0; i < 100; i++ {
			actor := flaggo.Actor{
				Attributes: map[string]string{
					"user_id":    "user-" + strconv.Itoa(i),
					"session_id": "session-" + strconv.Itoa(i),
				},
			}
			ctx := flaggo.WithActor(context.Background(), actor)
			enabled := ff.IsEnabled(ctx, flagKey)
			results[i] = userResult{ID: i, Enabled: enabled}
			if enabled {
				enabledCount++
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, simulationHTML, flagKey, enabledCount, 100-enabledCount, renderGrid(results))
	}
}

func renderGrid(results []userResult) string {
	html := ""
	for i, r := range results {
		class := "cell off"
		label := "OFF"
		if r.Enabled {
			class = "cell on"
			label = "ON"
		}
		html += fmt.Sprintf(`<div class="%s" title="user-%d (session-%d)"><span class="uid">%d</span><span class="status">%s</span></div>`, class, i, i, i, label)
	}
	return html
}

const simulationHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Flaggo - Feature Flag Simulation</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #0f172a;
  color: #e2e8f0;
  min-height: 100vh;
  padding: 2rem;
}
.container { max-width: 720px; margin: 0 auto; }
h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
.meta {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
  color: #94a3b8;
}
.meta .on-count { color: #4ade80; }
.meta .off-count { color: #f87171; }
.controls {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
  align-items: center;
}
.controls input {
  background: #1e293b;
  border: 1px solid #334155;
  color: #e2e8f0;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  font-size: 0.875rem;
  width: 200px;
}
.controls button {
  background: #3b82f6;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
}
.controls button:hover { background: #2563eb; }
.grid {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 4px;
}
.cell {
  aspect-ratio: 1;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  transition: transform 0.1s;
}
.cell:hover { transform: scale(1.1); z-index: 1; }
.cell.on { background: #166534; border: 1px solid #4ade80; }
.cell.off { background: #1e293b; border: 1px solid #334155; }
.uid { font-weight: 600; }
.status { font-size: 0.6rem; opacity: 0.7; }
.legend {
  display: flex;
  gap: 1.5rem;
  margin-top: 1rem;
  font-size: 0.75rem;
  color: #94a3b8;
}
.legend span { display: flex; align-items: center; gap: 0.4rem; }
.legend .dot {
  width: 10px; height: 10px; border-radius: 3px;
}
.legend .dot.on { background: #166534; border: 1px solid #4ade80; }
.legend .dot.off { background: #1e293b; border: 1px solid #334155; }
a { color: #60a5fa; text-decoration: none; }
a:hover { text-decoration: underline; }
.nav { margin-bottom: 1.5rem; font-size: 0.875rem; }
</style>
</head>
<body>
<div class="container">
  <div class="nav"><a href="/flags">Open Dashboard</a></div>
  <h1>Feature Flag Simulation</h1>
  <div class="meta">
    <span>Flag: <strong>%[1]s</strong></span>
    <span class="on-count">ON: %[2]d</span>
    <span class="off-count">OFF: %[3]d</span>
  </div>
  <div class="controls">
    <form method="GET" action="/simulation" style="display:flex;gap:0.75rem;">
      <input type="text" name="key" placeholder="Flag key..." value="%[1]s" />
      <button type="submit">Simulate</button>
    </form>
  </div>
  <div class="grid">%[4]s</div>
  <div class="legend">
    <span><span class="dot on"></span> Enabled</span>
    <span><span class="dot off"></span> Disabled</span>
  </div>
</div>
</body>
</html>
`