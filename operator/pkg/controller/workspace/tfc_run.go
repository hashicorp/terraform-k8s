package workspace

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	getter "github.com/hashicorp/go-getter"
	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
)

var (
	autoQueueRuns           = false
	speculative             = false
	isDestroy               = true
	basepath                = "/tmp"
	moduleDirectory         = fmt.Sprintf("%s/%s", basepath, "module")
	pluginDirectory         = fmt.Sprintf("%s/%s", moduleDirectory, ".terraform/plugins/linux_amd64")
	configurationFilePath   = fmt.Sprintf("%s/%s", moduleDirectory, "main.tf")
	terraformIgnoreFilePath = fmt.Sprintf("%s/%s", moduleDirectory, ".terraformignore")
	interval                = 30 * time.Second
)

// UploadConfigurationFile uploads the main.tf to a configuration version
func (t *TerraformCloudClient) UploadConfigurationFile(uploadURL string) error {
	if err := t.Client.ConfigurationVersions.Upload(context.TODO(), uploadURL, moduleDirectory); err != nil {
		return fmt.Errorf("error, %v, %v", err, moduleDirectory)
	}
	return nil
}

// CreateConfigurationVersion creates a configuration version for a workspace
func (t *TerraformCloudClient) CreateConfigurationVersion(workspaceID string) (*tfc.ConfigurationVersion, error) {
	options := tfc.ConfigurationVersionCreateOptions{
		AutoQueueRuns: &autoQueueRuns,
		Speculative:   &speculative,
	}
	configVersion, err := t.Client.ConfigurationVersions.Create(context.TODO(), workspaceID, options)
	if err != nil {
		return nil, err
	}
	return configVersion, nil
}

// DownloadProviders download custom provider binaries.
func (t *TerraformCloudClient) DownloadProviders(providers []*v1alpha1.Provider) error {
	wg := sync.WaitGroup{}
	wg.Add(len(providers))
	errChan := make(chan error)
	for _, elem := range providers {
		os.MkdirAll(pluginDirectory, 0777)

		client := &getter.Client{
			Ctx:     context.TODO(),
			Src:     elem.Source,
			Dst:     pluginDirectory,
			Mode:    getter.ClientModeAny,
			Options: nil,
		}
		go func() {
			defer wg.Done()
			if err := client.Get(); err != nil {
				log.Error(err, "Download Failed.")
				errChan <- err
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// CreateRun runs a new Terraform Cloud configuration
func (t *TerraformCloudClient) CreateRun(workspace *v1alpha1.Workspace, terraform []byte) error {
	configVersion, err := t.CreateConfigurationVersion(workspace.Status.WorkspaceID)
	if err != nil {
		return err
	}

	os.Mkdir(moduleDirectory, 0777)

	if err := t.DownloadProviders(workspace.Spec.AdditionalProviders); err != nil {
		return err
	}

	if err := ioutil.WriteFile(configurationFilePath, terraform, 0777); err != nil {
		return err
	}

	// By default plugins are ignored by the tfc slug.  We write an explicit .terraformignore to override this.
	if err := ioutil.WriteFile(terraformIgnoreFilePath, []byte("!.terraform/plugins/linux_amd64/"), 0644); err != nil {
		return err
	}

	if err := t.UploadConfigurationFile(configVersion.UploadURL); err != nil {
		return err
	}

	message := fmt.Sprintf("%s, apply", TerraformOperator)
	options := tfc.RunCreateOptions{
		Message:              &message,
		ConfigurationVersion: configVersion,
		Workspace: &tfc.Workspace{
			ID: workspace.Status.WorkspaceID,
		},
	}
	run, err := t.Client.Runs.Create(context.TODO(), options)
	if err != nil {
		return err
	}
	workspace.Status.RunID = run.ID
	workspace.Status.RunStatus = string(run.Status)
	return nil
}

// CheckRun gets the run status
func (t *TerraformCloudClient) CheckRun(runID string) (string, error) {
	if runID == "" {
		return "", nil
	}
	run, err := t.Client.Runs.Read(context.TODO(), runID)
	if err != nil {
		return "", err
	}
	return string(run.Status), nil
}

func isPending(status string) bool {
	state := tfc.RunStatus(status)
	switch state {
	case tfc.RunApplied:
		return false
	case tfc.RunPlannedAndFinished:
		return false
	case tfc.RunErrored:
		return false
	case tfc.RunCanceled:
		return false
	case tfc.RunDiscarded:
		return false
	case "":
		return false
	default:
		return true
	}
}

func isError(status string) bool {
	return tfc.RunStatus(status) == tfc.RunErrored
}

// DeleteRuns cancels runs that haven't been applied or planned
func (t *TerraformCloudClient) DeleteRuns(workspaceID string) error {
	message := "operator, finalizer, cancelling run"
	runs, err := t.Client.Runs.List(context.TODO(), workspaceID, tfc.RunListOptions{})
	if err != nil {
		return err
	}
	for _, run := range runs.Items {
		if isPending(string(run.Status)) {
			err := t.Client.Runs.ForceCancel(context.TODO(), run.ID, tfc.RunForceCancelOptions{
				Comment: &message,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteResources destroys the resources in a workspace
func (t *TerraformCloudClient) DeleteResources(workspaceID string) error {
	ws, err := t.Client.Workspaces.ReadByID(context.TODO(), workspaceID)
	if err != nil {
		return err
	}
	if ws.CurrentRun == nil {
		return nil
	}
	message := fmt.Sprintf("%s, destroy", TerraformOperator)
	options := tfc.RunCreateOptions{
		IsDestroy: &isDestroy,
		Message:   &message,
		Workspace: ws,
	}
	run, err := t.Client.Runs.Create(context.TODO(), options)
	if err != nil {
		return err
	}
	for {
		checkRun, err := t.Client.Runs.Read(context.TODO(), run.ID)
		if err != nil || checkRun.Status == tfc.RunErrored {
			return fmt.Errorf("destroy had error: %v", err)
		}
		if !isPending(string(checkRun.Status)) {
			return nil
		}
		time.Sleep(interval)
	}
}
