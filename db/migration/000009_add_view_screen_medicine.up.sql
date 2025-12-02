INSERT INTO permissions (name, description) VALUES ('VIEW_SCREEN_MEDICINE', 'Permission to view medicine screen');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name = 'VIEW_SCREEN_MEDICINE';
