package access

import "testing"

// BR-8.1 — the matrix, spelled out. Every cell is asserted, so adding a
// permission without deciding it for every role fails here.
func TestBR8_1_Matrix(t *testing.T) {
	all := []Permission{ContentRead, ContentWrite, MediaWrite, ParamsContent, ParamsSystem,
		QuotationsRead, ApplicationsRead, UsersManage, AuditRead}
	want := map[Role][]Permission{
		RoleSuperadmin: all,
		RoleEditor:     {ContentRead, ContentWrite, MediaWrite, ParamsContent},
		RoleSales:      {QuotationsRead},
		RoleHR:         {ApplicationsRead},
	}
	for _, r := range Roles {
		allowed := map[Permission]bool{}
		for _, p := range want[r] {
			allowed[p] = true
		}
		for _, p := range all {
			if Can(r, p) != allowed[p] {
				t.Errorf("Can(%s, %s) = %v, want %v", r, p, Can(r, p), allowed[p])
			}
		}
	}
}

func TestDenyByDefault(t *testing.T) {
	if Can("intern", ContentRead) {
		t.Fatal("an unknown role was granted a permission")
	}
	if Can(RoleEditor, "anything.else") {
		t.Fatal("an unknown permission was granted")
	}
}

// CVs are personal data: sales and editors must never see them.
func TestCVsAreHROnly(t *testing.T) {
	for _, r := range []Role{RoleEditor, RoleSales} {
		if Can(r, ApplicationsRead) {
			t.Errorf("%s can read job applications", r)
		}
	}
}
