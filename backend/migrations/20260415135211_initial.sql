-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" character varying(255) NOT NULL,
  "password_hash" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email")
);
-- Create "habits" table
CREATE TABLE "habits" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "name" character varying(50) NOT NULL,
  "description" character varying(200) NULL,
  "color" character varying(7) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "habits_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_habits_user_id" to table: "habits"
CREATE INDEX "idx_habits_user_id" ON "habits" ("user_id");
-- Create "check_ins" table
CREATE TABLE "check_ins" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "habit_id" uuid NOT NULL,
  "checked_date" date NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "check_ins_habit_id_checked_date_key" UNIQUE ("habit_id", "checked_date"),
  CONSTRAINT "check_ins_habit_id_fkey" FOREIGN KEY ("habit_id") REFERENCES "habits" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_check_ins_checked_date" to table: "check_ins"
CREATE INDEX "idx_check_ins_checked_date" ON "check_ins" ("checked_date");
-- Create index "idx_check_ins_habit_id" to table: "check_ins"
CREATE INDEX "idx_check_ins_habit_id" ON "check_ins" ("habit_id");
