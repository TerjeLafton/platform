-- Modify "users" table
ALTER TABLE "id"."users" ADD COLUMN "name" text NOT NULL DEFAULT 'User', ADD COLUMN "avatar" bytea NULL, ADD COLUMN "avatar_content_type" text NULL;
ALTER TABLE "id"."users" ADD CONSTRAINT "users_name_check" CHECK ((length(name) >= 1) AND (length(name) <= 100));
ALTER TABLE "id"."users" ALTER COLUMN "name" DROP DEFAULT;
