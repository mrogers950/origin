package policy

import (
	"reflect"
	"testing"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	kapi "k8s.io/kubernetes/pkg/api"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	"github.com/openshift/origin/pkg/client/testclient"
	"github.com/openshift/origin/pkg/oc/admin/policy"
)

func TestModifyRoleBindingName(t *testing.T) {
	var fake *testclient.Fake
	fake = testclient.NewSimpleFake()

	tests := map[string]struct {
		role                        string
		rolenamespace               string
		rolebindingname             string
		users                       []string
		accessor                    policy.RoleBindingAccessor
		existingRoleBindings        *authorizationapi.RoleBindingList
		existingClusterRoleBindings *authorizationapi.ClusterRoleBindingList
		clusterTest                 bool
	}{
		// no rolebinding name provided - create "edit" for role "edit"
		"create-rolebinding": {
			role:            "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "",
			users: []string{
				"foo",
			},
			accessor: policy.NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			existingRoleBindings: &authorizationapi.RoleBindingList{
				Items: []authorizationapi.RoleBinding{},
			},
		},
		// rolebinding name provided - create "custom" for role "edit"
		"create-named-binding": {
			role:            "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "custom",
			users: []string{
				"foo",
			},
			accessor: policy.NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			existingRoleBindings: &authorizationapi.RoleBindingList{
				Items: []authorizationapi.RoleBinding{},
			},
		},
		// no rolebinding name provided - modify "edit"
		"update-default-binding": {
			role:            "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "",
			users: []string{
				"foo",
				"bar",
			},
			accessor: policy.NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			existingRoleBindings: &authorizationapi.RoleBindingList{
				Items: []authorizationapi.RoleBinding{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					},
					Subjects: []kapi.ObjectReference{{
						Name: "foo",
						Kind: authorizationapi.UserKind,
					}},
					RoleRef: kapi.ObjectReference{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					}}, {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "custom",
						Namespace: metav1.NamespaceDefault,
					},
					Subjects: []kapi.ObjectReference{{
						Name: "baz",
						Kind: authorizationapi.UserKind,
					}},
					RoleRef: kapi.ObjectReference{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					}},
				},
			},
		},
		// rolebinding name provided - modify "custom"
		"update-named-binding": {
			role:            "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "custom",
			users: []string{
				"bar",
				"baz",
			},
			accessor: policy.NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			existingRoleBindings: &authorizationapi.RoleBindingList{
				Items: []authorizationapi.RoleBinding{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					},
					Subjects: []kapi.ObjectReference{{
						Name: "foo",
						Kind: authorizationapi.UserKind,
					}},
					RoleRef: kapi.ObjectReference{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					}}, {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "custom",
						Namespace: metav1.NamespaceDefault,
					},
					Subjects: []kapi.ObjectReference{{
						Name: "bar",
						Kind: authorizationapi.UserKind,
					}},
					RoleRef: kapi.ObjectReference{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					}},
				},
			},
		},
		// cluster variant with rolebinding name provided - modify "custom"
		"update-named-clusterbinding": {
			role:            "edit",
			rolenamespace:   "",
			rolebindingname: "custom",
			users: []string{
				"bar",
				"baz",
			},
			accessor: policy.NewClusterRoleBindingAccessor(fake),
			existingClusterRoleBindings: &authorizationapi.ClusterRoleBindingList{
				Items: []authorizationapi.ClusterRoleBinding{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edit",
					},
					Subjects: []kapi.ObjectReference{{
						Name: "foo",
						Kind: authorizationapi.UserKind,
					}},
					RoleRef: kapi.ObjectReference{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					}}, {
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
					Subjects: []kapi.ObjectReference{{
						Name: "bar",
						Kind: authorizationapi.UserKind,
					}},
					RoleRef: kapi.ObjectReference{
						Name:      "edit",
						Namespace: metav1.NamespaceDefault,
					}},
				},
			},
			clusterTest: true,
		},
	}

	for tcName, tc := range tests {
		var binding *authorizationapi.RoleBinding
		var clusterBinding *authorizationapi.ClusterRoleBinding

		// Set up fakeclient actions for rolebindings or clusterrolebindings
		if tc.clusterTest {
			fake.PrependReactor("list", "clusterrolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tc.existingClusterRoleBindings, nil
			})

			fake.PrependReactor("get", "clusterrolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				for i, rb := range tc.existingClusterRoleBindings.Items {
					if tc.rolebindingname == rb.Name {
						return true, &tc.existingClusterRoleBindings.Items[i], nil
					}
				}
				return true, nil, kapierrors.NewNotFound(authorizationapi.Resource("clusterrolebinding"), tc.rolebindingname)
			})

			fake.PrependReactor("update", "clusterrolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				clusterBinding = action.(clientgotesting.UpdateAction).GetObject().(*authorizationapi.ClusterRoleBinding)
				return true, clusterBinding, nil
			})
		} else {
			fake.PrependReactor("list", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tc.existingRoleBindings, nil
			})
			fake.PrependReactor("get", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				for i, rb := range tc.existingRoleBindings.Items {
					if tc.rolebindingname == rb.Name {
						return true, &tc.existingRoleBindings.Items[i], nil
					}
				}
				return true, nil, kapierrors.NewNotFound(authorizationapi.Resource("rolebinding"), tc.rolebindingname)
			})
			fake.PrependReactor("update", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				binding = action.(clientgotesting.UpdateAction).GetObject().(*authorizationapi.RoleBinding)
				return true, binding, nil
			})

			fake.PrependReactor("create", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				binding = action.(clientgotesting.CreateAction).GetObject().(*authorizationapi.RoleBinding)
				return true, binding, nil
			})
		}

		// Set up modifier options and run AddRole()
		o := &policy.RoleModificationOptions{
			RoleName:            tc.role,
			RoleBindingName:     tc.rolebindingname,
			Users:               tc.users,
			RoleNamespace:       tc.rolenamespace,
			RoleBindingAccessor: tc.accessor,
		}

		err := o.AddRole()
		if err != nil {
			t.Errorf("%s: unexpected err %v", tcName, err)
		}

		expectedName := tc.role
		if len(tc.rolebindingname) > 0 {
			expectedName = tc.rolebindingname
		}

		if tc.clusterTest {
			binding = authorizationapi.ToRoleBinding(clusterBinding)
		}

		// check that the desired rolebinding was updated
		if binding.Name != expectedName {
			t.Errorf("%s: wrong rolebinding, expected: %v, actual: %v", tcName, expectedName, binding.Name)
		}

		if binding.RoleRef.Name != tc.role {
			t.Errorf("%s: wrong role, expected: %v, actual: %v", tcName, tc.role, binding.RoleRef.Name)
		}

		subs, _ := authorizationapi.StringSubjectsFor(binding.Namespace, binding.Subjects)
		if !reflect.DeepEqual(tc.users, subs) {
			t.Errorf("%s: err expected users: %v, actual: %v", tcName, tc.users, subs)
		}
	}
}
