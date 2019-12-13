package workspace

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hashicorp/consul/command/flags"
	tfc "github.com/hashicorp/go-tfe"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/mitchellh/cli"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Command is the command for syncing the K8S and Consul service
// catalogs (one or both directions).
type Command struct {
	UI cli.Ui

	flags        *flag.FlagSet
	flagLogLevel string

	tfcClient *tfc.Client
	clientset kubernetes.Interface

	once  sync.Once
	sigCh chan os.Signal
	help  string
}

func (c *Command) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.flagLogLevel, "log-level", "info",
		"Log verbosity level. Supported values (in order of detail) are \"trace\", "+
			"\"debug\", \"info\", \"warn\", and \"error\".")

	c.help = flags.Usage(help, c.flags)
}

func (c *Command) Run(args []string) int {
	c.once.Do(c.init)
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if len(c.flags.Args()) > 0 {
		c.UI.Error(fmt.Sprintf("Should have no non-flag arguments."))
		return 1
	}

	// Setup TerraformCloud client
	if c.tfcClient == nil {
		var err error
		tfConfig, diag := cliconfig.LoadConfig()
		if diag.Err() != nil {
			c.UI.Error(fmt.Sprintf("Error finding Terraform Cloud configuration: %s", diag.Err()))
			return 1
		}

		config := &tfc.Config{
			Token: fmt.Sprintf("%v", tfConfig.Credentials["app.terraform.io"]["token"]),
		}

		c.tfcClient, err = tfc.NewClient(config)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error connecting to Terraform Cloud: %s", err))
			return 1
		}
	}

	level := hclog.LevelFromString(c.flagLogLevel)
	if level == hclog.NoLevel {
		c.UI.Error(fmt.Sprintf("Unknown log level: %s", c.flagLogLevel))
		return 1
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Level:  level,
		Output: os.Stderr,
	})

	// Get the sync interval
	var syncInterval time.Duration

	// Create the context we'll use to cancel everything
	ctx, cancelF := context.WithCancel(context.Background())

	// Start the K8S-to-TFC syncer
	var toTFCCh chan struct{}

	// Build the TFC workspace sync and start it
	syncer := &catalogtoconsul.ConsulSyncer{
		Client:            c.consulClient,
		Log:               logger.Named("to-consul/sink"),
		Namespace:         c.flagK8SSourceNamespace,
		SyncPeriod:        syncInterval,
		ServicePollPeriod: syncInterval * 2,
		ConsulK8STag:      c.flagConsulK8STag,
	}
	go syncer.Run(ctx)

	// Build the controller and start it
	ctl := &controller.Controller{
		Log: logger.Named("to-consul/controller"),
		Resource: &catalogtoconsul.ServiceResource{
			Log:                   logger.Named("to-consul/source"),
			Client:                c.clientset,
			Syncer:                syncer,
			Namespace:             c.flagK8SSourceNamespace,
			ExplicitEnable:        !c.flagK8SDefault,
			ClusterIPSync:         c.flagSyncClusterIPServices,
			NodePortSync:          catalogtoconsul.NodePortSyncType(c.flagNodePortSyncType),
			ConsulK8STag:          c.flagConsulK8STag,
			ConsulServicePrefix:   c.flagConsulServicePrefix,
			AddK8SNamespaceSuffix: c.flagAddK8SNamespaceSuffix,
		},
	}

	toConsulCh = make(chan struct{})
	go func() {
		defer close(toConsulCh)
		ctl.Run(ctx.Done())
	}()

	// Start healthcheck handler
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health/ready", c.handleReady)
		var handler http.Handler = mux

		c.UI.Info(fmt.Sprintf("Listening on %q...", c.flagListen))
		if err := http.ListenAndServe(c.flagListen, handler); err != nil {
			c.UI.Error(fmt.Sprintf("Error listening: %s", err))
		}
	}()

	// Wait on an interrupt to exit
	c.sigCh = make(chan os.Signal, 1)
	signal.Notify(c.sigCh, os.Interrupt)
	select {
	// Unexpected exit
	case <-toConsulCh:
		cancelF()
		if toK8SCh != nil {
			<-toK8SCh
		}
		return 1

	// Unexpected exit
	case <-toK8SCh:
		cancelF()
		if toConsulCh != nil {
			<-toConsulCh
		}
		return 1

	// Interrupted, gracefully exit
	case <-c.sigCh:
		cancelF()
		if toConsulCh != nil {
			<-toConsulCh
		}
		if toK8SCh != nil {
			<-toK8SCh
		}
		return 0
	}
}

func (c *Command) handleReady(rw http.ResponseWriter, req *http.Request) {
	// The main readiness check is whether sync can talk to
	// the consul cluster, in this case querying for the leader
	_, err := c.consulClient.Status().Leader()
	if err != nil {
		c.UI.Error(fmt.Sprintf("[GET /health/ready] Error getting leader status: %s", err))
		rw.WriteHeader(500)
		return
	}
	rw.WriteHeader(204)
}

func (c *Command) Synopsis() string { return synopsis }
func (c *Command) Help() string {
	c.once.Do(c.init)
	return c.help
}

// interrupt sends os.Interrupt signal to the command
// so it can exit gracefully. This function is needed for tests
func (c *Command) interrupt() {
	c.sigCh <- os.Interrupt
}

const synopsis = "Sync Workspace and Terraform Cloud."
const help = `
Usage: terraform-k8s sync-workspace [options]

  Sync K8s TFC Workspace resource with Terraform Cloud.
	This enables Workspaces in Kubernetes to manage infrastructure resources
	created by Terraform Cloud.
`
