/*


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

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toxiv1alpha1 "github.com/claudioed/toxi-operator/api/v1alpha1"
)

// KillerReconciler reconciles a Killer object
type KillerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=toxi.tech.claudioed,resources=killers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=toxi.tech.claudioed,resources=killers/status,verbs=get;update;patch

func (r *KillerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Starting reconcile", "killer_name", req.Name, "killer_namespace", req.Namespace)
	instance := &toxiv1alpha1.Killer{}
	r.Log.Info("Finding updated killer ", "killer_name", req.Name, "killer_namespace", req.Namespace)
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("Killer deleted nothing to do", "killer_name", req.Name, "killer_namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
	}
	r.Log.Info("Finding pods to delete ", "killer_name", req.Name, "killer_namespace", req.Namespace)
	sel := labels.NewSelector()
	for key, value := range instance.Spec.Selector.MatchLabels {
		r, _ := labels.NewRequirement(key, selection.Equals, []string{value})
		sel.Add(*r)
	}
	pods := &v1.PodList{}
	r.Client.List(context.TODO(), pods, &client.ListOptions{
		LabelSelector: client.MatchingLabelsSelector{Selector: sel},
		Namespace:     req.Namespace,
	})
	if len(pods.Items) > 0 {
		r.Log.Info("There are pods to delete. ", "Number:", len(pods.Items))
		err := r.KillPods(pods)
		if err != nil {
			r.Log.Error(err, "Error to delete pods ", "killer_name", req.Name, "killer_namespace", req.Namespace)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true, RequeueAfter: instance.Spec.Rule.Every}, nil
	} else {
		r.Log.Info("There are no pods to delete ")
	}
	return ctrl.Result{}, nil
}

func (r *KillerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toxiv1alpha1.Killer{}).
		Complete(r)
}
