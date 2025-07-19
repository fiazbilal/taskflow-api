-- +goose Up
-- +goose StatementBegin

-- Create projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_gen_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6366f1',
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status project_status DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for projects table
CREATE INDEX idx_projects_owner_id ON projects(owner_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX idx_projects_created_at ON projects(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop projects table
DROP TABLE IF EXISTS projects;

-- +goose StatementEnd