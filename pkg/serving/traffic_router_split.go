// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serving

import (
	"encoding/json"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

var (
	modelPathSeparator = ":"
	regexp4serviceName = "^[a-z0-9A-Z_-]+$"
)

func RunTrafficRouterSplit(namespace string, args *types.TrafficRouterSplitArgs) (err error) {
	istioClient, err := initIstioClient()
	if err != nil {
		return err
	}
	preprocessObject := types.PreprocesObject{
		ServiceName:     args.ServingName,
		Namespace:       namespace,
		DestinationRule: generateDestinationRule(namespace, args.ServingName, args.VersionWeights),
		VirtualService:  generateVirtualService(namespace, args.ServingName, args.VersionWeights),
	}
	log.Debugf("serviceName: %s", preprocessObject.ServiceName)
	jsonDestinationRule, err := json.Marshal(preprocessObject.DestinationRule)
	log.Debugf("destination rule: %s", jsonDestinationRule)
	jsonVirtualService, err := json.Marshal(preprocessObject.VirtualService)
	log.Debugf("virtual service: %s", jsonVirtualService)
	virtualServiceName := preprocessObject.ServiceName
	log.Debugf("virtualServiceName:%s", virtualServiceName)
	destinationRuleName := preprocessObject.ServiceName
	log.Debugf("destinationRuleName:%s", virtualServiceName)
	err = createOrUpdateDestinationRule(istioClient, preprocessObject, destinationRuleName)
	err = createOrUpdateVirtualService(namespace, istioClient, preprocessObject, virtualServiceName)
	return err
}

func generateDestinationRule(namespace string, serviceName string, versionWeights []types.ServingVersionWeight) types.DestinationRuleCRD {
	destinationRule := types.DestinationRuleCRD{
		Kind:       "DestinationRule",
		APIVersion: "networking.istio.io/v1alpha3",
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceName,
			Namespace:   namespace,
			Labels:      nil,
			Annotations: nil,
		},
		Spec: istiov1alpha3.DestinationRule{
			Host: serviceName,
		},
	}
	for _, vw := range versionWeights {
		labels := map[string]string{}
		labels["servingVersion"] = vw.Version
		destinationRule.Spec.Subsets = append(destinationRule.Spec.Subsets, &istiov1alpha3.Subset{
			Name:   "subset-" + vw.Version,
			Labels: labels,
		})
	}
	return destinationRule
}

func generateVirtualService(namespace string, serviceName string, versionWeights []types.ServingVersionWeight) types.VirtualServiceCRD {
	routes := []*types.DestinationWeight{}
	for _, vw := range versionWeights {
		routes = append(routes, &types.DestinationWeight{
			Destination: &types.Destination{
				Destination: &istiov1alpha3.Destination{
					Subset: "subset-" + vw.Version,
					Host:   serviceName,
				},
			},
			Weight: int32(vw.Weight),
		})
	}
	httpMatchRequests := []*types.HTTPMatchRequest{
		&types.HTTPMatchRequest{
			Uri: &types.StringMatchPrefix{
				Prefix: "/",
			},
		},
	}
	virtualService := types.VirtualServiceCRD{
		Kind:       "VirtualService",
		APIVersion: "networking.istio.io/v1alpha3",
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceName,
			Namespace:   namespace,
			Labels:      nil,
			Annotations: nil,
		},
		Spec: types.VirtualService{
			VirtualService: &istiov1alpha3.VirtualService{
				Hosts: []string{serviceName},
			},
			Http: []*types.HTTPRoute{
				{
					HTTPRoute: &istiov1alpha3.HTTPRoute{
						Rewrite: &istiov1alpha3.HTTPRewrite{
							Uri:       "/",
							Authority: "",
						},
					},
					Match: httpMatchRequests,
					Route: routes,
				},
			},
		},
	}
	return virtualService
}

func createOrUpdateDestinationRule(istioClient *rest.RESTClient, preprocessObject types.PreprocesObject, destinationRuleName string) (err error) {
	request := istioClient.Get().Namespace(preprocessObject.Namespace).Resource("destinationrules").Name(destinationRuleName)
	request.SetHeader("Accept", "application/json")
	request.SetHeader("Content-Type", "application/json")
	log.Debugf("request URL: %s", request.URL())
	result2, err := request.Do().Raw()
	if err != nil {
		log.Debugf("will create new destinationrule \"%s\"", destinationRuleName)
		convertedjson, err := json.Marshal(preprocessObject.DestinationRule)
		if err != nil {
			return err
		}
		log.Debugf("create destinationrule: %s", (convertedjson))
		postRequest := istioClient.Post().Namespace(preprocessObject.Namespace).Resource("destinationrules")
		log.Debugf("postRequest URL: %s", postRequest.URL())
		newbody, err := postRequest.Body(convertedjson).Do().Raw()
		if err != nil {
			return err
		}
		log.Debugf(string(newbody))
		return nil
	}
	log.Debugf("original destinationrule: %s", result2)
	var originalDestinationRule types.DestinationRuleCRD
	err = json.Unmarshal(result2, &originalDestinationRule)
	if err != nil {
		return err
	}
	originalDestinationRule.Spec = preprocessObject.DestinationRule.Spec
	updatedjson, err := json.Marshal(originalDestinationRule)
	if err != nil {
		return err
	}
	log.Debugf("updated destinationrule: %s", updatedjson)
	updateRequest := istioClient.Put().Namespace(preprocessObject.Namespace).Resource("destinationrules").Name(destinationRuleName)
	log.Debugf("updateRequest URL: %s", updateRequest.URL())
	newbody, err := updateRequest.Body(updatedjson).Do().Raw()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debugf(string(newbody))
	return nil
}

func createOrUpdateVirtualService(namespace string, istioClient *rest.RESTClient, preprocessObject types.PreprocesObject, virtualServiceName string) (err error) {
	request := istioClient.Get().Namespace(preprocessObject.Namespace).Resource("virtualservices").Name(virtualServiceName)
	request.SetHeader("Accept", "application/json")
	request.SetHeader("Content-Type", "application/json")
	log.Debugf("request URL: %s", request.URL())
	result2, err := request.Do().Raw()
	if err != nil {
		log.Debugf("will create new virtualservice \"%s\"", virtualServiceName)
		convertedjson, err := json.Marshal(preprocessObject.VirtualService)
		if err != nil {
			return err
		}
		log.Debugf("create virtualservice: %s", (convertedjson))
		newbody, err := istioClient.Post().Namespace(namespace).Resource("virtualservices").Body(convertedjson).Do().Raw()
		if err != nil {
			return err
		}
		log.Debugf(string(newbody))
		return nil
	}
	log.Debugf("original virtualservice: %s", result2)
	var originalVirtualService types.VirtualServiceCRD
	err = json.Unmarshal(result2, &originalVirtualService)
	if err != nil {
		return err
	}
	originalVirtualService.Spec = preprocessObject.VirtualService.Spec
	updatedjson, err := json.Marshal(originalVirtualService)
	if err != nil {
		return err
	}
	log.Debugf("updated virtualservice: %s", updatedjson)
	newbody, err := istioClient.Put().Namespace(namespace).Resource("virtualservices").Name(virtualServiceName).Body(updatedjson).Do().Raw()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debugf(string(newbody))
	return nil
}

func initIstioClient() (*rest.RESTClient, error) {
	restConfig := config.GetArenaConfiger().GetRestConfig()
	istioAPIGroupVersion := schema.GroupVersion{
		Group:   "networking.istio.io",
		Version: "v1alpha3",
	}
	restConfig.GroupVersion = &istioAPIGroupVersion
	restConfig.APIPath = "/apis"
	restConfig.ContentType = runtime.ContentTypeJSON
	types := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			metav1.AddToGroupVersion(scheme, istioAPIGroupVersion)
			return nil
		})
	err := schemeBuilder.AddToScheme(types)
	if err != nil {
		return nil, err
	}
	restConfig.NegotiatedSerializer = serializer.NewCodecFactory(types)

	// create the clientset

	return rest.RESTClientFor(restConfig)
}
