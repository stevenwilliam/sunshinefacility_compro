// Package access is the permissions matrix (BR-8.1). Deny by default: a
// permission not listed for a role is refused, and every admin handler
// declares the permission it needs.
package access

type Role string

const (
	RoleSuperadmin Role = "superadmin"
	RoleEditor     Role = "editor"
	RoleSales      Role = "sales"
	RoleHR         Role = "hr"
)

var Roles = []Role{RoleSuperadmin, RoleEditor, RoleSales, RoleHR}

func ParseRole(s string) (Role, bool) {
	for _, r := range Roles {
		if string(r) == s {
			return r, true
		}
	}
	return "", false
}

type Permission string

const (
	// Website content and the media library.
	ContentRead  Permission = "content.read"
	ContentWrite Permission = "content.write"
	MediaWrite   Permission = "media.write"
	// Option and contact parameters (building types, locations, contact
	// details, social links) vs system parameters (recipients, thresholds,
	// analytics ID) which change how the system behaves.
	ParamsContent Permission = "params.content"
	ParamsSystem  Permission = "params.system"
	// Inbox. CVs are personal data: only HR and superadmin can open them.
	QuotationsRead   Permission = "quotations.read"
	ApplicationsRead Permission = "applications.read"
	// Administration.
	UsersManage Permission = "users.manage"
	AuditRead   Permission = "audit.read"
)

var matrix = map[Role]map[Permission]bool{
	RoleSuperadmin: {
		ContentRead: true, ContentWrite: true, MediaWrite: true,
		ParamsContent: true, ParamsSystem: true,
		QuotationsRead: true, ApplicationsRead: true,
		UsersManage: true, AuditRead: true,
	},
	RoleEditor: {
		ContentRead: true, ContentWrite: true, MediaWrite: true, ParamsContent: true,
	},
	RoleSales: {QuotationsRead: true},
	RoleHR:    {ApplicationsRead: true},
}

// Can is the only authorisation question in the product.
func Can(r Role, p Permission) bool {
	return matrix[r][p]
}

// Permissions lists what a role holds, for the CMS to hide what it cannot do.
func Permissions(r Role) []Permission {
	var out []Permission
	for _, p := range []Permission{ContentRead, ContentWrite, MediaWrite, ParamsContent, ParamsSystem,
		QuotationsRead, ApplicationsRead, UsersManage, AuditRead} {
		if matrix[r][p] {
			out = append(out, p)
		}
	}
	return out
}
