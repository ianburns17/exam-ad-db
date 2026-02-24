-- =========================
-- LOCATIONS RELATION
-- =========================
CREATE TABLE Locations (
    location_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- CUSTOMERS RELATION
-- =========================
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

-- =========================
-- CUSTOMER VIOLATIONS RELATION
-- =========================
CREATE TABLE CustomerViolations (
    violation_id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL,
    description TEXT,
    fee_amount NUMERIC(10,2),
    is_paid BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (customer_id)
        REFERENCES Customers(customer_id)
        ON DELETE CASCADE
);

-- =========================
-- VEHICLES RELATION
-- =========================
CREATE TABLE Vehicles (
    vehicle_id SERIAL PRIMARY KEY,
    vin VARCHAR(50) UNIQUE NOT NULL,
    make VARCHAR(100),
    model VARCHAR(100),
    year INT,
    category VARCHAR(50),
    daily_rate NUMERIC(10,2),
    mileage INT DEFAULT 0,
    fuel_capacity NUMERIC(5,2),
    current_location_id INT,
    status VARCHAR(20) CHECK (status IN ('available','rented','maintenance','inactive')) DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (current_location_id)
        REFERENCES Locations(location_id)
);

-- =========================
-- RENTALS RELATION
-- =========================
CREATE TABLE Rentals (
    rental_id SERIAL PRIMARY KEY,
    vehicle_id INT NOT NULL,
    customer_id INT NOT NULL,

    pickup_location_id INT NOT NULL,
    dropoff_location_id INT NOT NULL,

    pickup_datetime TIMESTAMP NOT NULL,
    dropoff_datetime TIMESTAMP NOT NULL,

    start_mileage INT,
    end_mileage INT,
    start_fuel_level NUMERIC(5,2),
    end_fuel_level NUMERIC(5,2),

    daily_rate NUMERIC(10,2),
    mileage_limit_per_day INT,
    extra_mileage_fee NUMERIC(10,2),

    total_cost NUMERIC(12,2),
    status VARCHAR(20) CHECK (status IN ('reserved','active','completed','cancelled')),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (vehicle_id)
        REFERENCES Vehicles(vehicle_id),

    FOREIGN KEY (customer_id)
        REFERENCES Customers(customer_id),

    FOREIGN KEY (pickup_location_id)
        REFERENCES Locations(location_id),

    FOREIGN KEY (dropoff_location_id)
        REFERENCES Locations(location_id)
);

-- =========================
-- MAINTENANCE RECORDS RELATION
-- =========================
CREATE TABLE MaintenanceRecords (
    maintenance_id SERIAL PRIMARY KEY,
    vehicle_id INT NOT NULL,
    service_type VARCHAR(100),
    service_date DATE,
    mileage_at_service INT,
    next_service_mileage INT,
    next_service_date DATE,
    cost NUMERIC(12,2),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (vehicle_id)
        REFERENCES Vehicles(vehicle_id)
        ON DELETE CASCADE
);

-- =========================
-- INSURANCE POLICIES RELATION
-- =========================
CREATE TABLE InsurancePolicies (
    policy_id SERIAL PRIMARY KEY,
    vehicle_id INT NOT NULL,
    provider VARCHAR(100),
    policy_number VARCHAR(100),
    coverage_details TEXT,
    start_date DATE,
    end_date DATE,
    premium NUMERIC(12,2),

    FOREIGN KEY (vehicle_id)
        REFERENCES Vehicles(vehicle_id)
        ON DELETE CASCADE
);

-- =========================
-- INSURANCE CLAIMS RELATION
-- =========================
CREATE TABLE InsuranceClaims (
    claim_id SERIAL PRIMARY KEY,
    vehicle_id INT NOT NULL,
    rental_id INT,
    claim_date DATE,
    description TEXT,
    claim_amount NUMERIC(12,2),
    status VARCHAR(20) CHECK (status IN ('open','approved','rejected','paid')),

    FOREIGN KEY (vehicle_id)
        REFERENCES Vehicles(vehicle_id),

    FOREIGN KEY (rental_id)
        REFERENCES Rentals(rental_id)
);

-- =========================
-- INSPECTIONS RELATION
-- =========================
CREATE TABLE Inspections (
    inspection_id SERIAL PRIMARY KEY,
    vehicle_id INT NOT NULL,
    rental_id INT,
    inspection_date TIMESTAMP,
    inspector_name VARCHAR(100),
    damage_found BOOLEAN,
    notes TEXT,

    FOREIGN KEY (vehicle_id)
        REFERENCES Vehicles(vehicle_id),

    FOREIGN KEY (rental_id)
        REFERENCES Rentals(rental_id)
);

-- =========================
-- FEES RELATION
-- =========================
CREATE TABLE Fees (
    fee_id SERIAL PRIMARY KEY,
    rental_id INT,
    customer_id INT,
    fee_type VARCHAR(50),
    amount NUMERIC(10,2),
    description TEXT,
    is_paid BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (rental_id)
        REFERENCES Rentals(rental_id),

    FOREIGN KEY (customer_id)
        REFERENCES Customers(customer_id)
);