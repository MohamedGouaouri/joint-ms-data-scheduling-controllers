/*
Copyright 2025.

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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	crdv1 "github.com/MohamedGouaouri/ms-app-controller/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EdgeNetworkTopologyReconciler reconciles a EdgeNetworkTopology object
type EdgeNetworkTopologyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=crd.cs.phd.uqtr,resources=edgenetworktopologies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.cs.phd.uqtr,resources=edgenetworktopologies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crd.cs.phd.uqtr,resources=edgenetworktopologies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EdgeNetworkTopology object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
const finalizerName = "edgenetworktopology.cleanup.finalizers.cs.phd.uqtr"

func (r *EdgeNetworkTopologyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var topo crdv1.EdgeNetworkTopology
	if err := r.Get(ctx, req.NamespacedName, &topo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if topo.ObjectMeta.DeletionTimestamp.IsZero() {
		// Not being deleted, ensure finalizer is present
		if !controllerutil.ContainsFinalizer(&topo, finalizerName) {
			controllerutil.AddFinalizer(&topo, finalizerName)
			if err := r.Update(ctx, &topo); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Being deleted
		if controllerutil.ContainsFinalizer(&topo, finalizerName) {
			log.Info("Cleaning up resources for", "name", topo.Name)
			if err := r.cleanupResources(ctx, &topo); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(&topo, finalizerName)
			if err := r.Update(ctx, &topo); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	for _, edge := range topo.Spec.Edges {
		// Create or update ConfigMap
		log.Info("Applying config to", "edge", edge.Name)
		cm := generateConfigMap(edge, topo.Name)
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error { return nil })
		if err != nil {
			log.Error(err, "failed to apply ConfigMap", "edge", edge.Name)
			return ctrl.Result{}, err
		}
		// Create or update DaemonSet
		ds := generateBandwidthDaemonSet(edge, topo.Name)
		// Search for ds
		objName := types.NamespacedName{
			Name:      ds.Name,
			Namespace: ds.Namespace,
		}
		var dummyRef appsv1.DaemonSet
		err = r.Client.Get(ctx, objName, &dummyRef)
		if err != nil {
			if err := r.Client.Create(ctx, ds); err != nil {
				log.Error(err, "failed to apply DaemonSet", "edge", edge.Name)
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func generateConfigMap(edge crdv1.EdgeNode, topologyName string) *corev1.ConfigMap {
	data := "interfaces:\n"
	for _, link := range edge.Links {
		data += fmt.Sprintf("  - interface: %s\n    bandwidth: %d\n", link.InterfaceName, link.BandwidthCapacity)
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("netconfig-%s", edge.Name),
			Namespace: "default",
			Labels: map[string]string{
				"topology": topologyName,
			},
		},
		Data: map[string]string{
			"network_config.yaml": data,
		},
	}
}

func generateBandwidthDaemonSet(edge crdv1.EdgeNode, topologyName string) *appsv1.DaemonSet {
	labels := map[string]string{
		"app":       "bandwidth-limiter",
		"topology":  topologyName,
		"edge-node": edge.Name,
	}

	configsMountPath := "/confs"
	env := []corev1.EnvVar{
		{Name: "CONFIG_FILE", Value: configsMountPath + "/network_config.yaml"},
	}

	// command := []string{"sleep 36000"}
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("bw-limiter-%s", edge.Name),
			Namespace: "default",
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					HostNetwork: true,
					Containers: []corev1.Container{
						{
							Name:  "bandwidth-limiter",
							Image: "linuxtitan/network-proxy",
							SecurityContext: &corev1.SecurityContext{
								Privileged: pointer(true),
							},
							Env: env,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "net-config",
									MountPath: configsMountPath,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "net-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("netconfig-%s", edge.Name),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *EdgeNetworkTopologyReconciler) cleanupResources(ctx context.Context, topo *crdv1.EdgeNetworkTopology) error {
	for _, edge := range topo.Spec.Edges {
		cmName := fmt.Sprintf("netconfig-%s", edge.Name)
		dsName := fmt.Sprintf("bw-limiter-%s", edge.Name)

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: "default",
			},
		}
		if err := r.Client.Delete(ctx, cm); client.IgnoreNotFound(err) != nil {
			return err
		}

		ds := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dsName,
				Namespace: "default",
			},
		}
		if err := r.Client.Delete(ctx, ds); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func pointer(b bool) *bool {
	return &b
}

// SetupWithManager sets up the controller with the Manager.
func (r *EdgeNetworkTopologyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.EdgeNetworkTopology{}).
		Named("edgenetworktopology").
		Complete(r)
}
