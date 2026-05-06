CREATE TYPE sex AS ENUM ('male', 'female');

CREATE TABLE users (
    id UUID PRIMARY KEY,
    first_name VARCHAR(80) NOT NULL,
    last_name VARCHAR(80) NOT NULL,
    email VARCHAR(255) NOT NULL,
    birth_date DATE NOT NULL,
    bio VARCHAR(500),
    toiler_score SMALLINT,
    alcohol_info VARCHAR(120),
    smoking_info VARCHAR(120),
    sex sex NOT NULL,
    height_cm SMALLINT,
    city_id BIGINT REFERENCES cities(id) ON DELETE SET NULL,
    primary_photo_object_key TEXT NOT NULL,
    primary_photo_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
