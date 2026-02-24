
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
