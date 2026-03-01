package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// adminPage is the self-contained admin HTML dashboard with drag-drop.
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
table{width:100%;border-collapse:collapse;margin-bottom:1rem;table-layout:fixed}
th{text-align:left;padding:.5rem .75rem;color:#64748b;font-size:.8rem;border-bottom:1px solid #334155}
td{padding:.5rem .75rem;border-bottom:1px solid rgba(255,255,255,.05);font-size:.9rem}
td.actions { white-space: nowrap; }
code{background:rgba(0,0,0,.3);padding:2px 6px;border-radius:4px;font-size:.8rem}
.swatch{width:24px;height:24px;border-radius:6px;border:2px solid rgba(255,255,255,.15);display:inline-block;vertical-align:middle;background-size:cover;background-position:center}
.del-btn{background:none;border:none;cursor:pointer;font-size:1.2rem;opacity:.5;padding:4px 8px;border-radius:4px}
.del-btn:hover{opacity:1;background:rgba(239,68,68,.2)}
.status{position:fixed;bottom:1rem;right:1rem;padding:.5rem 1rem;border-radius:8px;font-size:.85rem;background:#22c55e33;color:#22c55e;opacity:0;transition:opacity .3s;z-index:999}
.status.show{opacity:1}
.status.error{background:#ef444433;color:#ef4444}

/* Drag Drop Zones */
.drop-zones { display: flex; gap: 1rem; margin-bottom: 2rem; }
.drop-zone {
  flex: 1; border: 2px dashed #475569; border-radius: 8px; padding: 2rem 1rem;
  text-align: center; color: #94a3b8; transition: .2s; cursor: pointer;
  background: rgba(255,255,255,0.02); position: relative;
}
.drop-zone:hover, .drop-zone.dragover { border-color: #6366f1; background: rgba(99,102,241,0.05); color: #c7d2fe; }
.drop-zone h3 { font-size: 1rem; margin-bottom: .5rem; color: #e2e8f0; }
.drop-zone p { font-size: .85rem; }
.drop-zone input[type=file] { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
.progress-bar { height: 4px; background: #6366f1; width: 0%; position: absolute; bottom: 0; left: 0; border-radius: 0 0 8px 8px; transition: width .2s; }

/* Layout Editor Modal */
.modal-bg { position: fixed; inset: 0; background: rgba(0,0,0,0.8); z-index: 1000; display: none; align-items: center; justify-content: center; backdrop-filter: blur(4px); }
.modal-bg.show { display: flex; }
.modal { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 1.5rem; width: 90vw; max-width: 1200px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.modal-header h2 { font-size: 1.25rem; color: #e2e8f0; }
.modal-close { background: none; border: none; color: #94a3b8; cursor: pointer; font-size: 1.5rem; }
.modal-close:hover { color: #fff; }
.modal-body { display: flex; gap: 2rem; flex: 1; min-height: 0; }

.editor-sidebar { width: 300px; overflow-y: auto; padding-right: 1rem; }
.layout-card { background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.1); border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
.layout-card h4 { margin-bottom: .5rem; font-size: .9rem; color: #94a3b8; }
.input-grid { display: grid; grid-template-columns: 1fr 1fr; gap: .5rem; }
.input-group { display: flex; flex-direction: column; }
.input-group label { font-size: .75rem; color: #64748b; margin-bottom: .25rem; }
.input-group input { background: #0f172a; border: 1px solid #334155; color: #e2e8f0; padding: .4rem; border-radius: 4px; font-family: monospace; font-size: .85rem; width: 100%; }
.input-group input:focus { border-color: #6366f1; outline: none; }

.editor-canvas-wrap { flex: 1; background: #0f172a; border-radius: 8px; border: 1px solid #334155; position: relative; overflow: auto; display: flex; align-items: center; justify-content: center; padding: 2rem; }
.editor-canvas { position: relative; background: #fff; box-shadow: 0 4px 20px rgba(0,0,0,0.5); transform-origin: center; transition: transform 0.1s; }
.editor-canvas img { display: block; width: 100%; height: 100%; pointer-events: none; }
.canvas-box { position: absolute; border: 2px dashed rgba(255,42,109,0.5); background: rgba(255,42,109,0.1); display: flex; align-items: center; justify-content: center; font-weight: bold; color: rgba(255,255,255,0.5); text-shadow: 0 1px 3px rgba(0,0,0,0.8); font-size: 1.5rem; cursor: move; transition: border-color .2s, background .2s, color .2s; }
.canvas-box:hover { border-color: #05d9e8; background: rgba(5,217,232,0.3); color: #fff; z-index: 10; }
.resize-handle { position: absolute; width: 30px; height: 30px; right: -10px; bottom: -10px; cursor: nwse-resize; z-index: 15; }
.resize-handle::after { content: ''; position: absolute; width: 12px; height: 12px; border-right: 4px solid #05d9e8; border-bottom: 4px solid #05d9e8; right: 10px; bottom: 10px; opacity: 0; transition: opacity .2s; }
.canvas-box:hover .resize-handle::after, .resize-handle:hover::after { opacity: 1; }

.btn { background: #6366f1; color: #fff; border: none; padding: .5rem 1rem; border-radius: 6px; cursor: pointer; font-weight: 600; width: 100%; margin-top: 1rem; transition: background .2s; }
.btn:hover { background: #4f46e5; }
.btn-sm { padding: .25rem .5rem; font-size: .8rem; width: auto; margin-top: 0; }
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
  <h2>Upload PNG Frame Context</h2>
  <div class="drop-zones">
    
    <!-- Strip 2x6 -->
    <div class="drop-zone" id="dz1">
      <h3>Photostrip (2x6)</h3>
      <p>Drag PNG here<br>Required size: 600×1800 px</p>
      <input type="file" accept=".png" onchange="handleUpload(this, 'strip_2x6')">
      <div class="progress-bar" id="pb1"></div>
    </div>

    <!-- Postcard 4x6 -->
    <div class="drop-zone" id="dz2">
      <h3>Postcard (4x6)</h3>
      <p>Drag PNG here<br>Required size: 1200×1800 px</p>
      <input type="file" accept=".png" onchange="handleUpload(this, 'postcard_4x6')">
      <div class="progress-bar" id="pb2"></div>
    </div>

  </div>

  <h2>Active Frames</h2>
  <table>
    <thead><tr><th style="width:25%">ID / File</th><th style="width:20%">Label</th><th style="width:15%">Template</th><th style="width:15%">Preview</th><th style="width:25%">Actions</th></tr></thead>
    <tbody id="framesBody"></tbody>
  </table>
</div>

<div id="status" class="status"></div>

<!-- Layout Editor Modal -->
<div class="modal-bg" id="layoutModal">
  <div class="modal">
    <div class="modal-header">
      <h2 id="modalTitle">Edit Layout</h2>
      <button class="modal-close" onclick="closeEditor()">×</button>
    </div>
    <div class="modal-body">
      <div class="editor-sidebar">
        <p style="color:#94a3b8; font-size:0.85rem; margin-bottom:1rem;">Adjust the coordinates for each photo. Values are in pixels relative to the frame's top-left corner.</p>
        <div id="sidebarInputs"></div>
        <button class="btn" onclick="saveLayouts()">Save Changes</button>
      </div>
      <div class="editor-canvas-wrap" id="canvasWrap">
        <!-- Scale controls -->
        <div style="position:absolute; top:1rem; left:1rem; z-index:9; display:flex; gap:.5rem;">
          <button class="btn btn-sm" onclick="zoomCanvas(0.1)">Zoom In</button>
          <button class="btn btn-sm" onclick="zoomCanvas(-0.1)">Zoom Out</button>
        </div>
        <div class="editor-canvas" id="canvas">
          <img id="canvasImg" src="" alt="Frame Preview">
          <div id="canvasOverlays"></div>
        </div>
      </div>
    </div>
  </div>
</div>

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
  tbody.innerHTML = config.frames.map(f => {
    let previewHtml = '';
    if (f.id === 'none') {
      previewHtml = '<div class="swatch" style="background:transparent"></div> <code>none</code>';
    } else {
      // Append timestamp to bust cache for thumbnail
      const ts = new Date().getTime();
      previewHtml = '<div class="swatch" style="background-image:url(\'/frames/' + esc(f.id) + '.png?t=' + ts + '\')"></div> <code>png</code>';
    }

    let templateBadge = f.template ? '<span class="badge on">' + esc(f.template) + '</span>' : '<span class="badge off">Any</span>';

    return '<tr>' +
      '<td><code style="word-break: break-all;">' + esc(f.id) + '</code></td>' +
      '<td>' + esc(f.label) + '</td>' +
      '<td>' + templateBadge + '</td>' +
      '<td>' + previewHtml + '</td>' +
      '<td class="actions" style="display:flex; gap:0.5rem; align-items:center;">' +
      '<button class="btn btn-sm" onclick="openEditor(\'' + esc(f.id) + '\')">Edit Layout</button>' +
      '<button class="del-btn" onclick="removeFrame(\'' + esc(f.id) + '\')" title="Delete">🗑</button>' +
      '</td>' +
      '</tr>';
  }).join('');
}

function esc(s) { const d = document.createElement('div'); d.textContent = s || ''; return d.innerHTML; }

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

async function handleUpload(input, template) {
  if (!input.files || input.files.length === 0) return;
  const file = input.files[0];
  if (!file.name.toLowerCase().endsWith('.png')) {
    flash('Only .png files are supported', true);
    input.value = '';
    return;
  }

  // Get progress bar element
  const pbId = template === 'strip_2x6' ? 'pb1' : 'pb2';
  const pb = document.getElementById(pbId);
  pb.style.width = '50%';

  const formData = new FormData();
  formData.append('frame', file);
  formData.append('template', template);
  formData.append('label', file.name.replace('.png', ''));

  try {
    const res = await fetch('/api/frames/upload', {
      method: 'POST',
      body: formData
    });
    
    if (!res.ok) {
      const errText = await res.text();
      throw new Error(errText);
    }
    
    pb.style.width = '100%';
    flash('Frame uploaded successfully');
    await load();
  } catch (err) {
    flash(err.message || 'Upload failed', true);
  } finally {
    input.value = '';
    setTimeout(() => { pb.style.width = '0%'; }, 500);
  }
}

async function removeFrame(id) {
  if (!confirm('Delete frame ' + id + '?')) return;
  await fetch('/api/frames/' + encodeURIComponent(id), { method: 'DELETE' });
  await load();
  flash('Frame removed');
}

function flash(msg, isError) {
  const s = document.getElementById('status');
  s.textContent = (isError ? '✕ ' : '✓ ') + msg;
  s.className = 'status show' + (isError ? ' error' : '');
  setTimeout(() => s.classList.remove('show'), 3000);
}

// Drag over effects
document.querySelectorAll('.drop-zone').forEach(z => {
  z.addEventListener('dragover', (e) => { e.preventDefault(); z.classList.add('dragover'); });
  z.addEventListener('dragleave', () => z.classList.remove('dragover'));
  z.addEventListener('drop', () => z.classList.remove('dragover'));
});

// --- Layout Editor Logic ---
let activeFrameId = null;
let activeLayouts = [];
let activeTemplateScale = 1;
let activeIndex = -1;

function openEditor(id) {
  const f = config.frames.find(x => x.id === id);
  if (!f) return;
  activeFrameId = id;
  // Deep copy layouts so we don't modify config until save
  activeLayouts = JSON.parse(JSON.stringify(f.layouts || []));
  
  if (activeLayouts.length !== 4) {
    flash('Warning: Frame does not have exactly 4 layouts defined.', true);
  }

  document.getElementById('modalTitle').textContent = 'Edit Layout: ' + f.label;
  
  // Setup canvas
  const canvas = document.getElementById('canvas');
  const img = document.getElementById('canvasImg');
  const ts = new Date().getTime();
  img.src = '/frames/' + f.id + '.png?t=' + ts;
  
  // Real dimensions
  let w = 600, h = 1800;
  if (f.template === 'postcard_4x6') { w = 1200; h = 1800; }
  
  canvas.style.width = w + 'px';
  canvas.style.height = h + 'px';
  
  // Calculate initial zoom to fit wrap
  const wrap = document.getElementById('canvasWrap');
  const scaleX = (wrap.clientWidth - 40) / w;
  const scaleY = (wrap.clientHeight - 40) / h;
  activeTemplateScale = Math.min(scaleX, scaleY, 1);
  updateCanvasTransform();

  renderSidebarInputs();
  renderCanvasOverlays();
  
  document.getElementById('layoutModal').classList.add('show');
}

function closeEditor() {
  document.getElementById('layoutModal').classList.remove('show');
  activeFrameId = null;
}

function zoomCanvas(delta) {
  activeTemplateScale = Math.max(0.1, Math.min(3, activeTemplateScale + delta));
  updateCanvasTransform();
}

function updateCanvasTransform() {
  document.getElementById('canvas').style.transform = 'scale(' + activeTemplateScale + ')';
}

function renderSidebarInputs() {
  const s = document.getElementById('sidebarInputs');
  s.innerHTML = activeLayouts.map((lo, i) =>
    '<div class="layout-card">' +
      '<h4>Photo ' + (i+1) + '</h4>' +
      '<div class="input-grid">' +
        '<div class="input-group"><label>X</label><input type="number" value="' + lo.x + '" onchange="updateVal(' + i + ', \'x\', this.value)"></div>' +
        '<div class="input-group"><label>Y</label><input type="number" value="' + lo.y + '" onchange="updateVal(' + i + ', \'y\', this.value)"></div>' +
        '<div class="input-group"><label>Width</label><input type="number" value="' + lo.width + '" onchange="updateVal(' + i + ', \'width\', this.value)"></div>' +
        '<div class="input-group"><label>Height</label><input type="number" value="' + lo.height + '" onchange="updateVal(' + i + ', \'height\', this.value)"></div>' +
      '</div>' +
    '</div>'
  ).join('');
}

function updateVal(idx, key, val) {
  activeLayouts[idx][key] = parseInt(val, 10) || 0;
  renderCanvasOverlays();
}

function renderCanvasOverlays() {
  const container = document.getElementById('canvasOverlays');
  container.innerHTML = activeLayouts.map((lo, i) =>
    '<div class="canvas-box" ' +
         'style="left:' + lo.x + 'px; top:' + lo.y + 'px; width:' + lo.width + 'px; height:' + lo.height + 'px;" ' +
         'data-idx="' + i + '">' +
      (i+1) +
      '<div class="resize-handle" data-idx="' + i + '"></div>' +
    '</div>'
  ).join('');
}

// --- Drag & Drop / Resize Logic for Canvas ---
let isDragging = false;
let isResizing = false;
let dragStartX, dragStartY;
let initialLayout = null;

const canvasWrap = document.getElementById('canvasWrap');

canvasWrap.addEventListener('mousedown', (e) => {
  if (e.target.classList.contains('resize-handle')) {
    isResizing = true;
    activeIndex = parseInt(e.target.dataset.idx, 10);
    startDrag(e);
  } else if (e.target.classList.contains('canvas-box')) {
    isDragging = true;
    activeIndex = parseInt(e.target.dataset.idx, 10);
    startDrag(e);
  }
});

function startDrag(e) {
  e.preventDefault();
  dragStartX = e.clientX;
  dragStartY = e.clientY;
  initialLayout = { ...activeLayouts[activeIndex] };
  
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', onMouseUp);
}

function onMouseMove(e) {
  if (!isDragging && !isResizing) return;
  e.preventDefault();
  
  const dx = (e.clientX - dragStartX) / activeTemplateScale;
  const dy = (e.clientY - dragStartY) / activeTemplateScale;
  
  const lo = activeLayouts[activeIndex];
  
  if (isDragging) {
    lo.x = Math.round(initialLayout.x + dx);
    lo.y = Math.round(initialLayout.y + dy);
  } else if (isResizing) {
    const ratio = initialLayout.width / initialLayout.height;
    if (Math.abs(dx) > Math.abs(dy)) {
      lo.width = Math.max(10, Math.round(initialLayout.width + dx));
      lo.height = Math.max(10, Math.round(lo.width / ratio));
    } else {
      lo.height = Math.max(10, Math.round(initialLayout.height + dy));
      lo.width = Math.max(10, Math.round(lo.height * ratio));
    }
  }
  
  // Directly manipulate DOM elements for performance during drag instead of re-rendering innerHTML completely
  const box = document.querySelector('.canvas-box[data-idx="' + activeIndex + '"]');
  if (box) {
    box.style.left = lo.x + 'px';
    box.style.top = lo.y + 'px';
    box.style.width = lo.width + 'px';
    box.style.height = lo.height + 'px';
  }
  
  // Debounce sidebar input updates slightly to avoid destroying focus if user is typing
  renderSidebarInputs();
}

function onMouseUp(e) {
  isDragging = false;
  isResizing = false;
  activeIndex = -1;
  document.removeEventListener('mousemove', onMouseMove);
  document.removeEventListener('mouseup', onMouseUp);
  renderCanvasOverlays(); // final precise sync
}

async function saveLayouts() {
  if (!activeFrameId) return;
  try {
    const res = await fetch('/api/frames/' + encodeURIComponent(activeFrameId) + '/layouts', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(activeLayouts)
    });
    if (!res.ok) throw new Error(await res.text());
    flash('Layouts saved successfully!');
    closeEditor();
    await load();
  } catch (err) {
    flash(err.message || 'Failed to save layouts', true);
  }
}

let loadTimer = setInterval(() => {
  if (!activeFrameId) load(); // only auto-refresh if editor is closed
}, 5000);
load();
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

	// 1. Serve the admin dashboard HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(adminPage))
	})

	// 2. Serve uploaded PNG frame files directly to the admin UI for thumbnails
	// E.g., /frames/my_cool_frame.png
	mux.HandleFunc("/frames/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/frames/")
		if !strings.HasSuffix(id, ".png") {
			http.Error(w, "Not found", 404)
			return
		}

		// Map the ID back to a file path
		frameID := strings.TrimSuffix(id, ".png")
		frames := cfg.GetFrames()
		var filePath string
		for _, f := range frames {
			if f.ID == frameID && f.FilePath != "" {
				filePath = f.FilePath
				break
			}
		}

		if filePath == "" {
			http.Error(w, "Frame not found", 404)
			return
		}

		// Disable caching for thumbnail preview to see instant updates
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, filePath)
	})

	// 3. API: GET/POST /api/config
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

	// 4. API: POST /api/frames/upload
	mux.HandleFunc("/api/frames/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", 405)
			return
		}

		// 10 MB max memory for upload parsing
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), 400)
			return
		}

		file, handler, err := r.FormFile("frame")
		if err != nil {
			http.Error(w, "Missing 'frame' file part", 400)
			return
		}
		defer file.Close()

		if !strings.HasSuffix(strings.ToLower(handler.Filename), ".png") {
			http.Error(w, "Only .png files are allowed", 400)
			return
		}

		template := r.FormValue("template") // "strip_2x6" or "postcard_4x6"
		label := r.FormValue("label")
		if label == "" {
			label = handler.Filename
		}

		// Generate an ID from the filename (strip extension, clean spaces)
		id := strings.TrimSuffix(handler.Filename, filepath.Ext(handler.Filename))
		id = strings.ReplaceAll(id, " ", "_")
		id = strings.ToLower(id)

		// Append a timestamp to ID to avoid collisions and cache issues
		id = fmt.Sprintf("%s_%d", id, time.Now().Unix())

		// Ensure the frames directory exists
		if err := os.MkdirAll(cfg.FramesDir(), 0755); err != nil {
			http.Error(w, "Failed to create frames directory", 500)
			return
		}

		// Save the file
		destPath := filepath.Join(cfg.FramesDir(), fmt.Sprintf("%s.png", id))
		destFile, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Failed to save file", 500)
			return
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, file); err != nil {
			http.Error(w, "Failed to write file", 500)
			return
		}

		// Provide default layouts based on template so user has a starting point
		var layouts []PhotoLayout
		if template == "strip_2x6" {
			// 600x1800, stack of 4
			h := 1800 / 4
			layouts = []PhotoLayout{
				{X: 0, Y: 0 * h, Width: 600, Height: h},
				{X: 0, Y: 1 * h, Width: 600, Height: h},
				{X: 0, Y: 2 * h, Width: 600, Height: h},
				{X: 0, Y: 3 * h, Width: 600, Height: h},
			}
		} else if template == "postcard_4x6" {
			// 1200x1800, 2x2 grid
			layouts = []PhotoLayout{
				{X: 0, Y: 0, Width: 600, Height: 900},
				{X: 600, Y: 0, Width: 600, Height: 900},
				{X: 0, Y: 900, Width: 600, Height: 900},
				{X: 600, Y: 900, Width: 600, Height: 900},
			}
		}

		// Add it to the config
		newFrame := Frame{
			ID:       id,
			Label:    label,
			FilePath: destPath,
			Template: template,
			Layouts:  layouts,
		}
		cfg.AddFrame(newFrame)

		w.WriteHeader(http.StatusOK)
	})

	// 5. API: PUT /api/frames/{id}/layouts
	mux.HandleFunc("/api/frames/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/frames/")

		// Handle DELETE /api/frames/{id}
		if r.Method == http.MethodDelete && !strings.Contains(path, "/") {
			id := path
			if id == "" {
				http.Error(w, "missing frame id", 400)
				return
			}

			// Optional: delete the file from disk
			frames := cfg.GetFrames()
			for _, f := range frames {
				if f.ID == id && f.FilePath != "" {
					os.Remove(f.FilePath) // ignoring error if file is already gone
					break
				}
			}

			cfg.RemoveFrame(id)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle PUT /api/frames/{id}/layouts
		if r.Method == http.MethodPut && strings.HasSuffix(path, "/layouts") {
			id := strings.TrimSuffix(path, "/layouts")
			if id == "" {
				http.Error(w, "missing frame id", 400)
				return
			}

			var layouts []PhotoLayout
			if err := json.NewDecoder(r.Body).Decode(&layouts); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			cfg.UpdateFrameLayouts(id, layouts)
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
