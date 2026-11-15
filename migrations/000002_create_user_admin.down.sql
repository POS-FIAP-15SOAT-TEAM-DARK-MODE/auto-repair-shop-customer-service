-- Remove user Admin and relations
DELETE FROM "user_role" WHERE id = '0b4b3baf-ff7b-461f-9b90-81f6b0c5e5b2';

-- Remove user from table
DELETE FROM "user" WHERE id = '3b123072-5ad5-4df8-8d3b-3c2770f11b2a';
