package deployment

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	appLabelKey = "app"
)

type Manifest struct {
	Deployment *appsv1.Deployment
	Secret     *corev1.Secret
	OnProgress bool
}

type Deployer interface {
	Deploy(ctx context.Context, data *models.WorkerData) error
	GetDeployedManifest(ctx context.Context, data *models.WorkerData) (*Manifest, error)
	Undeploy(ctx context.Context, data *models.WorkerData) error
}

type deployer struct {
	kubeClient     kubernetes.Interface
	consumerConfig config.ObservabilityPublisher

	resourceRequest corev1.ResourceList
	resourceLimit   corev1.ResourceList
}

func New(restConfig *rest.Config, consumerConfig config.ObservabilityPublisher) (*deployer, error) {
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	resourceRequest, err := parseResourceList(consumerConfig.DefaultResources.Requests)
	if err != nil {
		return nil, err
	}
	resourceLimit, err := parseResourceList(consumerConfig.DefaultResources.Limits)
	if err != nil {
		return nil, err
	}

	return &deployer{
		kubeClient:      kubeClient,
		consumerConfig:  consumerConfig,
		resourceRequest: resourceRequest,
		resourceLimit:   resourceLimit,
	}, nil
}

func parseResourceList(resourceCfg config.Resource) (corev1.ResourceList, error) {
	resourceList := corev1.ResourceList{}
	if resourceCfg.CPU != "" {
		quantity, err := resource.ParseQuantity(resourceCfg.CPU)
		if err != nil {
			return corev1.ResourceList{}, err
		}
		resourceList[corev1.ResourceCPU] = quantity
	}

	if resourceCfg.Memory != "" {
		quantity, err := resource.ParseQuantity(resourceCfg.Memory)
		if err != nil {
			return corev1.ResourceList{}, err
		}
		resourceList[corev1.ResourceMemory] = quantity
	}

	return resourceList, nil
}

func (c *deployer) targetNamespace() string {
	return c.consumerConfig.TargetNamespace
}

func (c *deployer) GetDeployedManifest(ctx context.Context, data *models.WorkerData) (*Manifest, error) {
	secretName := c.getSecretName(data)
	secret, err := c.getSecret(ctx, secretName, c.targetNamespace())
	if err != nil {
		return nil, err
	}
	deploymentName := c.getDeploymentName(data)
	depl, err := c.getDeployment(ctx, deploymentName, c.targetNamespace())
	if err != nil {
		return nil, err
	}

	if depl == nil {
		if secret != nil {
			log.Warnf("deployment %s is not exist but secret name exist %s", deploymentName, secretName)
		}
		return nil, nil
	}

	if secret == nil {
		log.Warnf("deployment %s exist without associated secret", deploymentName)
		return nil, nil
	}

	isDeploymentRolledOut, err := deploymentRolledOut(depl, data.Revision, false)
	if err != nil {
		return nil, err
	}
	return &Manifest{Secret: secret, Deployment: depl, OnProgress: !isDeploymentRolledOut}, nil
}

func deploymentRolledOut(depl *appsv1.Deployment, revision int, strictCheck bool) (bool, error) {
	deploymentRev, err := getDeploymentRevision(depl)
	if err != nil {
		return false, err
	}

	if strictCheck && deploymentRev != int64(revision) {
		return false, fmt.Errorf("revision is not matched, requested: %d - actual: %d", revision, deploymentRev)
	}

	if depl.Generation <= depl.Status.ObservedGeneration {
		cond := getDeploymentCondition(depl.Status, appsv1.DeploymentProgressing)
		if cond != nil && cond.Reason == timeoutReason {
			return false, fmt.Errorf("deployment %q exceeded its progress deadline", depl.Name)
		}
		if depl.Spec.Replicas != nil && depl.Status.UpdatedReplicas < *depl.Spec.Replicas {
			return false, nil
		}
		if depl.Status.Replicas > depl.Status.UpdatedReplicas {
			return false, nil
		}
		if depl.Status.AvailableReplicas < depl.Status.UpdatedReplicas {
			return false, nil
		}
		return true, nil
	}

	return false, nil
}

func (c *deployer) Deploy(ctx context.Context, data *models.WorkerData) (err error) {
	secret, previousSecret, err := c.applySecret(ctx, data)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			// meaning that we need to rollback to previous secret
			if previousSecret != nil {
				if _, err := c.rollbackSecret(ctx, previousSecret); err != nil {
					log.Warnf("failed rollback secret to previous with err: %v", err)
				}
			} else {
				// delete current secret
				if err := c.deleteSecret(ctx, secret.Name, secret.Namespace); err != nil {
					log.Warnf("failed delete secret with err: %v", err)
				}
				// delete current deployment
				if err := c.deleteDeployment(ctx, c.getDeploymentName(data), c.targetNamespace()); err != nil {
					log.Warnf("failed delete deployment with err: %v", err)
				}
			}
		}
	}()
	deployment, err := c.applyDeployment(ctx, data, secret.Name)
	if err != nil {
		return err
	}

	if err := c.waitUntilDeploymentReady(ctx, deployment, data.Revision); err != nil {
		return err
	}

	return nil
}

func (c *deployer) rollbackSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	coreV1 := c.kubeClient.CoreV1()
	secretV1 := coreV1.Secrets(secret.Namespace)
	return secretV1.Update(ctx, secret, metav1.UpdateOptions{})
}

func (c *deployer) getSecret(ctx context.Context, secretName string, namespace string) (*corev1.Secret, error) {
	coreV1 := c.kubeClient.CoreV1()
	secretV1 := coreV1.Secrets(namespace)
	secret, err := secretV1.Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		return nil, nil
	}
	return secret, nil
}

func (c *deployer) deleteSecret(ctx context.Context, secretName string, namespace string) error {
	coreV1 := c.kubeClient.CoreV1()
	secretV1 := coreV1.Secrets(namespace)
	return secretV1.Delete(ctx, secretName, metav1.DeleteOptions{})
}

func (c *deployer) getDeployment(ctx context.Context, deploymentName string, namespace string) (*appsv1.Deployment, error) {
	appV1 := c.kubeClient.AppsV1()
	deploymentV1 := appV1.Deployments(namespace)
	deployment, err := deploymentV1.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		return nil, nil
	}
	return deployment, nil
}

func (c *deployer) deleteDeployment(ctx context.Context, deploymentName string, namespace string) error {
	appV1 := c.kubeClient.AppsV1()
	deploymentV1 := appV1.Deployments(namespace)
	return deploymentV1.Delete(ctx, deploymentName, metav1.DeleteOptions{})
}

func (c *deployer) applySecret(ctx context.Context, data *models.WorkerData) (secret *corev1.Secret, previousSecret *corev1.Secret, err error) {
	// Create secret
	coreV1 := c.kubeClient.CoreV1()
	secretV1 := coreV1.Secrets(c.targetNamespace())
	secretName := c.getSecretName(data)
	applySecretFunc := func(data *models.WorkerData, isExistingSecret bool) (*corev1.Secret, error) {
		secretSpec, err := c.createSecretSpec(data)
		if err != nil {
			return nil, err
		}
		if isExistingSecret {
			return secretV1.Update(ctx, secretSpec, metav1.UpdateOptions{})
		}
		return secretV1.Create(ctx, secretSpec, metav1.CreateOptions{})
	}
	previousSecret, err = secretV1.Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, nil, err
		}
		secret, err = applySecretFunc(data, false)
		return secret, nil, err
	}
	secret, err = applySecretFunc(data, true)
	return secret, previousSecret, err
}

func (c *deployer) createSecretSpec(data *models.WorkerData) (*corev1.Secret, error) {
	observationSinks := []ObservationSink{}
	if c.consumerConfig.BigQuerySink.Enabled {
		observationSinks = append(observationSinks, ObservationSink{
			Type: BQ,
			Config: BigQuerySink{
				Project: c.consumerConfig.BigQuerySink.Project,
				Dataset: c.consumerConfig.BigQuerySink.Dataset,
				TTLDays: c.consumerConfig.BigQuerySink.TTLDays,
			},
		})
	}
	if c.consumerConfig.MaxComputeSink.Enabled {
		observationSinks = append(observationSinks, ObservationSink{
			Type: MaxCompute,
			Config: MaxComputeSink{
				Project:         c.consumerConfig.MaxComputeSink.Project,
				Dataset:         c.consumerConfig.MaxComputeSink.Dataset,
				TTLDays:         c.consumerConfig.MaxComputeSink.TTLDays,
				AccessKeyID:     c.consumerConfig.MaxComputeSink.AccessKeyID,
				AccessKeySecret: c.consumerConfig.MaxComputeSink.AccessKeySecret,
				AccessUrl:       c.consumerConfig.MaxComputeSink.AccessUrl,
			},
		})
	}

	if c.consumerConfig.ArizeSink.IsEnabled(data.GetModelSerial()) {
		observationSinks = append(observationSinks, ObservationSink{
			Type: Arize,
			Config: ArizeSink{
				APIKey:   c.consumerConfig.ArizeSink.APIKey,
				SpaceKey: c.consumerConfig.ArizeSink.SpaceKey,
			},
		})
	}

	consumerCfg := &ConsumerConfig{
		Project:          data.Project,
		ModelID:          data.ModelName,
		ModelVersion:     data.ModelVersion,
		InferenceSchema:  data.ModelSchemaSpec,
		ObservationSinks: observationSinks,
		ObservationSource: &ObserVationSource{
			Type: Kafka,
			Config: &KafkaSource{
				Topic:                    fmt.Sprintf("%s%s", c.consumerConfig.KafkaConsumer.TopicPrefix, data.TopicSource),
				BootstrapServers:         c.consumerConfig.KafkaConsumer.Brokers,
				GroupID:                  c.consumerConfig.KafkaConsumer.GroupID,
				BatchSize:                c.consumerConfig.KafkaConsumer.BatchSize,
				AdditionalConsumerConfig: c.consumerConfig.KafkaConsumer.AdditionalConsumerConfig,
			},
		},
	}
	consumerCfgStr, err := yaml.Marshal(consumerCfg)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.getSecretName(data),
			Namespace: c.targetNamespace(),
			Labels:    c.getLabels(data),
		},
		StringData: map[string]string{
			"config.yaml": string(consumerCfgStr),
		},
	}, nil
}

func (c *deployer) applyDeployment(ctx context.Context, data *models.WorkerData, secretName string) (*appsv1.Deployment, error) {
	appV1 := c.kubeClient.AppsV1()
	deploymentName := c.getDeploymentName(data)
	deploymentV1 := appV1.Deployments(c.targetNamespace())

	applyDeploymentFunc := func(data *models.WorkerData, secretName string, isExistingDeployment bool) (*appsv1.Deployment, error) {
		deployment, err := c.createDeploymentSpec(data, secretName)
		if err != nil {
			return nil, err
		}
		if isExistingDeployment {
			return deploymentV1.Update(ctx, deployment, metav1.UpdateOptions{})
		}
		return deploymentV1.Create(ctx, deployment, metav1.CreateOptions{})
	}
	_, err := deploymentV1.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		return applyDeploymentFunc(data, secretName, false)
	}

	return applyDeploymentFunc(data, secretName, true)
}

func (c *deployer) getLabels(data *models.WorkerData) map[string]string {
	labels := data.Metadata.ToLabel()
	labels[appLabelKey] = data.Metadata.App
	return labels
}

func (c *deployer) getResources(data *models.WorkerData) (corev1.ResourceList, corev1.ResourceList) {
	requests := c.resourceRequest.DeepCopy()
	limits := c.resourceLimit.DeepCopy()
	if data.ResourceRequest != nil {
		if cpuReq := data.ResourceRequest.CPURequest; cpuReq != nil {
			requests[corev1.ResourceCPU] = *cpuReq
			// remove default limits
			delete(limits, corev1.ResourceCPU)
		}
		if memoryReq := data.ResourceRequest.MemoryRequest; memoryReq != nil {
			requests[corev1.ResourceMemory] = *memoryReq
			limits[corev1.ResourceMemory] = *memoryReq
		}
	}
	return requests, limits
}

func (c *deployer) createDeploymentSpec(data *models.WorkerData, secretName string) (*appsv1.Deployment, error) {
	labels := c.getLabels(data)

	cfgVolName := "config-volume"
	workerContainer := "worker"

	requestsResources, limitsResources := c.getResources(data)
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  workerContainer,
				Image: c.consumerConfig.ImageName,
				Command: []string{
					"python",
					"-m",
					"publisher",
					"+environment=config",
				},
				ImagePullPolicy: corev1.PullIfNotPresent,

				Resources: corev1.ResourceRequirements{
					Requests: requestsResources,
					Limits:   limitsResources,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      cfgVolName,
						MountPath: "/mlobs/observation-publisher/conf/environment",
						ReadOnly:  true,
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "prom-metric",
						ContainerPort: 8000,
						Protocol:      corev1.ProtocolTCP,
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: cfgVolName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			},
		},
	}
	podSpecWithIdentity := enrichIdentityToPod(podSpec, c.consumerConfig.ServiceAccountSecretName, []string{workerContainer})
	numReplicas := c.consumerConfig.Replicas
	if data.ResourceRequest != nil && data.ResourceRequest.Replica > 0 {
		numReplicas = data.ResourceRequest.Replica
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.getDeploymentName(data),
			Namespace: c.targetNamespace(),
			Labels:    labels,
			Annotations: map[string]string{
				PublisherRevisionAnnotationKey: strconv.Itoa(data.Revision),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					appLabelKey: data.Metadata.App,
				},
			},
			Replicas: &numReplicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						PublisherRevisionAnnotationKey: strconv.Itoa(data.Revision),
					},
				},
				Spec: podSpecWithIdentity,
			},
		},
	}, nil
}

func (c *deployer) waitUntilDeploymentReady(ctx context.Context, deployment *appsv1.Deployment, revision int) error {
	deploymentv1 := c.kubeClient.AppsV1().Deployments(deployment.Namespace)
	timeout := time.After(c.consumerConfig.DeploymentTimeout)
	watcher, err := deploymentv1.Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", deployment.Name),
	})
	if err != nil {
		return err
	}
	for {
		select {
		case <-timeout:
			watcher.Stop()
			return fmt.Errorf("timeout waiting deployment ready")
		case watchRes := <-watcher.ResultChan():
			deployManifest, ok := watchRes.Object.(*appsv1.Deployment)
			if !ok {
				return fmt.Errorf("watch result is not deployment")
			}

			rolledOut, err := deploymentRolledOut(deployManifest, revision, true)
			if err != nil {
				return err
			}
			if rolledOut {
				return nil
			}
		}
	}
}

func (c *deployer) getDeploymentName(data *models.WorkerData) string {
	return fmt.Sprintf("%s-%s-mlobs", data.Project, data.ModelName)
}

func (c *deployer) getSecretName(data *models.WorkerData) string {
	return fmt.Sprintf("%s-%s-config", data.Project, data.ModelName)
}

func (c *deployer) Undeploy(ctx context.Context, data *models.WorkerData) error {
	deploymentName := c.getDeploymentName(data)
	if err := c.deleteDeployment(ctx, deploymentName, c.targetNamespace()); err != nil {
		return err
	}

	secretName := c.getSecretName(data)
	if err := c.deleteSecret(ctx, secretName, c.targetNamespace()); err != nil {
		return err
	}

	return nil
}
