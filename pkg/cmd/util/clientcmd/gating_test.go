package clientcmd

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fake "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
)

func TestResourceGate(t *testing.T) {
	legacyTests := map[string]struct {
		//existingAPIResources []*metav1.APIResourceList
		//inputGVR             []runtime.Object
		existingAPIResources []runtime.Object
		expectedErr          string
	}{
		"legacy-policy-all-supported": {
			existingAPIResources: []runtime.Object{
				&metav1.APIResourceList{
					GroupVersion: "a/v1",
					APIResources: []metav1.APIResource{
						{
							Name: "clusterpolicies",
						},
						{
							Name: "policybindings",
						},
						{
							Name: "policies",
						},
						{
							Name: "clusterpolicybindings",
						},
						{
							Name: "foo1",
						},
						{
							Name: "foo2",
						},
					},
				},
				&metav1.APIResourceList{
					GroupVersion: "b/v1",
					APIResources: []metav1.APIResource{
						{
							Name: "bar1",
						},
						{
							Name: "bar2",
						},
					},
				},
			},
		},
		"legacy-policy-none-supported": {
			existingAPIResources: []runtime.Object{
				&metav1.APIResourceList{
					GroupVersion: "a/v1",
					APIResources: []metav1.APIResource{
						{
							Name: "foo1",
						},
						{
							Name: "foo2",
						},
					},
				},
				&metav1.APIResourceList{
					GroupVersion: "b/v1",
					APIResources: []metav1.APIResource{
						{
							Name: "bar1",
						},
						{
							Name: "bar2",
						},
					},
				},
			},
			expectedErr: "the server does not support legacy policy resources",
		},
	}
	discoverTests := map[string]struct {
		existingAPIResources []runtime.Object
		inputGVR             []schema.GroupVersionResource
		expectedGVR          []schema.GroupVersionResource
	}{
		"regular": {
			existingAPIResources: []runtime.Object{
				&metav1.APIResourceList{
					GroupVersion: "a/v1",
					APIResources: []metav1.APIResource{
						{
							Name: "foo",
						},
					},
				},
				&metav1.APIResourceList{
					GroupVersion: "b/v1",
					APIResources: []metav1.APIResource{
						{
							Name: "bar",
						},
					},
				},
			},
			inputGVR: []schema.GroupVersionResource{
				{Group: "a", Version: "v1", Resource: "foo"},
			},
			expectedGVR: []schema.GroupVersionResource{
				{Group: "a", Version: "v1", Resource: "foo"},
			},
		},
	}

	for tcName, tc := range legacyTests {
		err := LegacyPolicyResourceGate(fake.NewSimpleClientset(tc.existingAPIResources...))
		if err != nil && tc.expectedErr != err.Error() {
			t.Fatalf("%s: expected err %s, got %s", tcName, tc.expectedErr, err.Error())
		}
	}

	for tcName, tc := range discoverTests {
		result, err := DiscoverGroupVersionResources(fake.NewSimpleClientset(tc.existingAPIResources...), tc.inputGVR)
		if err != nil {
			t.Fatalf("%s: unexpected err %s", tcName, err.Error())
		}
		if !reflect.DeepEqual(result, tc.expectedGVR) {
			t.Fatalf("%s: expected %v, got %v", tcName, tc.expectedGVR, result)
		}
	}
}

/* Example stuff from discovery_client_test.go

import (
	"encoding/json"
	"reflect"
	"testing"
	"net/http"
	"net/http/httptest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	//fake "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/apimachinery/pkg/util/sets"
	//"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

func TestGetServerResources(t *testing.T) {
	stable := metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{Name: "pods", Namespaced: true, Kind: "Pod"},
			{Name: "services", Namespaced: true, Kind: "Service"},
			{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
		},
	}
	beta := metav1.APIResourceList{
		GroupVersion: "extensions/v1beta1",
		APIResources: []metav1.APIResource{
			{Name: "deployments", Namespaced: true, Kind: "Deployment"},
			{Name: "ingresses", Namespaced: true, Kind: "Ingress"},
			{Name: "jobs", Namespaced: true, Kind: "Job"},
		},
	}
	tests := []struct {
		resourcesList *metav1.APIResourceList
		path          string
		request       string
		expectErr     bool
	}{
		{
			resourcesList: &stable,
			path:          "/api/v1",
			request:       "v1",
			expectErr:     false,
		},
		{
			resourcesList: &beta,
			path:          "/apis/extensions/v1beta1",
			request:       "extensions/v1beta1",
			expectErr:     false,
		},
		{
			resourcesList: &stable,
			path:          "/api/v1",
			request:       "foobar",
			expectErr:     true,
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var list interface{}
		switch req.URL.Path {
		case "/api/v1":
			list = &stable
		case "/apis/extensions/v1beta1":
			list = &beta
		case "/api":
			list = &metav1.APIVersions{
				Versions: []string{
					"v1",
				},
			}
		case "/apis":
			list = &metav1.APIGroupList{
				Groups: []metav1.APIGroup{
					{
						Versions: []metav1.GroupVersionForDiscovery{
							{GroupVersion: "extensions/v1beta1"},
						},
					},
				},
			}
		default:
			t.Logf("unexpected request: %s", req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		output, err := json.Marshal(list)
		if err != nil {
			t.Errorf("unexpected encoding error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}))
	defer server.Close()
	client := internalclientset.NewForConfigOrDie(&restclient.Config{Host: server.URL})
	//client := discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})

	got, err := DiscoverGroupVersionResources(client, []schema.GroupVersionResource{{Version: "v1", Resource: "pods"}})
	if err != nil {
		t.Fatalf("myerr %s", err.Error())
	}
	t.Errorf("%v", got)
	for _, test := range tests {
		got, err := client.ServerResourcesForGroupVersion(test.request)
		if test.expectErr {
			if err == nil {
				t.Error("unexpected non-error")
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if !reflect.DeepEqual(got, test.resourcesList) {
			t.Errorf("expected:\n%v\ngot:\n%v\n", test.resourcesList, got)
		}
	}

	serverResources, err := client.ServerResources()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	serverGroupVersions := sets.NewString(groupVersions(serverResources)...)
	for _, api := range []string{"v1", "extensions/v1beta1"} {
		if !serverGroupVersions.Has(api) {
			t.Errorf("missing expected api %q in %v", api, serverResources)
		}
	}
}

func groupVersions(resources []*metav1.APIResourceList) []string {
	result := []string{}
	for _, resourceList := range resources {
		result = append(result, resourceList.GroupVersion)
	}
	return result
}
*/
