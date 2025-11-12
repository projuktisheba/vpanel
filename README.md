# VPanel — Modern Lightweight VPS & App Management Control Panel

VPanel is a next-generation open-source control panel designed for developers and hosting providers to manage VPS infrastructure easily and efficiently.It automates deployment, database management, backups, and app synchronization — all from a unified dashboard.

---

## Overview

VPanel allows you to:

* Deploy PHP (Laravel), Go, and Next.js applications automatically
* Manage databases, users, and associations with hosted sites
* Create and restore full server backups
* Deploy pre-built application templates in one click
* Transfer websites between servers seamlessly

---

## Features

### Core Features

* **Server & App Deployment**
  Deploy PHP (Laravel), Go, or Next.js apps automatically in isolated environments.

* **Smart Agent System**
  Lightweight Go agent runs on each VPS to handle communication, deployment, and health checks.

* **Database Management**
  Create, list, import/export databases and assign them to specific websites.

* **One-click Templates**
  Deploy pre-configured site templates instantly for clients or rapid testing.

* **Backup & Sync Manager**
  Backup all or specific websites and transfer them between servers automatically.

* **Multi-Domain Support**
  Manage multiple domains and subdomains under one unified panel.

* **Secure Architecture**
  Token-based authentication, agent verification, and strict access controls.

---

## Tech Stack

| Component  | Technology                                  |
| ---------- | ------------------------------------------- |
| Backend    | Go (RESTful API, Agents, Deployment Engine) |
| Frontend   | Next.js / React                             |
| Database   | PostgreSQL                                  |
| Web Server | Nginx                                       |
| Optional   | Docker, systemd integration for agents      |

---

## System Architecture

**VPanel consists of three main components:**

1. **Control Panel Backend (Go)**

   * Provides RESTful APIs for user, server, and app management.
   * Handles authentication, permissions, and deployment orchestration.

2. **Frontend Dashboard (Next.js)**

   * Web-based interface for managing deployments, backups, and monitoring.

3. **VPS Agent (Go binary)**

   * Installed on each VPS server.
   * Communicates securely with the main panel.
   * Executes deployment, sync, and backup tasks.

---

## Folder Structure

```
vpanel/
├── backend/                 # Go API server
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   └── main.go
├── frontend/                # React dashboard
│   ├── pages/
│   ├── components/
│   └── package.json
├── agent/                   # Lightweight Go agent for VPS
│   ├── cmd/
│   ├── internal/
│   └── main.go
├── scripts/                 # Deployment & setup scripts
└── README.md
```

---

## API Overview

| Method | Endpoint                   | Description                |
| ------ | -------------------------- | -------------------------- |
| POST   | `/api/v1/auth/signin`       | User authentication        |
| GET    | `/api/v1/servers`          | List all connected servers |
| POST   | `/api/v1/servers/register` | Register a new VPS agent   |
| POST   | `/api/v1/apps/deploy`      | Deploy an application      |
| GET    | `/api/v1/databases`        | List all databases         |
| POST   | `/api/v1/databases/create` | Create a new database      |
| POST   | `/api/v1/backup/create`    | Create a backup            |
| POST   | `/api/v1/backup/restore`   | Restore from a backup      |

---

## Deployment Setup

### 1. Backend (Go)

```bash
cd backend
go mod tidy
go build -o vpanel-backend ./cmd/server
./vpanel-backend
```

### 2. Frontend (Next.js)

```bash
cd frontend
npm install
npm run build
npm run start
```

### 3. VPS Agent (Go)

```bash
cd agent
go build -o vpanel-agent ./cmd/agent
sudo ./vpanel-agent --register --server=https://your-vpanel-server.com
```

### 4. Access Panel

Visit `http://your-server-ip:3000` to access the VPanel dashboard.

---

## Backup and Sync Flow

1. The agent triggers a backup of all website files and databases.
2. The backup is uploaded to the control panel or external storage (e.g., S3).
3. The target server agent pulls and restores it automatically.
4. Configuration files and domains are re-mapped automatically.

---

## Roadmap

* User & Role Management
* API Token Access
* Docker Integration
* Real-time Logs & Notifications
* Marketplace for Templates

---

## Contributing

Contributions are welcome!
Please open an issue or submit a pull request for feature requests, bug fixes, or documentation improvements.

---

## License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---

## Vision

VPanel aims to become a developer-friendly alternative to traditional control panels —
lightweight, fast, open-source, and fully customizable for modern deployment workflows.
# vpanel
# vpanel
# vpanel
