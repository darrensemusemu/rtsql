CREATE TABLE IF NOT EXISTS "user" (
	id serial primary key,
	email text not null,
	first_name text,
	last_name text
);

