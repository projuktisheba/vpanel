CREATE TABLE databases (
    id SERIAL PRIMARY KEY,
    db_name TEXT NOT NULL UNIQUE,
    user_id INT NOT NULL,  -- foreign key to db_users
    db_type TEXT NOT NULL DEFAULT 'mysql',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT fk_user
        FOREIGN KEY (user_id) REFERENCES db_users(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

-- Indexes
CREATE INDEX idx_databases_db_name ON databases(db_name);
CREATE INDEX idx_databases_user_id ON databases(user_id);
CREATE INDEX idx_databases_db_type ON databases(db_type);
