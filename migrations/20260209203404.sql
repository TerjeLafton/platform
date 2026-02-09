-- Create "list_members" table
CREATE TABLE "todo"."list_members" (
  "list_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "name" text NOT NULL,
  "email" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("list_id", "user_id"),
  CONSTRAINT "list_members_list_id_fkey" FOREIGN KEY ("list_id") REFERENCES "todo"."lists" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_list_members_user_id" to table: "list_members"
CREATE INDEX "idx_list_members_user_id" ON "todo"."list_members" ("user_id");
