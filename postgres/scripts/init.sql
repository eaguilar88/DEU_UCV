CREATE SCHEMA IF NOT EXISTS "deu";

CREATE TYPE IF NOT EXISTS "provider_type" AS ENUM (
  'user',
  'extension_group'
);

CREATE TYPE IF NOT EXISTS "education_level" AS ENUM (
  'bachiller',
  'universitario_incompleta',
  'universitaria_completa',
  'posgrado'
);

CREATE TYPE IF NOT EXISTS "request_status" AS ENUM (
  'created',
  'under_review',
  'rejected',
  'approved'
);

CREATE TYPE IF NOT EXISTS "contact_type" AS ENUM (
  'user',
  'course',
  'extension_group'
);

CREATE TYPE IF NOT EXISTS "course_participant_status" AS ENUM (
  'abandoned',
  'approved',
  'failed'
);

CREATE TABLE IF NOT EXISTS "deu"."users" (
  "id" integer PRIMARY KEY,
  "ci" integer UNIQUE,
  "username" varchar UNIQUE,
  "first_name" varchar,
  "last_name" varchar,
  "date_of_birth" date,
  "gender" varchar,
  "education" education_level,
  "adress" varchar,
  "created_at" timestamp,
  "updated_at" timestamp,
  "salt" char(22)
  "password" char(31)
);

INSERT INTO deu.users(username, first_name, last_name, password, created_at, updated_at)
VALUES  ("superadmin", "John","Doe", "$2y$10$mCZ9kIFBJ27JRYz3UhXWZ.e2/G3NLoTT0.2QTKYW.4vL5WETXvly6", "2024-02-10 07:40:00", "2024-02-10 07:40:00");

CREATE TABLE IF NOT EXISTS "deu"."providers" (
  "id" integer PRIMARY KEY,
  "entity_id" integer NOT NULL,
  "entity_type" provider_type,
  "registration_date" date,
  "code" varchar,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE IF NOT EXISTS "deu"."extension_group" (
  "id" integer PRIMARY KEY,
  "name" varchar,
  "description" text,
  "objective" text,
  "action" text,
  "reach" text,
  "path" varchar
);

CREATE TABLE IF NOT EXISTS "deu"."endorsment_requests" (
  "id" integer PRIMARY KEY,
  "user_id" integer,
  "status" request_status,
  "path" string,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE IF NOT EXISTS "deu"."courses" (
  "id" integer PRIMARY KEY,
  "user_id" integer,
  "endorsment_id" integer,
  "endorsed_by" varchar,
  "objectives" text,
  "content" text,
  "cost" float,
  "location" varchar,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE IF NOT EXISTS "deu"."contact_info" (
  "id" integer PRIMARY KEY,
  "entity_id" integer,
  "entity_type" contact_type,
  "email" varchar,
  "phone" varchar
);

CREATE TABLE IF NOT EXISTS "deu"."course_cycles" (
  "id" integer PRIMARY KEY,
  "course_id" integer,
  "start_date" datetime,
  "end_date" datetime,
  "inscription_date" datetime,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE IF NOT EXISTS "deu"."course_participants" (
  "id" integer PRIMARY KEY,
  "user_id" integer,
  "course_cycle_id" integer,
  "participant_status" course_participant_status,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE IF NOT EXISTS "deu"."roles" (
  "id" integer PRIMARY KEY,
  "name" varchar UNIQUE,
  "description" text,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE IF NOT EXISTS "deu"."user_roles" (
  "user_id" integer,
  "role_id" integer,
  "created_at" timestamp,
  "updated_at" timestamp,
  PRIMARY KEY ("user_id", "role_id")
);

CREATE TABLE IF NOT EXISTS "deu"."activities" (
  "id" integer PRIMARY KEY,
  "name" varchar,
  "description" text,
  "date" datetime,
  "path" varchar,
  "location" varchar,
  "assistants" integer
);

CREATE UNIQUE INDEX ON "deu"."providers" ("entity_id", "entity_type");

CREATE UNIQUE INDEX ON "deu"."course_participants" ("user_id", "course_cycle_id");

COMMENT ON TABLE "deu"."roles" IS 'Los roles son: root - admin - facilitador - estudiante - extensión';

ALTER TABLE "deu"."endorsment_requests" ADD FOREIGN KEY ("user_id") REFERENCES "deu"."users" ("id") ON DELETE CASCADE;

ALTER TABLE "deu"."courses" ADD FOREIGN KEY ("user_id") REFERENCES "deu"."users" ("id") ON DELETE CASCADE;

ALTER TABLE "deu"."courses" ADD FOREIGN KEY ("endorsment_id") REFERENCES "deu"."endorsment_requests" ("id") ON DELETE CASCADE;

ALTER TABLE "deu"."user_roles" ADD FOREIGN KEY ("role_id") REFERENCES "deu"."roles" ("id");

ALTER TABLE "deu"."user_roles" ADD FOREIGN KEY ("user_id") REFERENCES "deu"."users" ("id");

ALTER TABLE "deu"."course_cycles" ADD FOREIGN KEY ("course_id") REFERENCES "deu"."courses" ("id");

ALTER TABLE "deu"."course_participants" ADD FOREIGN KEY ("course_cycle_id") REFERENCES "deu"."course_cycles" ("id");

ALTER TABLE "deu"."course_participants" ADD FOREIGN KEY ("user_id") REFERENCES "deu"."users" ("id");

ALTER TABLE "deu"."contact_info" ADD FOREIGN KEY ("entity_id") REFERENCES "deu"."courses" ("id");

ALTER TABLE "deu"."contact_info" ADD FOREIGN KEY ("entity_id") REFERENCES "deu"."users" ("id");

ALTER TABLE "deu"."contact_info" ADD FOREIGN KEY ("entity_id") REFERENCES "deu"."extension_group" ("id");

ALTER TABLE "deu"."providers" ADD FOREIGN KEY ("entity_id") REFERENCES "deu"."users" ("id");

ALTER TABLE "deu"."providers" ADD FOREIGN KEY ("entity_id") REFERENCES "deu"."extension_group" ("id");

