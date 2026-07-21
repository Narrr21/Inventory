# Inventory Management System

A full-stack inventory management application built with **React (Vite)** on the frontend and **Go (Golang)** on the backend. Fully containerized using **Docker**. Deployed to **Vercel**.

---

## Project Structure

This repository is managed as a monorepo structured as follows:

- `/frontend` - React application powered by Vite.
- `/backend` - REST API built with Go.

---

## Setup & Installation

### 1. Prerequisites

Ensure you have the following installed on your local machine:

- [Node.js](https://nodejs.org/) (v20 or higher)
- [Go](https://go.dev/) (v1.24 or higher)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) & Docker Compose

### 2. Clone the Repository

```bash
git clone [https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git](https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git)
cd YOUR_REPO_NAME
```

### 3. Environment Variables

```bash
cp backend/.env.example backend/.env
```

## How to Run

### Local Testing (Native)

Run this option if you want fast development cycles and hot-reloading without container overhead.

#### 1. Start MongoDB

The backend needs a reachable MongoDB instance before it will start. Either:

```bash
docker-compose up mongo
# or: podman run -d --name inventory-mongo -p 27017:27017 mongo:7
# or: your own local mongod
```

#### 2. Run the backend (Go)

`backend/cmd/server` is the real entrypoint — it serves the full `/api/v1/items` /
`/api/v1/projects` contract (see `docs/design/api-contract.md`) backed by MongoDB. It
automatically loads `backend/.env` on startup (via [godotenv](https://github.com/joho/godotenv);
already-exported environment variables still take precedence). `backend/main.go` at the repo
root is an unrelated legacy Vercel demo handler (`/api/hello`) — not part of this API, don't use
it to run the server.

```bash
cd backend
go run ./cmd/server
```

- The API will be live at: http://localhost:8080 (override with `PORT=<port> go run ./cmd/server`)

Every item references a project via `idProyek`, so an empty database has nothing valid to reference yet. Populate sample data first:

```bash
cd backend
go run ./cmd/seed          # insert sample projects + items (additive)
go run ./cmd/seed --reset  # wipe items/projects first, then insert
```

Every route in `docs/design/api-contract.md` can be hit with curl, e.g.:

```bash
curl http://localhost:8080/api/v1/items
curl http://localhost:8080/api/v1/items/507f191e810c19729de860ea

# filter items by project — by id or by name, either works:
curl "http://localhost:8080/api/v1/items?idProyek=<project _id>"
curl "http://localhost:8080/api/v1/items?namaProyek=ALPHA"

# dropdown data (jenis, status, and the full project list) in one call:
curl http://localhost:8080/api/v1/items/filter-options

# or create your own project + item:
curl -X POST http://localhost:8080/api/v1/projects -H "Content-Type: application/json" -d '{"namaProyek":"ALPHA","lokasi":"Jakarta HQ"}'
curl -X POST http://localhost:8080/api/v1/items -H "Content-Type: application/json" -d '{"jenis":"Laptop","serialNumber":"SN-00123","nama":"RTI-ALPHA-005","status":"Healthy","idProyek":"<_id from the project response above>"}'
```

Every Item response (create/list/get/update) also carries a read-only `projectName`, resolved
server-side from `idProyek` — no separate `/projects` lookup needed just to display it.

#### 3. Run Frontend (React + Vite)

```bash
cd frontend
npm install
npm run dev
```

- The frontend development server will be live at: http://localhost:5173

### Docker Compose (Production Replicas)

#### 1. Build and Run Containers

```bash
docker-compose up --build
```

- Frontend (React via Nginx): http://localhost:3000

- Backend API (Go): http://localhost:8080

#### 2. Stop and Remove Containers

```bash
docker-compose down
```
