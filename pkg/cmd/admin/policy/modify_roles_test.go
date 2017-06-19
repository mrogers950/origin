package policy

import (
	"testing"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"github.com/openshift/origin/pkg/client/testclient"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	clientgotesting "k8s.io/client-go/testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api"
)

func TestModifyRoleBindingName(t *testing.T) {
	var fake *testclient.Fake
	fake = testclient.NewSimpleFake()

	tests := map[string]struct {
		rolename        string
		rolebindingname string
		users           []string
		accessor LocalRoleBindingAccessor
		startPolicyBinding *authorizationapi.PolicyBinding
	}{
		"create-default-binding": {
			rolename: "edit",
			rolebindingname: "",
			users: []string{
				"foo",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding {
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name: "policybinding",
				},
				PolicyRef: kapi.ObjectReference{
					Namespace: "namespace",
				},
				RoleBindings: map[string]*authorizationapi.RoleBinding{},
			},
		},
		"create-named-binding": {
			rolename: "edit",
			rolebindingname: "rb1",
			users: []string{
				"foo",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding {
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name: "policybinding",
				},
				PolicyRef: kapi.ObjectReference{
					Namespace: "namespace",
				},
				RoleBindings: map[string]*authorizationapi.RoleBinding{},
			},
		},
		"update-default-binding": {
			rolename: "edit",
			rolebindingname: "",
			users: []string{
				"foo",
				"bar",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding {
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name: "policybinding",
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
		},
		"update-named-binding": {
			rolename: "edit",
			rolebindingname: "rb1",
			users: []string{
				"bar",
				"baz",
			},
			accessor: NewLocalRoleBindingAccessor(metav1.NamespaceDefault, fake),
			startPolicyBinding: &authorizationapi.PolicyBinding {
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name: "policybinding",
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
		},

	}

	for tcName, tc := range tests {

		fake.PrependReactor("get", "policybindings", func(action clientgotesting.Action)(handled bool, ret runtime.Object, err error) {
			return true, tc.startPolicyBinding, nil
		})

		pbList := &authorizationapi.PolicyBindingList{}
		fake.PrependReactor("list", "policybindings", func(action clientgotesting.Action)(handled bool, ret runtime.Object, err error) {
			pbList.Items = []authorizationapi.PolicyBinding{*tc.startPolicyBinding}
			return true, pbList, nil
		})

		var actualRoleBinding *authorizationapi.RoleBinding
		fake.PrependReactor("update","rolebindings", func(action clientgotesting.Action)(handled bool, ret runtime.Object, err error) {
			actualRoleBinding = action.(clientgotesting.UpdateAction).GetObject().(*authorizationapi.RoleBinding)
			return true, actualRoleBinding, nil
		})
		fake.PrependReactor("create","rolebindings", func(action clientgotesting.Action)(handled bool, ret runtime.Object, err error) {
			actualRoleBinding = action.(clientgotesting.CreateAction).GetObject().(*authorizationapi.RoleBinding)
			return true, actualRoleBinding, nil
		})

		o := &RoleModificationOptions{
			RoleName:            tc.rolename,
			RoleBindingName:     tc.rolebindingname,
			Users:               tc.users,
			RoleNamespace:       "default",
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
