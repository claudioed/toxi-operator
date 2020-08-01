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
	toxiv1alpha1 "github.com/claudioed/toxi-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// KillerReconciler reconciles a Killer object
type KillerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Er     record.EventRecorder
}

// +kubebuilder:rbac:groups=toxi.tech.claudioed,resources=killers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods;events,verbs=get;list;watch;create;update;patch;delete
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
	if err := r.EnsurePodsKilled(instance); err != nil {
		return ctrl.Result{}, err
	} else {
		when, _ := time.ParseDuration(instance.Spec.Rule.Every)
		return ctrl.Result{Requeue: true, RequeueAfter: when}, err
	}
}

func (r *KillerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toxiv1alpha1.Killer{}).
		Complete(r)
}
