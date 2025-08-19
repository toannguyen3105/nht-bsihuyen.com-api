DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP TRIGGER IF EXISTS trg_user_roles_updated_at ON user_roles;
DROP TRIGGER IF EXISTS trg_role_permissions_updated_at ON role_permissions;
DROP TRIGGER IF EXISTS trg_permissions_updated_at ON permissions;
DROP TRIGGER IF EXISTS trg_roles_updated_at ON roles;

DROP FUNCTION IF EXISTS set_updated_at;

DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;