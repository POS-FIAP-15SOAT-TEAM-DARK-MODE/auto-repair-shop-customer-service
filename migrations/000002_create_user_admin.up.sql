-- Create the first user admin of the application
INSERT INTO "user" (id, name, email, password_hash)
VALUES ('3b123072-5ad5-4df8-8d3b-3c2770f11b2a', 'admin', 'admin@email.com', '$2a$10$BCYYfulkmNYzgaMhpe3FAeWApyiQC6YAI6nzgsb8aWbT2LfgFi18C');

-- Insert role to this user
INSERT INTO "user_role" (id, user_id, role_id)
VALUES ('0b4b3baf-ff7b-461f-9b90-81f6b0c5e5b2', '3b123072-5ad5-4df8-8d3b-3c2770f11b2a', '1');

-- TODO: Move this to a seed file
