# Inventory Management System

...

---

## 1. Project Structure

This repository is managed as a monorepo structured as follows:

- `/frontend` - React using Vite.
- `/backend` - Go.

---

## 2. Setup & Installation

### a. Prerequisites

Ensure you have the following installed on your local machine:

- [Node.js](https://nodejs.org/) (v20 or higher)
- [Go](https://go.dev/) (v1.24 or higher)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) & Docker Compose

### b. Clone the Repository

```bash
git clone [https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git](https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git)
cd YOUR_REPO_NAME
```

### c. Environment Variables

```bash
cp backend/.env.example backend/.env
```

## 3. How to Run

### 3.1. Local Testing (Native)

open 3 terminal windows/tabs, one for the mongo, one for backend and one for the frontend.

#### a. Start MongoDB

In terminal 1, start a local MongoDB instance.

```bash
docker-compose up mongo
# or: podman run -d --name inventory-mongo -p 27017:27017 mongo:7
# or: your own local mongod
```

#### b. Run the backend (Go)

In Terminal 2, navigate to the backend directory and run the server.

```bash
cd backend
go run ./cmd/server
```

- The API will be live at: http://localhost:8080

If you want to override the default port, set the `PORT` environment variable in `backend/.env` before running.

#### c. Populate the database with sample data

```bash
cd backend
go run ./cmd/seed          # insert sample projects + items + item types (additive)
go run ./cmd/seed --reset  # wipe items/projects/itemTypes first, then insert
```

#### d. Run Frontend

In Terminal 3, navigate to the frontend directory and start the development server.

```bash
cd frontend
npm install
npm run dev
```

- The frontend development server will be live at: http://localhost:5173

### 3.2. Docker

#### a. Build and Run Containers

```bash
docker-compose up --build
```

- Frontend (React via Nginx): http://localhost:3000

- Backend API (Go): http://localhost:8080

#### b. Stop and Remove Containers

```bash
docker-compose down
```
