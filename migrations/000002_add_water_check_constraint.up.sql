-- Filename: migrations/000002_add_water_check_constraint.up.sql

ALTER TABLE water_system ADD CONSTRAINT status_length_check CHECK (array_length(status, 1) BETWEEN 1 AND 5);