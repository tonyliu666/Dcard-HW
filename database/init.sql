-- Create the advertisement table
CREATE TABLE IF NOT EXISTS advertisement (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL,
    conditions JSONB NOT NULL
);

-- Inserting some initial data
INSERT INTO advertisement (title, start_at, end_at, conditions)
VALUES ('AD55', '2023-12-10 03:00:00', '2023-12-31 16:00:00', '{"ageStart": 20, "ageEnd": 30, "country": ["TW", "JP"], "platform": ["android", "ios"]}');
INSERT INTO advertisement (title, start_at, end_at, conditions)
VALUES ('AD56', '2023-12-10 03:00:00', '2023-12-31 16:00:00', '{"ageStart": 20, "ageEnd": 30, "country": ["TW", "JP"], "platform": ["android", "ios"]}');
INSERT INTO advertisement (title, start_at, end_at, conditions)
VALUES ('AD403', '2024-12-27 13:00:00', '2023-12-27 16:23:15', '{"ageStart": 20, "ageEnd": 40,"gender": "F",  "country": ["TW", "JP"], "platform": ["android", "ios"]}');




