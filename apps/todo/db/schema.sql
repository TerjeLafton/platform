CREATE SCHEMA IF NOT EXISTS todo;

CREATE TABLE todo.lists (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    title TEXT NOT NULL CHECK (length(title) BETWEEN 1 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, title)
);

CREATE INDEX idx_lists_user_id ON todo.lists(user_id);

CREATE TABLE todo.items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    list_id UUID NOT NULL
        REFERENCES todo.lists (id)
        ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (length(title) BETWEEN 1 AND 500),
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_list_id ON todo.items (list_id);
