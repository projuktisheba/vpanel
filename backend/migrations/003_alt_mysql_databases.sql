CREATE TABLE mysql_databases (
    id SERIAL PRIMARY KEY,
    db_name TEXT NOT NULL UNIQUE,
    user_id INT NOT NULL,  -- foreign key to mysql_db_users
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT fk_user
        FOREIGN KEY (user_id) REFERENCES mysql_db_users(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

-- Indexes
CREATE INDEX idx_mysql_databases_db_name ON mysql_databases(db_name);
CREATE INDEX idx_mysql_databases_user_id ON mysql_databases(user_id);
