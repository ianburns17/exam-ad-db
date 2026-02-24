
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
