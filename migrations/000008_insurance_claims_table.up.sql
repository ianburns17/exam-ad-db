
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
