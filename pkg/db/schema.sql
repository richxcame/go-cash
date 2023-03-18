CREATE TABLE cashes (
	uuid uuid PRIMARY KEY,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	client varchar(255) NOT NULL,
	contact varchar(255) NOT NULL,
	amount numeric NOT NULL,
	detail varchar(255),
	note varchar(255)
);

CREATE TABLE ranges (
	uuid uuid PRIMARY KEY,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	client varchar(255) NOT NULL,
	detail varchar(255),
	note  varchar(255)
);

CREATE TABLE users (
	username varchar(255) PRIMARY KEY,
	password varchar(255) NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL
);

CREATE TABLE clients (
	api_key uuid PRIMARY KEY,
	name varchar(255) NOT NULL
);
