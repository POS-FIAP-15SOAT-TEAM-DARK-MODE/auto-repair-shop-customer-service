package repository

const (
	getUserByEmail   = `SELECT id, name, email, password_hash FROM "user" WHERE email = $1`
	getRolesById     = `SELECT r.name FROM "user_role" ur JOIN "role" r ON r.id = ur.role_id WHERE ur.user_id = $1`
	createUserQuery  = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	deleteRolesQuery = `DELETE FROM user_role WHERE user_id = $1`
	updateRoleQuery  = `INSERT INTO user_role (id, user_id, role_id) SELECT $1, $2, r.id FROM "role" r WHERE r.name = $3`
	updateUserQuery  = `UPDATE "user" SET name = $1, email = $2 WHERE id = $3`
	deleteUserQuery  = `DELETE FROM "user" WHERE id = $1`
)
