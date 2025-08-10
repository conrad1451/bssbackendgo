-- CHQ: Gemini AI generated
CREATE TABLE gamecheckpoints (
    id SERIAL PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL,
    checkpoint_data TEXT NOT NULL
);