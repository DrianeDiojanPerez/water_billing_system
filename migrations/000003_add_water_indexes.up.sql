-- Filename: migrations/000003_add_water_indexes.up.sql

CREATE INDEX IF NOT EXISTS water_system_task_name_idx ON water_system USING GIN(to_tsvector('simple', task_name));
CREATE INDEX IF NOT EXISTS water_system_priority_idx ON water_system USING GIN(to_tsvector('simple', priority));
CREATE INDEX IF NOT EXISTS water_system_status_idx ON water_system USING GIN(status);