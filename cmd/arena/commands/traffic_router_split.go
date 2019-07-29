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

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewTrafficRouterSplitCommand() *cobra.Command {
	var (
		submitArgs runTrafficRouterSplitArgs
	)

	var command = &cobra.Command{
		Use:     "traffic-split",
		Short:   "Adjust traffic routing dynamically for tfserving jobs",
		Aliases: []string{"trs", "traffic-router", "traffic-router-split", "traffic-shift", "traffic-shifting"},
		Run: func(cmd *cobra.Command, args []string) {
			/*if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}*/

			util.SetLogLevel(logLevel)
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			istioClient, errIstio := initIstioClient()
			if errIstio != nil {
				fmt.Println(errIstio)
				os.Exit(1)
			}
			err = ensureNamespace(client, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = runTrafficRouterSplit(client, istioClient, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	// Traffic Routing Args
	command.Flags().StringVar(&submitArgs.ServingName, "servingName", "", "the serving name")
	command.Flags().StringVar(&submitArgs.Namespace, "namespace", "", "namespace (default \"default\")")
	command.Flags().StringVar(&submitArgs.Versions, "servingVersions", "", "Model versions which the traffic will be routed to, e.g. 1,2,3")
	command.Flags().StringVar(&submitArgs.Weights, "weights", "", "Weight percentage values for each model version which the traffic will be routed to,e.g. 70,20,10")
	command.MarkFlagRequired("servingName")
	command.MarkFlagRequired("servingVersions")
	command.MarkFlagRequired("weights")

	return command
}

type runTrafficRouterSplitArgs struct {
	ServingName string `yaml:"servingName,omitempty"` //--servingName
	Namespace   string `yaml:"namespace,omitempty"`   //--namespace
	Versions    string `yaml:"versions,omitempty"`    //--versions
	Weights     string `yaml:"weights,omitempty"`     //--weights
}

type PreprocesObject struct {
	ServiceName     string
	Namespace       string
	DestinationRule DestinationRuleCRD
	VirtualService  VirtualServiceCRD
}

type DestinationRuleCRD struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
	// +optional
	APIVersion        string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              istiov1alpha3.DestinationRule `json:"spec,omitempty" yaml:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type VirtualServiceCRD struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
	// +optional
	APIVersion        string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              VirtualService `json:"spec,omitempty" yaml:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type VirtualService struct {
	*istiov1alpha3.VirtualService
	Http []*HTTPRoute `protobuf:"bytes,3,rep,name=http" json:"http,omitempty"`
}

type HTTPRoute struct {
	*istiov1alpha3.HTTPRoute
	Match []*HTTPMatchRequest  `protobuf:"bytes,1,rep,name=match" json:"match,omitempty"`
	Route []*DestinationWeight `protobuf:"bytes,2,rep,name=route" json:"route,omitempty"`
}

type HTTPMatchRequest struct {
	*istiov1alpha3.HTTPMatchRequest
	Uri *StringMatchPrefix `protobuf:"bytes,1,opt,name=uri" json:"uri,omitempty"`
}

type StringMatchPrefix struct {
	Prefix string `protobuf:"bytes,2,opt,name=prefix,proto3,oneof" json:"prefix,omitempty"`
}

type DestinationWeight struct {
	Destination *Destination `protobuf:"bytes,1,opt,name=destination" json:"destination,omitempty"`
	Weight      int32        `protobuf:"varint,2,opt,name=weight,proto3" json:"weight"`
}

type Destination struct {
	*istiov1alpha3.Destination
	Port *PortSelector `protobuf:"bytes,3,opt,name=port" json:"port,omitempty"`
}

type PortSelector struct {
	*istiov1alpha3.PortSelector
	Number uint32 `protobuf:"varint,1,opt,name=number,proto3,oneof" json:"number,omitempty"`
}

func generateDestinationRule(namespace string, serviceName string, versionArray []string) DestinationRuleCRD {
	destinationRule := DestinationRuleCRD{
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

	length := len(versionArray)
	subsets := make([]*istiov1alpha3.Subset, length)

	for i := 0; i < length; i++ {
		label := map[string]string{}
		label["servingVersion"] = versionArray[i]
		subsets[i] = &istiov1alpha3.Subset{
			Name:   "subset-" + versionArray[i],
			Labels: label,
		}

		destinationRule.Spec.Subsets = append(destinationRule.Spec.Subsets, subsets[i])
	}
	return destinationRule
}

func generateVirtualService(namespace string, serviceName string, versionArray []string, iweightArray []int32) VirtualServiceCRD {
	length := len(versionArray)
	routes := make([]*DestinationWeight, length)

	for i := 0; i < length; i++ {
		routes[i] = &DestinationWeight{
			Destination: &Destination{
				Destination: &istiov1alpha3.Destination{
					Subset: "subset-" + versionArray[i],
					Host:   serviceName,
				},
			},
			Weight: iweightArray[i],
		}

	}

	httpMatchRequests := make([]*HTTPMatchRequest, 1)
	httpMatchRequests[0] = &HTTPMatchRequest{
		Uri: &StringMatchPrefix{
			Prefix: "/",
		},
	}

	virtualService := VirtualServiceCRD{
		Kind:       "VirtualService",
		APIVersion: "networking.istio.io/v1alpha3",
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceName,
			Namespace:   namespace,
			Labels:      nil,
			Annotations: nil,
		},
		Spec: VirtualService{
			VirtualService: &istiov1alpha3.VirtualService{
				Hosts: []string{serviceName},
			},
			Http: []*HTTPRoute{
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

func (runTrafficRouterSplitArgs *runTrafficRouterSplitArgs) preprocess(client *kubernetes.Clientset) (preprocessObject PreprocesObject, err error) {
	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(runTrafficRouterSplitArgs.ServingName)
	if !matched {
		return preprocessObject, fmt.Errorf("parameter modelName should be numbers, letters, dashes, and underscores ONLY")
	}

	serviceName := strings.Trim(runTrafficRouterSplitArgs.ServingName, " ")
	serviceName = strings.Trim(serviceName, "\"")
	log.Debugf("serviceName: %s", serviceName)
	preprocessObject.ServiceName = serviceName
	if serviceName == "" {
		return preprocessObject, fmt.Errorf("parameter modelName should be specified")
	}
	versions := strings.Trim(runTrafficRouterSplitArgs.Versions, " ")
	versions = strings.Trim(versions, "\"")
	log.Debugf("versions: %s", versions)
	if versions == "" {
		return preprocessObject, fmt.Errorf("versions should be specified")
	}
	weights := strings.Trim(runTrafficRouterSplitArgs.Weights, " ")
	weights = strings.Trim(weights, "\"")
	log.Debugf("weights: %s", weights)
	if weights == "" {
		return preprocessObject, fmt.Errorf("weights should be specified")
	}

	versionArray := strings.Split(versions, ",")
	if len(versionArray) == 0 {
		return preprocessObject, fmt.Errorf("versions should be specified following the format: \"1,2,3\" ")
	}

	weightArray := strings.Split(weights, ",")
	if len(weightArray) == 0 {
		return preprocessObject, fmt.Errorf("weights should be specified following the format: \"60,30,10\" ")
	}

	if len(versionArray) != len(weightArray) {
		return preprocessObject, fmt.Errorf("the number of versions and weights should be equal")
	}

	iweightArray := make([]int32, len(weightArray))
	totalweight := 0
	for i := 0; i < len(weightArray); i++ {
		iweight, err := strconv.Atoi(weightArray[i])
		if err != nil {
			return preprocessObject, fmt.Errorf("weights should be specified following the format and integer type: \"60,30,10\" ")
		}
		iweightArray[i] = int32(iweight)
		totalweight += iweight
	}

	if totalweight != 100 {
		return preprocessObject, fmt.Errorf("configuration is invalid: total weight %d != 100", totalweight)
	}

	if namespace == "" {
		namespace = "default"
	}
	preprocessObject.Namespace = namespace
	destinationRule := generateDestinationRule(namespace, serviceName, versionArray)

	//data, err := json.Marshal(destinationRule)
	//if err != nil {
	//	return para, fmt.Errorf("cannot create destination rule: %s", err)
	//}
	preprocessObject.DestinationRule = destinationRule
	//log.Debugf("destination rule: %s", string(data))

	virtualService := generateVirtualService(namespace, serviceName, versionArray, iweightArray)
	//data, err = json.Marshal(virtualService)
	//if err != nil {
	//	return para, fmt.Errorf("cannot create virtual service: %s", err)
	//}
	preprocessObject.VirtualService = virtualService
	//log.Debugf("virtual service: %s", string(data))

	return preprocessObject, nil
}

func createOrUpdateDestinationRule(istioClient *rest.RESTClient, preprocessObject PreprocesObject, destinationRuleName string) (err error) {
	request := istioClient.Get().Namespace(preprocessObject.Namespace).Resource("destinationrules").Name(destinationRuleName)
	request.SetHeader("Accept", "application/json")
	request.SetHeader("Content-Type", "application/json")
	log.Debugf("request URL: %s", request.URL())
	result2, err := request.Do().Raw()
	if err != nil {
		log.Debugf("will create new destinationrule \"%s\"", destinationRuleName)
		convertedjson, err := json.Marshal(preprocessObject.DestinationRule)
		if err == nil {
			log.Debugf("create destinationrule: %s", (convertedjson))
		} else {
			return err
		}
		postRequest := istioClient.Post().Namespace(preprocessObject.Namespace).Resource("destinationrules")
		log.Debugf("postRequest URL: %s", postRequest.URL())
		newbody, err := postRequest.Body(convertedjson).Do().Raw()
		if err == nil {
			log.Debugf(string(newbody))
		} else {
			return err
		}
	} else {
		log.Debugf("original destinationrule: %s", result2)
		var originalDestinationRule DestinationRuleCRD
		err = json.Unmarshal(result2, &originalDestinationRule)
		if err != nil {
			return err
		}
		originalDestinationRule.Spec = preprocessObject.DestinationRule.Spec
		updatedjson, err := json.Marshal(originalDestinationRule)
		if err == nil {
			log.Debugf("updated destinationrule: %s", updatedjson)
		} else {
			return err
		}

		updateRequest := istioClient.Put().Namespace(preprocessObject.Namespace).Resource("destinationrules").Name(destinationRuleName)
		log.Debugf("updateRequest URL: %s", updateRequest.URL())
		newbody, err := updateRequest.Body(updatedjson).Do().Raw()
		if err == nil {
			log.Debugf(string(newbody))
		} else {
			log.Errorf(string(newbody))
			log.Error(err)
			return err
		}
	}
	return nil
}

func createOrUpdateVirtualService(istioClient *rest.RESTClient, preprocessObject PreprocesObject, virtualServiceName string) (err error) {
	request := istioClient.Get().Namespace(preprocessObject.Namespace).Resource("virtualservices").Name(virtualServiceName)
	request.SetHeader("Accept", "application/json")
	request.SetHeader("Content-Type", "application/json")
	log.Debugf("request URL: %s", request.URL())
	result2, err := request.Do().Raw()
	if err != nil {
		log.Debugf("will create new virtualservice \"%s\"", virtualServiceName)
		convertedjson, err := json.Marshal(preprocessObject.VirtualService)
		if err == nil {
			log.Debugf("create virtualservice: %s", (convertedjson))
		} else {
			return err
		}
		newbody, err := istioClient.Post().Namespace(namespace).Resource("virtualservices").Body(convertedjson).Do().Raw()
		if err == nil {
			log.Debugf(string(newbody))
		} else {
			return err
		}
	} else {
		log.Debugf("original virtualservice: %s", result2)
		var originalVirtualService VirtualServiceCRD
		err = json.Unmarshal(result2, &originalVirtualService)
		if err != nil {
			return err
		}
		originalVirtualService.Spec = preprocessObject.VirtualService.Spec
		updatedjson, err := json.Marshal(originalVirtualService)
		if err == nil {
			log.Debugf("updated virtualservice: %s", updatedjson)
		} else {
			return err
		}

		newbody, err := istioClient.Put().Namespace(namespace).Resource("virtualservices").Name(virtualServiceName).Body(updatedjson).Do().Raw()
		if err == nil {
			log.Debugf(string(newbody))
		} else {
			log.Errorf(string(newbody))
			log.Error(err)
			return err
		}
	}
	return nil
}

func runTrafficRouterSplit(client *kubernetes.Clientset, istioClient *rest.RESTClient, submitArgs *runTrafficRouterSplitArgs) (err error) {
	preprocesObject, err := submitArgs.preprocess(client)
	if err != nil {
		return err
	}

	log.Debugf("serviceName: %s", preprocesObject.ServiceName)
	jsonDestinationRule, err := json.Marshal(preprocesObject.DestinationRule)
	log.Debugf("destination rule: %s", jsonDestinationRule)
	jsonVirtualService, err := json.Marshal(preprocesObject.VirtualService)
	log.Debugf("virtual service: %s", jsonVirtualService)

	virtualServiceName := preprocesObject.ServiceName
	log.Debugf("virtualServiceName:%s", virtualServiceName)
	destinationRuleName := preprocesObject.ServiceName
	log.Debugf("destinationRuleName:%s", virtualServiceName)

	err = createOrUpdateDestinationRule(istioClient, preprocesObject, destinationRuleName)
	err = createOrUpdateVirtualService(istioClient, preprocesObject, virtualServiceName)

	return err
}

func initIstioClient() (*rest.RESTClient, error) {
	var err error
	restConfig, err = clientConfig.ClientConfig()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	istioAPIGroupVersion := schema.GroupVersion{
		Group:   "networking.istio.io",
		Version: "v1alpha3",
	}
	//istioAPIGroupVersion := schema.GroupVersion{
	//	Group:   "config.istio.io",
	//	Version: "v1alpha2",
	//}

	restConfig.GroupVersion = &istioAPIGroupVersion

	restConfig.APIPath = "/apis"
	restConfig.ContentType = runtime.ContentTypeJSON

	types := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			metav1.AddToGroupVersion(scheme, istioAPIGroupVersion)
			return nil
		})
	err = schemeBuilder.AddToScheme(types)
	restConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(types)}
	// create the clientset

	return rest.RESTClientFor(restConfig)
}
