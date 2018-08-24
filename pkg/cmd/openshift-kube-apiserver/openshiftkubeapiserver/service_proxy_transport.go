package openshiftkubeapiserver

import (
	"net/http"

	listersv1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
)

// serviceCABundleRoundTripper uses its own RoundTripper configured with a CA bundle containing both
// handlerCA and the SSCS CA bundle read from the SSCS CA bundle configMap.
type serviceCABundleRoundTripper struct {
	serverName string
	handlerCA  []byte
	lister     listersv1.ConfigMapLister
}

func (r *serviceCABundleRoundTripper) getServiceCABundle() string {
	configMap, err := r.lister.ConfigMaps("openshift-service-cert-signer").Get("signing-cabundle")
	if err != nil {
		return ""
	}
	bundle, ok := configMap.Data["cabundle.crt"]
	if !ok {
		return ""
	}
	return bundle
}

func (r *serviceCABundleRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	combinedCA := make([]byte, len(r.handlerCA))
	copy(combinedCA, r.handlerCA)
	bundle := r.getServiceCABundle()
	if len(bundle) != 0 {
		combinedCA = append(combinedCA, []byte(bundle)...)
	}
	newRestConfig := &restclient.Config{
		TLSClientConfig: restclient.TLSClientConfig{
			ServerName: r.serverName,
			CAData:     combinedCA,
		},
	}

	rt, err := restclient.TransportFor(newRestConfig)
	if err != nil {
		return nil, err
	}
	return rt.RoundTrip(req)
}
