-- Create "point_ledger" table
CREATE TABLE "point_ledger" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "habit_id" uuid NULL,
  "type" character varying(10) NOT NULL,
  "amount" integer NOT NULL,
  "reason" character varying(100) NOT NULL,
  "idempotency_key" character varying(255) NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "point_ledger_idempotency_key_key" UNIQUE ("idempotency_key"),
  CONSTRAINT "point_ledger_habit_id_fkey" FOREIGN KEY ("habit_id") REFERENCES "habits" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "point_ledger_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "point_ledger_amount_check" CHECK (amount > 0)
);
-- Create index "idx_point_ledger_user_id" to table: "point_ledger"
CREATE INDEX "idx_point_ledger_user_id" ON "point_ledger" ("user_id");
