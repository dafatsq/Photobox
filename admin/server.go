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
<title>Photobox Admin Workspace</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
:root {
  --bg-deep: #0a0a0f;
  --bg-surface: #13141c;
  --bg-card: rgba(255, 255, 255, 0.02);
  --border: rgba(255, 255, 255, 0.08);
  --border-hover: rgba(255, 255, 255, 0.15);
  --primary: #6366f1;
  --primary-hover: #4f46e5;
  --accent: #0ea5e9;
  --text-main: #f8fafc;
  --text-muted: #94a3b8;
  --success: #10b981;
  --danger: #ef4444;
  --danger-bg: rgba(239, 68, 68, 0.1);
  --radius-lg: 16px;
  --radius-md: 10px;
  --radius-sm: 6px;
}

* { box-sizing: border-box; margin: 0; padding: 0; }
body { 
  font-family: 'Inter', system-ui, -apple-system, sans-serif; 
  background: var(--bg-deep); 
  color: var(--text-main); 
  line-height: 1.5;
  min-height: 100vh;
  background-image: 
    radial-gradient(circle at 15% 50%, rgba(99, 102, 241, 0.05), transparent 25%),
    radial-gradient(circle at 85% 30%, rgba(14, 165, 233, 0.05), transparent 25%);
  background-attachment: fixed;
}

/* Page Layout */
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 3rem 2rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 2.5rem;
  border-bottom: 1px solid var(--border);
  padding-bottom: 1.5rem;
}

.header-brand h1 {
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  background: linear-gradient(135deg, #fff 0%, #cbd5e1 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-brand h1::before {
  content: '';
  display: inline-block;
  width: 24px; height: 24px;
  background: linear-gradient(135deg, var(--primary), var(--accent));
  border-radius: 6px;
  box-shadow: 0 0 15px rgba(99, 102, 241, 0.4);
}

.header-brand p {
  color: var(--text-muted);
  font-size: 0.95rem;
  margin-top: 0.25rem;
}

/* Base Card Style */
.card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 2rem;
  margin-bottom: 2rem;
  box-shadow: 0 10px 30px -10px rgba(0,0,0,0.5);
}

.card-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-main);
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* Settings Grid */
.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
}

.setting-item {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: all 0.2s ease;
}

.setting-item:hover {
  border-color: var(--border-hover);
  background: rgba(255, 255, 255, 0.03);
}

.setting-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.setting-info h3 {
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--text-main);
}

.setting-info p {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}

/* Toggles and Inputs */
.switch { position: relative; width: 44px; height: 24px; background: #334155; border-radius: 12px; cursor: pointer; transition: 0.3s; }
.switch.on { background: var(--success); box-shadow: 0 0 10px rgba(16, 185, 129, 0.3); }
.switch::after { content: ''; position: absolute; top: 2px; left: 2px; width: 20px; height: 20px; border-radius: 50%; background: #fff; transition: 0.3s cubic-bezier(0.4, 0.0, 0.2, 1); box-shadow: 0 2px 4px rgba(0,0,0,0.2); }
.switch.on::after { transform: translateX(20px); }

select, input[type="text"], input[type="number"] {
  width: 100%;
  background: rgba(0,0,0,0.2);
  border: 1px solid #334155;
  color: var(--text-main);
  padding: 0.6rem 0.8rem;
  border-radius: var(--radius-sm);
  font-family: inherit;
  font-size: 0.9rem;
  transition: all 0.2s;
  outline: none;
}

select option {
  background: #1e293b;
  color: var(--text-main);
}

select:focus, input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
}

/* Drag Drop Zones */
.drop-zones { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; }
.drop-zone {
  position: relative;
  background: var(--bg-card);
  border: 2px dashed #334155;
  border-radius: var(--radius-md);
  padding: 2.5rem 1.5rem;
  text-align: center;
  transition: all 0.2s ease;
  cursor: pointer;
  overflow: hidden;
}

.drop-zone:hover, .drop-zone.dragover {
  border-color: var(--primary);
  background: rgba(99, 102, 241, 0.05);
  transform: translateY(-2px);
}

.dz-icon { font-size: 2rem; margin-bottom: 1rem; opacity: 0.8; }
.dz-title { font-size: 1rem; font-weight: 600; color: var(--text-main); margin-bottom: 0.4rem; }
.dz-desc { font-size: 0.85rem; color: var(--text-muted); }
.dz-req { display: inline-block; margin-top: 0.75rem; font-size: 0.75rem; padding: 0.25em 0.75em; border-radius: 100px; background: rgba(255,255,255,0.05); border: 1px solid var(--border); }
.drop-zone input[type=file] { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; z-index: 10; }
.progress-bar { height: 4px; background: linear-gradient(90deg, var(--primary), var(--accent)); width: 0%; position: absolute; bottom: 0; left: 0; transition: width 0.3s ease; }

/* Frames Grid */
.frames-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 1.5rem; }
.frame-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.frame-card:hover {
  transform: translateY(-4px);
  border-color: rgba(255,255,255,0.2);
  box-shadow: 0 10px 20px -10px rgba(0,0,0,0.5);
  background: rgba(255,255,255,0.04);
}
.frame-preview {
  height: 260px;
  background: rgba(0,0,0,0.4);
  border-radius: var(--radius-sm);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  position: relative;
}
.frame-preview img { height: 100%; width: 100%; object-fit: contain; padding: 0.5rem; filter: drop-shadow(0 4px 6px rgba(0,0,0,0.3)); }
.frame-meta {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
.frame-title { font-size: 0.95rem; font-weight: 600; color: var(--text-main); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.frame-badge { align-self: flex-start; font-size: 0.7rem; font-weight: 600; padding: 2px 8px; border-radius: 4px; background: rgba(99, 102, 241, 0.15); color: #818cf8; border: 1px solid rgba(99, 102, 241, 0.3); }
.frame-actions { display: flex; gap: 0.5rem; margin-top: auto; }

/* Buttons */
.btn {
  background: var(--primary);
  color: #fff;
  border: 1px solid rgba(255,255,255,0.1);
  padding: 0.5rem 1rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-weight: 500;
  font-size: 0.85rem;
  font-family: inherit;
  transition: all 0.2s;
  flex-shrink: 0;
  text-align: center;
}
.btn-sm { flex: 0 0 auto !important; width: auto !important; padding: 0.25rem 0.75rem !important; font-size: 0.78rem !important; }
.btn:hover { background: var(--primary-hover); transform: translateY(-1px); }
.btn-icon {
  background: var(--danger-bg);
  color: var(--danger);
  border: 1px solid rgba(239, 68, 68, 0.2);
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
  display: flex; align-items: center; justify-content: center;
}
.btn-icon:hover { background: rgba(239, 68, 68, 0.2); color: #f87171; border-color: rgba(239, 68, 68, 0.4); }

/* Status Toast */
.status {
  position: fixed; bottom: 2rem; right: 2rem;
  padding: 1rem 1.5rem; border-radius: var(--radius-md); font-size: 0.9rem; font-weight: 500;
  background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.3); color: #34d399;
  backdrop-filter: blur(8px);
  transform: translateY(20px); opacity: 0; transition: all 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55);
  z-index: 9999; box-shadow: 0 10px 25px -5px rgba(0,0,0,0.5);
  display: flex; align-items: center; gap: 0.75rem;
}
.status.show { transform: translateY(0); opacity: 1; }
.status.error { background: rgba(239, 68, 68, 0.1); border-color: rgba(239, 68, 68, 0.3); color: #f87171; }

/* Modal Redesign */
.modal-bg { position: fixed; inset: 0; background: rgba(0,0,0,0.7); z-index: 1000; display: none; align-items: center; justify-content: center; backdrop-filter: blur(8px); }
.modal-bg.show { display: flex; animation: fadeIn 0.2s ease; }
.modal { background: var(--bg-surface); border: 1px solid var(--border); box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5); border-radius: var(--radius-lg); width: 95vw; max-width: 1400px; height: 90vh; display: flex; flex-direction: column; overflow: hidden; }
.modal-header { padding: 1.5rem 2rem; border-bottom: 1px solid var(--border); display: flex; justify-content: space-between; align-items: center; background: rgba(0,0,0,0.2); }
.modal-header h2 { font-size: 1.2rem; font-weight: 600; }
.modal-close { background: none; border: none; color: var(--text-muted); cursor: pointer; font-size: 1.5rem; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 50%; transition: 0.2s; }
.modal-close:hover { background: rgba(255,255,255,0.1); color: #fff; }
.modal-body { display: flex; flex: 1; min-height: 0; }
.editor-sidebar { width: 320px; background: rgba(0,0,0,0.2); border-right: 1px solid var(--border); padding: 1.5rem; display: flex; flex-direction: column; overflow-y: auto; }
.layout-card { background: rgba(255,255,255,0.03); border: 1px solid var(--border); border-radius: var(--radius-md); padding: 1.25rem; margin-bottom: 1rem; }
.layout-card h4 { font-size: 0.9rem; color: var(--text-main); margin-bottom: 1rem; display: flex; justify-content: space-between; align-items: center; }
.input-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.input-group label { display: block; font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin-bottom: 0.3rem; }
.input-group input { padding: 0.4rem; font-size: 0.85rem; text-align: center; }
.editor-canvas-wrap { flex: 1; position: relative; overflow: auto; background: #000; padding: 2rem; }
.editor-canvas { position: relative; background: transparent; margin: 0 auto; }
.editor-canvas img { display: block; width: 100%; height: auto; pointer-events: none; }
.canvas-box { position: absolute; border: 2px dashed rgba(99, 102, 241, 0.8); background: rgba(99, 102, 241, 0.15); display: flex; align-items: center; justify-content: center; font-weight: 700; color: rgba(255,255,255,0.6); font-size: 2rem; cursor: move; transition: border-color .2s, background .2s, color .2s; }
.canvas-box:hover { border-color: #0ea5e9; background: rgba(14, 165, 233, 0.25); color: #fff; z-index: 10; box-shadow: inset 0 0 0 2px rgba(255,255,255,0.2); }
.resize-handle { position: absolute; width: 14px; height: 14px; z-index: 15; background: #0ea5e9; border: 2px solid #fff; border-radius: 50%; opacity: 0; transition: opacity .2s, transform .2s; box-shadow: 0 2px 5px rgba(0,0,0,0.5); }
.canvas-box:hover .resize-handle { opacity: 1; }
.resize-handle:hover { transform: scale(1.4); }
.resize-handle.tl { top: -7px; left: -7px; cursor: nw-resize; }
.resize-handle.tr { top: -7px; right: -7px; cursor: ne-resize; }
.resize-handle.bl { bottom: -7px; left: -7px; cursor: sw-resize; }
.resize-handle.br { bottom: -7px; right: -7px; cursor: se-resize; }
.toolbar { position: absolute; top: 1.5rem; left: 1.5rem; z-index: 50; display: flex; gap: 0.5rem; background: rgba(15, 23, 42, 0.8); backdrop-filter: blur(8px); padding: 0.5rem; border-radius: var(--radius-md); border: 1px solid var(--border); }

@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
</head>
<body>

<div class="container">
  <header class="header">
    <div class="header-brand">
      <h1>Photobox Workspace</h1>
      <p>Configure hardware, capture settings, and design layouts.</p>
    </div>
  </header>

  <div class="card">
    <h2 class="card-title">Station Preferences</h2>
    <div class="settings-grid">
      
      <!-- Payment Bypass -->
      <div class="setting-item">
        <div class="setting-header">
          <div class="setting-info">
            <h3>Bypass Payment</h3>
            <p>Allow users to capture without QRIS</p>
          </div>
          <div id="bypassToggle" class="switch" onclick="toggleBypass()"></div>
        </div>
      </div>

      <!-- Fullscreen -->
      <div class="setting-item">
        <div class="setting-header">
          <div class="setting-info">
            <h3>Kiosk Fullscreen</h3>
            <p>Lock frontend to borderless fullscreen</p>
          </div>
          <div id="fullscreenToggle" class="switch" onclick="toggleFullscreen()"></div>
        </div>
      </div>

      <!-- Camera Mode -->
      <div class="setting-item">
        <div class="setting-info" style="margin-bottom: 0.75rem;">
          <h3>Camera Mode</h3>
          <p>Select capture hardware type</p>
        </div>
        <select id="cameraTypeSelect" onchange="updateCameraSettings()">
          <option value="dslr">DSLR Camera</option>
          <option value="webcam">Webcam</option>
        </select>
      </div>

      <!-- DSLR Driver -->
      <div class="setting-item" id="dslrModeRow">
        <div class="setting-info" style="margin-bottom: 0.75rem;">
          <h3>DSLR Integration</h3>
          <p style="color:#f59e0b">Restart app to apply changes</p>
        </div>
        <select id="dslrModeSelect" onchange="updateCameraSettings()">
          <option value="integrated">Integrated Engine (DSLRBridge)</option>
          <option value="legacy">Legacy Mode (DigiCamControl GUI)</option>
        </select>
      </div>

      <!-- Webcam Device -->
      <div class="setting-item" id="webcamRow" style="display:none;">
        <div class="setting-info" style="margin-bottom: 0.75rem;">
          <h3>Webcam Device</h3>
          <p>Select video source</p>
        </div>
        <select id="webcamSelect" onchange="updateCameraSettings()">
          <option value="">Default System Webcam</option>
        </select>
      </div>

    </div>
  </div>

  <style>
    #r2Card summary::-webkit-details-marker { display: none; } /* Hide default Safari arrow */
    .r2-chevron { transition: transform 0.3s ease; color: var(--text-muted); }
    #r2Card details[open] .r2-chevron { transform: rotate(180deg); }
  </style>
  <div class="card" id="r2Card">
    <details style="cursor: pointer;">
      <summary style="outline:none; list-style:none; display:flex; justify-content:space-between; align-items:center;">
        <h2 class="card-title" style="margin:0;">Photo Sharing (Cloudflare R2)</h2>
        <svg class="r2-chevron" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>
      </summary>
      <div style="margin-top: 1.5rem; cursor: default;">
        <p style="color:var(--text-muted); font-size:0.9rem; margin-bottom:1.5rem;">After a session, the app uploads the photo to your R2 bucket and shows a QR code so guests can scan and download their photo. Each admin enters their own Cloudflare credentials.</p>
        <style>
          .info-icon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            color: var(--text-muted);
            cursor: help;
            position: relative;
            margin-left: 0.4rem;
          }
          .info-icon svg {
            width: 14px;
            height: 14px;
          }
          .info-tooltip {
            visibility: hidden;
            opacity: 0;
            width: 340px;
            background-color: #1e293b;
            color: #f8fafc;
            text-align: left;
            border-radius: 8px;
            padding: 1rem;
            position: absolute;
            z-index: 50;
            bottom: 125%;
            left: 50%;
            transform: translateX(-50%);
            transition: opacity 0.2s, visibility 0.2s;
            font-size: 0.8rem;
            line-height: 1.5;
            font-weight: normal;
            text-transform: none;
            letter-spacing: normal;
            border: 1px solid #334155;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.8);
          }
          .info-tooltip::after {
            content: "";
            position: absolute;
            top: 100%;
            left: 50%;
            margin-left: -6px;
            border-width: 6px;
            border-style: solid;
            border-color: #334155 transparent transparent transparent;
          }
          .info-icon:hover .info-tooltip {
            visibility: visible;
            opacity: 1;
          }
          .info-tooltip ol {
            margin: 0.5rem 0 0 0;
            padding-left: 1.25rem;
          }
          .info-tooltip li {
            margin-bottom: 0.25rem;
          }
        </style>
        <div class="settings-grid">
          <div class="setting-item" style="flex-direction:column; align-items:flex-start; gap:0.5rem;">
            <label style="font-size:0.8rem; text-transform:uppercase; letter-spacing:.05em; color:var(--text-muted); display:flex; align-items:center;">
              Account ID
              <span class="info-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4"></path><path d="M12 8h.01"></path></svg>
                <div class="info-tooltip">
                  <strong>How to get your Account ID:</strong>
                  <ol>
                    <li>Log in to your Cloudflare dashboard at dash.cloudflare.com.</li>
                    <li>Look at the URL in your web browser's address bar.</li>
                    <li>The URL will look like <code>https://dash.cloudflare.com/2a63bbbcc87...</code></li>
                    <li>The long string between <code>dash.cloudflare.com/</code> and the next <code>/</code> is your <strong>Account ID</strong>.</li>
                  </ol>
                </div>
              </span>
            </label>
            <input type="text" id="r2AccountId" placeholder="e.g. abc123def456..." style="width:100%;">
          </div>
          <div class="setting-item" style="flex-direction:column; align-items:flex-start; gap:0.5rem;">
            <label style="font-size:0.8rem; text-transform:uppercase; letter-spacing:.05em; color:var(--text-muted); display:flex; align-items:center;">
              Access Key ID
              <span class="info-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4"></path><path d="M12 8h.01"></path></svg>
                <div class="info-tooltip">
                  <strong>How to get your Access Key & Secret:</strong>
                  <ol>
                    <li>On the left sidebar, click <strong>R2 object storage</strong>.</li>
                    <li>On the right side, click <strong>Manage R2 API Tokens</strong>.</li>
                    <li>Click <strong>Create Account API token</strong>.</li>
                    <li>Set Permissions to <strong>Admin Read & Write</strong> (or Object Read & Write).</li>
                    <li>Scroll to the bottom and click <strong>Create API Token</strong>.</li>
                    <li>Copy the <strong>Access Key ID</strong> shown here.</li>
                  </ol>
                </div>
              </span>
            </label>
            <input type="text" id="r2AccessKeyId" placeholder="R2 API token key ID..." style="width:100%;">
          </div>
          <div class="setting-item" style="flex-direction:column; align-items:flex-start; gap:0.5rem;">
            <label style="font-size:0.8rem; text-transform:uppercase; letter-spacing:.05em; color:var(--text-muted); display:flex; align-items:center;">
              Secret Access Key
              <span class="info-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4"></path><path d="M12 8h.01"></path></svg>
                <div class="info-tooltip">
                  <strong>How to get your Secret Access Key:</strong>
                  <ol>
                    <li>Follow the steps above to create an API token.</li>
                    <li>Copy the <strong>Secret Access Key</strong> shown on that page.</li>
                    <li style="color:#f87171;"><strong>WARNING:</strong> Cloudflare will only show this to you ONCE. Do not close the tab until you save it here! If you lose it, create a new token.</li>
                  </ol>
                </div>
              </span>
            </label>
            <input type="password" id="r2SecretKey" placeholder="Leave blank to keep current..." style="width:100%;">
          </div>
          <div class="setting-item" style="flex-direction:column; align-items:flex-start; gap:0.5rem;">
            <label style="font-size:0.8rem; text-transform:uppercase; letter-spacing:.05em; color:var(--text-muted); display:flex; align-items:center;">
              Bucket Name
              <span class="info-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4"></path><path d="M12 8h.01"></path></svg>
                <div class="info-tooltip">
                  <strong>How to create a Bucket Name:</strong>
                  <ol>
                    <li>On the left sidebar, click <strong>R2 object storage</strong>.</li>
                    <li>Click <strong>Create bucket</strong> on the right.</li>
                    <li>Enter a name (e.g. <code>photobox</code>). Must be lowercase.</li>
                    <li>Leave default settings and click <strong>Create bucket</strong>.</li>
                    <li>Type the exact name you used here.</li>
                  </ol>
                </div>
              </span>
            </label>
            <input type="text" id="r2BucketName" placeholder="my-photobox-bucket" style="width:100%;">
          </div>
          <div class="setting-item" style="flex-direction:column; align-items:flex-start; gap:0.5rem; grid-column: 1 / -1;">
            <label style="font-size:0.8rem; text-transform:uppercase; letter-spacing:.05em; color:var(--text-muted); display:flex; align-items:center;">
              Public Base URL
              <span class="info-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4"></path><path d="M12 8h.01"></path></svg>
                <div class="info-tooltip">
                  <strong>How to get your Public Base URL:</strong>
                  <ol>
                    <li>Click into the bucket you just created.</li>
                    <li>Go to the <strong>Settings</strong> tab.</li>
                    <li>Under the <strong>Public Access</strong> section, click <strong>Allow Access</strong> and type <code>allow</code> to confirm.</li>
                    <li>A green URL will appear (e.g. <code>https://pub-044dfe1...r2.dev</code>).</li>
                    <li>Copy that URL here <strong>(do NOT include a trailing slash <code>/</code>)</strong>.</li>
                  </ol>
                </div>
              </span>
            </label>
            <input type="text" id="r2PublicBaseUrl" placeholder="https://pub-xxx.r2.dev or https://photos.yourdomain.com" style="width:100%;">
          </div>
        </div>
        <button class="btn" onclick="saveR2Config()" style="margin-top:1.5rem; width:auto; flex:none; padding:0.6rem 2rem;">Save R2 Settings</button>
      </div>
    </details>
  </div>

  <div class="card">
    <h2 class="card-title">Frame Management</h2>
    <p style="color:var(--text-muted); font-size:0.9rem; margin-bottom:1.5rem;">Upload transparent PNG overlays for the photobooth. The system will automatically detect the template type based on the drop zone.</p>
    
    <div class="drop-zones">
      <div class="drop-zone" id="dz1" ondrop="handleDrop(event, '3strip_2x6')" ondragover="event.preventDefault()">
        <div class="dz-icon">🎞️</div>
        <h3 class="dz-title">Photostrip (2x6)</h3>
        <p class="dz-desc">Vertical classic strip format</p>
        <span class="dz-req">600 × 1800 px PNG</span>
        <input type="file" accept=".png" onchange="handleUpload(this, '3strip_2x6')">
        <div class="progress-bar" id="pb1"></div>
      </div>

      <div class="drop-zone" id="dz3" ondrop="handleDrop(event, '6strip_4x6')" ondragover="event.preventDefault()">
        <div class="dz-icon">🎞️🎞️</div>
        <h3 class="dz-title">Double Strip (4x6)</h3>
        <p class="dz-desc">Two side-by-side vertical strips</p>
        <span class="dz-req">1200 × 1800 px PNG</span>
        <input type="file" accept=".png" onchange="handleUpload(this, '6strip_4x6')">
        <div class="progress-bar" id="pb3"></div>
      </div>

      <div class="drop-zone" id="dz2" ondrop="handleDrop(event, '4postcard_4x6')" ondragover="event.preventDefault()">
        <h3 class="dz-title">Postcard (4x6)</h3>
        <p class="dz-desc">2x2 grid landscape format</p>
        <span class="dz-req">1200 × 1800 px PNG</span>
        <input type="file" accept=".png" onchange="handleUpload(this, '4postcard_4x6')">
        <div class="progress-bar" id="pb2"></div>
      </div>
    </div>
  </div>

  <div class="card">
    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:1.5rem;">
      <h2 class="card-title" style="margin-bottom:0;">Active Gallery</h2>
      <div style="display:flex; gap:1rem;">
        <input type="text" id="frameSearch" placeholder="Search frames..." style="width:200px; padding:0.4rem 0.8rem;" oninput="render()">
        <select id="frameTemplateFilter" style="width:180px; padding:0.4rem 0.8rem;" onchange="render()">
          <option value="all">All Types</option>
          <option value="3strip_2x6">Photostrip (2x6)</option>
          <option value="6strip_4x6">Double Strip (4x6)</option>
          <option value="4postcard_4x6">Postcard (4x6)</option>
        </select>
      </div>
    </div>
    <div id="framesGrid" class="frames-grid"></div>
  </div>

</div>


<div id="status" class="status"></div>

<!-- Layout Editor Modal -->
<div class="modal-bg" id="layoutModal">
  <div class="modal">
    <div class="modal-header">
      <h2 id="modalTitle">Layout Mapping</h2>
      <button class="modal-close" onclick="closeEditor()">×</button>
    </div>
    <div class="modal-body">
      <div class="editor-sidebar">
        <p style="color:var(--text-muted); font-size:0.85rem; margin-bottom:1.5rem; line-height:1.4;">Adjust capture zones. Drag the boxes on the canvas, drag corners to resize, or type precise pixel values below.</p>
        <div id="sidebarInputs"></div>
        <div style="flex: 1"></div>
        <button class="btn" onclick="saveLayouts()" style="padding: 1rem; font-size: 0.95rem; margin-top:2rem;">💾 Save Layout Settings</button>
      </div>
      <div class="editor-canvas-wrap" id="canvasWrap">
        <div class="toolbar">
          <button class="btn btn-sm" onclick="zoomCanvas(0.1)" style="min-width:40px; background:rgba(255,255,255,0.1)">+</button>
          <button class="btn btn-sm" onclick="zoomCanvas(-0.1)" style="min-width:40px; background:rgba(255,255,255,0.1)">-</button>
          <span style="display:flex; align-items:center; padding:0 0.5rem; font-size:0.8rem; color:var(--text-muted)" id="zoomIndicator">100%</span>
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
let config = { bypassPayment: false, fullscreen: false, frames: [] };
let lastConfigJson = "";

async function load() {
  const r = await fetch('/api/config');
  config = await r.json();
  render();
}

async function updateCameraSettings() {
  config.cameraType = document.getElementById('cameraTypeSelect').value;
  config.webcamId = document.getElementById('webcamSelect').value;
  config.dslrMode = document.getElementById('dslrModeSelect').value;
  await fetch('/api/config', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bypassPayment: config.bypassPayment, cameraType: config.cameraType, webcamId: config.webcamId, fullscreen: config.fullscreen, dslrMode: config.dslrMode })
  });
  render();
  flash('Camera settings saved — restart app to apply DSLR driver change');
}

function render() {
  var sqInput = document.getElementById('frameSearch');
  var tfSelect = document.getElementById('frameTemplateFilter');
  var currentSearch = (sqInput ? sqInput.value : '').toLowerCase();
  var currentType = tfSelect ? tfSelect.value : 'all';
  
  // Combine config + UI state for change detection to prevent flickering
  var stateSnapshot = JSON.stringify({ config, currentSearch, currentType });
  if (stateSnapshot === lastConfigJson) return; 
  lastConfigJson = stateSnapshot;

  const t = document.getElementById('bypassToggle');
  const b = document.getElementById('bypassBadge');
  if (config.bypassPayment) {
    t.classList.add('on');
  } else {
    t.classList.remove('on');
  }
  
  const ft = document.getElementById('fullscreenToggle');
  if (config.fullscreen) {
    ft.classList.add('on');
  } else {
    ft.classList.remove('on');
  }

  const cType = document.getElementById('cameraTypeSelect');
  cType.value = config.cameraType || 'dslr';
  
  const dslrModeRow = document.getElementById('dslrModeRow');
  const dslrModeSel = document.getElementById('dslrModeSelect');
  dslrModeRow.style.display = config.cameraType === 'dslr' ? 'flex' : 'none';
  dslrModeSel.value = config.dslrMode || 'integrated';
  
  const wRow = document.getElementById('webcamRow');
  const wSel = document.getElementById('webcamSelect');
  
  if (config.cameraType === 'webcam') {
    wRow.style.display = 'flex';
    // Preserve current selection if possible, otherwise build
    const tempVal = config.webcamId || '';
    wSel.innerHTML = '<option value="">Default Webcam</option>';
    if (config.availableCameras && config.availableCameras.length > 0) {
       config.availableCameras.forEach(cam => {
         wSel.innerHTML += '<option value="' + esc(cam.id) + '">' + esc(cam.label || 'Camera ' + cam.id.substring(0,8)) + '</option>';
       });
    }
    wSel.value = tempVal;
  } else {
    wRow.style.display = 'none';
  }

  // Populate R2 fields (secret key is never shown for security)
  var r2AccId = document.getElementById('r2AccountId');
  var r2AkId = document.getElementById('r2AccessKeyId');
  var r2Bkt = document.getElementById('r2BucketName');
  var r2Url = document.getElementById('r2PublicBaseUrl');
  if (r2AccId && !r2AccId.matches(':focus')) r2AccId.value = config.r2AccountId || '';
  if (r2AkId && !r2AkId.matches(':focus')) r2AkId.value = config.r2AccessKeyId || '';
  if (r2Bkt && !r2Bkt.matches(':focus')) r2Bkt.value = config.r2BucketName || '';
  if (r2Url && !r2Url.matches(':focus')) r2Url.value = config.r2PublicBaseUrl || '';
  
  var sqInput = document.getElementById('frameSearch');
  var tfSelect = document.getElementById('frameTemplateFilter');
  var searchQuery = (sqInput ? sqInput.value : '').toLowerCase();
  var typeFilter = tfSelect ? tfSelect.value : 'all';

  const filteredFrames = config.frames.filter(f => {
    if (f.id === 'none') return true;
    var label = f.label || '';
    var matchesSearch = label.toLowerCase().includes(searchQuery);
    var matchesType = typeFilter === 'all' || f.template === typeFilter;
    return matchesSearch && matchesType;
  });

  const grid = document.getElementById('framesGrid');
  grid.innerHTML = filteredFrames.map(f => {
    let previewHtml = '';
    // The image id already contains a timestamp from the backend so cache busting ?t= is not needed anymore
    if (f.id === 'none') {
      previewHtml = '<div class="frame-preview" style="background:rgba(0,0,0,0.5); color:#64748b; font-size:2rem;">🚫</div>';
    } else {
      previewHtml = '<div class="frame-preview"><img src="/frames/' + esc(f.id) + '.png" alt="' + esc(f.label) + '"></div>';
    }

    let templateBadge = f.template ? '<span class="frame-badge">' + esc(f.template) + '</span>' : '<span class="frame-badge">Any</span>';

    return '<div class="frame-card">' +
      previewHtml +
      '<div class="frame-meta">' +
        '<h4 class="frame-title" title="' + esc(f.label) + '">' + esc(f.label) + '</h4>' +
        templateBadge +
      '</div>' +
      '<div class="frame-actions">' +
        '<button class="btn" onclick="openEditor(\'' + esc(f.id) + '\')">Configure Layout</button>' +
        '<button class="btn-icon" onclick="removeFrame(\'' + esc(f.id) + '\')">' +
          '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path></svg>' +
        '</button>' +
      '</div>' +
    '</div>';
  }).join('');
}

function esc(s) { const d = document.createElement('div'); d.textContent = s || ''; return d.innerHTML; }

async function toggleFullscreen() {
  config.fullscreen = !config.fullscreen;
  await fetch('/api/config', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bypassPayment: config.bypassPayment, cameraType: config.cameraType, webcamId: config.webcamId, fullscreen: config.fullscreen, dslrMode: config.dslrMode })
  });
  render();
  flash('Fullscreen ' + (config.fullscreen ? 'enabled' : 'disabled'));
}

async function toggleBypass() {
  config.bypassPayment = !config.bypassPayment;
  await fetch('/api/config', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bypassPayment: config.bypassPayment, cameraType: config.cameraType, webcamId: config.webcamId, dslrMode: config.dslrMode })
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
  const pbMap = { '3strip_2x6': 'pb1', '6strip_4x6': 'pb3', '4postcard_4x6': 'pb2' };
  const pbId = pbMap[template] || 'pb1';
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
let activeCanvasW = 600;
let activeCanvasH = 1800;

function openEditor(id) {
  const f = config.frames.find(x => x.id === id);
  if (!f) return;
  activeFrameId = id;
  // Deep copy layouts so we don't modify config until save
  activeLayouts = JSON.parse(JSON.stringify(f.layouts || []));
  
  const slotMap = { '3strip_2x6': 3, '6strip_4x6': 6, '4postcard_4x6': 4 };
  const expectedSlots = slotMap[f.template] || 4;
  if (activeLayouts.length !== expectedSlots) {
    flash('Warning: Frame does not have exactly ' + expectedSlots + ' layouts defined.', true);
  }

  document.getElementById('modalTitle').textContent = 'Edit Layout: ' + f.label;
  
  // Setup canvas
  const canvas = document.getElementById('canvas');
  const img = document.getElementById('canvasImg');
  img.src = '/frames/' + f.id + '.png';
  
  // Real dimensions
  let w = 600, h = 1800;
  if (f.template === '4postcard_4x6' || f.template === '6strip_4x6') { w = 1200; h = 1800; }
  
  activeCanvasW = w;
  activeCanvasH = h;
  
  document.getElementById('layoutModal').classList.add('show');

  // Need to wait until modal is visible so clientWidth isn't 0
  setTimeout(() => {
    const wrap = document.getElementById('canvasWrap');
    let wrapW = wrap.clientWidth - 60;
    let wrapH = wrap.clientHeight - 60;
    if (wrapW < 100) wrapW = window.innerWidth - 400;
    if (wrapH < 100) wrapH = window.innerHeight - 200;
    
    // Auto-fit: scale so the canvas fits within the viewport
    const scaleW = wrapW / w;
    const scaleH = wrapH / h;
    activeTemplateScale = Math.min(scaleW, scaleH, 1.0);
    activeTemplateScale = Math.max(0.15, Math.round(activeTemplateScale * 20) / 20);
    updateCanvasTransform();
  }, 50);

  renderSidebarInputs();
  renderCanvasOverlays();
}

function closeEditor() {
  document.getElementById('layoutModal').classList.remove('show');
  activeFrameId = null;
}

function zoomCanvas(delta) {
  activeTemplateScale = Math.max(0.15, Math.min(3.0, activeTemplateScale + delta));
  updateCanvasTransform();
}

function updateCanvasTransform() {
  const canvas = document.getElementById('canvas');
  canvas.style.width = (activeCanvasW * activeTemplateScale) + 'px';
  canvas.style.height = (activeCanvasH * activeTemplateScale) + 'px';
  document.getElementById('zoomIndicator').textContent = Math.round(activeTemplateScale * 100) + '%';
}

function renderSidebarInputs() {
  const s = document.getElementById('sidebarInputs');
  s.innerHTML = activeLayouts.map((lo, i) =>
    '<div class="layout-card">' +
      '<h4>Slot ' + (i+1) + ' <button class="btn btn-sm" style="background: rgba(255,255,255,0.1); border:none; width:auto; margin:0;" onclick="resetLayout(' + i + ')">Reset</button></h4>' +
      '<div class="input-grid">' +
        '<div class="input-group"><label>X Pos</label><input type="number" value="' + lo.x + '" onchange="updateVal(' + i + ', \'x\', this.value)"></div>' +
        '<div class="input-group"><label>Y Pos</label><input type="number" value="' + lo.y + '" onchange="updateVal(' + i + ', \'y\', this.value)"></div>' +
        '<div class="input-group"><label>Width</label><input type="number" value="' + lo.width + '" onchange="updateVal(' + i + ', \'width\', this.value)"></div>' +
        '<div class="input-group"><label>Height</label><input type="number" value="' + lo.height + '" onchange="updateVal(' + i + ', \'height\', this.value)"></div>' +
      '</div>' +
    '</div>'
  ).join('');
}

function resetLayout(idx) {
  const f = config.frames.find(x => x.id === activeFrameId);
  if (!f) return;
  
  if (f.template === '3strip_2x6') {
    activeLayouts[idx] = { x: 0, y: 100 + (idx * 500), width: 600, height: 400 };
  } else if (f.template === '6strip_4x6') {
    const col = idx < 3 ? 0 : 1;
    const row = idx % 3;
    activeLayouts[idx] = { x: col * 600, y: 100 + (row * 500), width: 600, height: 400 };
  } else if (f.template === '4postcard_4x6') {
    const col = idx % 2;
    const row = Math.floor(idx / 2);
    // matches the 540x360 logic with 40px offsets
    activeLayouts[idx] = { 
      x: col === 0 ? 40 : 620, 
      y: row === 0 ? 40 : 440, 
      width: 540, 
      height: 360 
    };
  } else {
    // fallback generic
    activeLayouts[idx] = { x: 0, y: 0, width: 600, height: 400 };
  }
  
  renderSidebarInputs();
  renderCanvasOverlays();
}

function updateVal(idx, key, val) {
  activeLayouts[idx][key] = parseInt(val, 10) || 0;
  renderCanvasOverlays();
}

function renderCanvasOverlays() {
  const container = document.getElementById('canvasOverlays');
  container.innerHTML = activeLayouts.map((lo, i) => {
    const leftPct = (lo.x / activeCanvasW * 100);
    const topPct = (lo.y / activeCanvasH * 100);
    const widthPct = (lo.width / activeCanvasW * 100);
    const heightPct = (lo.height / activeCanvasH * 100);
    return '<div class="canvas-box" ' +
         'style="left:' + leftPct + '%; top:' + topPct + '%; width:' + widthPct + '%; height:' + heightPct + '%;" ' +
         'data-idx="' + i + '">' +
      (i+1) +
      '<div class="resize-handle tl" data-idx="' + i + '" data-handle="tl"></div>' +
      '<div class="resize-handle tr" data-idx="' + i + '" data-handle="tr"></div>' +
      '<div class="resize-handle bl" data-idx="' + i + '" data-handle="bl"></div>' +
      '<div class="resize-handle br" data-idx="' + i + '" data-handle="br"></div>' +
    '</div>';
  }).join('');
}

// --- Drag & Drop / Resize / Pan Logic for Canvas ---
let isDragging = false;
let isResizing = false;
let isPanning = false;
let dragStartX, dragStartY;
let initialLayout = null;

// Panning variables
let panStartX, panStartY, panScrollLeft, panScrollTop;

const canvasWrap = document.getElementById('canvasWrap');
let activeHandle = null;

// Intercept Ctrl+Scroll on the editor to prevent browser zoom
// and use the app's own zoom instead, keeping frame and overlays in sync
canvasWrap.addEventListener('wheel', (e) => {
  if (e.ctrlKey) {
    e.preventDefault();
    const delta = e.deltaY > 0 ? -0.05 : 0.05;
    zoomCanvas(delta);
  }
}, { passive: false });

canvasWrap.addEventListener('mousedown', (e) => {
  if (e.target.classList.contains('resize-handle')) {
    isResizing = true;
    activeIndex = parseInt(e.target.dataset.idx, 10);
    activeHandle = e.target.dataset.handle;
    startDrag(e);
  } else if (e.target.classList.contains('canvas-box')) {
    isDragging = true;
    activeIndex = parseInt(e.target.dataset.idx, 10);
    startDrag(e);
  } else {
    // Start panning if clicking the empty canvas space
    e.preventDefault(); // Prevent native browser dragging or text selection
    isPanning = true;
    panStartX = e.clientX;
    panStartY = e.clientY;
    panScrollLeft = canvasWrap.scrollLeft;
    panScrollTop = canvasWrap.scrollTop;
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
    canvasWrap.style.cursor = 'grabbing';
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
  if (!isDragging && !isResizing && !isPanning) return;
  e.preventDefault();
  
  if (isPanning) {
    const dx = e.clientX - panStartX;
    const dy = e.clientY - panStartY;
    canvasWrap.scrollLeft = panScrollLeft - dx;
    canvasWrap.scrollTop = panScrollTop - dy;
    return;
  }
  
  const dx = (e.clientX - dragStartX) / activeTemplateScale;
  const dy = (e.clientY - dragStartY) / activeTemplateScale;
  
  const lo = activeLayouts[activeIndex];
  
  if (isDragging) {
    lo.x = Math.round(initialLayout.x + dx);
    lo.y = Math.round(initialLayout.y + dy);
  } else if (isResizing) {
    const ratio = initialLayout.width / initialLayout.height;
    
    // Handle different corners
    if (activeHandle === 'br') {
      if (Math.abs(dx) > Math.abs(dy)) {
        lo.width = Math.max(10, Math.round(initialLayout.width + dx));
        lo.height = Math.round(lo.width / ratio);
      } else {
        lo.height = Math.max(10, Math.round(initialLayout.height + dy));
        lo.width = Math.round(lo.height * ratio);
      }
    } else if (activeHandle === 'bl') {
      // Bottom-Left
      if (Math.abs(dx) > Math.abs(dy)) {
        lo.width = Math.max(10, Math.round(initialLayout.width - dx));
        lo.height = Math.round(lo.width / ratio);
        lo.x = Math.round(initialLayout.x + (initialLayout.width - lo.width));
      } else {
        lo.height = Math.max(10, Math.round(initialLayout.height + dy));
        lo.width = Math.round(lo.height * ratio);
        lo.x = Math.round(initialLayout.x + (initialLayout.width - lo.width));
      }
    } else if (activeHandle === 'tr') {
      // Top-Right
      if (Math.abs(dx) > Math.abs(dy)) {
        lo.width = Math.max(10, Math.round(initialLayout.width + dx));
        lo.height = Math.round(lo.width / ratio);
        lo.y = Math.round(initialLayout.y + (initialLayout.height - lo.height));
      } else {
        lo.height = Math.max(10, Math.round(initialLayout.height - dy));
        lo.width = Math.round(lo.height * ratio);
        lo.y = Math.round(initialLayout.y + (initialLayout.height - lo.height));
      }
    } else if (activeHandle === 'tl') {
      // Top-Left
      if (Math.abs(dx) > Math.abs(dy)) {
        lo.width = Math.max(10, Math.round(initialLayout.width - dx));
        lo.height = Math.round(lo.width / ratio);
        lo.x = Math.round(initialLayout.x + (initialLayout.width - lo.width));
        lo.y = Math.round(initialLayout.y + (initialLayout.height - lo.height));
      } else {
        lo.height = Math.max(10, Math.round(initialLayout.height - dy));
        lo.width = Math.round(lo.height * ratio);
        lo.x = Math.round(initialLayout.x + (initialLayout.width - lo.width));
        lo.y = Math.round(initialLayout.y + (initialLayout.height - lo.height));
      }
    }
  }
  
  // Directly manipulate DOM elements for performance during drag
  const box = document.querySelector('.canvas-box[data-idx="' + activeIndex + '"]');
  if (box) {
    box.style.left = (lo.x / activeCanvasW * 100) + '%';
    box.style.top = (lo.y / activeCanvasH * 100) + '%';
    box.style.width = (lo.width / activeCanvasW * 100) + '%';
    box.style.height = (lo.height / activeCanvasH * 100) + '%';
  }
  
  // Debounce sidebar input updates slightly to avoid destroying focus if user is typing
  renderSidebarInputs();
}

function onMouseUp(e) {
  isDragging = false;
  isResizing = false;
  isPanning = false;
  activeIndex = -1;
  document.removeEventListener('mousemove', onMouseMove);
  document.removeEventListener('mouseup', onMouseUp);
  canvasWrap.style.cursor = ''; // reset cursor if we were grab-panning
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

function handleDrop(e, template) {
  e.preventDefault();
  const dzMap = { '3strip_2x6': 'dz1', '6strip_4x6': 'dz3', '4postcard_4x6': 'dz2' };
  const dz = document.getElementById(dzMap[template] || 'dz1');
  dz.classList.remove('dragover');
  
  if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
    const input = dz.querySelector('input[type="file"]');
    input.files = e.dataTransfer.files;
    handleUpload(input, template);
  }
}

async function saveR2Config() {
  var accountId = document.getElementById('r2AccountId').value.trim();
  var accessKeyId = document.getElementById('r2AccessKeyId').value.trim();
  var secretKey = document.getElementById('r2SecretKey').value.trim();
  var bucketName = document.getElementById('r2BucketName').value.trim();
  var publicBaseUrl = document.getElementById('r2PublicBaseUrl').value.trim();
  try {
    var body = {
      bypassPayment: config.bypassPayment,
      cameraType: config.cameraType,
      webcamId: config.webcamId,
      fullscreen: config.fullscreen,
      dslrMode: config.dslrMode,
      r2AccountId: accountId,
      r2AccessKeyId: accessKeyId,
      r2BucketName: bucketName,
      r2PublicBaseUrl: publicBaseUrl
    };
    if (secretKey) body.r2SecretKey = secretKey;
    var res = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    if (!res.ok) throw new Error(await res.text());
    flash('R2 settings saved!');
    document.getElementById('r2SecretKey').value = ''; // clear after save
    await load();
  } catch (err) {
    flash(err.message || 'Failed to save R2 settings', true);
  }
}

</script>
</body>
</html>`

// configResponse is used for JSON serialization of the admin config.
type configResponse struct {
	BypassPayment    bool           `json:"bypassPayment"`
	Frames           []Frame        `json:"frames"`
	CameraType       string         `json:"cameraType"`
	WebcamID         string         `json:"webcamId"`
	Fullscreen       bool           `json:"fullscreen"`
	DSLRMode         string         `json:"dslrMode"`
	AvailableCameras []CameraDevice `json:"availableCameras"`
	R2AccountID      string         `json:"r2AccountId"`
	R2AccessKeyID    string         `json:"r2AccessKeyId"`
	R2BucketName     string         `json:"r2BucketName"`
	R2PublicBaseURL  string         `json:"r2PublicBaseUrl"`
	// Secret key is intentionally omitted from GET response for security
}

// configUpdateRequest is used for JSON deserialization of config updates.
type configUpdateRequest struct {
	BypassPayment   bool   `json:"bypassPayment"`
	CameraType      string `json:"cameraType"`
	WebcamID        string `json:"webcamId"`
	Fullscreen      bool   `json:"fullscreen"`
	DSLRMode        string `json:"dslrMode"`
	R2AccountID     string `json:"r2AccountId"`
	R2AccessKeyID   string `json:"r2AccessKeyId"`
	R2SecretKey     string `json:"r2SecretKey"`
	R2BucketName    string `json:"r2BucketName"`
	R2PublicBaseURL string `json:"r2PublicBaseUrl"`
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
			r2 := cfg.GetR2Config()
			resp := configResponse{
				BypassPayment:    cfg.GetBypassPayment(),
				Frames:           cfg.GetFrames(),
				CameraType:       cfg.GetCameraType(),
				WebcamID:         cfg.GetWebcamID(),
				Fullscreen:       cfg.GetFullscreen(),
				DSLRMode:         cfg.GetDSLRMode(),
				AvailableCameras: cfg.GetAvailableCameras(),
				R2AccountID:      r2.AccountID,
				R2AccessKeyID:    r2.AccessKeyID,
				R2BucketName:     r2.BucketName,
				R2PublicBaseURL:  r2.PublicBaseURL,
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
			if body.CameraType != "" {
				cfg.SetCameraType(body.CameraType)
			}
			cfg.SetWebcamID(body.WebcamID)
			cfg.SetFullscreen(body.Fullscreen)
			if body.DSLRMode == "legacy" || body.DSLRMode == "integrated" {
				cfg.SetDSLRMode(body.DSLRMode)
			}
			// Save R2 config if any field is provided
			if body.R2AccountID != "" || body.R2AccessKeyID != "" || body.R2SecretKey != "" || body.R2BucketName != "" || body.R2PublicBaseURL != "" {
				cfg.SetR2Config(R2Config{
					AccountID:       body.R2AccountID,
					AccessKeyID:     body.R2AccessKeyID,
					SecretAccessKey: body.R2SecretKey,
					BucketName:      body.R2BucketName,
					PublicBaseURL:   body.R2PublicBaseURL,
				})
			}
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

		template := r.FormValue("template") // "3strip_2x6" or "4postcard_4x6"
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
		if template == "3strip_2x6" {
			// 600x1800, stack of 3. Each slot is 600x400 (3:2 ratio).
			// We stagger them evenly in the 1800px vertical space.
			// Example: y=100, y=600, y=1100
			layouts = []PhotoLayout{
				{X: 0, Y: 100, Width: 600, Height: 400},
				{X: 0, Y: 600, Width: 600, Height: 400},
				{X: 0, Y: 1100, Width: 600, Height: 400},
			}
		} else if template == "6strip_4x6" {
			// 1200x1800, two columns of 3 photos each 600x400 (3:2 DSLR ratio).
			// 3 photos × 400px height = 1200px, centered in 1800px (300px top/bottom)
			layouts = []PhotoLayout{
				{X: 0, Y: 100, Width: 600, Height: 400},
				{X: 0, Y: 600, Width: 600, Height: 400},
				{X: 0, Y: 1100, Width: 600, Height: 400},
				{X: 600, Y: 100, Width: 600, Height: 400},
				{X: 600, Y: 600, Width: 600, Height: 400},
				{X: 600, Y: 1100, Width: 600, Height: 400},
			}
		} else if template == "4postcard_4x6" {
			// 1200x1800. Needs 4 boxes in 3:2 ratio.
			// If we put two side-by-side, width = 540 each (leaves margins), height = 360.
			layouts = []PhotoLayout{
				{X: 40, Y: 40, Width: 540, Height: 360},
				{X: 620, Y: 40, Width: 540, Height: 360},
				{X: 40, Y: 440, Width: 540, Height: 360},
				{X: 620, Y: 440, Width: 540, Height: 360},
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
