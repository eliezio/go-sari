package sari

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"gopkg.in/yaml.v3"
)

const (
	StdGrantClassNone  = "none"
	StdGrantClassQuery = "query"
	StdGrantClassCru   = "cru"
	StdGrantClassCrud  = "crud"

	DefaultGrantClass = StdGrantClassQuery
)

type Permission struct {
	database         string
	schemas          []string
	grantClass       string
	actualGrantClass string
	validity         ValidityPeriod
}

func (p *Permission) UnmarshalYAML(node *yaml.Node) error {
	raw := &struct {
		Database   string         `yaml:"db" valid:"required,lowercase"`
		Schemas    []string       `yaml:"schemas,omitempty"`
		GrantClass string         `yaml:"grant_class,omitempty"`
		Validity   ValidityPeriod `yaml:",inline" valid:"period"`
	}{Schemas: []string{}, GrantClass: DefaultGrantClass}
	if err := node.Decode(raw); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(raw); err != nil {
		return fmt.Errorf("line %d: %w", node.Line, err)
	}
	p.database = raw.Database
	p.schemas = raw.Schemas
	p.grantClass = raw.GrantClass
	p.validity = raw.Validity
	return nil
}
