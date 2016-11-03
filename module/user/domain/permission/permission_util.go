package permission

import "github.com/fatih/set"

// 创建权限集合
func PermissionSet(perms []string) set.Interface {
	s := set.NewNonTS()
	for _, perm := range perms {
		s.Add(perm)
	}
	return s
}

// 创建权限集
func Permissions(perms []string) []string {
	return set.StringSlice(PermissionSet(perms))
}

// 是否包含权限集
func IsInclude(actualPerms, expectedPerms []string) bool {
	expectedPermSet := PermissionSet(expectedPerms)
	return set.Intersection(PermissionSet(actualPerms), expectedPermSet).Size() == expectedPermSet.Size()
}
