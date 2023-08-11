package cluster

import (
	"context"
	"fmt"
	"math"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	metav1cfg "k8s.io/client-go/applyconfigurations/meta/v1"
	policyv1cfg "k8s.io/client-go/applyconfigurations/policy/v1"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/models"
)

type PodDisruptionBudget struct {
	Name                     string
	Namespace                string
	Labels                   map[string]string
	MaxUnavailablePercentage *int
	MinAvailablePercentage   *int
}

func NewPodDisruptionBudget(modelService *models.Service, componentType string, pdbConfig config.PodDisruptionBudgetConfig) *PodDisruptionBudget {
	return &PodDisruptionBudget{
		Name:                     fmt.Sprintf("%s-%s-%s", modelService.ModelName, componentType, models.PDBComponentType),
		Namespace:                modelService.Namespace,
		Labels:                   modelService.Metadata.ToLabel(),
		MaxUnavailablePercentage: pdbConfig.MaxUnavailablePercentage,
		MinAvailablePercentage:   pdbConfig.MinAvailablePercentage,
	}
}

func (cfg PodDisruptionBudget) BuildPDBSpec() (*policyv1cfg.PodDisruptionBudgetSpecApplyConfiguration, error) {
	if cfg.MaxUnavailablePercentage == nil && cfg.MinAvailablePercentage == nil {
		return nil, fmt.Errorf("one of maxUnavailable and minAvailable must be specified")
	}

	pdbSpec := &policyv1cfg.PodDisruptionBudgetSpecApplyConfiguration{
		Selector: &metav1cfg.LabelSelectorApplyConfiguration{
			MatchLabels: cfg.Labels,
		},
	}

	// Since we can specify only one of maxUnavailable and minAvailable, minAvailable takes precedence
	// https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
	if cfg.MinAvailablePercentage != nil {
		minAvailable := intstr.FromString(fmt.Sprintf("%d%%", *cfg.MinAvailablePercentage))
		pdbSpec.MinAvailable = &minAvailable
	} else if cfg.MaxUnavailablePercentage != nil {
		maxUnavailable := intstr.FromString(fmt.Sprintf("%d%%", *cfg.MaxUnavailablePercentage))
		pdbSpec.MaxUnavailable = &maxUnavailable
	}

	return pdbSpec, nil
}

func createPodDisruptionBudgets(modelService *models.Service, pdbConfig config.PodDisruptionBudgetConfig) []*PodDisruptionBudget {
	pdbs := []*PodDisruptionBudget{}

	// Only create PDB if: ceil(minReplica * minAvailablePercent) < minReplica
	// If not, the replicas may be unable to be removed.
	// Note that we only care about minReplica here because, for active replicas > minReplica,
	// the condition will be satisfied if it was satisfied for the minReplica case.

	var minAvailablePercent float64
	if pdbConfig.MinAvailablePercentage != nil {
		minAvailablePercent = float64(*pdbConfig.MinAvailablePercentage) / 100.0
	} else if pdbConfig.MaxUnavailablePercentage != nil {
		minAvailablePercent = float64(100-*pdbConfig.MaxUnavailablePercentage) / 100.0
	} else {
		// PDB config is not set properly, we can't apply it.
		return pdbs
	}

	if modelService.ResourceRequest != nil &&
		math.Ceil(float64(modelService.ResourceRequest.MinReplica)*
			minAvailablePercent) < float64(modelService.ResourceRequest.MinReplica) {
		predictorPdb := NewPodDisruptionBudget(modelService, models.ModelComponentType, pdbConfig)
		pdbs = append(pdbs, predictorPdb)
	}

	if modelService.Transformer != nil &&
		modelService.Transformer.Enabled &&
		modelService.Transformer.ResourceRequest != nil &&
		math.Ceil(float64(modelService.Transformer.ResourceRequest.MinReplica)*
			minAvailablePercent) < float64(modelService.Transformer.ResourceRequest.MinReplica) {
		transformerPdb := NewPodDisruptionBudget(modelService, models.TransformerComponentType, pdbConfig)
		pdbs = append(pdbs, transformerPdb)
	}

	return pdbs
}

func (c *controller) deployPodDisruptionBudgets(ctx context.Context, pdbs []*PodDisruptionBudget) error {
	for _, pdb := range pdbs {
		if err := c.deployPodDisruptionBudget(ctx, pdb); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) deployPodDisruptionBudget(ctx context.Context, pdb *PodDisruptionBudget) error {
	pdbSpec, err := pdb.BuildPDBSpec()
	if err != nil {
		return err
	}

	pdbCfg := policyv1cfg.PodDisruptionBudget(pdb.Name, pdb.Namespace)
	pdbCfg.WithLabels(pdb.Labels)
	pdbCfg.WithSpec(pdbSpec)

	_, err = c.policyClient.PodDisruptionBudgets(pdb.Namespace).Apply(ctx, pdbCfg, metav1.ApplyOptions{FieldManager: "application/apply-patch"})
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) deletePodDisruptionBudgets(ctx context.Context, pdbs []*PodDisruptionBudget) error {
	for _, pdb := range pdbs {
		if err := c.deletePodDisruptionBudget(ctx, pdb); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) deletePodDisruptionBudget(ctx context.Context, pdb *PodDisruptionBudget) error {
	err := c.policyClient.PodDisruptionBudgets(pdb.Namespace).Delete(ctx, pdb.Name, metav1.DeleteOptions{})
	if !kerrors.IsNotFound(err) {
		return err
	}
	return nil
}
