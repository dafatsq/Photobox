package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// adminPage is the self-contained admin HTML dashboard.
const adminPage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Photobox Admin</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Segoe UI',system-ui,sans-serif;background:#0f172a;color:#e2e8f0;padding:2rem}
h1{font-size:1.5rem;margin-bottom:2rem;color:#94a3b8}
h1 span{color:#e2e8f0}
.card{background:#1e293b;border:1px solid #334155;border-radius:12px;padding:1.5rem;margin-bottom:1.5rem}
.card h2{font-size:1rem;color:#64748b;text-transform:uppercase;letter-spacing:.05em;margin-bottom:1rem}
.toggle-row{display:flex;align-items:center;justify-content:space-between;padding:.75rem 1rem;background:rgba(255,255,255,.04);border-radius:8px}
.toggle-row label{cursor:pointer;display:flex;align-items:center;gap:.75rem;font-size:1rem}
.switch{position:relative;width:48px;height:26px;background:#475569;border-radius:13px;transition:.2s;cursor:pointer}
.switch.on{background:#22c55e}
.switch::after{content:'';position:absolute;top:3px;left:3px;width:20px;height:20px;border-radius:50%;background:#fff;transition:.2s}
.switch.on::after{transform:translateX(22px)}
.badge{display:inline-block;padding:2px 8px;border-radius:4px;font-size:.75rem;font-weight:600}
.badge.on{background:#22c55e33;color:#22c55e}
.badge.off{background:#ef444433;color:#ef4444}
table{width:100%;border-collapse:collapse;margin-bottom:1rem}
th{text-align:left;padding:.5rem .75rem;color:#64748b;font-size:.8rem;border-bottom:1px solid #334155}
td{padding:.5rem .75rem;border-bottom:1px solid rgba(255,255,255,.05);font-size:.9rem}
code{background:rgba(0,0,0,.3);padding:2px 6px;border-radius:4px;font-size:.8rem}
.swatch{width:24px;height:24px;border-radius:6px;border:2px solid rgba(255,255,255,.15);display:inline-block;vertical-align:middle}
.del-btn{background:none;border:none;cursor:pointer;font-size:1rem;opacity:.5;padding:4px 8px;border-radius:4px}
.del-btn:hover{opacity:1;background:rgba(239,68,68,.2)}
.add-form{display:flex;gap:.5rem;align-items:center;flex-wrap:wrap}
.add-form input[type=text]{flex:1;min-width:120px;padding:.5rem .75rem;background:#0f172a;border:1px solid #334155;border-radius:6px;color:#e2e8f0;font-size:.85rem}
.add-form input[type=text]:focus{outline:none;border-color:#6366f1}
.add-form input[type=color]{width:36px;height:36px;border:none;border-radius:6px;cursor:pointer;background:none;padding:0}
.add-btn{padding:.5rem 1rem;background:#6366f1;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600;font-size:.85rem;white-space:nowrap}
.add-btn:hover{background:#4f46e5}
.status{position:fixed;bottom:1rem;right:1rem;padding:.5rem 1rem;border-radius:8px;font-size:.85rem;background:#22c55e33;color:#22c55e;opacity:0;transition:opacity .3s}
.status.show{opacity:1}
</style>
</head>
<body>
<h1>⚙️ <span>Photobox Admin</span></h1>

<div class="card">
  <h2>Settings</h2>
  <div class="toggle-row">
    <label>
      Bypass Payment
      <span id="bypassBadge" class="badge off">OFF</span>
    </label>
    <div id="bypassToggle" class="switch" onclick="toggleBypass()"></div>
  </div>
</div>

<div class="card">
  <h2>Frame Manager</h2>
  <table>
    <thead><tr><th>ID</th><th>Label</th><th>Color</th><th></th><th></th></tr></thead>
    <tbody id="framesBody"></tbody>
  </table>
  <form class="add-form" onsubmit="addFrame(event)">
    <input type="text" id="newId" placeholder="ID (e.g. pastel_pink)" required>
    <input type="text" id="newLabel" placeholder="Label" required>
    <input type="color" id="newColor" value="#ffffff">
    <button type="submit" class="add-btn">+ Add Frame</button>
  </form>
</div>

<div id="status" class="status"></div>

<script>
let config = { bypassPayment: false, frames: [] };

async function load() {
  const r = await fetch('/api/config');
  config = await r.json();
  render();
}

function render() {
  const t = document.getElementById('bypassToggle');
  const b = document.getElementById('bypassBadge');
  if (config.bypassPayment) {
    t.classList.add('on');
    b.className = 'badge on'; b.textContent = 'ON';
  } else {
    t.classList.remove('on');
    b.className = 'badge off'; b.textContent = 'OFF';
  }
  const tbody = document.getElementById('framesBody');
  tbody.innerHTML = config.frames.map(f =>
    '<tr>' +
    '<td><code>' + esc(f.id) + '</code></td>' +
    '<td>' + esc(f.label) + '</td>' +
    '<td><div class="swatch" style="background:' + esc(f.color) + '"></div> <code>' + esc(f.color) + '</code></td>' +
    '<td><button class="del-btn" onclick="removeFrame(\'' + esc(f.id) + '\')">🗑</button></td>' +
    '</tr>'
  ).join('');
}

function esc(s) { const d = document.createElement('div'); d.textContent = s; return d.innerHTML; }

async function toggleBypass() {
  config.bypassPayment = !config.bypassPayment;
  await fetch('/api/config', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bypassPayment: config.bypassPayment })
  });
  render();
  flash('Payment bypass ' + (config.bypassPayment ? 'enabled' : 'disabled'));
}

async function addFrame(e) {
  e.preventDefault();
  const id = document.getElementById('newId').value.trim();
  const label = document.getElementById('newLabel').value.trim();
  const color = document.getElementById('newColor').value;
  if (!id || !label) return;
  await fetch('/api/frames', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, label, color })
  });
  document.getElementById('newId').value = '';
  document.getElementById('newLabel').value = '';
  await load();
  flash('Frame "' + label + '" added');
}

async function removeFrame(id) {
  await fetch('/api/frames/' + encodeURIComponent(id), { method: 'DELETE' });
  await load();
  flash('Frame removed');
}

function flash(msg) {
  const s = document.getElementById('status');
  s.textContent = '✓ ' + msg;
  s.classList.add('show');
  setTimeout(() => s.classList.remove('show'), 2000);
}

load();
setInterval(load, 5000); // auto-refresh every 5s
</script>
</body>
</html>`

// configResponse is used for JSON serialization of the admin config.
type configResponse struct {
	BypassPayment bool    `json:"bypassPayment"`
	Frames        []Frame `json:"frames"`
}

// configUpdateRequest is used for JSON deserialization of config updates.
type configUpdateRequest struct {
	BypassPayment bool `json:"bypassPayment"`
}

// StartAdminServer launches the admin HTTP server on the given port.
func StartAdminServer(cfg *AdminConfig, port int) {
	mux := http.NewServeMux()

	// Serve the admin dashboard HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(adminPage))
	})

	// GET /api/config — return current config
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := configResponse{
				BypassPayment: cfg.GetBypassPayment(),
				Frames:        cfg.GetFrames(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case http.MethodPost:
			var body configUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			cfg.SetBypassPayment(body.BypassPayment)
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", 405)
		}
	})

	// POST /api/frames — add a frame
	// DELETE /api/frames/{id} — remove a frame
	mux.HandleFunc("/api/frames", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var f Frame
			if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			cfg.AddFrame(f)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Method not allowed", 405)
	})

	mux.HandleFunc("/api/frames/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			id := strings.TrimPrefix(r.URL.Path, "/api/frames/")
			if id == "" {
				http.Error(w, "missing frame id", 400)
				return
			}
			cfg.RemoveFrame(id)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Method not allowed", 405)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("[Admin] Dashboard available at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("[Admin] Server error: %v\n", err)
	}
}
