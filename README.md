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

### 3. Environtment Variables

```bash
# For Backend
cp backend/.env.example backend/.env

# For Frontend
cp frontend/.env.example frontend/.env
```

## How to Run

### Local Testing (Native)

Run this option if you want fast development cycles and hot-reloading without container overhead.

#### 1. Run Backend (Go)

```bash
cd backend
go run main.go
```

- The backend API will be live at: http://localhost:8080
- Test endpoint: http://localhost:8080/api/hello

#### 1b. Run Backend Items API (Go)

This is a separate entrypoint that serves the full `/api/v1/items` contract (see `API_CONTRACT.md`) backed by a real MongoDB instance — set `MONGODB_URI` / `MONGODB_DB` in `backend/.env` (or export them) before running, and make sure a MongoDB instance is reachable (e.g. `docker-compose up mongo` or a local `mongod`).

```bash
cd backend
go run ./cmd/server
```

- The API will be live at: http://localhost:8080 (override with `PORT=<port> go run ./cmd/server`)
- Test endpoint: http://localhost:8080/api/v1/items

Every route in `API_CONTRACT.md` can be hit with curl, e.g.:

```bash
curl http://localhost:8080/api/v1/items
curl http://localhost:8080/api/v1/items/507f191e810c19729de860ea
curl -X POST http://localhost:8080/api/v1/items -H "Content-Type: application/json" -d '{"name":"RTI-ALPHA-005"}'
```

#### 2. Run Frontend (React + Vite)

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
