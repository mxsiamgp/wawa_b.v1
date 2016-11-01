package permission

import "github.com/deckarep/golang-set"

func IsInclude(actualPerms, expectedPerms []string) bool {
	expectedPermSet := makeFlatPermissionSet(expectedPerms)
	return makeFlatPermissionSet(actualPerms).Intersect(expectedPermSet).Cardinality() == expectedPermSet.Cardinality()
}

func makeFlatPermissionSet(perms []string) mapset.Set {
	permSet := mapset.NewThreadUnsafeSet()
	for _, perm := range perms {
		permSet.Add(perm)
	}
	return permSet
}
