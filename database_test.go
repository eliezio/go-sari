package sari

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/assert"
)

func TestLoadDatabasesConfig(t *testing.T) {
	var data = `
- id: blackwells
  master_password: "ssm:blackwells.master_password"

- id: foyles
  enabled: false
  master_password: "ssm:foyles.master_password"
`
	pwdResolver := NewAwsPasswordResolver(mockGetParameter{})
	databases, err := LoadDatabasesConfig(strings.NewReader(data), pwdResolver)
	assert.NoError(t, err)
	assert.Len(t, databases, 2)
	assert.ElementsMatch(t, []*Database{
		{id: "blackwells", status: DbStatusEnabled, masterPasswordURL: "ssm:blackwells.master_password", masterPassword: "focused_mendel"},
		{id: "foyles", status: DbStatusDisabled, masterPasswordURL: "ssm:foyles.master_password"},
	}, databases)
}

type mockGetParameter struct {
	ssmiface.SSMAPI
}

func (m mockGetParameter) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	var value string
	switch *input.Name {
	case "blackwells.master_password":
		value = "focused_mendel"
	default:
		return nil, fmt.Errorf("unexpected parameter name: %s", *input.Name)
	}
	return &ssm.GetParameterOutput{Parameter: &ssm.Parameter{Value: aws.String(value)}}, nil
}
