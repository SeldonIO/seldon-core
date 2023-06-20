/*
Copyright 2023 Seldon Technologies Ltd.

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

package server

import (
	"context"

	v1 "k8s.io/api/core/v1"
	auth "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

const (
	AgentRoleName               = "agent-role"
	HodometerServiceAccountName = "hodometer"
	HodometerRoleName           = "hodometer-role"
	SchedulerServiceAccountName = "seldon-scheduler"
	SchedulerRoleName           = "seldon-scheduler-role"
	ServerServiceAccountName    = "seldon-server"
	AgentRoleBindingName        = "agent-rolebinding"
	HodometerRoleBindingName    = "hodometer-rolebinding"
	SchedulerRoleBindingName    = "seldon-scheduler-rolebinding"
)

type ComponentRBACReconciler struct {
	common.ReconcilerConfig
	meta            metav1.ObjectMeta
	Roles           []*auth.Role
	RoleBindings    []*auth.RoleBinding
	ServiceAccounts []*v1.ServiceAccount
}

func NewComponentRBACReconciler(
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta) *ComponentRBACReconciler {
	return &ComponentRBACReconciler{
		ReconcilerConfig: common,
		Roles:            getRoles(meta),
		RoleBindings:     getRoleBindings(meta),
		ServiceAccounts:  getServiceAccounts(meta),
	}
}

func (s *ComponentRBACReconciler) GetResources() []client.Object {
	var objs []client.Object
	for _, role := range s.Roles {
		objs = append(objs, role)
	}
	for _, roleBinding := range s.RoleBindings {
		objs = append(objs, roleBinding)
	}
	for _, serviceAccount := range s.ServiceAccounts {
		objs = append(objs, serviceAccount)
	}
	return objs
}

func getServiceAccounts(meta metav1.ObjectMeta) []*v1.ServiceAccount {
	return []*v1.ServiceAccount{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      HodometerServiceAccountName,
				Namespace: meta.Namespace,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      SchedulerServiceAccountName,
				Namespace: meta.Namespace,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServerServiceAccountName,
				Namespace: meta.Namespace,
			},
		},
	}
}

func getRoleBindings(meta metav1.ObjectMeta) []*auth.RoleBinding {
	return []*auth.RoleBinding{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      AgentRoleBindingName,
				Namespace: meta.Namespace,
			},
			Subjects: []auth.Subject{
				{
					Kind:      auth.ServiceAccountKind,
					Name:      ServerServiceAccountName,
					Namespace: meta.Namespace,
				},
			},
			RoleRef: auth.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     AgentRoleName,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      HodometerRoleBindingName,
				Namespace: meta.Namespace,
			},
			Subjects: []auth.Subject{
				{
					Kind:      auth.ServiceAccountKind,
					Name:      HodometerServiceAccountName,
					Namespace: meta.Namespace,
				},
			},
			RoleRef: auth.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     HodometerRoleName,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      SchedulerRoleBindingName,
				Namespace: meta.Namespace,
			},
			Subjects: []auth.Subject{
				{
					Kind:      auth.ServiceAccountKind,
					Name:      SchedulerServiceAccountName,
					Namespace: meta.Namespace,
				},
			},
			RoleRef: auth.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     SchedulerRoleName,
			},
		},
	}
}

func getRoles(meta metav1.ObjectMeta) []*auth.Role {
	return []*auth.Role{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      AgentRoleName,
				Namespace: meta.Namespace,
			},
			Rules: []auth.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"configmaps"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      HodometerRoleName,
				Namespace: meta.Namespace,
			},
			Rules: []auth.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      SchedulerRoleName,
				Namespace: meta.Namespace,
			},
			Rules: []auth.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"configmaps"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
	}
}

func (s *ComponentRBACReconciler) getReconcileOperationForRole(idx int, role *auth.Role) (constants.ReconcileOperation, error) {
	found := &auth.Role{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      role.GetName(),
		Namespace: role.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}
	if equality.Semantic.DeepEqual(role.Rules, found.Rules) {
		// Update our version so we have Status if needed
		s.Roles[idx] = found
		return constants.ReconcileNoChange, nil
	}
	// Update resource version so the client Update will succeed
	s.Roles[idx].SetResourceVersion(found.ResourceVersion)
	return constants.ReconcileUpdateNeeded, nil
}

func (s *ComponentRBACReconciler) getReconcileOperationForRoleBinding(idx int, roleBinding *auth.RoleBinding) (constants.ReconcileOperation, error) {
	found := &auth.RoleBinding{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      roleBinding.GetName(),
		Namespace: roleBinding.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}
	if equality.Semantic.DeepEqual(roleBinding.Subjects, found.Subjects) &&
		equality.Semantic.DeepEqual(roleBinding.RoleRef, found.RoleRef) {
		// Update our version so we have Status if needed
		s.RoleBindings[idx] = found
		return constants.ReconcileNoChange, nil
	}
	// Update resource version so the client Update will succeed
	s.Roles[idx].SetResourceVersion(found.ResourceVersion)
	return constants.ReconcileUpdateNeeded, nil
}
func (s *ComponentRBACReconciler) getReconcileOperationForServiceAccount(idx int, serviceAccount *v1.ServiceAccount) (constants.ReconcileOperation, error) {
	found := &v1.ServiceAccount{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      serviceAccount.GetName(),
		Namespace: serviceAccount.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}

	return constants.ReconcileNoChange, nil
}

func (s *ComponentRBACReconciler) reconcileRoles() error {
	logger := s.Logger.WithName("ReconcileRoles")
	for idx, role := range s.Roles {
		op, err := s.getReconcileOperationForRole(idx, role)
		switch op {
		case constants.ReconcileCreateNeeded:
			logger.V(1).Info("Role Create", "Name", role.GetName(), "Namespace", role.GetNamespace())
			err = s.Client.Create(s.Ctx, role)
			if err != nil {
				logger.Error(err, "Failed to create service", "Name", role.GetName(), "Namespace", role.GetNamespace())
				return err
			}
		case constants.ReconcileUpdateNeeded:
			logger.V(1).Info("Role Update", "Name", role.GetName(), "Namespace", role.GetNamespace())
			err = s.Client.Update(s.Ctx, role)
			if err != nil {
				logger.Error(err, "Failed to update service", "Name", role.GetName(), "Namespace", role.GetNamespace())
				return err
			}
		case constants.ReconcileNoChange:
			logger.V(1).Info("Role No Change", "Name", role.GetName(), "Namespace", role.GetNamespace())
		case constants.ReconcileUnknown:
			logger.Error(err, "Failed to get reconcile operation for Role", "Name", role.GetName(), "Namespace", role.GetNamespace())
			return err
		}
	}
	return nil
}

func (s *ComponentRBACReconciler) reconcileRoleBindings() error {
	logger := s.Logger.WithName("ReconcileRoles")
	for idx, roleBinding := range s.RoleBindings {
		op, err := s.getReconcileOperationForRoleBinding(idx, roleBinding)
		switch op {
		case constants.ReconcileCreateNeeded:
			logger.V(1).Info("RoleBinding Create", "Name", roleBinding.GetName(), "Namespace", roleBinding.GetNamespace())
			err = s.Client.Create(s.Ctx, roleBinding)
			if err != nil {
				logger.Error(err, "Failed to create service", "Name", roleBinding.GetName(), "Namespace", roleBinding.GetNamespace())
				return err
			}
		case constants.ReconcileUpdateNeeded:
			logger.V(1).Info("RoleBinding Update", "Name", roleBinding.GetName(), "Namespace", roleBinding.GetNamespace())
			err = s.Client.Update(s.Ctx, roleBinding)
			if err != nil {
				logger.Error(err, "Failed to update service", "Name", roleBinding.GetName(), "Namespace", roleBinding.GetNamespace())
				return err
			}
		case constants.ReconcileNoChange:
			logger.V(1).Info("RoleBinding No Change", "Name", roleBinding.GetName(), "Namespace", roleBinding.GetNamespace())
		case constants.ReconcileUnknown:
			logger.Error(err, "Failed to get reconcile operation for RoleBinding", "Name", roleBinding.GetName(), "Namespace", roleBinding.GetNamespace())
			return err
		}
	}
	return nil
}

func (s *ComponentRBACReconciler) reconcileServiceAccounts() error {
	logger := s.Logger.WithName("ReconcileServiceAccounts")
	for idx, serviceAccount := range s.ServiceAccounts {
		op, err := s.getReconcileOperationForServiceAccount(idx, serviceAccount)
		switch op {
		case constants.ReconcileCreateNeeded:
			logger.V(1).Info("ServiceAccount Create", "Name", serviceAccount.GetName(), "Namespace", serviceAccount.GetNamespace())
			err = s.Client.Create(s.Ctx, serviceAccount)
			if err != nil {
				logger.Error(err, "Failed to create service", "Name", serviceAccount.GetName(), "Namespace", serviceAccount.GetNamespace())
				return err
			}
		case constants.ReconcileUpdateNeeded:
			logger.V(1).Info("ServiceAccount Update", "Name", serviceAccount.GetName(), "Namespace", serviceAccount.GetNamespace())
			err = s.Client.Update(s.Ctx, serviceAccount)
			if err != nil {
				logger.Error(err, "Failed to update service", "Name", serviceAccount.GetName(), "Namespace", serviceAccount.GetNamespace())
				return err
			}
		case constants.ReconcileNoChange:
			logger.V(1).Info("ServiceAccount No Change", "Name", serviceAccount.GetName(), "Namespace", serviceAccount.GetNamespace())
		case constants.ReconcileUnknown:
			logger.Error(err, "Failed to get reconcile operation for ServiceAccount", "Name", serviceAccount.GetName(), "Namespace", serviceAccount.GetNamespace())
			return err
		}
	}
	return nil
}

func (s *ComponentRBACReconciler) Reconcile() error {
	err := s.reconcileServiceAccounts()
	if err != nil {
		return err
	}
	err = s.reconcileRoles()
	if err != nil {
		return err
	}
	err = s.reconcileRoleBindings()
	if err != nil {
		return err
	}
	return nil
}

func (s *ComponentRBACReconciler) GetConditions() []*apis.Condition {
	return nil
}
