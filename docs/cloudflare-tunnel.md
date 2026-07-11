# Setup Cloudflare Tunnel (Gratis + HTTPS)

API di VPS dapat URL HTTPS gratis **tanpa beli domain**.

```
Frontend Vercel          Cloudflare Tunnel           VPS
https://xxx.vercel.app  →  https://xxxx.trycloudflare.com  →  localhost:8080
```

Ada 2 cara:

| Cara | Domain | Persistensi | Cocok untuk |
|------|--------|-------------|-------------|
| **Quick Tunnel** | `*.trycloudflare.com` (random) | URL berubah tiap restart | Coba-coba / demo |
| **Named Tunnel** | subdomain gratis Cloudflare / domain sendiri | Tetap | Lebih stabil |

Mulai dari **Quick Tunnel** dulu (paling sederhana).

---

## Prasyarat

- VPS sudah jalan, API container aktif di port `8080`
- Cek dulu:

```bash
curl http://localhost:8080/health
```

Harus dapat response JSON sukses.

---

## Cara A — Quick Tunnel (paling mudah)

### 1. Install `cloudflared` di VPS

**Ubuntu/Debian:**

```bash
curl -L --output cloudflared.deb \
  https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb

sudo dpkg -i cloudflared.deb
cloudflared --version
rm cloudflared.deb
```

### 2. Jalankan tunnel ke API

```bash
cloudflared tunnel --url http://localhost:8080
```

Output mirip:

```text
Your quick Tunnel has been created! Visit it at:
https://random-words-here.trycloudflare.com
```

**Itu URL publik API kamu.**

### 3. Tes dari laptop / browser

```bash
curl https://random-words-here.trycloudflare.com/health
curl "https://random-words-here.trycloudflare.com/api/entries?page=1&limit=5"
```

### 4. Hubungkan ke frontend Vercel

Di env frontend:

```env
VITE_API_BASE_URL=https://random-words-here.trycloudflare.com
```

Di VPS `.env` API:

```env
CORS_ORIGIN=https://your-app.vercel.app
```

Lalu restart API:

```bash
cd /opt/mindex-api
docker compose up -d
```

> Catatan Quick Tunnel: URL **berubah** setiap kali proses `cloudflared` di-restart.

### 5. Biar tetap jalan di background (sementara)

```bash
# Pakai screen
sudo apt install -y screen
screen -S cf-tunnel
cloudflared tunnel --url http://localhost:8080
# Lepas screen: Ctrl+A lalu D
# Balik lagi: screen -r cf-tunnel
```

Atau pakai `tmux`:

```bash
tmux new -s cf-tunnel
cloudflared tunnel --url http://localhost:8080
# Lepas: Ctrl+B lalu D
# Balik: tmux attach -t cf-tunnel
```

---

## Cara B — Named Tunnel (lebih stabil, masih gratis)

Butuh akun Cloudflare gratis: [https://dash.cloudflare.com/sign-up](https://dash.cloudflare.com/sign-up)

### 1. Login cloudflared

```bash
cloudflared tunnel login
```

Perintah ini memberi link. Buka di browser → pilih domain (kalau belum punya domain, Quick Tunnel lebih cocok; Named Tunnel biasanya butuh zone/domain di Cloudflare).

> Tanpa domain sendiri, **Quick Tunnel (Cara A)** tetap opsi gratis terbaik.  
> Named Tunnel paling berguna kalau nanti kamu punya domain (bahkan yang murah).

### 2. Buat tunnel

```bash
cloudflared tunnel create mindex-api
```

Catat **Tunnel ID** yang muncul.

### 3. Buat config

```bash
mkdir -p ~/.cloudflared
nano ~/.cloudflared/config.yml
```

Isi (ganti `TUNNEL_ID` dan hostname):

```yaml
tunnel: TUNNEL_ID
credentials-file: /root/.cloudflared/TUNNEL_ID.json

ingress:
  - hostname: api.yourdomain.com
    service: http://localhost:8080
  - service: http_status:404
```

### 4. DNS route

```bash
cloudflared tunnel route dns mindex-api api.yourdomain.com
```

### 5. Jalankan sebagai service

```bash
sudo cloudflared service install
sudo systemctl enable --now cloudflared
sudo systemctl status cloudflared
```

Tes:

```bash
curl https://api.yourdomain.com/health
```

---

## Setup CORS (penting untuk Vercel)

Di `/opt/mindex-api/.env`:

```env
CORS_ORIGIN=https://your-frontend.vercel.app
```

Kalau masih development lokal + Vercel:

```env
# Satu origin saja di API saat ini
CORS_ORIGIN=https://your-frontend.vercel.app
```

Restart:

```bash
cd /opt/mindex-api
docker compose up -d
```

---

## Flow akhir yang disarankan (gratis)

```text
1. API jalan di VPS: docker compose up -d
2. Install cloudflared
3. cloudflared tunnel --url http://localhost:8080
4. Copy URL https://xxxx.trycloudflare.com
5. Set VITE_API_BASE_URL di Vercel = URL itu
6. Set CORS_ORIGIN di VPS = https://app-kamu.vercel.app
7. Redeploy frontend Vercel
```

---

## Update API (image baru)

Tunnel **tidak perlu diubah** selama tetap ke `localhost:8080`:

```bash
cd /opt/mindex-api
./deploy.sh
# atau
docker compose pull && docker compose up -d
```

Quick Tunnel tetap jalan selama proses `cloudflared` tidak mati.

---

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| `502 Bad Gateway` | API belum jalan. `curl localhost:8080/health` |
| CORS error di browser | `CORS_ORIGIN` harus exact match URL Vercel (pakai `https://`) |
| URL tunnel berubah | Normal untuk Quick Tunnel. Update `VITE_API_BASE_URL` di Vercel |
| Tunnel mati setelah logout SSH | Jalankan di `screen`/`tmux`, atau pakai Named Tunnel + systemd |
| Mixed content | Frontend HTTPS harus panggil API HTTPS (tunnel), jangan `http://IP` |

---

## Alternatif lain (juga gratis)

| Tool | URL contoh | Catatan |
|------|------------|---------|
| **Cloudflare Quick Tunnel** | `*.trycloudflare.com` | Direkomendasikan |
| **ngrok** | `*.ngrok-free.app` | Perlu daftar akun, limit free tier |
| **localhost.run** | SSH tunnel | Lebih sederhana, kurang stabil |

### ngrok (opsional)

```bash
# Install di VPS (cek docs ngrok untuk cara terbaru)
ngrok http 8080
```

---

## Ringkas

- **Gratis + manual + tanpa beli domain** → Cloudflare Quick Tunnel
- Frontend tetap di **Vercel** (`*.vercel.app`)
- API dapat HTTPS lewat **`*.trycloudflare.com`**
- Jangan harapkan API “nempel” ke domain `vercel.app` — itu tidak bisa

File terkait:

- `docs/vps-setup.md` — setup VPS + Docker Hub
- `docs/deployment.md` — CI/CD penuh
