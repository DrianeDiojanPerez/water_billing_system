-- Filename: migrations/000001_create_water_table.up.sql

CREATE TABLE IF NOT EXISTS water_system (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    waterbill text NOT NULL,
    description text NOT NULL,
    notes text NOT NULL,
    category text NOT NULL,
    priority text NOT NULL,
    status text[] NOT NULL,
    version int NOT NULL DEFAULT 1
);