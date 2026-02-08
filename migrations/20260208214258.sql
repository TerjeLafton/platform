-- Add new schema named "id"
CREATE SCHEMA "id";
-- Add new schema named "todo"
CREATE SCHEMA "todo";
-- Create "users" table
CREATE TABLE "id"."users" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email"),
  CONSTRAINT "users_email_check" CHECK (length(email) <= 255)
);
-- Create "lists" table
CREATE TABLE "todo"."lists" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "user_id" uuid NOT NULL,
  "title" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "lists_user_id_title_key" UNIQUE ("user_id", "title"),
  CONSTRAINT "lists_title_check" CHECK ((length(title) >= 1) AND (length(title) <= 100))
);
-- Create index "idx_lists_user_id" to table: "lists"
CREATE INDEX "idx_lists_user_id" ON "todo"."lists" ("user_id");
-- Create "items" table
CREATE TABLE "todo"."items" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "list_id" uuid NOT NULL,
  "title" text NOT NULL,
  "completed" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "items_list_id_fkey" FOREIGN KEY ("list_id") REFERENCES "todo"."lists" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "items_title_check" CHECK ((length(title) >= 1) AND (length(title) <= 500))
);
-- Create index "idx_items_list_id" to table: "items"
CREATE INDEX "idx_items_list_id" ON "todo"."items" ("list_id");
