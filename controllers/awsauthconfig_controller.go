/*
Copyright 2021.

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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	authv1alpha1 "github.com/gp42/aws-auth-operator/api/v1alpha1"
	"github.com/gp42/aws-auth-operator/controllers/common"
	"github.com/gp42/aws-auth-operator/controllers/model"
)

const (
	AUTH_CM_NAME = "aws-auth"
)

// AwsAuthSyncConfigReconciler reconciles a AwsAuthSyncConfig object
type AwsAuthSyncConfigReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	AwsAuthSyncConfigName string
	SyncInterval          int
	IamClient             model.IamClientInterface
}

//+kubebuilder:rbac:groups="",namespace=kube-system,resources=configmaps,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=auth.ops42.org,namespace=kube-system,resources=awsauthsyncconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=auth.ops42.org,namespace=kube-system,resources=awsauthsyncconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=auth.ops42.org,namespace=kube-system,resources=awsauthsyncconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *AwsAuthSyncConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get current AwsAuth ConfigMap
	awsAuthCm := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: req.NamespacedName.Namespace,
		Name:      AUTH_CM_NAME,
	}, awsAuthCm)
	if err != nil {
		ctrl.Log.Error(err, "Failed to get AwsAuth ConfigMap", "Namespace", req.NamespacedName.Namespace, "Name", AUTH_CM_NAME)
		return ctrl.Result{Requeue: true}, err
	}
	oldSum := common.Sha512SumFromStringSorted(awsAuthCm.Data["mapUsers"])

	// Get new state from AWS
	awsAuthConfig := &authv1alpha1.AwsAuthSyncConfig{}
	err = r.Get(ctx, req.NamespacedName, awsAuthConfig)
	if err != nil {
		ctrl.Log.Error(err, "Failed to get AwsAuthSyncConfig")
		return ctrl.Result{Requeue: true}, err
	}

	logger.Info("Reconciling AwsAuthSyncConfig: " + awsAuthConfig.Name)

	mapUsersObject := model.NewMapUsers()
	for _, syncIamGroup := range awsAuthConfig.Spec.SyncIamGroups {
		logger.V(1).Info("Reconciling IAM Group '" + syncIamGroup.Source + "' to RBAC groups.")
		err := mapUsersObject.LoadFromIamGroup(ctx, r.IamClient, syncIamGroup.Source, syncIamGroup.Dest)
		if err != nil {
			ctrl.Log.Error(err, "problem with loading MapUsers from IAM group")
			return ctrl.Result{Requeue: true}, err
		}
	}

	patch, newSum, err := mapUsersObject.ToPatch()
	if err != nil {
		ctrl.Log.Error(err, "yaml Marshal error")
		return ctrl.Result{Requeue: true}, err
	}

	syncStatus := "Failure"
	if newSum == oldSum {
		ctrl.Log.V(1).Info("data did not change, will skip updating", "oldSum", oldSum, "newSum", newSum)
		syncStatus = "No Change"
	} else {
		ctrl.Log.V(1).Info("applying patch", "patch", string(patch), "oldSum", oldSum, "newSum", newSum)

		err = r.Patch(context.Background(), &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: req.NamespacedName.Namespace,
				Name:      AUTH_CM_NAME,
			},
		}, client.RawPatch(types.StrategicMergePatchType, patch))
		if err != nil {
			ctrl.Log.Error(err, "kubernetes patch error")
			return ctrl.Result{Requeue: true}, err
		}
		syncStatus = "Success"
	}

	// Update status on AwsAuthSyncConfig
	awsAuthConfig.Status.LastSyncTime = &metav1.Time{Time: time.Now()}
	awsAuthConfig.Status.Status = syncStatus
	if err := r.Status().Update(ctx, awsAuthConfig); err != nil {
		ctrl.Log.Error(err, "unable to update AwsAuthSyncConfig status")
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{RequeueAfter: (time.Duration(r.SyncInterval) * time.Second)}, nil
}

func onlyProcessAwsAuthSyncConfigWithName(awsAuthConfigName string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetName() == awsAuthConfigName
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Object.GetName() == awsAuthConfigName
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew.GetName() == awsAuthConfigName
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return e.Object.GetName() == awsAuthConfigName
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsAuthSyncConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1alpha1.AwsAuthSyncConfig{}).
		WithEventFilter(onlyProcessAwsAuthSyncConfigWithName(r.AwsAuthSyncConfigName)).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
