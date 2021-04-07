package sari

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type PasswordResolver interface {
	Resolve(passwordSpec string) (string, error)
}

type AwsPasswordResolver struct {
	ssmiface.SSMAPI
}

var (
	ErrUnsupportedPasswordSpec = errors.New("unsupported password spec")
)

func NewAwsPasswordResolver(api ssmiface.SSMAPI) *AwsPasswordResolver {
	return &AwsPasswordResolver{SSMAPI: api}
}

func (r *AwsPasswordResolver) Resolve(passwordSpec string) (password string, err error) {
	if strings.HasPrefix(passwordSpec, "ssm:") {
		password, err = r.ssmGetEncryptedParameter(passwordSpec[4:])
		if err != nil {
			return "", err
		}
		return password, nil
	}
	return "", ErrUnsupportedPasswordSpec
}

func (r *AwsPasswordResolver) ssmGetEncryptedParameter(name string) (string, error) {
	input := ssm.GetParameterInput{Name: aws.String(name), WithDecryption: aws.Bool(true)}
	output, err := r.GetParameter(&input)
	if err != nil {
		return "", err
	}
	return *output.Parameter.Value, nil
}
