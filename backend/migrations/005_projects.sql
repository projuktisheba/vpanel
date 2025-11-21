-- Drop an existing table 'projects'
DROP TABLE IF EXISTS projects;

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    project_name VARCHAR(255) NOT NULL DEFAULT '',
    domain_name VARCHAR(255) NOT NULL UNIQUE REFERENCES domains(domain) ON DELETE CASCADE,    
    db_name TEXT NOT NULL DEFAULT '',
    project_framework VARCHAR(100) NOT NULL DEFAULT '',
    template_path TEXT NOT NULL DEFAULT '',
    project_directory TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'build',  --build, uploaded, active, suspend, inactive
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_projects_database_name ON projects(db_name);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_project_framework ON projects(project_framework);