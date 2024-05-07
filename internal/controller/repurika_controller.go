/*
Copyright 2024.

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
	"math"
	"math/rand"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	batchv1 "github.com/bonavadeur/repurika/api/v1"
	"github.com/bonavadeur/repurika/internal/bonalib"
)

const (
	typeAvailableRepurika = "Available"
)

// RepurikaReconciler reconciles a Repurika object
type RepurikaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.bonavadeur.io,resources=repurikas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.bonavadeur.io,resources=repurikas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.bonavadeur.io,resources=repurikas/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Repurika object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *RepurikaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	repurika := &batchv1.Repurika{}
	if err := r.Get(ctx, req.NamespacedName, repurika); err != nil {
		bonalib.Warn("unable to fetch repurika")
		return ctrl.Result{}, nil
	}
	bonalib.Succ("Reconcile", repurika.Name)

	if repurika.Status.Conditions == nil || len(repurika.Status.Conditions) == 0 {
		bonalib.Info("repurika.Status.Conditions is nil")
		meta.SetStatusCondition(&repurika.Status.Conditions, metav1.Condition{
			Type:    typeAvailableRepurika,
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation",
		})
		if err := r.Status().Update(ctx, repurika); err != nil {
			bonalib.Warn("failed to update Repurika Status")
			return ctrl.Result{}, err
		}
		if err := r.Get(ctx, req.NamespacedName, repurika); err != nil {
			bonalib.Warn("Failed to re-fetch repurika")
		}
	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(repurika.Namespace),
		client.MatchingLabels(repurika.Spec.Template.ObjectMeta.Labels),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		bonalib.Warn("Failed to list pods")
		return ctrl.Result{}, err
	}

	podNames := getPodNames(podList.Items)
	podsCount := len(podList.Items)
	size := repurika.Spec.Size
	countDiff := int(math.Abs(float64(podsCount - int(size))))
	bonalib.Log("", podsCount, size, countDiff)
	rand.Seed(time.Now().Unix())
	if podsCount < int(size) {
		bonalib.Info("Need to create more pods")
		for i := 0; i < countDiff; i++ {
			pod, err := r.createPodTemplate(repurika)
			if err != nil {
				bonalib.Warn("Failed to define new Pod for Repurika")
				meta.SetStatusCondition(&repurika.Status.Conditions, metav1.Condition{
					Type:    typeAvailableRepurika,
					Status:  metav1.ConditionFalse,
					Reason:  "Reconciling",
					Message: fmt.Sprintf("Failed to define new pod for the custom resource (%s): (%s)", repurika.Name, err),
				})
				if err := r.Status().Update(ctx, repurika); err != nil {
					bonalib.Warn("Failed to update Repurika Status")
					return ctrl.Result{}, nil
				}
				return ctrl.Result{}, nil
			}
			if err := r.Create(ctx, pod); err != nil {
				bonalib.Warn("Failed to create new Pod", err)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{Requeue: true}, nil
	} else if podsCount > int(size) {
		bonalib.Info("Need to delete less pods")
		for i := 0; i < countDiff; i++ {
			pod := &corev1.Pod{}
			podTobeDeletedName := podNames[i]
			if err := r.Get(ctx, types.NamespacedName{Name: podTobeDeletedName, Namespace: repurika.Namespace}, pod); err != nil {
				bonalib.Warn("Pod doesnt exist", podTobeDeletedName)
				meta.SetStatusCondition(&repurika.Status.Conditions, metav1.Condition{
					Type:    typeAvailableRepurika,
					Status:  metav1.ConditionFalse,
					Reason:  "Reconciling",
					Message: fmt.Sprintf("Failed to fetch the pod to be deleted for the custom resource (%s): (%s)", repurika.Name, err),
				})
				if err = r.Status().Update(ctx, repurika); err != nil {
					bonalib.Warn("Failed to update Repurika status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}
			if err := r.Delete(ctx, pod, client.GracePeriodSeconds(0)); err != nil {
				bonalib.Warn("Failed to delete the pod", podTobeDeletedName)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	meta.SetStatusCondition(&repurika.Status.Conditions, metav1.Condition{
		Type:    typeAvailableRepurika,
		Status:  metav1.ConditionTrue,
		Reason:  "Reconciling",
		Message: fmt.Sprintf("Pods for custom resource (%s) created/deleted successfully", repurika.Name),
	})

	repurika.Status.Pods = podNames
	if err := r.Status().Update(ctx, repurika); err != nil {
		bonalib.Warn("Failed to update Repurika status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func generatePodName() string {
	letters := []byte("1234567890abcdefghijklmnopqrstuvwxyz")
	ranStr := make([]byte, 5)
	for i := 0; i < 5; i++ {
		ranStr[i] = letters[rand.Intn(len(letters))]
	}
	str := string(ranStr)
	return str
}

func getPodNames(podList []corev1.Pod) []string {
	podNames := []string{}
	for _, pod := range podList {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *RepurikaReconciler) createPodTemplate(repurika *batchv1.Repurika) (*corev1.Pod, error) {
	template := repurika.Spec.Template

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-%v", repurika.Name, generatePodName()),
			Namespace: repurika.Namespace,
			Labels:    template.ObjectMeta.Labels,
		},
		Spec: template.Spec,
	}

	if err := ctrl.SetControllerReference(repurika, pod, r.Scheme); err != nil {
		return nil, err
	}

	return pod, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepurikaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	bonalib.Succ("SetupWithManager")

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Repurika{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
