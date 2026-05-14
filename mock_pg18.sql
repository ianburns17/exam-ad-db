CREATE OR REPLACE FUNCTION uuidv7() RETURNS uuid AS $$
BEGIN
    RETURN gen_random_uuid();
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION uuidv4() RETURNS uuid AS $$
BEGIN
    RETURN gen_random_uuid();
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION uuid_extract_timestamp(uuid) RETURNS timestamptz AS $$
BEGIN
    RETURN now();
END;
$$ LANGUAGE plpgsql;
