<img src="web/frontend/public/favicon.svg" alt="TeamSphere" width="20%" height="20%" align="center" style="display: block; margin: 0 auto;"> 

# TeamSphere

Enterprise chat system for teams, offering real time rooms and direct messaging with notifications and file sharing.

[中文](README.zh-CN.md)

[![License](https://img.shields.io/badge/License-MIT-blue)](License)
[![Go](https://img.shields.io/badge/Go-1.25.0-00ADD8)](https://go.dev/dl/)
[![Node](https://img.shields.io/badge/Node-22.21.1-339933)](https://nodejs.org/)



## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Screenshots](#screenshots)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Common Commands](#common-commands)
- [Project Structure](#project-structure)
- [License](#license)


## Overview

TeamSphere uses a Go backend and a React frontend, featuring WebSocket real time messaging, room management, direct messages, friends, and invite links.


## Key Features

- Real time chat and room management
- Direct messages and friends
- Reactions and notifications
- Invite links and access control
- File upload and download



## Screenshots

- ![Screenshot 1](docs/assets/screenshot-1.png)

- ![Screenshot 2](docs/assets/screenshot-2.png)

- ![Screenshot 3](docs/assets/screenshot-3.png)
## Tech Stack

- Backend: Go, Gin, PostgreSQL, Redis, JWT, WebSocket
- Frontend: React, Vite, TypeScript, Zustand


## Quick Start

### Docker Compose

1. Prepare `.env` for PostgreSQL.
2. Start services.
3. Open the web UI and finish the setup wizard. It will generate `config.yaml` at `/app/data/config.yaml`.

```bash
mkdir -p data

# Create .env with:
# POSTGRES_USER=team_sphere
# POSTGRES_PASSWORD=your_db_password
# POSTGRES_DB=team_sphere

docker compose up --build
```

The backend listens on port 8080 by default.

### Local Development

1. Start the backend.
2. Open the web UI and finish the setup wizard (it generates `config.yaml`).
3. Start the frontend.

```bash
# Backend
go run ./cmd/server/main.go

# Frontend
cd web\frontend
npm install
npm run dev
```

The frontend runs at http://localhost:5173 by default.
## Configuration

The setup wizard generates `config.yaml`, so you usually do not need to create or edit it manually.
The default path is `config.yaml`, and you can override it with `TEAMSPHERE_CONFIG_PATH`.

Config values can be overridden by environment variables with the `TEAMSPHERE_` prefix, such as `TEAMSPHERE_DATABASE_HOST`.

Generate security keys as noted in the config comments:

```bash
openssl rand -hex 32
openssl rand -hex 64
```

The default config file is `config.yaml`, and you can override the path with `TEAMSPHERE_CONFIG_PATH`.

Config values can be overridden by environment variables with the `TEAMSPHERE_` prefix, such as `TEAMSPHERE_DATABASE_HOST`.

Generate security keys as noted in the config comments:

```bash
openssl rand -hex 32
openssl rand -hex 64
```


## Common Commands

```bash
# Build
make build

# Migrate
make migrate

# Clean
make clean
```


## Project Structure

- `cmd/`: entry points and startup
- `internal/`: core business logic and services
- `web/`: frontend
- `docs/`: docs
- `uploads/`: uploaded files


## License

MIT License.
