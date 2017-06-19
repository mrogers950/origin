package policy

import (
	"reflect"
	"testing"

	"github.com/openshift/origin/pkg/client/testclient"
	"k8s.io/apimachinery/pkg/runtime"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgotesting "k8s.io/client-go/testing"
	kapi "k8s.io/kubernetes/pkg/api"
)

func TestModifyRoleBindingName(t *testing.T) {
	var fake *testclient.Fake
	fake = testclient.NewSimpleFake()

	tests := map[string]struct {
		rolename                  string
		rolenamespace             string
		rolebindingname           string
		users                     []string
		accessor                  RoleBindingAccessor
		startPolicyBinding        *authorizationapi.PolicyBinding
		startClusterPolicyBinding *authorizationapi.ClusterPolicyBinding
		cluster                   bool
	}{
		"create-default-binding": {
			rolename:        "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "",
			users: []string{
				"foo",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "policybinding",
				},
				PolicyRef: kapi.ObjectReference{
					Namespace: "namespace",
				},
				RoleBindings: map[string]*authorizationapi.RoleBinding{},
			},
			startClusterPolicyBinding: nil,
			cluster:                   false,
		},
		"create-named-binding": {
			rolename:        "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "rb1",
			users: []string{
				"foo",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "policybinding",
				},
				PolicyRef: kapi.ObjectReference{
					Namespace: "namespace",
				},
				RoleBindings: map[string]*authorizationapi.RoleBinding{},
			},
			startClusterPolicyBinding: nil,
			cluster:                   false,
		},
		"update-default-binding": {
			rolename:        "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "",
			users: []string{
				"foo",
				"bar",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "policybinding",
				},
				PolicyRef: kapi.ObjectReference{
					Namespace: "namespace",
				},
				RoleBindings: map[string]*authorizationapi.RoleBinding{
					"edit": {
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
						},
					},
					"custom": {
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
						},
					},
				},
			},
			startClusterPolicyBinding: nil,
			cluster:                   false,
		},
		"update-named-binding": {
			rolename:        "edit",
			rolenamespace:   metav1.NamespaceDefault,
			rolebindingname: "custom",
			users: []string{
				"bar",
				"baz",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "policybinding",
				},
				PolicyRef: kapi.ObjectReference{
					Namespace: "namespace",
				},
				RoleBindings: map[string]*authorizationapi.RoleBinding{
					"edit": {
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
						},
					},
					"custom": {
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
						},
					},
				},
			},
			startClusterPolicyBinding: nil,
			cluster:                   false,
		},
		"update-named-clusterbinding": {
			rolename:        "edit",
			rolenamespace:   "",
			rolebindingname: "custom",
			users: []string{
				"bar",
				"baz",
			},
			accessor:           NewClusterRoleBindingAccessor(fake),
			startPolicyBinding: nil,
			startClusterPolicyBinding: &authorizationapi.ClusterPolicyBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "clusterpolicybinding",
				},
				PolicyRef: kapi.ObjectReference{
					Namespace: "namespace",
				},
				RoleBindings: map[string]*authorizationapi.ClusterRoleBinding{
					"edit": {
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
						},
					},
					"custom": {
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
						},
					},
				},
			},
			cluster: true,
		},
	}

	for tcName, tc := range tests {
		pbList := &authorizationapi.PolicyBindingList{}
		cpbList := &authorizationapi.ClusterPolicyBindingList{}
		if tc.cluster {
			fake.PrependReactor("get", "clusterpolicybindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tc.startClusterPolicyBinding, nil
			})

			fake.PrependReactor("list", "clusterpolicybindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				cpbList.Items = []authorizationapi.ClusterPolicyBinding{*tc.startClusterPolicyBinding}
				return true, cpbList, nil
			})
		} else {
			fake.PrependReactor("get", "policybindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tc.startPolicyBinding, nil
			})

			fake.PrependReactor("list", "policybindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				pbList.Items = []authorizationapi.PolicyBinding{*tc.startPolicyBinding}
				return true, pbList, nil
			})
		}

		var actualRoleBinding *authorizationapi.RoleBinding
		var clusterRoleBinding *authorizationapi.ClusterRoleBinding
		if tc.cluster {
			fake.PrependReactor("get", "clusterrolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				if len(tc.rolebindingname) > 0 && tc.startClusterPolicyBinding.RoleBindings[tc.rolebindingname] != nil {
					return true, tc.startClusterPolicyBinding.RoleBindings[tc.rolebindingname], nil
				}
				return true, nil, kapierrors.NewNotFound(authorizationapi.Resource("clusterrolebinding"), tc.rolebindingname)
			})

			fake.PrependReactor("update", "clusterrolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				clusterRoleBinding = action.(clientgotesting.UpdateAction).GetObject().(*authorizationapi.ClusterRoleBinding)
				return true, clusterRoleBinding, nil
			})
		}

		fake.PrependReactor("get", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			if len(tc.rolebindingname) > 0 && tc.startPolicyBinding.RoleBindings[tc.rolebindingname] != nil {
				return true, tc.startPolicyBinding.RoleBindings[tc.rolebindingname], nil
			}
			return true, nil, kapierrors.NewNotFound(authorizationapi.Resource("rolebinding"), tc.rolebindingname)
		})

		fake.PrependReactor("update", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			actualRoleBinding = action.(clientgotesting.UpdateAction).GetObject().(*authorizationapi.RoleBinding)
			return true, actualRoleBinding, nil
		})

		fake.PrependReactor("create", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			actualRoleBinding = action.(clientgotesting.CreateAction).GetObject().(*authorizationapi.RoleBinding)
			return true, actualRoleBinding, nil
		})

		o := &RoleModificationOptions{
			RoleName:            tc.rolename,
			RoleBindingName:     tc.rolebindingname,
			Users:               tc.users,
			RoleNamespace:       tc.rolenamespace,
			RoleBindingAccessor: tc.accessor,
		}

		var err error

		err = o.AddRole()
		if err != nil {
			t.Errorf("%s: unexpected err %v", tcName, err)
		}

		if len(tc.rolebindingname) < 1 {
			// Check default case for rolebindingname which is the role name.
			tc.rolebindingname = tc.rolename
		}

		if tc.cluster {
			actualRoleBinding = authorizationapi.ToRoleBinding(clusterRoleBinding)
		}

		if tc.rolebindingname != actualRoleBinding.Name {
			t.Errorf("%s: wrong rolebinding, expected: %v, actual: %v", tcName, tc.rolebindingname, actualRoleBinding.Name)
		}

		if tc.rolename != actualRoleBinding.RoleRef.Name {
			t.Errorf("%s: wrong role, expected: %v, actual: %v", tcName, tc.rolename, actualRoleBinding.RoleRef.Name)
		}

		subs, _ := authorizationapi.StringSubjectsFor(actualRoleBinding.Namespace, actualRoleBinding.Subjects)
		if !reflect.DeepEqual(tc.users, subs) {
			t.Errorf("%s: err expected users: %v, actual: %v", tcName, tc.users, subs)
		}
	}
}
