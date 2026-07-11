# Setup VPS — Pull Image dari Docker Hub

Panduan ini untuk kondisi **image sudah ada di Docker Hub**.  
VPS **tidak build** — hanya **pull + run**.

```
Docker Hub (ristudev/mindex-go-server:latest)
              │
              │  docker compose pull
              ▼
            VPS
              │
              │  docker compose up -d
              ▼
     API jalan di port 8080
```

Image yang dipakai:

```
ristudev/mindex-go-server:latest
```

---

## Prasyarat

- VPS Ubuntu/Debian (atau distro Linux lain)
- Akses SSH ke VPS
- Image sudah tersedia di Docker Hub (public atau private)
- Port `8080` (atau port yang kamu pilih) terbuka di firewall

---

## Step 1 — SSH ke VPS

```bash
ssh root@YOUR_VPS_IP
# atau
ssh user@YOUR_VPS_IP
```

---

## Step 2 — Install Docker

```bash
curl -fsSL https://get.docker.com | sh

# Jika login sebagai non-root:
sudo usermod -aG docker $USER
# logout lalu login lagi
```

Cek:

```bash
docker --version
docker compose version
```

---

## Step 3 — Buat folder deploy

```bash
sudo mkdir -p /opt/mindex-api
sudo chown $USER:$USER /opt/mindex-api
cd /opt/mindex-api
```

---

## Step 4 — Copy file deploy ke VPS

Dari **laptop lokal** (bukan di VPS):

```bash
cd /path/to/mindex-api

scp deploy/docker-compose.yml deploy/.env.example deploy/deploy.sh \
  user@YOUR_VPS_IP:/opt/mindex-api/
```

Atau clone repo di VPS lalu copy folder `deploy/`:

```bash
# Di VPS
git clone https://github.com/crosbydoo/mindex-go.git /tmp/mindex-go
cp /tmp/mindex-go/deploy/docker-compose.yml /opt/mindex-api/
cp /tmp/mindex-go/deploy/.env.example /opt/mindex-api/
cp /tmp/mindex-go/deploy/deploy.sh /opt/mindex-api/
rm -rf /tmp/mindex-go
```

Isi folder VPS harus seperti ini:

```text
/opt/mindex-api/
├── docker-compose.yml
├── .env.example
└── deploy.sh
```

---

## Step 5 — Buat & isi `.env`

```bash
cd /opt/mindex-api
cp .env.example .env
nano .env
```

Isi minimal:

```env
# Image dari Docker Hub (sudah ada)
DOCKER_IMAGE=ristudev/mindex-go-server
IMAGE_TAG=latest

# App
PORT=8080
ADMIN_PASSWORD=ganti-password-kuat
CORS_ORIGIN=https://your-frontend-domain.com

# Database (pakai Postgres di docker-compose)
POSTGRES_URL=postgres://mindex:mindex_secret@postgres:5432/mindex?sslmode=disable
POSTGRES_USER=mindex
POSTGRES_PASSWORD=mindex_secret
POSTGRES_DB=mindex
```

> Ganti `ADMIN_PASSWORD` dan `POSTGRES_PASSWORD` sebelum production.

---

## Step 6 — Login Docker Hub (hanya jika image private)

Kalau image **public**, skip step ini.

```bash
docker login
# Username: ristudev
# Password: Docker Hub Access Token
```

---

## Step 7 — Pull image & jalankan

```bash
cd /opt/mindex-api
chmod +x deploy.sh
./deploy.sh
```

Atau manual:

```bash
cd /opt/mindex-api

# Pull image dari Docker Hub (tidak build)
docker compose pull

# Jalankan container
docker compose up -d

# Cek status
docker compose ps
docker compose logs -f api
```

Yang terjadi:

1. `docker compose pull` → download `ristudev/mindex-go-server:latest`
2. `docker compose up -d` → start `api` + `postgres`
3. API migrate + seed otomatis saat startup

---

## Step 8 — Verifikasi

```bash
# Health
curl http://localhost:8080/health

# List entries
curl "http://localhost:8080/api/entries?page=1&limit=5"

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"password":"ganti-password-kuat"}'
```

Dari luar VPS (jika firewall terbuka):

```bash
curl http://YOUR_VPS_IP:8080/health
```

---

## Update image (saat ada versi baru di Docker Hub)

Tidak perlu rebuild di VPS. Cukup pull ulang:

```bash
cd /opt/mindex-api
./deploy.sh
```

Atau:

```bash
docker compose pull
docker compose up -d
docker image prune -f
```

Ganti tag versi tertentu di `.env`:

```env
IMAGE_TAG=v1.0.0
# atau
IMAGE_TAG=abc1234   # git sha
```

Lalu:

```bash
./deploy.sh
```

---

## Perintah operasional sehari-hari

```bash
cd /opt/mindex-api

# Status
docker compose ps

# Logs API
docker compose logs -f api

# Logs Postgres
docker compose logs -f postgres

# Restart
docker compose restart

# Stop
docker compose down

# Stop + hapus volume DB (HATI-HATI: data hilang)
docker compose down -v
```

---

## Firewall (opsional)

```bash
# UFW
sudo ufw allow OpenSSH
sudo ufw allow 8080/tcp
sudo ufw enable
```

Untuk production, lebih baik buka hanya `80`/`443` dan pakai Nginx reverse proxy.

### Contoh Nginx

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| `no matching manifest for linux/amd64` | Image di-build di Mac ARM. Rebuild untuk amd64: di laptop jalankan `make docker-push`, lalu di VPS `docker compose pull` |
| `pull access denied` | Image private → `docker login`. Cek `DOCKER_IMAGE` di `.env` |
| `manifest unknown` | Tag salah. Cek di Docker Hub: `ristudev/mindex-go-server` |
| Port already in use | Ubah `PORT=8081` di `.env`, lalu `./deploy.sh` |
| DB connection failed | Tunggu postgres healthy: `docker compose logs postgres` |
| `ADMIN_PASSWORD` 503 | Pastikan `ADMIN_PASSWORD` terisi di `.env` |
| Container exit loop | `docker compose logs api` untuk lihat error |

Cek image lokal:

```bash
docker images | grep mindex-go-server
```

Cek image remote:

```bash
docker pull ristudev/mindex-go-server:latest
```

---

## Ringkasan flow

```text
1. Install Docker di VPS
2. Buat /opt/mindex-api
3. Copy docker-compose.yml + .env.example + deploy.sh
4. cp .env.example .env  →  isi password & CORS
5. docker compose pull   →  ambil image dari Docker Hub
6. docker compose up -d  →  jalankan API + Postgres
7. curl /health          →  verifikasi
```

**Tidak ada `docker build` di VPS.**  
Semua image datang dari Docker Hub.

---

## File terkait

| File | Fungsi |
|------|--------|
| `deploy/docker-compose.yml` | Compose pull-only (API + Postgres) |
| `deploy/.env.example` | Template env VPS |
| `deploy/deploy.sh` | Script pull + up |
| `docs/deployment.md` | Full CI/CD GitHub → Docker Hub → VPS |
