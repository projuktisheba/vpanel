# Projuktisheba Folder Structure

This document explains the folder structure used in the `projuktisheba` workspace for managing projects and templates.

```
$USERHOME/projuktisheba/
│
├── projects/                     # All user projects
│   ├── <projectFramework>/       # e.g., Laravel, CodeIgniter
│   │   └── <projectName>/       # e.g., example.com
│   │       ├── src/
│   │       ├── config/
│   │       └── assets/
│   └── ...
│
├── templates/                    # Reusable templates for new projects
│   ├── databases/                # Predefined database schemas or SQL scripts
│   │   └── <templateName>.sql
│   └── projects/                 # Project boilerplates
│       ├── <projectFramework>/   # e.g., Laravel, CodeIgniter
│       │   └── <templateName>/  # e.g., API Boilerplate
│       │       ├── src/
│       │       ├── config/
│       │       └── assets/
│       └── ...
│
└── README.md                     # This documentation

```

## Folder Details

* **projects/** – Contains all project folders, each named by its domain.

  * Example: `example.com/` contains all files related to that specific project.

* **templates/** – Stores reusable templates for faster project setup.

  * **databases/** – Database schemas, SQL scripts, or ready-to-use database files.
  * **projects/** – Project boilerplates (frontend, backend, or full-stack templates).

## Guidelines

1. **Project Naming**: Always name project folders under `projects/` by the domain or project identifier.
2. **Templates**: Use `templates/` to clone new projects or databases for consistency.
3. **Organization**: Keep project files structured inside each project folder (`src/`, `config/`, `assets/`, etc.) for easy maintenance.


#VPS setup guide

1. mkdir -p /var/log (in case if dir not exist)