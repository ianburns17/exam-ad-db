
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
