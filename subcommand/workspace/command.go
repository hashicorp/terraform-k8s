package workspace

import (
	"context"
	"fmt"
	"os"
	"sync"

	flag "github.com/spf13/pflag"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis"
	"github.com/hashicorp/terraform-k8s/operator/pkg/controller"
	"github.com/mitchellh/cli"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	appv1alpha1 "github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
	log                       = logf.Log.WithName("operator")
)

// Command is the command for syncing the K8S and Terraform
// Cloud workspaces.
type Command struct {
	UI cli.Ui

	flags                 *flag.FlagSet
	flagLogLevel          string
	flagK8sWatchNamespace string

	tfcClient *tfc.Client
	clientset kubernetes.Interface

	once  sync.Once
	sigCh chan os.Signal
	help  string
}

func (c *Command) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.flagK8sWatchNamespace, "k8s-watch-namespace", metav1.NamespaceAll,
		"The Kubernetes namespace to watch for service changes and sync to Terraform Cloud. "+
			"If this is not set then it will default to all namespaces.")

	zapFlags := zap.FlagSet()
	c.help = fmt.Sprintf("%s\n%s\n%s", help, c.flags.FlagUsages(), zapFlags.FlagUsages())

	flag.CommandLine.AddFlagSet(zapFlags)
	flag.CommandLine.AddFlagSet(c.flags)
	flag.Parse()

	logf.SetLogger(zap.Logger())
}

// Run starts the operator to synchronize workspaces.
func (c *Command) Run(args []string) int {
	c.init()

	namespace := c.flagK8sWatchNamespace

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		return 1
	}

	ctx := context.TODO()
	// Become the leader before proceeding
	err = leader.Become(ctx, "workspace-lock")
	if err != nil {
		log.Error(err, "")
		return 1
	}

	// Create CRDs
	if err := createCustomResourceDefinitions(cfg); err != nil {
		log.Error(err, "Could not create CRDs")
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		return 1
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		return 1
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
		return 1
	}

	if err = serveCRMetrics(cfg); err != nil {
		log.Info("Error generating and serving metrics", "error", err.Error())
	}

	// Add to the below struct any other metrics ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}
	// Create Service object to expose the metrics port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Error(err, "Could not create metrics Service")
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metrics from this operator.
	services := []*v1.Service{service}
	_, err = metrics.CreateServiceMonitors(cfg, namespace, services)
	if err != nil {
		log.Info("Could not create ServiceMonitor object", "error", err.Error())
		// If this operator is deployed to a cluster without the prometheus-operator running, it will return
		// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
		if err == metrics.ErrServiceMonitorNotPresent {
			log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects", "error", err.Error())
		}
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		return 1
	}

	return 0
}

func (c *Command) Synopsis() string { return synopsis }
func (c *Command) Help() string {
	c.once.Do(c.init)
	return c.help
}

const synopsis = "Sync Workspace and Terraform Cloud."
const help = `
Usage: terraform-k8s sync-workspace [options]

  Sync K8s TFC Workspace resource with Terraform Cloud.
	This enables Workspaces in Kubernetes to manage infrastructure resources
	created by Terraform Cloud.
`

// serveCRMetrics gets the Operator/CustomResource GVKs and generates metrics based on those types.
// It serves those metrics on "http://metricsHost:operatorMetricsPort".
func serveCRMetrics(cfg *rest.Config) error {
	// Below function returns filtered operator/CustomResource specific GVKs.
	// For more control override the below GVK list with your own custom logic.
	filteredGVK, err := k8sutil.GetGVKsFromAddToScheme(apis.AddToScheme)
	if err != nil {
		return err
	}
	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	// To generate metrics in other namespaces, add the values below.
	ns := []string{operatorNs}
	// Generate and serve custom resource specific metrics.
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, ns, filteredGVK, metricsHost, operatorMetricsPort)
	if err != nil {
		return err
	}
	return nil
}

func createCustomResourceDefinitions(cfg *rest.Config) error {
	apiextensionsClientSet, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return err
	}

	if _, err := appv1alpha1.CreateCustomResourceDefinition(apiextensionsClientSet); err != nil {
		return err
	}

	return nil
}
