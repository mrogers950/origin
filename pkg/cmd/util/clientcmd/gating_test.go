package clientcmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	//fake "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	//"k8s.io/client-go/discovery"
	"github.com/openshift/origin/pkg/authorization/apis/authorization"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	//"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

func TestGetServerResources(t *testing.T) {
	legacyAPIResourcesA := []metav1.APIResource{
		{Name: "clusterpolicies", Kind: "ClusterPolicies"},
		{Name: "clusterpolicybindings", Kind: "ClusterPolicyBindings"},
		{Name: "policies", Kind: "Policies"},
		{Name: "policybindings", Kind: "PolicyBindings"},
		{Name: "foo", Kind: "Foo"},
	}
	legacyAPIResourcesB := []metav1.APIResource{
		{Name: "clusterpolicies", Kind: "ClusterPolicies"},
		{Name: "clusterpoliybindings", Kind: "ClusterPolicyBindings"},
		{Name: "foo", Kind: "Foo"},
	}
	legacyAPIResourcesC := []metav1.APIResource{
		{Name: "policies", Kind: "Policies"},
		{Name: "policybindings", Kind: "PolicyBindings"},
		{Name: "bar", Kind: "Bar"},
	}
	/*
		legacyAPIResourcesD := []metav1.APIResource{
			{Name: "foo", Kind: "Foo"},
			{Name: "bar", Kind: "Bar"},
			{Name: "baz", Kind: "Baz"},
		}
	*/
	discoverListA := metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: legacyAPIResourcesA,
	}
	resourceListA := metav1.APIResourceList{
		GroupVersion: authorization.LegacySchemeGroupVersion.String(),
		APIResources: legacyAPIResourcesA,
	}
	resourceListB := metav1.APIResourceList{
		GroupVersion: authorization.LegacySchemeGroupVersion.String(),
		APIResources: legacyAPIResourcesB,
	}
	resourceListC := metav1.APIResourceList{
		GroupVersion: authorization.LegacySchemeGroupVersion.String(),
		APIResources: legacyAPIResourcesC,
	}
	/*
		resourceListD := metav1.APIResourceList{
			GroupVersion: authorization.LegacySchemeGroupVersion.String(),
			APIResources: legacyAPIResourcesD,
		}
		resourceListE := metav1.APIResourceList{
			GroupVersion: authorization.SchemeGroupVersion.String(),
			APIResources: legacyAPIResourcesB,
		}
		resourceListF := metav1.APIResourceList{
			GroupVersion: authorization.SchemeGroupVersion.String(),
			APIResources: legacyAPIResourcesC,
		}
	*/
	discoverTests := map[string]struct {
		resourcesList *metav1.APIResourceList
		inputGVR      []schema.GroupVersionResource
		expectedGVR   []schema.GroupVersionResource
		path          string
		request       string
		expectErr     bool
	}{
		"discover": {
			resourcesList: &resourceListA,
			inputGVR: []schema.GroupVersionResource{
				{
					Group:    "",
					Version:  "v1",
					Resource: "foo",
				},
				{
					Group:    "",
					Version:  "v1",
					Resource: "noexist",
				},
			},
			expectedGVR: []schema.GroupVersionResource{
				{
					Group:    "",
					Version:  "v1",
					Resource: "foo",
				},
			},
			path:      "/api/v1",
			request:   "v1",
			expectErr: false,
		},
	}
	/*
		discoverTests := map[string]struct {
			resourcesList *metav1.APIResourceList
			inputGVR      []schema.GroupVersionResource
			path          string
			request       string
			expectErr     bool
		}{
			"disc-one": {
				resourcesList: &resourceListA,
				path:          "/api/" + authorization.LegacySchemeGroupVersion.String(),
				request:       authorization.LegacySchemeGroupVersion.String(),
				expectErr:     false,
			},
			"disc-two": {
				resourcesList: &resourceListB,
				path:          "/api/discover/v1",
				request:       "discover/v1",
				expectErr:     false,
			},
		}
	*/
	/*
		legacyTests := map[string]struct {
			resourcesList *metav1.APIResourceList
			inputGVR      []schema.GroupVersionResource
			path          string
			request       string
			expectErr     bool
		}{
			"legacy-one": {
				resourcesList: &hasLegacyResources,
				path:          "/api/v1",
				request:       "v1",
				expectErr:     false,
			},
			"legacy-two": {
				resourcesList: &testDiscoveryResourcesV1,
				path:          "/api/discover/v1",
				request:       "discover/v1",
				expectErr:     false,
			},
		}
	*/
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var list interface{}
		switch req.URL.Path {
		case "/api/v1":
			list = &discoverListA
		case "/api/policy/A":
			list = &resourceListA
		case "/api/policy/B":
			list = &resourceListB
		case "/api/policy/C":
			list = &resourceListC
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
	client := discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})

	/*
		for _, tc := range legacyTests {
			err := LegacyPolicyResourceGate(client)
			if err != nil {
				t.Fatalf("legacy err")
			}
		}
	*/
	for tcName, tc := range discoverTests {
		got, err := DiscoverGroupVersionResources(client, tc.inputGVR...)
		if err != nil {
			t.Fatalf("myerr %s", err.Error())
		}
		if !reflect.DeepEqual(got, tc.expectedGVR) {
			t.Errorf("%s got %v, expected %v", tcName, got, tc.inputGVR)
		}
	}

}
