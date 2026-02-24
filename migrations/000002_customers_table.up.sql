
CREATE TABLE Customers (
	customer_id SERIAL PRIMARY KEY,
	first_name VARCHAR(100),
	last_name VARCHAR(100),
	email VARCHAR(150) UNIQUE,
	phone VARCHAR(30),
	driver_license_number VARCHAR(50) UNIQUE NOT NULL,
	license_expiry DATE NOT NULL,
	is_active BOOLEAN DEFAULT TRUE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
