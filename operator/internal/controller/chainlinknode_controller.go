/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oraclev1alpha1 "github.com/JooKop/chainlink-kubernetes-operator/api/v1alpha1"
)

// Definitions to manage status conditions
const (
	// typeAvailableChainlinkNode represents the status of the Deployment reconciliation
	typeAvailableChainlinkNode = "Available"
)

// ChainlinkNodeReconciler reconciles a ChainlinkNode object
type ChainlinkNodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oracle.example.com,resources=chainlinknodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oracle.example.com,resources=chainlinknodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oracle.example.com,resources=chainlinknodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ChainlinkNode object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ChainlinkNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	reQueue := false

	//Fetch the changed ChainlinkNode resource
	chainlinkNode := &oraclev1alpha1.ChainlinkNode{}
	err := r.Get(ctx, req.NamespacedName, chainlinkNode)
	if err != nil {
		log.Error(err, "An error occurred")

		if apierrors.IsNotFound(err) {
			log.Info("ChainlinkNode resource not found. Ignoring error because it means the resource was deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the updated object. Retry.
		log.Error(err, "Failed to get ChainlinkNode")
		return ctrl.Result{}, err
	}

	// Create a new Chainlink Node deployment if one doesn't exist
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: chainlinkNode.Name, Namespace: chainlinkNode.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Chainlink node deployment didn't exist, create a new deployment
		dep, err := r.deploymentForChainlinkNode(chainlinkNode)
		if err != nil {
			log.Error(err, "Failed to define new Deployment resource for ChainlinkNode")

			// update ChainlinkNode object status
			meta.SetStatusCondition(&chainlinkNode.Status.Conditions, metav1.Condition{Type: typeAvailableChainlinkNode,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create Deployment for the custom resource (%s): (%s)", chainlinkNode.Name, err)})

			if err := r.Status().Update(ctx, chainlinkNode); err != nil {
				log.Error(err, "Failed to update ChainlinkNode status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new Deployment",
			"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		if err = r.Create(ctx, dep); err != nil {
			log.Error(err, "Failed to create new Deployment",
				"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}

		// Deployment created successfully
		reQueue = true
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		// Return error and requeue to try again
		return ctrl.Result{}, err
	}

	// Create a new Chainlink Node Service if one doesn't exist
	foundSvc := &v1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: chainlinkNode.Name + "-service", Namespace: chainlinkNode.Namespace}, foundSvc)
	if err != nil && apierrors.IsNotFound(err) {
		// Chainlink node service didn't exist, create a new service
		svc, err := r.serviceForChainlinkNode(chainlinkNode)
		if err != nil {
			log.Error(err, "Failed to define new Service resource for ChainlinkNode")

			// update ChainlinkNode object status
			meta.SetStatusCondition(&chainlinkNode.Status.Conditions, metav1.Condition{Type: typeAvailableChainlinkNode,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create Service for the custom resource deployment (%s): (%s)", chainlinkNode.Name, err)})

			if err := r.Status().Update(ctx, chainlinkNode); err != nil {
				log.Error(err, "Failed to update ChainlinkNode status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new Service",
			"Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		if err = r.Create(ctx, svc); err != nil {
			log.Error(err, "Failed to create new Deployment",
				"Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}

		// Service created successfully
		reQueue = true
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		// Return error and requeue to try again
		return ctrl.Result{}, err
	}

	if reQueue {
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainlinkNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oraclev1alpha1.ChainlinkNode{}).
		Complete(r)
}

// serviceForChainlinkNode returns a Chainlink node Service object
func (r *ChainlinkNodeReconciler) serviceForChainlinkNode(
	chainlinkNode *oraclev1alpha1.ChainlinkNode) (*v1.Service, error) {
	ls := labelsForChainlinkNode(chainlinkNode.Name)
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chainlinkNode.Name + "-service",
			Namespace: chainlinkNode.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Name:       "operator-api",
					Port:       80,
					TargetPort: intstr.FromInt(6688),
					Protocol:   "TCP",
				},
			},
		},
	}

	// Set the ownerRef for the Deployment
	if err := ctrl.SetControllerReference(chainlinkNode, svc, r.Scheme); err != nil {
		return nil, err
	}
	return svc, nil
}

// deploymentForChainlinkNode returns a Chainlink node Deployment object
func (r *ChainlinkNodeReconciler) deploymentForChainlinkNode(
	chainlinkNode *oraclev1alpha1.ChainlinkNode) (*appsv1.Deployment, error) {
	ls := labelsForChainlinkNode(chainlinkNode.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chainlinkNode.Name,
			Namespace: chainlinkNode.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "secret-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: chainlinkNode.Name + "-secrets",
							},
						},
					}, {
						Name: "config-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: chainlinkNode.Name + "-config",
								},
							},
						},
					},
					},
					Containers: []corev1.Container{{
						Image:   "smartcontract/chainlink:2.1.1",
						Name:    "chainlink-node",
						Command: []string{"chainlink"},
						Args:    []string{"node", "-config", "/chainlink/config/config.toml", "-secrets", "/chainlink/secrets/secrets.toml", "start", "-a", "/chainlink/secrets/apiuser.txt"},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "secret-volume",
							ReadOnly:  true,
							MountPath: "/chainlink/secrets/",
						}, {
							Name:      "config-volume",
							ReadOnly:  true,
							MountPath: "/chainlink/config",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 6688,
							Name:          "operator-api",
						}},
					}, {
						Image: "postgres:latest",
						Name:  "chainlink-postgres",
						Env: []corev1.EnvVar{{
							Name:  "POSTGRES_PASSWORD",
							Value: "mysecretpassword",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 5432,
							Name:          "postgres",
						}},
					},
					},
				},
			},
		},
	}

	// Set the ownerRef for the Deployment
	if err := ctrl.SetControllerReference(chainlinkNode, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}

// labelsForChainlinkNode returns the labels for selecting the resources
func labelsForChainlinkNode(name string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": "Chainlink",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/part-of":    "chainlink-kubernetes-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}
