-- Function để tự động update cột updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Bảng roles
CREATE TABLE roles (
  id SERIAL PRIMARY KEY,
  name VARCHAR UNIQUE NOT NULL,        -- admin, doctor, nurse
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_roles_updated_at
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Bảng permissions
CREATE TABLE permissions (
  id SERIAL PRIMARY KEY,
  name VARCHAR UNIQUE NOT NULL,        -- VIEW_SCREEN_1, VIEW_SCREEN_2, ...
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_permissions_updated_at
BEFORE UPDATE ON permissions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Mapping role ↔ permission (many-to-many)
CREATE TABLE role_permissions (
  role_id INT REFERENCES roles(id) ON DELETE CASCADE,
  permission_id INT REFERENCES permissions(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (role_id, permission_id)
);

CREATE TRIGGER trg_role_permissions_updated_at
BEFORE UPDATE ON role_permissions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Mapping user ↔ role (many-to-many)
CREATE TABLE user_roles (
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, role_id)
);

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_user_roles_updated_at
BEFORE UPDATE ON user_roles
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Seed dữ liệu mẫu cho roles
INSERT INTO roles (name, description) VALUES
('admin', 'Quản trị hệ thống'),
('doctor', 'Bác sĩ'),
('nurse', 'Điều dưỡng');

-- Seed dữ liệu mẫu cho permissions (màn hình)
INSERT INTO permissions (name, description) VALUES
('VIEW_SCREEN_1', 'Truy cập màn hình 1'),
('VIEW_SCREEN_2', 'Truy cập màn hình 2'),
('VIEW_SCREEN_3', 'Truy cập màn hình 3');

-- Mapping role ↔ permission
-- admin: screen 1,2,3
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin';

-- doctor: screen 1,2
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('VIEW_SCREEN_1','VIEW_SCREEN_2')
WHERE r.name = 'doctor';

-- nurse: screen 2,3
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('VIEW_SCREEN_2','VIEW_SCREEN_3')
WHERE r.name = 'nurse';