-- ==========================================
-- PostgreSQL Database & User Creation Script
-- ==========================================
-- How to run this script:
--   psql -U postgres -f database.sql
-- If shows FATAL:  Peer authentication failed for user "postgres"
--   sudo -i -u postgres
--   psql -f database.sql

-- (Run as a PostgreSQL superuser, e.g. 'postgres')
CREATE USER vpnale_sudo_user WITH PASSWORD '58yubDDQ6EeVvnXxeAU8GXwiP3j8xtjx';

-- Create a new database owned by that user
CREATE DATABASE vpanel_main_db OWNER vpnale_sudo_user;

-- Grant all privileges on the database to the user
GRANT ALL PRIVILEGES ON DATABASE vpanel_main_db TO vpnale_sudo_user;

-- (Optional) Verify ownership later:
--   \l   → list databases
--   \du  → list roles/users
