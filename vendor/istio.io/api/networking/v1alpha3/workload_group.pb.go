// Copyright 2020 Istio Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: networking/v1alpha3/workload_group.proto

// $schema: istio.networking.v1alpha3.WorkloadGroup
// $title: Workload Group
// $description: Describes a collection of workload instances.
// $location: https://istio.io/docs/reference/config/networking/workload-group.html
// $aliases: [/docs/reference/config/networking/v1alpha3/workload-group]

// `WorkloadGroup` describes a collection of workload instances.
// It provides a specification that the workload instances can use to bootstrap
// their proxies, including the metadata and identity. It is only intended to
// be used with non-k8s workloads like Virtual Machines, and is meant to mimic
// the existing sidecar injection and deployment specification model used for
// Kubernetes workloads to bootstrap Istio proxies.
//
// The following example declares a workload group representing a collection
// of workloads that will be registered under `reviews` in namespace
// `bookinfo`. The set of labels will be associated with each workload
// instance during the bootstrap process, and the ports 3550 and 8080
// will be associated with the workload group and use service account `default`.
// `app.kubernetes.io/version` is just an arbitrary example of a label.
//
// ```yaml
// apiVersion: networking.istio.io/v1
// kind: WorkloadGroup
// metadata:
//   name: reviews
//   namespace: bookinfo
// spec:
//   metadata:
//     labels:
//       app.kubernetes.io/name: reviews
//       app.kubernetes.io/version: "1.3.4"
//   template:
//     ports:
//       grpc: 3550
//       http: 8080
//     serviceAccount: default
//   probe:
//     initialDelaySeconds: 5
//     timeoutSeconds: 3
//     periodSeconds: 4
//     successThreshold: 3
//     failureThreshold: 3
//     httpGet:
//      path: /foo/bar
//      host: 127.0.0.1
//      port: 3100
//      scheme: HTTPS
//      httpHeaders:
//      - name: Lit-Header
//        value: Im-The-Best
// ```
//

package v1alpha3

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// `WorkloadGroup` enables specifying the properties of a single workload for bootstrap and
// provides a template for `WorkloadEntry`, similar to how `Deployment` specifies properties
// of workloads via `Pod` templates. A `WorkloadGroup` can have more than one `WorkloadEntry`.
// `WorkloadGroup` has no relationship to resources which control service registry like `ServiceEntry`
// and as such doesn't configure host name for these workloads.
//
// <!-- crd generation tags
// +cue-gen:WorkloadGroup:groupName:networking.istio.io
// +cue-gen:WorkloadGroup:versions:v1beta1,v1alpha3,v1
// +cue-gen:WorkloadGroup:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio
// +cue-gen:WorkloadGroup:subresource:status
// +cue-gen:WorkloadGroup:scope:Namespaced
// +cue-gen:WorkloadGroup:resource:categories=istio-io,networking-istio-io,shortNames=wg,plural=workloadgroups
// +cue-gen:WorkloadGroup:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp
// representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations.
// Clients may not set this value. It is represented in RFC3339 form and is in UTC.
// Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
// +cue-gen:WorkloadGroup:preserveUnknownFields:false
// +cue-gen:WorkloadGroup:spec:required
// -->
//
// <!-- go code generation tags
// +kubetype-gen
// +kubetype-gen:groupVersion=networking.istio.io/v1alpha3
// +genclient
// +k8s:deepcopy-gen=true
// -->
type WorkloadGroup struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Metadata that will be used for all corresponding `WorkloadEntries`.
	// User labels for a workload group should be set here in `metadata` rather than in `template`.
	Metadata *WorkloadGroup_ObjectMeta `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// Template to be used for the generation of `WorkloadEntry` resources that belong to this `WorkloadGroup`.
	// Please note that `address` and `labels` fields should not be set in the template, and an empty `serviceAccount`
	// should default to `default`. The workload identities (mTLS certificates) will be bootstrapped using the
	// specified service account's token. Workload entries in this group will be in the same namespace as the
	// workload group, and inherit the labels and annotations from the above `metadata` field.
	// +protoc-gen-crd:validation:IgnoreSubValidation:["Address is required"]
	Template *WorkloadEntry `protobuf:"bytes,2,opt,name=template,proto3" json:"template,omitempty"`
	// `ReadinessProbe` describes the configuration the user must provide for healthchecking on their workload.
	// This configuration mirrors K8S in both syntax and logic for the most part.
	Probe *ReadinessProbe `protobuf:"bytes,3,opt,name=probe,proto3" json:"probe,omitempty"`
}

func (x *WorkloadGroup) Reset() {
	*x = WorkloadGroup{}
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WorkloadGroup) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkloadGroup) ProtoMessage() {}

func (x *WorkloadGroup) ProtoReflect() protoreflect.Message {
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkloadGroup.ProtoReflect.Descriptor instead.
func (*WorkloadGroup) Descriptor() ([]byte, []int) {
	return file_networking_v1alpha3_workload_group_proto_rawDescGZIP(), []int{0}
}

func (x *WorkloadGroup) GetMetadata() *WorkloadGroup_ObjectMeta {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *WorkloadGroup) GetTemplate() *WorkloadEntry {
	if x != nil {
		return x.Template
	}
	return nil
}

func (x *WorkloadGroup) GetProbe() *ReadinessProbe {
	if x != nil {
		return x.Probe
	}
	return nil
}

type ReadinessProbe struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Number of seconds after the container has started before readiness probes are initiated.
	// +kubebuilder:validation:Minimum=0
	InitialDelaySeconds int32 `protobuf:"varint,2,opt,name=initial_delay_seconds,json=initialDelaySeconds,proto3" json:"initial_delay_seconds,omitempty"`
	// Number of seconds after which the probe times out.
	// Defaults to 1 second. Minimum value is 1 second.
	// +kubebuilder:validation:Minimum=0
	TimeoutSeconds int32 `protobuf:"varint,3,opt,name=timeout_seconds,json=timeoutSeconds,proto3" json:"timeout_seconds,omitempty"`
	// How often (in seconds) to perform the probe.
	// Default to 10 seconds. Minimum value is 1 second.
	// +kubebuilder:validation:Minimum=0
	PeriodSeconds int32 `protobuf:"varint,4,opt,name=period_seconds,json=periodSeconds,proto3" json:"period_seconds,omitempty"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1 second.
	// +kubebuilder:validation:Minimum=0
	SuccessThreshold int32 `protobuf:"varint,5,opt,name=success_threshold,json=successThreshold,proto3" json:"success_threshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Defaults to 3 seconds.
	// +kubebuilder:validation:Minimum=0
	FailureThreshold int32 `protobuf:"varint,6,opt,name=failure_threshold,json=failureThreshold,proto3" json:"failure_threshold,omitempty"`
	// Users can only provide one configuration for healthchecks (tcp, http, exec),
	// and this is expressed as a oneof. All of the other configuration values
	// hold true for any of the healthcheck methods.
	//
	// Types that are assignable to HealthCheckMethod:
	//
	//	*ReadinessProbe_HttpGet
	//	*ReadinessProbe_TcpSocket
	//	*ReadinessProbe_Exec
	HealthCheckMethod isReadinessProbe_HealthCheckMethod `protobuf_oneof:"health_check_method"`
}

func (x *ReadinessProbe) Reset() {
	*x = ReadinessProbe{}
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReadinessProbe) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadinessProbe) ProtoMessage() {}

func (x *ReadinessProbe) ProtoReflect() protoreflect.Message {
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadinessProbe.ProtoReflect.Descriptor instead.
func (*ReadinessProbe) Descriptor() ([]byte, []int) {
	return file_networking_v1alpha3_workload_group_proto_rawDescGZIP(), []int{1}
}

func (x *ReadinessProbe) GetInitialDelaySeconds() int32 {
	if x != nil {
		return x.InitialDelaySeconds
	}
	return 0
}

func (x *ReadinessProbe) GetTimeoutSeconds() int32 {
	if x != nil {
		return x.TimeoutSeconds
	}
	return 0
}

func (x *ReadinessProbe) GetPeriodSeconds() int32 {
	if x != nil {
		return x.PeriodSeconds
	}
	return 0
}

func (x *ReadinessProbe) GetSuccessThreshold() int32 {
	if x != nil {
		return x.SuccessThreshold
	}
	return 0
}

func (x *ReadinessProbe) GetFailureThreshold() int32 {
	if x != nil {
		return x.FailureThreshold
	}
	return 0
}

func (m *ReadinessProbe) GetHealthCheckMethod() isReadinessProbe_HealthCheckMethod {
	if m != nil {
		return m.HealthCheckMethod
	}
	return nil
}

func (x *ReadinessProbe) GetHttpGet() *HTTPHealthCheckConfig {
	if x, ok := x.GetHealthCheckMethod().(*ReadinessProbe_HttpGet); ok {
		return x.HttpGet
	}
	return nil
}

func (x *ReadinessProbe) GetTcpSocket() *TCPHealthCheckConfig {
	if x, ok := x.GetHealthCheckMethod().(*ReadinessProbe_TcpSocket); ok {
		return x.TcpSocket
	}
	return nil
}

func (x *ReadinessProbe) GetExec() *ExecHealthCheckConfig {
	if x, ok := x.GetHealthCheckMethod().(*ReadinessProbe_Exec); ok {
		return x.Exec
	}
	return nil
}

type isReadinessProbe_HealthCheckMethod interface {
	isReadinessProbe_HealthCheckMethod()
}

type ReadinessProbe_HttpGet struct {
	// `httpGet` is performed to a given endpoint
	// and the status/able to connect determines health.
	HttpGet *HTTPHealthCheckConfig `protobuf:"bytes,7,opt,name=http_get,json=httpGet,proto3,oneof"`
}

type ReadinessProbe_TcpSocket struct {
	// Health is determined by if the proxy is able to connect.
	TcpSocket *TCPHealthCheckConfig `protobuf:"bytes,8,opt,name=tcp_socket,json=tcpSocket,proto3,oneof"`
}

type ReadinessProbe_Exec struct {
	// Health is determined by how the command that is executed exited.
	Exec *ExecHealthCheckConfig `protobuf:"bytes,9,opt,name=exec,proto3,oneof"`
}

func (*ReadinessProbe_HttpGet) isReadinessProbe_HealthCheckMethod() {}

func (*ReadinessProbe_TcpSocket) isReadinessProbe_HealthCheckMethod() {}

func (*ReadinessProbe_Exec) isReadinessProbe_HealthCheckMethod() {}

type HTTPHealthCheckConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path to access on the HTTP server.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Port on which the endpoint lives.
	// +kubebuilder:validation:XValidation:message="port must be between 1-65535",rule="0 < self && self <= 65535"
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// Host name to connect to, defaults to the pod IP. You probably want to set
	// "Host" in httpHeaders instead.
	Host string `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty"`
	// HTTP or HTTPS, defaults to HTTP
	// +kubebuilder:validation:XValidation:message="scheme must be one of [HTTP, HTTPS]",rule="self in [”, 'HTTP', 'HTTPS']"
	Scheme string `protobuf:"bytes,4,opt,name=scheme,proto3" json:"scheme,omitempty"`
	// Headers the proxy will pass on to make the request.
	// Allows repeated headers.
	HttpHeaders []*HTTPHeader `protobuf:"bytes,5,rep,name=http_headers,json=httpHeaders,proto3" json:"http_headers,omitempty"`
}

func (x *HTTPHealthCheckConfig) Reset() {
	*x = HTTPHealthCheckConfig{}
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HTTPHealthCheckConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HTTPHealthCheckConfig) ProtoMessage() {}

func (x *HTTPHealthCheckConfig) ProtoReflect() protoreflect.Message {
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HTTPHealthCheckConfig.ProtoReflect.Descriptor instead.
func (*HTTPHealthCheckConfig) Descriptor() ([]byte, []int) {
	return file_networking_v1alpha3_workload_group_proto_rawDescGZIP(), []int{2}
}

func (x *HTTPHealthCheckConfig) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *HTTPHealthCheckConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *HTTPHealthCheckConfig) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *HTTPHealthCheckConfig) GetScheme() string {
	if x != nil {
		return x.Scheme
	}
	return ""
}

func (x *HTTPHealthCheckConfig) GetHttpHeaders() []*HTTPHeader {
	if x != nil {
		return x.HttpHeaders
	}
	return nil
}

type HTTPHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The header field name
	// +kubebuilder:validation:Pattern=^[-_A-Za-z0-9]+$
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The header field value
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *HTTPHeader) Reset() {
	*x = HTTPHeader{}
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HTTPHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HTTPHeader) ProtoMessage() {}

func (x *HTTPHeader) ProtoReflect() protoreflect.Message {
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HTTPHeader.ProtoReflect.Descriptor instead.
func (*HTTPHeader) Descriptor() ([]byte, []int) {
	return file_networking_v1alpha3_workload_group_proto_rawDescGZIP(), []int{3}
}

func (x *HTTPHeader) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *HTTPHeader) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type TCPHealthCheckConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Host to connect to, defaults to localhost
	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// Port of host
	// +kubebuilder:validation:XValidation:message="port must be between 1-65535",rule="0 < self && self <= 65535"
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *TCPHealthCheckConfig) Reset() {
	*x = TCPHealthCheckConfig{}
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TCPHealthCheckConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TCPHealthCheckConfig) ProtoMessage() {}

func (x *TCPHealthCheckConfig) ProtoReflect() protoreflect.Message {
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TCPHealthCheckConfig.ProtoReflect.Descriptor instead.
func (*TCPHealthCheckConfig) Descriptor() ([]byte, []int) {
	return file_networking_v1alpha3_workload_group_proto_rawDescGZIP(), []int{4}
}

func (x *TCPHealthCheckConfig) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *TCPHealthCheckConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

type ExecHealthCheckConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Command to run. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
	// +protoc-gen-crd:list-value-validation:MinLength=1
	Command []string `protobuf:"bytes,1,rep,name=command,proto3" json:"command,omitempty"`
}

func (x *ExecHealthCheckConfig) Reset() {
	*x = ExecHealthCheckConfig{}
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExecHealthCheckConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecHealthCheckConfig) ProtoMessage() {}

func (x *ExecHealthCheckConfig) ProtoReflect() protoreflect.Message {
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecHealthCheckConfig.ProtoReflect.Descriptor instead.
func (*ExecHealthCheckConfig) Descriptor() ([]byte, []int) {
	return file_networking_v1alpha3_workload_group_proto_rawDescGZIP(), []int{5}
}

func (x *ExecHealthCheckConfig) GetCommand() []string {
	if x != nil {
		return x.Command
	}
	return nil
}

// `ObjectMeta` describes metadata that will be attached to a `WorkloadEntry`.
// It is a subset of the supported Kubernetes metadata.
type WorkloadGroup_ObjectMeta struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Labels to attach
	// +kubebuilder:validation:MaxProperties=256
	Labels map[string]string `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Annotations to attach
	// +kubebuilder:validation:MaxProperties=256
	Annotations map[string]string `protobuf:"bytes,2,rep,name=annotations,proto3" json:"annotations,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *WorkloadGroup_ObjectMeta) Reset() {
	*x = WorkloadGroup_ObjectMeta{}
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WorkloadGroup_ObjectMeta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkloadGroup_ObjectMeta) ProtoMessage() {}

func (x *WorkloadGroup_ObjectMeta) ProtoReflect() protoreflect.Message {
	mi := &file_networking_v1alpha3_workload_group_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkloadGroup_ObjectMeta.ProtoReflect.Descriptor instead.
func (*WorkloadGroup_ObjectMeta) Descriptor() ([]byte, []int) {
	return file_networking_v1alpha3_workload_group_proto_rawDescGZIP(), []int{0, 0}
}

func (x *WorkloadGroup_ObjectMeta) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *WorkloadGroup_ObjectMeta) GetAnnotations() map[string]string {
	if x != nil {
		return x.Annotations
	}
	return nil
}

var File_networking_v1alpha3_workload_group_proto protoreflect.FileDescriptor

var file_networking_v1alpha3_workload_group_proto_rawDesc = []byte{
	0x0a, 0x28, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x33, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x67,
	0x72, 0x6f, 0x75, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x69, 0x73, 0x74, 0x69,
	0x6f, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x33, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x28, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69,
	0x6e, 0x67, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33, 0x2f, 0x77, 0x6f, 0x72, 0x6b,
	0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xb8, 0x04, 0x0a, 0x0d, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x12, 0x4f, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33,
	0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x4f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x4a, 0x0a, 0x08, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x33, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x42,
	0x04, 0xe2, 0x41, 0x01, 0x02, 0x52, 0x08, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12,
	0x3f, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29,
	0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e,
	0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x69,
	0x6e, 0x65, 0x73, 0x73, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x62, 0x65,
	0x1a, 0xc8, 0x02, 0x0a, 0x0a, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x12,
	0x57, 0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x3f, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69,
	0x6e, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33, 0x2e, 0x57, 0x6f, 0x72, 0x6b,
	0x6c, 0x6f, 0x61, 0x64, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x4d, 0x65, 0x74, 0x61, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x66, 0x0a, 0x0b, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x44, 0x2e,
	0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f,
	0x61, 0x64, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65,
	0x74, 0x61, 0x2e, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x0b, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x1a, 0x39, 0x0a, 0x0b, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3e, 0x0a, 0x10, 0x41,
	0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xee, 0x03, 0x0a, 0x0e,
	0x52, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x12, 0x32,
	0x0a, 0x15, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x5f, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x5f,
	0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x13, 0x69,
	0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x53, 0x65, 0x63, 0x6f, 0x6e,
	0x64, 0x73, 0x12, 0x27, 0x0a, 0x0f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65,
	0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x74, 0x69, 0x6d,
	0x65, 0x6f, 0x75, 0x74, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x70,
	0x65, 0x72, 0x69, 0x6f, 0x64, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0d, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x53, 0x65, 0x63, 0x6f, 0x6e,
	0x64, 0x73, 0x12, 0x2b, 0x0a, 0x11, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x68,
	0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x10, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x12,
	0x2b, 0x0a, 0x11, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x74, 0x68, 0x72, 0x65, 0x73,
	0x68, 0x6f, 0x6c, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x10, 0x66, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x12, 0x4d, 0x0a, 0x08,
	0x68, 0x74, 0x74, 0x70, 0x5f, 0x67, 0x65, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30,
	0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e,
	0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x48, 0x00, 0x52, 0x07, 0x68, 0x74, 0x74, 0x70, 0x47, 0x65, 0x74, 0x12, 0x50, 0x0a, 0x0a, 0x74,
	0x63, 0x70, 0x5f, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x2f, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69,
	0x6e, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33, 0x2e, 0x54, 0x43, 0x50, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x48, 0x00, 0x52, 0x09, 0x74, 0x63, 0x70, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x46, 0x0a,
	0x04, 0x65, 0x78, 0x65, 0x63, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x69, 0x73,
	0x74, 0x69, 0x6f, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x33, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x48, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48, 0x00, 0x52,
	0x04, 0x65, 0x78, 0x65, 0x63, 0x42, 0x15, 0x0a, 0x13, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x5f,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x22, 0xbb, 0x01, 0x0a,
	0x15, 0x48, 0x54, 0x54, 0x50, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x18, 0x0a, 0x04, 0x70, 0x6f,
	0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x04, 0xe2, 0x41, 0x01, 0x02, 0x52, 0x04,
	0x70, 0x6f, 0x72, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65,
	0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x65,
	0x12, 0x48, 0x0a, 0x0c, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x6e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x33, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x0b, 0x68,
	0x74, 0x74, 0x70, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x22, 0x36, 0x0a, 0x0a, 0x48, 0x54,
	0x54, 0x50, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x22, 0x44, 0x0a, 0x14, 0x54, 0x43, 0x50, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f,
	0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x18,
	0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x04, 0xe2, 0x41,
	0x01, 0x02, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x37, 0x0a, 0x15, 0x45, 0x78, 0x65, 0x63,
	0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x1e, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x09, 0x42, 0x04, 0xe2, 0x41, 0x01, 0x02, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x42, 0x22, 0x5a, 0x20, 0x69, 0x73, 0x74, 0x69, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_networking_v1alpha3_workload_group_proto_rawDescOnce sync.Once
	file_networking_v1alpha3_workload_group_proto_rawDescData = file_networking_v1alpha3_workload_group_proto_rawDesc
)

func file_networking_v1alpha3_workload_group_proto_rawDescGZIP() []byte {
	file_networking_v1alpha3_workload_group_proto_rawDescOnce.Do(func() {
		file_networking_v1alpha3_workload_group_proto_rawDescData = protoimpl.X.CompressGZIP(file_networking_v1alpha3_workload_group_proto_rawDescData)
	})
	return file_networking_v1alpha3_workload_group_proto_rawDescData
}

var file_networking_v1alpha3_workload_group_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_networking_v1alpha3_workload_group_proto_goTypes = []any{
	(*WorkloadGroup)(nil),            // 0: istio.networking.v1alpha3.WorkloadGroup
	(*ReadinessProbe)(nil),           // 1: istio.networking.v1alpha3.ReadinessProbe
	(*HTTPHealthCheckConfig)(nil),    // 2: istio.networking.v1alpha3.HTTPHealthCheckConfig
	(*HTTPHeader)(nil),               // 3: istio.networking.v1alpha3.HTTPHeader
	(*TCPHealthCheckConfig)(nil),     // 4: istio.networking.v1alpha3.TCPHealthCheckConfig
	(*ExecHealthCheckConfig)(nil),    // 5: istio.networking.v1alpha3.ExecHealthCheckConfig
	(*WorkloadGroup_ObjectMeta)(nil), // 6: istio.networking.v1alpha3.WorkloadGroup.ObjectMeta
	nil,                              // 7: istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry
	nil,                              // 8: istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry
	(*WorkloadEntry)(nil),            // 9: istio.networking.v1alpha3.WorkloadEntry
}
var file_networking_v1alpha3_workload_group_proto_depIdxs = []int32{
	6, // 0: istio.networking.v1alpha3.WorkloadGroup.metadata:type_name -> istio.networking.v1alpha3.WorkloadGroup.ObjectMeta
	9, // 1: istio.networking.v1alpha3.WorkloadGroup.template:type_name -> istio.networking.v1alpha3.WorkloadEntry
	1, // 2: istio.networking.v1alpha3.WorkloadGroup.probe:type_name -> istio.networking.v1alpha3.ReadinessProbe
	2, // 3: istio.networking.v1alpha3.ReadinessProbe.http_get:type_name -> istio.networking.v1alpha3.HTTPHealthCheckConfig
	4, // 4: istio.networking.v1alpha3.ReadinessProbe.tcp_socket:type_name -> istio.networking.v1alpha3.TCPHealthCheckConfig
	5, // 5: istio.networking.v1alpha3.ReadinessProbe.exec:type_name -> istio.networking.v1alpha3.ExecHealthCheckConfig
	3, // 6: istio.networking.v1alpha3.HTTPHealthCheckConfig.http_headers:type_name -> istio.networking.v1alpha3.HTTPHeader
	7, // 7: istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.labels:type_name -> istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry
	8, // 8: istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.annotations:type_name -> istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_networking_v1alpha3_workload_group_proto_init() }
func file_networking_v1alpha3_workload_group_proto_init() {
	if File_networking_v1alpha3_workload_group_proto != nil {
		return
	}
	file_networking_v1alpha3_workload_entry_proto_init()
	file_networking_v1alpha3_workload_group_proto_msgTypes[1].OneofWrappers = []any{
		(*ReadinessProbe_HttpGet)(nil),
		(*ReadinessProbe_TcpSocket)(nil),
		(*ReadinessProbe_Exec)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_networking_v1alpha3_workload_group_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_networking_v1alpha3_workload_group_proto_goTypes,
		DependencyIndexes: file_networking_v1alpha3_workload_group_proto_depIdxs,
		MessageInfos:      file_networking_v1alpha3_workload_group_proto_msgTypes,
	}.Build()
	File_networking_v1alpha3_workload_group_proto = out.File
	file_networking_v1alpha3_workload_group_proto_rawDesc = nil
	file_networking_v1alpha3_workload_group_proto_goTypes = nil
	file_networking_v1alpha3_workload_group_proto_depIdxs = nil
}
