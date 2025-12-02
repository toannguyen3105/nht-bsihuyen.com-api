-- Remove permissions (cascade delete will remove from role_permissions)
DELETE FROM permissions WHERE name = 'VIEW_SCREEN_PERMISSION';
