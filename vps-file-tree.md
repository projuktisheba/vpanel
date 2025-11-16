# Projuktisheba Folder Structure

This document explains the folder structure used in the `projuktisheba` workspace for managing projects and templates.

```
$USERHOME/projuktisheba/
│
├── projects/                         # All live project folders
│   ├── example.com/                  # Individual project folder named by domain
│   │   ├── src/                      # Source code of the project
│   │   ├── config/                   # Project-specific configuration files
│   │   ├── assets/                   # Static assets (images, fonts, icons)
│   │   ├── logs/                     # Log files for this project
│   │   └── README.md                 # Optional project-specific README
│   └── anotherdomain.com/
│
├── templates/                        # Templates for new projects
│   ├── databases/                    # Predefined database structures or SQL scripts
│   └── projects/                     # Boilerplate project templates
│       ├── <projectFramework>/       # Example template for a web project
│         └── <projectName>/          # Example template for an API project
│
└── README.md                         # This documentation
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
