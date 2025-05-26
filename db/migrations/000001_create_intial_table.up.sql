CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TYPE transaction_type AS ENUM ('expense', 'income');

CREATE TABLE IF NOT EXISTS users (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    email CITEXT UNIQUE NOT NULL,
    password_hash BYTEA NOT NULL,
    activated BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS token (
    hash BYTEA PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    scope TEXT NOT NULL
);

CREATE TABLE category (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transaction (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    type transaction_type NOT NULL,
    category_id uuid REFERENCES category(id) ON DELETE CASCADE NOT NULL,
    amount INTEGER NOT NULL,
    date DATE NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);


INSERT INTO category (name, description) VALUES
    ('food', 'Chicken feed and food expenses'),
    ('medicine', 'Veterinary and medicine expenses'),
    ('tools', 'Farm tools and equipment'),
    ('chicken_purchase', 'Chicken purchases'),
    ('chicken_sale', 'Chicken sales'),
    ('salary', 'Supervisor salary'),
    ('other', 'Other');