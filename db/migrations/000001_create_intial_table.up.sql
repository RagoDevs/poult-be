CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";


CREATE TYPE transaction_type AS ENUM ('expense', 'income');
CREATE TYPE chicken_type AS ENUM ('hen', 'cock', 'chicks');
CREATE TYPE reason_type AS ENUM ('purchase', 'sale', 'birth', 'death', 'gift', 'other');


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

CREATE TABLE chicken (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    type chicken_type NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT positive_quantity CHECK (quantity >= 0)
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

CREATE TABLE chicken_history (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    chicken_type chicken_type NOT NULL,
    quantity_change INTEGER NOT NULL,
    reason reason_type NOT NULL,
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

INSERT INTO chicken (type, quantity) VALUES
    ('hen', 0),
    ('cock', 0),
    ('chicks', 0);