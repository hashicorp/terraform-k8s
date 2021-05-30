package workspacehelper

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	tfc "github.com/hashicorp/go-tfe"
)

const (
	AccessKeyID     = "AWS_ACCESS_KEY_ID"
	SecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	SessionToken    = "AWS_SESSION_TOKEN"
)

func getCredentials() []*tfc.Variable {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	retrieve, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	tfcVariables := []*tfc.Variable{}
	tfcVariables = append(tfcVariables, &tfc.Variable{
		Key:       AccessKeyID,
		Value:     retrieve.AccessKeyID,
		Sensitive: true,
		Category:  setVariableType(true),
		HCL:       false,
	})
	tfcVariables = append(tfcVariables, &tfc.Variable{
		Key:       SecretAccessKey,
		Value:     retrieve.SecretAccessKey,
		Sensitive: true,
		Category:  setVariableType(true),
		HCL:       false,
	})
	tfcVariables = append(tfcVariables, &tfc.Variable{
		Key:       SessionToken,
		Value:     retrieve.SessionToken,
		Sensitive: true,
		Category:  setVariableType(true),
		HCL:       false,
	})
	return tfcVariables

}

func filterAwsCredentials(vars []*tfc.Variable) []*tfc.Variable {
	result := []*tfc.Variable{}
	for _, v := range vars {
		if v.Key == AccessKeyID || v.Key == SecretAccessKey || v.Key == SessionToken {
			continue
		}
		result = append(result, v)
	}
	return result
}
