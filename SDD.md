```markdown
# Software Design Document (SDD): Photobooth System

## 1. Ikhtisar Sistem
Sistem *photobooth* berbasis *desktop* (*kiosk*) yang mengintegrasikan UI interaktif dengan kontrol perangkat keras tingkat rendah (kamera DSLR/Mirrorless dan *printer thermal*). Dibangun menggunakan arsitektur hibrida Wails untuk memisahkan beban rendering visual di *frontend* dan eksekusi perangkat keras di *backend*.

## 2. Arsitektur Perangkat Keras (Hardware)
Sistem mendukung dua mode arsitektur fisik:

* **Mode Production (Kabel Ganda / Dual Cable):**
    * **Kamera:** Canon M50 / Sony ZV-E10 (Layar *vari-angle* dilipat, ditambah *dummy battery* dan pendingin Ulanzi CA25).
    * **Jalur 1 (Visual):** *Port* HDMI Kamera -> Kabel HDMI -> *Capture Card* -> PC (Diselesaikan via WebRTC API di *frontend*).
    * **Jalur 2 (Data):** *Port* USB Kamera -> PC (Ditembak via PTP CLI di *backend*).
* **Mode Development (Hybrid):**
    * **Kamera:** Canon 500D (Jalur data via USB) + *Webcam* standar (Jalur visual via USB).
* **Printer:** DNP RX1 (Dihubungkan via USB, dikendalikan via OS Print Spooler).

## 3. Stack Teknologi
* **Core Framework:** Wails v2 (Go + Webview).
* **Frontend (Mata):** React, TypeScript, Vite, WebRTC (`navigator.mediaDevices`), Zustand (State Management).
* **Backend (Otot):** Go (Golang).
* **Integrasi CLI Kamera:** `gphoto2` (macOS), `digiCamControl` (Windows).
* **Integrasi CLI Printer:** CUPS `lp` (macOS), `SumatraPDF.exe` (Windows).

---

## 4. Desain Antarmuka (Frontend Flow)


Alur antarmuka bersifat linear paksa (*forced progression*) tanpa tombol "Kembali" atau "Foto Ulang" untuk menjaga *turnover rate* mesin.

1.  **`AttractState`**: Layar siaga (Video/GIF *looping*). Sentuh untuk mulai.
2.  **`PaymentState`**: Menampilkan QRIS. UI melakukan *polling* IPC ke Go untuk status `payment_success`.
3.  **`TemplateState`**: Pemilihan tata letak (misal: *Strip* 2x6 atau *Postcard* 4x6).
4.  **`CaptureState`**: *Live Preview* 60fps via `<video>`. Menampilkan hitung mundur (3,2,1). Saat angka 0, layar berkedip putih, *frame* dibekukan (*freeze*), dan React mengirim perintah `Capture()` ke Go. Diulang $n$ kali sesuai *template*.
5.  **`ProcessingState`**: Menggabungkan foto mentah ke dalam *frame* PNG (via Canvas API atau *backend* Go). Mengirim perintah `Print()` ke Go. Menampilkan QR Code unduhan digital.
6.  **`DoneState`**: Pesan terima kasih. *Timer* 10 detik berjalan sebelum me-*reset* *state* kembali ke `AttractState`.

---

## 5. Desain Arsitektur Backend (Go) & Build Tags
*Backend* menggunakan *Strategy Pattern* berbasis OS pendeteksian saat proses kompilasi (*Build Tags*) untuk menghindari percabangan logika `if-else` yang kotor.

### Antarmuka Perangkat Keras (Interfaces)
```go
package hardware

type CameraDriver interface {
    Capture(filename string) error
}

type PrinterDriver interface {
    Print(filepath string) error
}

```

### Implementasi Lintas Platform (Cross-Platform)

File dipisahkan secara ketat agar *compiler* Go hanya membangun *binary* sesuai OS target.

**1. Modul Kamera (macOS)** -> `camera_darwin.go`

```go
//go:build darwin

package hardware
import "os/exec"

type MacCamera struct{}
func (m *MacCamera) Capture(filename string) error {
    cmd := exec.Command("gphoto2", "--capture-image-and-download", "--filename", filename)
    return cmd.Run()
}

```

**2. Modul Kamera (Windows)** -> `camera_windows.go`

```go
//go:build windows

package hardware
import "os/exec"

type WinCamera struct{}
func (w *WinCamera) Capture(filename string) error {
    cmd := exec.Command("CameraControlCmd.exe", "/capture", "/filename", filename)
    return cmd.Run()
}

```

*(Pola yang sama diterapkan untuk `printer_darwin.go` dan `printer_windows.go`)*

---

## 6. API Komunikasi (Wails IPC)

Daftar fungsi Go yang di-*ekspos* (di- *bind*) untuk dipanggil secara asinkron (sebagai *Promise*) oleh React TS:

* `CheckPaymentStatus(trxID string) (bool, error)`: Mengecek *webhook* QRIS.
* `TriggerCapture(sessionID string, sequence int) (string, error)`: Memerintahkan kamera menjepret dan mengembalikan *path* file *jpg* sementara.
* `ProcessComposite(images []string, templateID string) (string, error)`: Menggabungkan foto dengan *frame* PNG.
* `PrintPhoto(finalImagePath string) error`: Memicu proses *silent print* ke DNP RX1.

## 7. Penanganan Eror (Error Handling)

* **Camera Disconnect:** Jika `TriggerCapture` gagal (kabel tercabut/kamera *sleep*), IPC mengembalikan *error*. UI langsung melompat ke layar "Gangguan Teknis" dan membekukan antrean QRIS berikutnya.
* **Printer Out of Paper:** *Spooler* OS akan menahan antrean (*pending*). Go tidak akan *crash*, namun operator harus mereset antrean cetak secara manual di Control Panel Windows.

```