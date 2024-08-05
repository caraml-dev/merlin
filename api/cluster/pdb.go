package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	policyv1 "k8s.io/api/policy/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/models"
)

const (
	componentLabel      = "component"
	modelVersionIDLabel = "model-version-id"
	kserveIsvcLabel     = "serving.kserve.io/inferenceservice"
)

type PodDisruptionBudget struct {
	Name                     string
	Namespace                string
	Labels                   map[string]string
	Selectors                map[string]string
	MaxUnavailablePercentage *int
	MinAvailablePercentage   *int
}

func NewPodDisruptionBudget(modelService *models.Service, componentType string, pdbConfig config.PodDisruptionBudgetConfig) *PodDisruptionBudget {
	labels := buildPDBSelector(modelService, componentType)
	labels[modelVersionIDLabel] = modelService.ModelVersion

	return &PodDisruptionBudget{
		Name:                     buildPDBName(modelService, componentType),
		Namespace:                modelService.Namespace,
		Labels:                   labels,
		Selectors:                buildPDBSelector(modelService, componentType),
		MaxUnavailablePercentage: pdbConfig.MaxUnavailablePercentage,
		MinAvailablePercentage:   pdbConfig.MinAvailablePercentage,
	}
}

func (cfg PodDisruptionBudget) BuildPDBSpec() (*policyv1.PodDisruptionBudget, error) {
	if cfg.MaxUnavailablePercentage == nil && cfg.MinAvailablePercentage == nil {
		return nil, fmt.Errorf("one of maxUnavailable and minAvailable must be specified")
	}

	pdb := &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1",
			Kind:       "PodDisruptionBudget",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cfg.Selectors,
			},
		},
	}

	// Since we can specify only one of maxUnavailable and minAvailable, minAvailable takes precedence
	// https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
	if cfg.MinAvailablePercentage != nil {
		minAvailable := intstr.FromString(fmt.Sprintf("%d%%", *cfg.MinAvailablePercentage))
		pdb.Spec.MinAvailable = &minAvailable
	} else if cfg.MaxUnavailablePercentage != nil {
		maxUnavailable := intstr.FromString(fmt.Sprintf("%d%%", *cfg.MaxUnavailablePercentage))
		pdb.Spec.MaxUnavailable = &maxUnavailable
	}

	return pdb, nil
}

func buildPDBName(modelService *models.Service, componentType string) string {
	return fmt.Sprintf("%s-%s-%s", modelService.Name, componentType, models.PDBComponentType)
}

func buildPDBSelector(modelService *models.Service, componentType string) map[string]string {
	labelSelectors := modelService.Metadata.ToLabel()
	labelSelectors[componentLabel] = componentType
	labelSelectors[kserveIsvcLabel] = modelService.Name

	return labelSelectors
}

func generatePDBSpecs(modelService *models.Service, pdbConfig config.PodDisruptionBudgetConfig) []*PodDisruptionBudget {
	pdbs := []*PodDisruptionBudget{}

	// PDB config is not set properly, we can't apply it.
	if pdbConfig.MinAvailablePercentage == nil && pdbConfig.MaxUnavailablePercentage == nil {
		return pdbs
	}

	if (modelService.ResourceRequest != nil) &&
		doesPDBAllowDisruption(pdbConfig, modelService.ResourceRequest.MinReplica) {
		predictorPdb := NewPodDisruptionBudget(modelService, models.PredictorComponentType, pdbConfig)
		pdbs = append(pdbs, predictorPdb)
	}

	if (modelService.Transformer != nil && modelService.Transformer.Enabled && modelService.Transformer.ResourceRequest != nil) &&
		doesPDBAllowDisruption(pdbConfig, modelService.Transformer.ResourceRequest.MinReplica) {
		transformerPdb := NewPodDisruptionBudget(modelService, models.TransformerComponentType, pdbConfig)
		pdbs = append(pdbs, transformerPdb)
	}

	return pdbs
}

// doesPDBAllowDisruption determine if pdb config & minReplica will allow
// any disruption, this to avoid the replicas may be unable to be removed. rule:
// ceil(minReplica * minAvailablePercent) < minReplica || maxUnavailablePercent > 0.
// Note that we only care about minReplica for minAvailablePercent because, for
// active replicas > minReplica, the condition will be satisfied if it was
// satisfied for the minReplica case.
func doesPDBAllowDisruption(pdbConfig config.PodDisruptionBudgetConfig, minReplica int) bool {
	if pdbConfig.MinAvailablePercentage != nil &&
		math.Ceil(float64(minReplica)*(float64(*pdbConfig.MinAvailablePercentage)/100.0)) < float64(minReplica) {
		return true
	}

	if pdbConfig.MaxUnavailablePercentage != nil && *pdbConfig.MaxUnavailablePercentage > 0 {
		return true
	}

	return false
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

	pdbJSON, err := json.Marshal(pdbSpec)
	if err != nil {
		return err
	}

	forceEnabled := true

	_, err = c.policyClient.PodDisruptionBudgets(pdb.Namespace).
		Patch(ctx, pdb.Name, types.ApplyPatchType, pdbJSON, metav1.PatchOptions{FieldManager: "application/apply-patch", Force: &forceEnabled})
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

func (c *controller) getUnusedPodDisruptionBudgets(ctx context.Context, modelService *models.Service) ([]*PodDisruptionBudget, error) {
	pdbs := []*PodDisruptionBudget{}

	labels := modelService.Metadata.ToLabel()

	labelSelectors := []string{}
	for key, val := range labels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", key, val))
	}
	labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", modelVersionIDLabel, modelService.ModelVersion))
	labelSelectors = append(labelSelectors, fmt.Sprintf("%s!=%s", kserveIsvcLabel, modelService.Name))

	list, err := c.policyClient.PodDisruptionBudgets(modelService.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: strings.Join(labelSelectors, ","),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		pdbs = append(pdbs, &PodDisruptionBudget{
			Name:      item.Name,
			Namespace: modelService.Namespace,
		})
	}

	return pdbs, nil
}
