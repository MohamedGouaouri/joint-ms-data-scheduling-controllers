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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	crdv1 "github.com/MohamedGouaouri/ms-app-controller/api/v1"
)

// MicroserviceApplicationReconciler reconciles a MicroserviceApplication object
type MicroserviceApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=crd.cs.phd.uqtr,resources=microserviceapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.cs.phd.uqtr,resources=microserviceapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crd.cs.phd.uqtr,resources=microserviceapplications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MicroserviceApplication object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *MicroserviceApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// TODO(user): your logic here
	var app crdv1.MicroserviceApplication
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		logger.Error(err, "unable to fetch Application")
		// Ignore not-found errors (e.g., if the resource was deleted)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Traverse the microservices to calculate cpu demdands
	// var microserviceCPUDemands map[string]int64 = make(map[string]int64)
	var ranks map[string]int = make(map[string]int)
	for _, ms := range app.Spec.Microservices {

		demands, err := r.GetMicroserviceCPUDemands(ctx, ms)
		if err != nil {
			logger.Error(err, "unable to fetch Microservice demands")
			// Ignore not-found errors (e.g., if the resource was deleted)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		logger.Info("Total CPU for microservice", "name", ms.Name, "cpu(millicores)", demands)
		ranks[ms.Name] = r.Rank(app, ms, demands)

		// Add annotations to the pods
		r.AnnotatePod(ctx, ms, "topology-aware-scheduling.cs.phd.uqtr/microservice", ms.Name)
		r.AnnotatePod(ctx, ms, "topology-aware-scheduling.cs.phd.uqtr/rank", fmt.Sprintf("%s", ranks[ms.Name]))

	}
	app.Status.Ranks = ranks
	if err := r.Status().Update(ctx, &app); err != nil {
		logger.Error(err, "unable to update Application status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MicroserviceApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.MicroserviceApplication{}).
		Named("microserviceapplication").
		Complete(r)
}

func (r *MicroserviceApplicationReconciler) Rank(app crdv1.MicroserviceApplication, m crdv1.Microservice, demands int64) int {

	nerighours := app.Neighbours(m)
	if len(nerighours) == 0 {
		return int(demands)
	}
	var neighoursRanks []int = make([]int, 0)
	for _, n := range nerighours {
		neighbourMs := n.Microservice
		neighbourDemands, err := r.GetMicroserviceCPUDemands(context.TODO(), neighbourMs)
		if err != nil {
			return 0
		}
		neighoursRanks = append(neighoursRanks, n.Dependency.MinBandwidth+r.Rank(app, neighbourMs, neighbourDemands))
	}
	return int(demands) + maxOf(neighoursRanks)
}

func (r *MicroserviceApplicationReconciler) GetMicroserviceCPUDemands(ctx context.Context, ms crdv1.Microservice) (int64, error) {
	logger := logf.FromContext(ctx)

	var deployment appsv1.Deployment
	depNamespaced := types.NamespacedName{
		Namespace: ms.Namespace,
		Name:      ms.DeploymentRef,
	}
	if err := r.Get(ctx, depNamespaced, &deployment); err != nil {
		logger.Error(err, "unable to fetch deployment")
		return 0, err
	}
	podList := &corev1.PodList{}
	labelSelector := client.MatchingLabels(deployment.Spec.Selector.MatchLabels)
	if err := r.List(ctx, podList, client.InNamespace(ms.Namespace), labelSelector); err != nil {
		logger.Error(err, "unable to list pods for deployment", "deployment", depNamespaced)
		return 0, err
	}
	var totalCPU resource.Quantity
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			if cpuReq, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				totalCPU.Add(cpuReq)
			}
		}
	}
	return totalCPU.MilliValue(), nil
}

func (r *MicroserviceApplicationReconciler) AnnotatePod(ctx context.Context, ms crdv1.Microservice, annotationKey, annotationValue string) error {
	logger := logf.FromContext(ctx)

	var deployment appsv1.Deployment
	depNamespaced := types.NamespacedName{
		Namespace: ms.Namespace,
		Name:      ms.DeploymentRef,
	}
	if err := r.Get(ctx, depNamespaced, &deployment); err != nil {
		logger.Error(err, "unable to fetch deployment")
		return err
	}
	podList := &corev1.PodList{}
	labelSelector := client.MatchingLabels(deployment.Spec.Selector.MatchLabels)
	if err := r.List(ctx, podList, client.InNamespace(ms.Namespace), labelSelector); err != nil {
		logger.Error(err, "unable to list pods for deployment", "deployment", depNamespaced)
		return err
	}
	for _, pod := range podList.Items {
		// Annotate Pod
		p := pod

		if p.Annotations == nil {
			p.Annotations = make(map[string]string)
		}

		p.Annotations[annotationKey] = annotationValue

		if err := r.Update(ctx, &p); err != nil {
			logger.Error(err, "failed to update pod annotation", "pod", p.Name)
			// Optionally continue to next pod instead of failing all
			continue
		}
		logger.Info("Annotated pod", "pod", p.Name)
	}
	return nil
}

func maxOf(arr []int) int {
	if len(arr) == 0 {
		panic("Cannot find max of an empty slice")
	}
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max
}
