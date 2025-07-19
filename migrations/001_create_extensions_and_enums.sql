-- +goose Up
-- +goose StatementBegin

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create custom enum types
CREATE TYPE project_status AS ENUM ('active', 'archived', 'completed');
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done', 'cancelled');
CREATE TYPE task_priority AS ENUM ('low', 'medium', 'high', 'urgent');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop enum types
DROP TYPE IF EXISTS task_priority;
DROP TYPE IF EXISTS task_status;
DROP TYPE IF EXISTS project_status;

-- Drop UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd