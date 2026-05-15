-- Normalize legacy operator backoffice role to sales.
UPDATE users
SET role = 'sales'
WHERE role = 'operator';
