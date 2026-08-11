// Copyright 2026 The Kubeflow Authors
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

package training

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTensorboardURLUsesLoadBalancerHostname(t *testing.T) {
	url, err := tensorboardURL("job", "default", []*corev1.Service{{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Labels:    map[string]string{"release": "job"},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{{Port: 6006}},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{{Hostname: "a123456.elb.amazonaws.com"}},
			},
		},
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	const want = "http://a123456.elb.amazonaws.com:6006"
	if url != want {
		t.Fatalf("got %q, want %q", url, want)
	}
}

func TestTensorboardURLUsesLoadBalancerIP(t *testing.T) {
	url, err := tensorboardURL("job", "default", []*corev1.Service{{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Labels:    map[string]string{"release": "job"},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{{Port: 6006}},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{{IP: "192.0.2.10"}},
			},
		},
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	const want = "http://192.0.2.10:6006"
	if url != want {
		t.Fatalf("got %q, want %q", url, want)
	}
}
