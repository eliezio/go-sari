package sari

import (
	"fmt"
	"io"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/danwakefield/fnmatch"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/emirpasic/gods/sets/hashset"
	"gopkg.in/yaml.v3"
)

const (
	DefaultFallbackGrantClass = StdGrantClassNone
)

type User struct {
	login              string
	fallbackGrantClass string
	permissions        []*Permission
}

func (u *User) UnmarshalYAML(node *yaml.Node) error {
	raw := &struct {
		Login              string        `yaml:"login" valid:"email,lowercase,required"`
		FallbackGrantClass string        `yaml:"fallback_grant_class,omitempty"`
		Permissions        []*Permission `yaml:"permissions,omitempty"`
	}{FallbackGrantClass: DefaultFallbackGrantClass, Permissions: []*Permission{}}
	if err := node.Decode(raw); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(raw); err != nil {
		return fmt.Errorf("line %d: %w", node.Line, err)
	}
	u.login = raw.Login
	u.fallbackGrantClass = raw.FallbackGrantClass
	u.permissions = raw.Permissions
	return nil
}

func LoadUsersConfig(r io.Reader, databases []*Database, checker ValidityChecker) ([]*User, error) {
	var users []*User
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&users); err != nil {
		return nil, err
	}
	if err := validateUsers(users); err != nil {
		return nil, err
	}
	if err := expandPermissions(users, databases, checker); err != nil {
		return nil, err
	}
	return users, nil
}

func validateUsers(users []*User) error {
	ids := hashset.New()
	for _, user := range users {
		if ids.Contains(user.login) {
			return fmt.Errorf("duplicated user login '%s'", user.login)
		}
		ids.Add(user.login)
	}
	return nil
}

func expandPermissions(users []*User, databases []*Database, checker ValidityChecker) error {
	dbUsage, dbIndex := enumEnabledDatabases(databases)
	for _, user := range users {
		clearUsage(dbUsage)
		newPermissions := make([]*Permission, 0)
		for j, perm := range user.permissions {
			if checker.Check(perm.validity) {
				perm.actualGrantClass = perm.grantClass
			} else {
				perm.actualGrantClass = user.fallbackGrantClass
			}
			ids, err := expandSpec(perm.database, dbUsage)
			if err != nil {
				return fmt.Errorf("user[%s].permission[%d]: %w", user.login, j, err)
			}
			for _, id := range ids {
				newPerm := *perm
				newPerm.database = id
				db := dbIndex[id]
				if len(newPerm.schemas) == 0 {
					newPerm.schemas = []string{db.schemas[0]}
				} else if diff := diffSlices(newPerm.schemas, db.schemas); diff != nil {
					return fmt.Errorf("user[%s].permission[%d]: unknown schema(s)='%v' for database='%s'",
						user.login, j, diff, db.id)
				}
				newPermissions = append(newPermissions, &newPerm)
			}
		}
		user.permissions = newPermissions
	}
	return nil
}

func enumEnabledDatabases(databases []*Database) (*linkedhashmap.Map, map[string]*Database) {
	dbUsage := linkedhashmap.New()
	index := map[string]*Database{}
	for _, db := range databases {
		if db.status >= DbStatusEnabled {
			dbUsage.Put(db.id, false)
			index[db.id] = db
		}
	}
	return dbUsage, index
}

func clearUsage(usage *linkedhashmap.Map) {
	for _, key := range usage.Keys() {
		usage.Put(key, false)
	}
}

func expandSpec(spec string, usage *linkedhashmap.Map) (ids []string, err error) {
	if isWildCard(spec) {
		pattern := spec
		for it := usage.Iterator(); it.Next(); {
			if used := it.Value().(bool); used {
				continue
			}
			id := it.Key().(string)
			if fnmatch.Match(pattern, id, fnmatch.FNM_PATHNAME) {
				ids = append(ids, id)
				usage.Put(id, true)
			}
		}
	} else {
		id := spec
		if used, found := usage.Get(id); !found {
			return nil, fmt.Errorf("id='%s' not found", id)
		} else if used.(bool) {
			return nil, fmt.Errorf("duplicated match for id='%s'", id)
		}
		ids = append(ids, id)
		usage.Put(id, true)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no match found for spec '%s'", spec)
	}
	return ids, nil
}

// diffSlices function computes the difference between lists 'a' and 'b', i.e., 'a - b'.
// Any duplicated element in 'a' will appear multiple times in the resulting slice.
func diffSlices(a, b []string) []string {
	var diff []string
	for _, aItem := range a {
		if !findItem(b, aItem) {
			diff = append(diff, aItem)
		}
	}
	return diff
}

func findItem(slice []string, item string) bool {
	for _, s := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func isWildCard(s string) bool {
	return strings.ContainsAny(s, "*?")
}
