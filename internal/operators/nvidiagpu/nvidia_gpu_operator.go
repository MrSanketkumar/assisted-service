package nvidiagpu

import (
	"context"
	"text/template"

	"github.com/kelseyhightower/envconfig"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/internal/operators/api"
	"github.com/openshift/assisted-service/internal/operators/nodefeaturediscovery"
	"github.com/openshift/assisted-service/internal/templating"
	"github.com/openshift/assisted-service/models"
	"github.com/sirupsen/logrus"
)

var Operator = models.MonitoredOperator{
	Namespace:        "nvidia-gpu-operator",
	Name:             "nvidia-gpu",
	OperatorType:     models.OperatorTypeOlm,
	SubscriptionName: "gpu-operator-certified",
	TimeoutSeconds:   30 * 60,
}

// operator is an NVIDIA GPU OLM operator plugin.
type operator struct {
	log       logrus.FieldLogger
	config    *Config
	templates *template.Template
}

// NewNvidiaGPUOperator creates new NVIDIA GPU operator.
func NewNvidiaGPUOperator(log logrus.FieldLogger) *operator {
	config := &Config{}
	err := envconfig.Process(common.EnvConfigPrefix, config)
	if err != nil {
		log.Fatal(err.Error())
	}
	templates, err := templating.LoadTemplates(templatesRoot)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &operator{
		log:       log,
		config:    config,
		templates: templates,
	}
}

// GetName reports the name of an operator.
func (o *operator) GetName() string {
	return Operator.Name
}

func (o *operator) GetFullName() string {
	return "NVIDIA GPU"
}

// GetDependencies provides a list of dependencies of the Operator
func (o *operator) GetDependencies(c *common.Cluster) ([]string, error) {
	result := []string{
		nodefeaturediscovery.Operator.Name,
	}
	return result, nil
}

// GetClusterValidationID returns cluster validation ID for the operator.
func (o *operator) GetClusterValidationID() string {
	return string(models.ClusterValidationIDNvidiaGpuRequirementsSatisfied)
}

// GetHostValidationID returns host validation ID for the operator.
func (o *operator) GetHostValidationID() string {
	return string(models.HostValidationIDNvidiaGpuRequirementsSatisfied)
}

// ValidateCluster checks if the cluster satisfies the requirements to install the operator.
func (o *operator) ValidateCluster(ctx context.Context, cluster *common.Cluster) (result api.ValidationResult,
	err error) {
	result.ValidationId = o.GetClusterValidationID()
	result = api.ValidationResult{
		Status:       api.Success,
		ValidationId: o.GetClusterValidationID(),
	}
	return
}

// ValidateHost returns validationResult based on node type requirements such as memory and CPU.
func (o *operator) ValidateHost(ctx context.Context, cluster *common.Cluster, host *models.Host,
	hostRequirements *models.ClusterHostRequirementsDetails) (result api.ValidationResult, err error) {
	result = api.ValidationResult{
		Status:       api.Success,
		ValidationId: o.GetHostValidationID(),
	}
	return
}

// GetProperties provides description of operator properties.
func (o *operator) GetProperties() models.OperatorProperties {
	return models.OperatorProperties{}
}

// GetMonitoredOperator returns the information that describes how to monitor the operator.
func (o *operator) GetMonitoredOperator() *models.MonitoredOperator {
	return &Operator
}

// GetHostRequirements provides the requirements that the host needs to satisfy in order to be able to install the
// operator.
func (o *operator) GetHostRequirements(ctx context.Context, cluster *common.Cluster,
	host *models.Host) (*models.ClusterHostRequirementsDetails, error) {
	preflightRequirements, err := o.GetPreflightRequirements(ctx, cluster)
	if err != nil {
		o.log.WithError(err).Errorf("Cannot retrieve preflight requirements for cluster %s", cluster.ID)
		return nil, err
	}
	return preflightRequirements.Requirements.Worker.Quantitative, nil
}

// GetPreflightRequirements returns operator hardware requirements that can be determined with cluster data only.
func (o *operator) GetPreflightRequirements(context context.Context,
	cluster *common.Cluster) (result *models.OperatorHardwareRequirements, err error) {
	dependencies, err := o.GetDependencies(cluster)
	if err != nil {
		return
	}
	result = &models.OperatorHardwareRequirements{
		OperatorName: o.GetName(),
		Dependencies: dependencies,
		Requirements: &models.HostTypeHardwareRequirementsWrapper{
			Master: &models.HostTypeHardwareRequirements{
				Quantitative: &models.ClusterHostRequirementsDetails{},
			},
			Worker: &models.HostTypeHardwareRequirements{
				Quantitative: &models.ClusterHostRequirementsDetails{},
			},
		},
	}
	return
}

func (o *operator) GetFeatureSupportID() models.FeatureSupportLevelID {
	return models.FeatureSupportLevelIDNVIDIAGPU
}
