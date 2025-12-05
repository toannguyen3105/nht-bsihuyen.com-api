# Modern Deployment Masterplan (Docker + GHCR + Traefik)

**Target Infrastructure**: Netcup VPS (4 Core, 8GB RAM).
**Strategy**: The "Modern Standard" - Automated, Containerized, and Scalable.

---

## Phase 1: VPS Preparation (One-time Setup)

### 1.1. Install Docker & Docker Compose

SSH into your VPS and run:

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Enable Docker to start on boot
systemctl enable docker
systemctl start docker

# Create a shared network for Traefik and Apps to talk
docker network create web_network
```

### 1.2. Create Infrastructure Directory

We will keep infrastructure configs separate from application code.

```bash
mkdir -p /opt/infrastructure/traefik
cd /opt/infrastructure
```

---

## Phase 2: The "Traffic Controller" (Traefik Setup)

Traefik will handle all incoming traffic (Port 80/443) and automatically issue SSL certificates (Let's Encrypt).

### 2.1. Create `traefik/acme.json`

This file stores your SSL certificates.

```bash
touch traefik/acme.json
chmod 600 traefik/acme.json
```

### 2.2. Create `/opt/infrastructure/docker-compose.yml`

```yaml
version: "3.8"

services:
  traefik:
    image: traefik:v2.10
    container_name: traefik
    restart: always
    command:
      - "--api.insecure=true" # Dashboard (Don't expose to public without auth)
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.myresolver.acme.tlschallenge=true"
      - "--certificatesresolvers.myresolver.acme.email=your-email@example.com" # CHANGE THIS
      - "--certificatesresolvers.myresolver.acme.storage=/acme.json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "./traefik/acme.json:/acme.json"
    networks:
      - web_network

networks:
  web_network:
    external: true
```

### 2.3. Start Traefik

```bash
docker compose up -d
```

_Now your VPS is ready to host infinite websites with auto-SSL._

---

## Phase 3: Application Setup (Per Website)

For every new website (e.g., `vcb-solver`), follow this standard structure.

### 3.1. Dockerfile (Optimized)

Ensure your app has a `Dockerfile`.

### 3.2. GitHub Actions Workflow (`.github/workflows/deploy.yml`)

This automates the build and push to GitHub Container Registry (GHCR).

```yaml
name: Deploy to VPS

on:
  push:
    branches: ["main"]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout repository
        uses: actions/checkout@v3

      - name: Log in to the Container registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata (tags, labels) for Docker
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}

      - name: Deploy to VPS
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            cd /opt/apps/vcb-solver
            # Login to GHCR on VPS (only needed once or if token expires)
            echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin

            # Pull new image
            docker compose pull

            # Restart container
            docker compose up -d
```

### 3.3. App Deployment on VPS (`docker-compose.yml`)

On your VPS, create `/opt/apps/vcb-solver/docker-compose.yml`:

```yaml
version: "3.8"

services:
  app:
    image: ghcr.io/your-github-username/vcb-captcha-solver:main
    restart: always
    labels:
      - "traefik.enable=true"
      # Router for HTTP (Redirect to HTTPS)
      - "traefik.http.routers.vcb-http.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.vcb-http.entrypoints=web"
      - "traefik.http.routers.vcb-http.middlewares=https-redirect"
      - "traefik.http.middlewares.https-redirect.redirectscheme.scheme=https"
      # Router for HTTPS
      - "traefik.http.routers.vcb.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.vcb.entrypoints=websecure"
      - "traefik.http.routers.vcb.tls.certresolver=myresolver"
    networks:
      - web_network

networks:
  web_network:
    external: true
```

---

## Phase 4: Adding a New Website (Checklist)

Whenever you have a new project:

1.  [ ] **Local**: Add `Dockerfile` to project.
2.  [ ] **Local**: Add `.github/workflows/deploy.yml` (Copy-paste the template).
3.  [ ] **GitHub**: Add Secrets (`VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`) to the repo.
4.  [ ] **VPS**: Create folder `/opt/apps/new-project`.
5.  [ ] **VPS**: Create `docker-compose.yml` (Copy-paste template, change Domain & Image Name).
6.  [ ] **Local**: Push code to `main`.

**Result**: GitHub builds it -> Pushes to GHCR -> VPS pulls it -> Traefik detects it -> **Live with HTTPS in minutes.**
