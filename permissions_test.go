package sari

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadUsersConfig(t *testing.T) {
	var data = `
- login: leroy.trent@acme.com
  fallback_grant_class: query
  permissions:
    - db: "us-east-1/borders"
      schemas:
        - db_borders
        - ebooks
      grant_class: crud
      not_valid_after: 2020-05-01T20:32:48.0+01:00
    - db: "eu-west-2/whsmith"
      grant_class: crud
      not_valid_after: 2020-05-26T10:22:00.0+01:00
    - db: "*/*"

- login: bridget.huntington-whiteley@acme.com

- login: valerie.tennant@acme.com
  permissions:
    - db: "eu-west-2/blackwells"
      schemas: [db_blackwells]
      grant_class: cru
`
	timeRef, _ := time.Parse(time.RFC3339, "2020-05-15T22:24:51.0+01:00")
	validityChecker := NewTrackingValidityChecker(timeRef)
	databases := []*Database{
		{id: "us-east-1/borders", status: DbStatusAccessible, schemas: []string{"db_borders", "ebooks"}},
		{id: "eu-west-2/blackwells", status: DbStatusAccessible, schemas: []string{"db_blackwells"}},
		{id: "eu-west-2/foyles", status: DbStatusDisabled},
		{id: "eu-west-2/blackwells-recover", status: DbStatusAbsent},
		{id: "eu-west-2/whsmith", status: DbStatusEnabled, schemas: []string{"db_whsmith"}},
	}

	users, err := LoadUsersConfig(strings.NewReader(data), databases, validityChecker)
	assert.NoError(t, err)
	notValidAfter1, _ := time.Parse(time.RFC3339, "2020-05-01T20:32:48.0+01:00")
	notValidAfter2, _ := time.Parse(time.RFC3339, "2020-05-26T10:22:00.0+01:00")
	assert.Equal(t, notValidAfter2, validityChecker.GetNextTransition())
	assert.Len(t, users, 3)
	assert.ElementsMatch(t, []*User{
		{login: "leroy.trent@acme.com", fallbackGrantClass: StdGrantClassQuery, permissions: []*Permission{
			{database: "us-east-1/borders", schemas: []string{"db_borders", "ebooks"}, grantClass: StdGrantClassCrud, actualGrantClass: StdGrantClassQuery, validity: ValidityPeriod{NotValidAfter: notValidAfter1}},
			{database: "eu-west-2/whsmith", schemas: []string{"db_whsmith"}, grantClass: StdGrantClassCrud, actualGrantClass: StdGrantClassCrud, validity: ValidityPeriod{NotValidAfter: notValidAfter2}},
			{database: "eu-west-2/blackwells", schemas: []string{"db_blackwells"}, grantClass: StdGrantClassQuery, actualGrantClass: StdGrantClassQuery},
		}},
		{login: "bridget.huntington-whiteley@acme.com", fallbackGrantClass: StdGrantClassNone, permissions: []*Permission{}},
		{login: "valerie.tennant@acme.com", fallbackGrantClass: StdGrantClassNone, permissions: []*Permission{
			{database: "eu-west-2/blackwells", schemas: []string{"db_blackwells"}, grantClass: StdGrantClassCru, actualGrantClass: StdGrantClassCru},
		}},
	}, users)
}
