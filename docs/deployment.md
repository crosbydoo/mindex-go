# Deployment — GitHub → Docker Hub → VPS

Alur deploy **tanpa build di VPS**. VPS hanya **pull image** dari Docker Hub lalu jalankan container.

```
Developer push ke GitHub (main)
        │
        ▼
GitHub Actions — build image + push ke Docker Hub
        │
        ▼
GitHub Actions — SSH ke VPS → docker compose pull + up
        │
        ▼
VPS menjalankan container dari image Docker Hub
```

---

## Arsitektur

| Komponen | Peran |
|----------|-------|
| **GitHub** | Source code + trigger CI/CD |
| **GitHub Actions** | Build Docker image, push ke Hub, deploy ke VPS |
| **Docker Hub** | Registry image (`ristudev/mindex-go-server`) |
| **VPS** | Pull image + run container (tidak build) |

---

## 1. Setup Docker Hub

1. Buat akun di [hub.docker.com](https://hub.docker.com)
2. Buat **Access Token**:
   - Account Settings → Security → New Access Token
   - Permission: **Read & Write**
   - Simpan token-nya

Image akan tersedia di:
```
docker.io/ristudev/mindex-go-server:latest
docker.io/ristudev/mindex-go-server:<git-sha>
```

Contoh push manual:

```bash
docker build -t ristudev/mindex-go-server:latest .
docker push ristudev/mindex-go-server:latest
```

---

## 2. Setup GitHub Secrets

Di repo GitHub: **Settings → Secrets and variables → Actions → New repository secret**

| Secret | Contoh | Keterangan |
|--------|--------|------------|
| `DOCKERHUB_USERNAME` | `ristudev` | Username Docker Hub |
| `DOCKERHUB_TOKEN` | `dckr_pat_xxx` | Access token Docker Hub |
| `VPS_HOST` | `123.45.67.89` | IP atau domain VPS |
| `VPS_USER` | `root` | SSH user VPS |
| `VPS_SSH_KEY` | `-----BEGIN OPENSSH...` | Private key SSH (full content) |
| `VPS_DEPLOY_PATH` | `/opt/mindex-api` | Path folder deploy di VPS |

### Generate SSH key (jika belum ada)

Di laptop lokal:

```bash
ssh-keygen -t ed25519 -C "github-deploy-mindex" -f ~/.ssh/mindex_deploy -N ""
```

Copy public key ke VPS:

```bash
ssh-copy-id -i ~/.ssh/mindex_deploy.pub root@YOUR_VPS_IP
```

Copy private key ke GitHub secret `VPS_SSH_KEY`:

```bash
cat ~/.ssh/mindex_deploy
```

---

## 3. Setup VPS (sekali saja)

### Install Docker di VPS

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# logout & login ulang
```

### Clone hanya folder deploy (atau copy manual)

```bash
sudo mkdir -p /opt/mindex-api
sudo chown $USER:$USER /opt/mindex-api
cd /opt/mindex-api
```

Copy file-file ini ke VPS (bisa via `scp` atau clone repo lalu pakai folder `deploy/`):

```bash
# Dari laptop lokal
scp -r deploy/* user@YOUR_VPS_IP:/opt/mindex-api/
```

Atau clone repo dan symlink:

```bash
git clone https://github.com/YOUR_USER/mindex-api.git /opt/mindex-api-src
cp -r /opt/mindex-api-src/deploy/* /opt/mindex-api/
```

### Konfigurasi environment VPS

```bash
cd /opt/mindex-api
cp .env.example .env
nano .env
```

Isi minimal:

```env
DOCKER_IMAGE=ristudev/mindex-go-server
IMAGE_TAG=latest
PORT=8080
ADMIN_PASSWORD=strong-password-here
CORS_ORIGIN=https://your-frontend.com
POSTGRES_URL=postgres://mindex:mindex_secret@postgres:5432/mindex?sslmode=disable
POSTGRES_USER=mindex
POSTGRES_PASSWORD=mindex_secret
POSTGRES_DB=mindex
```

### Deploy pertama (manual)

```bash
chmod +x deploy.sh
./deploy.sh
```

VPS **tidak build** — hanya:

```bash
docker compose pull    # download image dari Docker Hub
docker compose up -d   # jalankan container
```

### Cek

```bash
curl http://localhost:8080/health
docker compose logs -f api
```

---

## 4. Alur otomatis (setelah setup)

Setiap push ke branch `main`:

1. GitHub Actions build image dari `Dockerfile`
2. Push ke Docker Hub (`latest` + commit SHA)
3. SSH ke VPS → jalankan `docker compose pull && docker compose up -d`

Tidak perlu SSH manual ke VPS untuk setiap deploy.

### Trigger manual

GitHub → **Actions** → **CI/CD — Build, Push & Deploy** → **Run workflow**

---

## 5. Deploy manual di VPS (tanpa GitHub Actions)

Kalau mau update manual tanpa CI:

```bash
cd /opt/mindex-api
docker compose pull
docker compose up -d
```

Atau:

```bash
./deploy.sh
```

---

## 6. Pakai Postgres eksternal (opsional)

Kalau Postgres sudah ada di VPS / managed DB (bukan container), edit `docker-compose.yml`:

1. Hapus service `postgres` dan `depends_on`
2. Set `POSTGRES_URL` ke database eksternal:

```env
POSTGRES_URL=postgres://user:pass@your-db-host:5432/mindex?sslmode=disable
```

Contoh `docker-compose.yml` minimal (API saja):

```yaml
services:
  api:
    image: ${DOCKER_IMAGE:-ristudev/mindex-go-server}:${IMAGE_TAG:-latest}
    container_name: mindex-go-server
    restart: unless-stopped
    ports:
      - "${PORT:-8080}:8080"
    environment:
      POSTGRES_URL: ${POSTGRES_URL}
      ADMIN_PASSWORD: ${ADMIN_PASSWORD}
      CORS_ORIGIN: ${CORS_ORIGIN}
      PORT: "8080"
```

---

## 7. Reverse proxy (production)

Disarankan pakai Nginx/Caddy di depan container:

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Tambahkan SSL dengan Certbot atau Caddy auto-HTTPS.

---

## 8. Rollback ke versi sebelumnya

Image di-tag dengan git SHA. Untuk rollback:

```bash
# Lihat image tags di Docker Hub, lalu di VPS:
cd /opt/mindex-api
# Edit .env: IMAGE_TAG=<commit-sha>
nano .env
docker compose pull
docker compose up -d
```

---

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| `pull access denied` | Cek `DOCKER_IMAGE=ristudev/mindex-go-server` di `.env` VPS |
| `connection refused` Postgres | Tunggu healthcheck postgres, cek `docker compose logs postgres` |
| GitHub Actions deploy gagal | Cek `VPS_SSH_KEY`, `VPS_HOST`, firewall port 22 |
| Port 8080 sudah dipakai | Ubah `PORT` di `.env` VPS |
| Image tidak update | Pastikan push ke `main` dan workflow sukses di GitHub Actions |

---

## File terkait

```
.github/workflows/ci-cd.yml   # Pipeline GitHub Actions
deploy/docker-compose.yml    # Compose VPS (pull only)
deploy/.env.example          # Template env VPS
deploy/deploy.sh             # Script deploy manual VPS
Dockerfile                   # Dipakai GitHub Actions saat build
```
