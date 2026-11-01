DROP TABLE IF EXISTS vehicle CASCADE;

DROP TABLE IF EXISTS customer CASCADE;

DROP TABLE IF EXISTS user_role CASCADE;

DROP TABLE IF EXISTS "role" CASCADE;

DROP TABLE IF EXISTS "user" CASCADE;

DROP INDEX IF EXISTS idx_customer_user_id;

DROP INDEX IF EXISTS idx_vehicle_customer_id;

DROP INDEX IF EXISTS idx_user_role_user_id;

DROP INDEX IF EXISTS idx_user_role_role_id;
