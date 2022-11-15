-- Filename: migrations/000003_add_water_indexes.down.sql

DROP INDEX IF EXISTS water_system_waterbill_idx;
DROP INDEX IF EXISTS water_system_priority_idx;
DROP INDEX IF EXISTS water_system_status_idx;