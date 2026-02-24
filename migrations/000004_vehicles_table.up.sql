
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
