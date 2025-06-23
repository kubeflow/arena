// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TrafficRouterSplitArgs struct {
	ServingName    string `yaml:"servingName,omitempty"` //--name
	Namespace      string `yaml:"namespace,omitempty"`   //--namespace
	Versions       string `yaml:"versions,omitempty"`    //--versions
	Weights        string `yaml:"weights,omitempty"`     //--weights
	VersionWeights []ServingVersionWeight
}

type ServingVersionWeight struct {
	Version string
	Weight  int
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
	metav1.ObjectMeta `json:"metadata,omitempty"   protobuf:"bytes,1,opt,name=metadata"   yaml:"metadata,omitempty"`
	Spec              *istiov1alpha3.DestinationRule `json:"spec,omitempty"       protobuf:"bytes,2,opt,name=spec"       yaml:"spec,omitempty"`
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
	metav1.ObjectMeta `json:"metadata,omitempty"   protobuf:"bytes,1,opt,name=metadata"   yaml:"metadata,omitempty"`
	Spec              VirtualService `json:"spec,omitempty"       protobuf:"bytes,2,opt,name=spec"       yaml:"spec,omitempty"`
}

type VirtualService struct {
	*istiov1alpha3.VirtualService
	Http []*HTTPRoute `json:"http,omitempty" protobuf:"bytes,3,rep,name=http"`
}

type HTTPRoute struct {
	*istiov1alpha3.HTTPRoute
	Match []*HTTPMatchRequest  `json:"match,omitempty" protobuf:"bytes,1,rep,name=match"`
	Route []*DestinationWeight `json:"route,omitempty" protobuf:"bytes,2,rep,name=route"`
}

type HTTPMatchRequest struct {
	*istiov1alpha3.HTTPMatchRequest
	Uri *StringMatchPrefix `json:"uri,omitempty" protobuf:"bytes,1,opt,name=uri"`
}

type StringMatchPrefix struct {
	Prefix string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix,proto3,oneof"`
}

type DestinationWeight struct {
	Destination *Destination `json:"destination,omitempty" protobuf:"bytes,1,opt,name=destination"`
	Weight      int32        `json:"weight"                protobuf:"varint,2,opt,name=weight,proto3"`
}

type Destination struct {
	*istiov1alpha3.Destination
	Port *PortSelector `json:"port,omitempty" protobuf:"bytes,3,opt,name=port"`
}

type PortSelector struct {
	*istiov1alpha3.PortSelector
	Number uint32 `json:"number,omitempty" protobuf:"varint,1,opt,name=number,proto3,oneof"`
}
