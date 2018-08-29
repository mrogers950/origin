package openshiftkubeapiserver

import (
	"net/http"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	listersv1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	cache "k8s.io/client-go/tools/cache"

	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

// serviceCABundleRoundTripper creates a new RoundTripper for each request with serverName and caBundle input into the
// TLSClientConfig. It is expected that caBundle is kept up to date by the serviceCABundleUpdater controller.
type serviceCABundleRoundTripper struct {
	serverName string
	caBundle   []byte
}

func (r *serviceCABundleRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newRestConfig := &restclient.Config{
		TLSClientConfig: restclient.TLSClientConfig{
			ServerName: r.serverName,
			CAData:     r.caBundle,
		},
	}

	rt, err := restclient.TransportFor(newRestConfig)
	if err != nil {
		return nil, err
	}
	return rt.RoundTrip(req)
}

const (
	caBundleDataKey              = "cabundle.crt"
	serviceCABundleNamespace     = "openshift-service-cert-signer"
	serviceCABundleConfigMapName = "signing-cabundle"
)

// serviceCABundleUpdater runs a simple controller to keep rt.caBundle updated with CAs from the service-ca controller.
type serviceCABundleUpdater struct {
	// Initial CA bundle that CA updates are tacked on to.
	startingHandlerCA []byte
	// RoundTripper that utilizes the updated CA bundle.
	rt *serviceCABundleRoundTripper

	lister      listersv1.ConfigMapLister
	queue       workqueue.RateLimitingInterface
	hasSynced   cache.InformerSynced
	syncHandler func(serviceKey string) error
}

func isServiceCABundleConfigMap(configMap *v1.ConfigMap) bool {
	return configMap.Namespace == serviceCABundleNamespace && configMap.Name == serviceCABundleConfigMapName
}

// addCABundle is the informer's AddFunc.
func (u *serviceCABundleUpdater) addCABundle(obj interface{}) {
	cm := obj.(*v1.ConfigMap)
	if !isServiceCABundleConfigMap(cm) {
		glog.Infof("serviceCABundleUpdater controllers: addCABundle not the configmap we want")
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cm)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", cm, err)
		return
	}

	glog.Infof("serviceCABundleUpdater controller: queuing an add of %v", key)
	u.queue.Add(key)
}

// updateCABundle is the informer's UpdateFunc.
func (u *serviceCABundleUpdater) updateCABundle(old, cur interface{}) {
	cm := cur.(*v1.ConfigMap)
	if !isServiceCABundleConfigMap(cm) {
		glog.Infof("serviceCABundleUpdater controllers: updateCABundle not the configmap we want")
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cm)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", cm, err)
		return
	}

	glog.Infof("serviceCABundleUpdater controller: queuing an update of %v", key)
	u.queue.Add(key)
}

// processNextWorkItem processes the queued items.
func (u *serviceCABundleUpdater) processNextWorkItem() bool {
	key, quit := u.queue.Get()
	if quit {
		return false
	}
	defer u.queue.Done(key)

	err := u.syncHandler(key.(string))
	if err == nil {
		u.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", key, err))
	u.queue.AddRateLimited(key)

	return true
}

// Run runs the controller until stopCh is closed.
func (u *serviceCABundleUpdater) Run(stopCh <-chan struct{}) {
	panic(fmt.Errorf("serviceCABundleUpdater test panic"))
	defer utilruntime.HandleCrash()
	defer u.queue.ShutDown()
	glog.Infof("serviceCABundleUpdater controller: Run() waiting for cache sync")

	if !cache.WaitForCacheSync(stopCh, u.hasSynced) {
		return
	}
	glog.Infof("serviceCABundleUpdater controller: Run() done with cache sync")

	glog.Infof("starting serviceCABundleUpdater controller")
	go wait.Until(u.runWorker, time.Second, stopCh)
	<-stopCh
	glog.Infof("stopping serviceCABundleUpdater controller")
}

// runWorker repeatedly calls processNextWorkItem until the latter wants to exit.
func (u *serviceCABundleUpdater) runWorker() {
	for u.processNextWorkItem() {
	}
}

// syncCABundle will update the RoundTripper's CA bundle by combining the starting CA with the updated CA from the
// service CA configMap.
func (u *serviceCABundleUpdater) syncCABundle(key string) error {
	glog.Infof("serviceCABundleUpdater controller: syncCABundle")

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	glog.Infof("serviceCABundleUpdater controller: syncCABundle getting NS:%v NAME:%v", namespace, name)

	sharedConfigMap, err := u.lister.ConfigMaps(namespace).Get(name)
	if kapierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	glog.Infof("serviceCABundleUpdater controller: syncCABundle got NS:%v NAME:%v", namespace, name)

	data, ok := sharedConfigMap.Data[caBundleDataKey]
	if !ok {
		glog.Infof("serviceCABundleUpdater controller: syncCABundle didnt get data")
		return nil
	}

	combinedCA := make([]byte, len(u.startingHandlerCA))
	copy(combinedCA, u.startingHandlerCA)
	combinedCA = append(combinedCA, data...)

	u.rt.caBundle = combinedCA
	glog.Infof("serviceCABundleUpdater controller: syncCABundle updated caBundle")
	return nil
}

// NewServiceCABundleUpdater creates a new serviceCABundleUpdater controller.
func NewServiceCABundleUpdater(kubeInformers informers.SharedInformerFactory, serverName string, caBundle []byte) (*serviceCABundleUpdater, error) {
	roundTripper := &serviceCABundleRoundTripper{
		serverName: serverName,
		caBundle:   caBundle,
	}

	updater := &serviceCABundleUpdater{
		rt:                roundTripper,
		lister:            kubeInformers.Core().V1().ConfigMaps().Lister(),
		startingHandlerCA: caBundle,
		queue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	kubeInformers.Core().V1().ConfigMaps().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    updater.addCABundle,
			UpdateFunc: updater.updateCABundle,
		},
	)

	updater.hasSynced = kubeInformers.Core().V1().ConfigMaps().Informer().HasSynced
	updater.syncHandler = updater.syncCABundle
	return updater, nil
}
