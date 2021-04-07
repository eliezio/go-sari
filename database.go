package sari

import (
	"fmt"
	"io"

	"github.com/asaskevich/govalidator"
	"github.com/emirpasic/gods/sets/hashset"
	"gopkg.in/yaml.v3"
)

const (
	DbStatusAbsent = iota
	DbStatusDisabled
	DbStatusEnabled
	DbStatusAutoEnabled
	DbStatusAccessible
)

type Database struct {
	id                string
	status            int
	masterPasswordURL string
	masterPassword    string
	schemas           []string
}

func (u *Database) UnmarshalYAML(node *yaml.Node) error {
	raw := &struct {
		ID                string `yaml:"id" valid:"required"`
		Enabled           bool   `yaml:"enabled,omitempty"`
		MasterPasswordURL string `yaml:"master_password" valid:"required"`
	}{Enabled: true}
	if err := node.Decode(raw); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(raw); err != nil {
		return fmt.Errorf("line %d: %w", node.Line, err)
	}
	u.id = raw.ID
	if raw.Enabled {
		u.status = DbStatusEnabled
	} else {
		u.status = DbStatusDisabled
	}
	u.masterPasswordURL = raw.MasterPasswordURL
	return nil
}

func LoadDatabasesConfig(r io.Reader, pwdResolver PasswordResolver) ([]*Database, error) {
	var databases []*Database
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&databases); err != nil {
		return nil, err
	}
	if err := validateDatabases(databases); err != nil {
		return nil, err
	}
	if err := resolvePasswords(databases, pwdResolver); err != nil {
		return nil, err
	}
	return databases, nil
}

func validateDatabases(databases []*Database) error {
	ids := hashset.New()
	for _, db := range databases {
		if ids.Contains(db.id) {
			return fmt.Errorf("duplicated database ID '%s'", db.id)
		}
		ids.Add(db.id)
	}
	return nil
}

func resolvePasswords(databases []*Database, pwdResolver PasswordResolver) error {
	for _, db := range databases {
		if db.status < DbStatusEnabled {
			continue
		}
		pwd, err := pwdResolver.Resolve(db.masterPasswordURL)
		if err != nil {
			return fmt.Errorf("database[%s]: %w", db.id, err)
		}
		db.masterPassword = pwd
	}
	return nil
}
