const ADMIN_ROLES = new Set(['owner', 'admin', 'system_admin']);

export const isAdminRole = (role?: string | null) => {
  if (!role) return false;
  return ADMIN_ROLES.has(role);
};
