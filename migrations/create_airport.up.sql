-- Migration: Create Airport table
CREATE TABLE IF NOT EXISTS airport (
    site_number VARCHAR(10) PRIMARY KEY,
    facility_name VARCHAR(255) NOT NULL,
    faa VARCHAR(10) NOT NULL,
    icao VARCHAR(10) NOT NULL,
    state_code VARCHAR(2) NOT NULL,
    state_full VARCHAR(100) NOT NULL,
    county VARCHAR(100),
    city VARCHAR(100) NOT NULL,
    ownership_type VARCHAR(100) NOT NULL,
    use_type VARCHAR(50) NOT NULL,
    manager VARCHAR(255),
    manager_phone VARCHAR(20),
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    airport_status VARCHAR(50) NOT NULL DEFAULT 'open',
    weather TEXT DEFAULT NULL
);