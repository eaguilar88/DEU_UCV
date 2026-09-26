package entities

import "slices"

// Viewer identifies who is making a request on endpoints that serve both anonymous and
// authenticated callers. The zero value is an anonymous caller.
type Viewer struct {
	UserID  string
	Roles   []string
	Faculty Faculty
}

func (v Viewer) IsAnonymous() bool {
	return v.UserID == ""
}

func (v Viewer) HasRole(role string) bool {
	return slices.Contains(v.Roles, role)
}

// IsGlobalAdmin reports whether the viewer administers every faculty (root or deu_admin).
func (v Viewer) IsGlobalAdmin() bool {
	return v.HasRole(RoleNameFromID(RoleRoot)) || v.HasRole(RoleNameFromID(RoleDeuAdmin))
}

// IsFacultyAdmin reports whether the viewer holds the faculty_admin role, regardless of faculty.
func (v Viewer) IsFacultyAdmin() bool {
	return v.HasRole(RoleNameFromID(RoleFacultyAdmin))
}

// IsFacultyAdminOf reports whether the viewer is a faculty_admin for any of the given
// faculties. A DEU faculty claim is DEU-wide and covers every faculty, matching
// facultyscope.IsGlobalAdmin.
func (v Viewer) IsFacultyAdminOf(faculties ...Faculty) bool {
	if !v.IsFacultyAdmin() {
		return false
	}
	return v.Faculty == FacultyDEU || slices.Contains(faculties, v.Faculty)
}

// CanManageGroup reports whether the viewer may see a group's private data (members, the
// project document, participant lists of its activities): the group's owner (its
// group_admin account after approval), root and deu_admin, or a faculty_admin of one of
// the group's faculties. group_admin is deliberately not an admin role here, so one
// group's admin can't read another group's private data.
func (v Viewer) CanManageGroup(group ExtensionGroup) bool {
	if v.IsAnonymous() {
		return false
	}
	if group.Owner != nil && group.Owner.ID == v.UserID {
		return true
	}
	return v.IsGlobalAdmin() || v.IsFacultyAdminOf(group.Faculty...)
}
