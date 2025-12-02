-- Remove VIEW_SCREEN_USER_ROLE from role_permissions
DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE name = 'VIEW_SCREEN_USER_ROLE');

-- Remove VIEW_SCREEN_USER_ROLE permission
DELETE FROM permissions WHERE name = 'VIEW_SCREEN_USER_ROLE';
