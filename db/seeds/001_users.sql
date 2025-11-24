INSERT INTO users (name, email)
VALUES
    ('Seed User 1', 'seed1@example.com'),
    ('Seed User 2', 'seed2@example.com')
ON CONFLICT (email) DO NOTHING;
