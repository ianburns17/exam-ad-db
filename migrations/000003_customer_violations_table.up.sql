
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
