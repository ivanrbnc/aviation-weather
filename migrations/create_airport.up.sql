-- Migration: Create Airport table
CREATE TABLE IF NOT EXISTS airport (
    site_number VARCHAR(10),
    facility_name VARCHAR(255),
    faa VARCHAR(10) PRIMARY KEY,
    icao VARCHAR(10),
    state_code VARCHAR(2),
    state_full VARCHAR(100),
    county VARCHAR(100),
    city VARCHAR(100),
    ownership_type VARCHAR(100),
    use_type VARCHAR(50),
    manager VARCHAR(255),
    manager_phone VARCHAR(20),
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    airport_status VARCHAR(50),
    weather VARCHAR(50)
);