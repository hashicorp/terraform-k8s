package workspacehelper

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	tfc "github.com/hashicorp/go-tfe"
)

func getToken() []*tfc.Variable {
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
		Key:       "AWS_ACCESS_KEY_ID",
		Value:     retrieve.AccessKeyID,
		Sensitive: true,
		Category:  setVariableType(true),
		HCL:       false,
	})
	tfcVariables = append(tfcVariables, &tfc.Variable{
		Key:       "AWS_SECRET_ACCESS_KEY",
		Value:     retrieve.SecretAccessKey,
		Sensitive: true,
		Category:  setVariableType(true),
		HCL:       false,
	})
	tfcVariables = append(tfcVariables, &tfc.Variable{
		Key:       "AWS_SESSION_TOKEN",
		Value:     retrieve.SessionToken,
		Sensitive: true,
		Category:  setVariableType(true),
		HCL:       false,
	})
	return tfcVariables

}
