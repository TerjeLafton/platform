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

CREATE TABLE todo.list_members (
    list_id UUID NOT NULL REFERENCES todo.lists(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (list_id, user_id)
);

CREATE INDEX idx_list_members_user_id ON todo.list_members(user_id);

CREATE TABLE todo.templates (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    title TEXT NOT NULL CHECK (length(title) BETWEEN 1 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, title)
);

CREATE INDEX idx_templates_user_id ON todo.templates(user_id);

CREATE TABLE todo.template_items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    template_id UUID NOT NULL
        REFERENCES todo.templates (id)
        ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (length(title) BETWEEN 1 AND 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_template_items_template_id ON todo.template_items(template_id);
