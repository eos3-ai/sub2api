-- Rename legacy backoffice role value from operator to sales.
-- No compatibility path: all backoffice non-admin accounts use role='sales'.
UPDATE users
SET role = 'sales'
WHERE role = 'operator';
