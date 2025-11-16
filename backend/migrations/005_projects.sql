CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    project_name VARCHAR(255) NOT NULL,
    domain_id INT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,    
    database_id INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    project_framework VARCHAR(100) NOT NULL,
    root_directory TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_projects_database_id ON projects(database_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_project_framework ON projects(project_framework);