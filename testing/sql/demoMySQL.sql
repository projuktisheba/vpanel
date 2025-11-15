-- -- Create database
-- CREATE DATABASE IF NOT EXISTS demo_db;
-- USE demo_db;

-- Drop tables if they exist
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS customers;

-- Create customers table
CREATE TABLE customers (
    customer_id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create products table
CREATE TABLE products (
    product_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create orders table
CREATE TABLE orders (
    order_id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT DEFAULT 1,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id),
    FOREIGN KEY (product_id) REFERENCES products(product_id)
);

-- Insert sample customers
INSERT INTO customers (first_name, last_name, email) VALUES
('John', 'Doe', 'john.doe@example.com'),
('Jane', 'Smith', 'jane.smith@example.com'),
('Alice', 'Johnson', 'alice.johnson@example.com'),
('Bob', 'Brown', 'bob.brown@example.com'),
('Charlie', 'Davis', 'charlie.davis@example.com');

-- Insert sample products
INSERT INTO products (name, description, price) VALUES
('Laptop', 'High performance laptop', 1200.00),
('Smartphone', 'Latest model smartphone', 800.00),
('Headphones', 'Noise-cancelling headphones', 150.00),
('Monitor', '24-inch LED monitor', 200.00),
('Keyboard', 'Mechanical keyboard', 100.00);

-- Insert sample orders
INSERT INTO orders (customer_id, product_id, quantity) VALUES
(1, 1, 1),
(1, 3, 2),
(2, 2, 1),
(3, 4, 1),
(4, 5, 3),
(5, 1, 2),
(2, 5, 1),
(3, 2, 1),
(4, 3, 2),
(5, 4, 1);

-- More sample inserts to reach ~100 lines
INSERT INTO customers (first_name, last_name, email) VALUES
('David', 'Lee', 'david.lee@example.com'),
('Eva', 'Martinez', 'eva.martinez@example.com'),
('Frank', 'Clark', 'frank.clark@example.com'),
('Grace', 'Lewis', 'grace.lewis@example.com'),
('Hannah', 'Walker', 'hannah.walker@example.com');

INSERT INTO products (name, description, price) VALUES
('Mouse', 'Wireless mouse', 40.00),
('Chair', 'Ergonomic office chair', 250.00),
('Desk', 'Wooden office desk', 350.00),
('Webcam', 'HD webcam', 80.00),
('Printer', 'All-in-one printer', 180.00);

INSERT INTO orders (customer_id, product_id, quantity) VALUES
(6, 6, 1),
(7, 7, 1),
(8, 8, 1),
(9, 9, 2),
(10, 10, 1),
(6, 1, 1),
(7, 2, 1),
(8, 3, 2),
(9, 4, 1),
(10, 5, 1);

-- Add more orders
INSERT INTO orders (customer_id, product_id, quantity) VALUES
(1, 6, 2),
(2, 7, 1),
(3, 8, 1),
(4, 9, 1),
(5, 10, 2),
(6, 2, 1),
(7, 3, 1),
(8, 4, 1),
(9, 5, 1),
(10, 1, 1);

-- End of demo SQL
