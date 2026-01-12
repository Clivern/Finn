
export const PLATFORM_ROLE_ADMIN = 'admin'
export const PLATFORM_ROLE_REGULAR = 'regular'
export const WORKSPACE_ROLE_READONLY = 'readonly'
export const WORKSPACE_ROLE_ADMIN = 'admin'
export const WORKSPACE_ROLE_OWNER = 'owner'

const PLATFORM_ROLES = new Set([
  PLATFORM_ROLE_ADMIN,
  PLATFORM_ROLE_REGULAR,
])

const WORKSPACE_MEMBERSHIP_ROLES = new Set([
  WORKSPACE_ROLE_READONLY,
  WORKSPACE_ROLE_ADMIN,
  WORKSPACE_ROLE_OWNER,
])

/**
 * Platform role for the logged-in user (`admin` | `regular`),
 * or `null` if missing / not recognized.
 */
export function getPlatformRole(user) {
  const r = user?.role
  if (r == null || r === '') return null
  return PLATFORM_ROLES.has(r) ? r : null
}

/**
 * Workspace membership role (`readonly` | `admin` | `owner`),
 * or `null` if missing / not a membership role.
 */
export function getWorkspaceMembershipRole(workspace) {
  const r = workspace?.role
  if (r == null || r === '') return null
  return WORKSPACE_MEMBERSHIP_ROLES.has(r) ? r : null
}
