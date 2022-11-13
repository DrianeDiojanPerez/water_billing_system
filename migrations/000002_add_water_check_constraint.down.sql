-- Filename: migrations/000002_add_water_check_constraint.down.sql

ALTER TABLE water_system DROP CONSTRAINT IF EXISTS status_length_check;