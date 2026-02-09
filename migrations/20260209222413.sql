-- Create "templates" table
CREATE TABLE "todo"."templates" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "user_id" uuid NOT NULL,
  "title" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "templates_user_id_title_key" UNIQUE ("user_id", "title"),
  CONSTRAINT "templates_title_check" CHECK ((length(title) >= 1) AND (length(title) <= 100))
);
-- Create index "idx_templates_user_id" to table: "templates"
CREATE INDEX "idx_templates_user_id" ON "todo"."templates" ("user_id");
-- Create "template_items" table
CREATE TABLE "todo"."template_items" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "template_id" uuid NOT NULL,
  "title" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "template_items_template_id_fkey" FOREIGN KEY ("template_id") REFERENCES "todo"."templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "template_items_title_check" CHECK ((length(title) >= 1) AND (length(title) <= 500))
);
-- Create index "idx_template_items_template_id" to table: "template_items"
CREATE INDEX "idx_template_items_template_id" ON "todo"."template_items" ("template_id");
