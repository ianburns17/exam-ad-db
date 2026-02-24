
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
