ALTER TABLE jobs ADD COLUMN public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid();
