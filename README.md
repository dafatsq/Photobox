# Photobox

A kiosk-style photobooth desktop app built with **Wails v2** (Go + React/TypeScript) for **Windows**.

The app controls a DSLR/mirrorless camera via **digiCamControl**, composites the captured photos into strip or postcard layouts, and prints them silently via the OS print spooler.

---

## Prerequisites

Install all of these before running the project.

| Tool | Download |
|---|---|
| Go ≥ 1.21 | https://go.dev/dl/ |
| Node.js ≥ 18 | https://nodejs.org/ |
| Wails CLI v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| digiCamControl | https://digicamcontrol.com/ |

---

## Getting Started

### 1. Clone the repo

```bash
git clone <repo-url>
cd Photobox
```

### 2. Configure digiCamControl

1. Connect your camera via USB and open **digiCamControl**.
2. Enable the web server: *Settings → Web Server → Enable* (default port `5513`).
3. Leave digiCamControl running in the background — the app will control live view and capture through it automatically.

> [!IMPORTANT]
> digiCamControl must be running with its web server enabled **before** launching Photobox.

### 3. Install dependencies

```bash
go mod tidy
```

### 4. Run in development mode

```bash
wails dev
```

The app window opens automatically. Hot-reload is active for frontend changes.

You can also open http://localhost:34115 in a browser to call Go methods from DevTools.

### 5. Build a production binary

```bash
wails build
```

The compiled executable is placed in `build/bin/`.

---

## Hardware Setup

**Dual-cable mode** (recommended for production):

- **Visual stream:** Camera HDMI → Capture Card → PC (WebRTC in the frontend)
- **Capture / shutter:** Camera USB → PC (controlled by digiCamControl)

**Single-cable mode** (development/testing):

- Camera USB only — digiCamControl handles both live view and capture.

---

## Project Structure

```
Photobox/
├── main.go                        # Entry point — Wails app config
├── app.go                         # Wails-bound Go methods (IPC layer)
├── hardware/
│   ├── interfaces.go              # CameraDriver & PrinterDriver interfaces
│   ├── camera_windows.go          # digiCamControl HTTP API integration
│   └── printer_windows.go         # SumatraPDF silent print
└── frontend/
    └── src/
        ├── screens/               # One component per app state/screen
        ├── store/                 # Zustand global state
        └── components/            # Shared UI components
```

---

## App Flow

```
Attract → Payment (QRIS) → Template Selection → Capture (n shots) → Frame Selection → Processing → Share/Print → Done
```

Captured photos are saved to `~/PhotoboxData/sessions/<session-id>/`.

---

## Configuration

Project settings (app name, window size, etc.) are in [`wails.json`](./wails.json).
