
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
