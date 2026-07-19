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

package config

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	authenticationapi "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getUserName(namespace string, clientConfig clientcmd.ClientConfig, restConfig *rest.Config, clientSet *kubernetes.Clientset, tr *tokenRetriever) (*string, error) {
	tc, err := restConfig.TransportConfig()
	if err != nil {
		return nil, err
	}
	if tc.HasBasicAuth() {
		userName := fmt.Sprintf("kubecfg:basicauth:%s", restConfig.Username)
		return &userName, nil
	}
	if tc.HasCertAuth() {
		// Prefer asking the API server who we are (same as `kubectl auth whoami`).
		// Historically this branch hard-coded "kubecfg:certauth:admin", which is wrong for
		// per-user client certificates (e.g. ACK RAM X509 kubeconfigs whose CN is the RAM UID).
		if userName, err := getUserNameBySelfSubjectReview(clientSet); err == nil {
			return userName, nil
		} else {
			log.Debugf("SelfSubjectReview failed for cert auth, falling back to certificate CN: %v", err)
		}
		// Offline fallback: Kubernetes maps the certificate CommonName to the username.
		if userName, err := getUserNameFromClientCertificate(restConfig); err == nil {
			return userName, nil
		} else {
			log.Debugf("failed to parse client certificate CN, falling back to legacy username: %v", err)
		}
		// Preserve legacy behavior for unusual certificate setups.
		userName := "kubecfg:certauth:admin"
		return &userName, nil
	}
	var token string
	if tc.HasTokenAuth() {
		if restConfig.BearerTokenFile != "" {
			tokenContent, err := os.ReadFile(restConfig.BearerTokenFile)
			if err != nil {
				return nil, err
			}
			token = string(tokenContent)
		}
		if restConfig.BearerToken != "" {
			token = restConfig.BearerToken
		}
	}
	if token == "" && restConfig.AuthProvider != nil {
		if err := createSubjectRulesReviews(namespace, clientSet); err != nil {
			return nil, err
		}
		token = tr.token
	}
	if token == "" {
		return nil, fmt.Errorf("not found user name for the current context,we don't know how to detect user name")
	}
	return getUserNameByToken(clientSet, token)
}

func getUserNameByToken(kubeclient kubernetes.Interface, token string) (*string, error) {
	result, err := kubeclient.AuthenticationV1().TokenReviews().Create(
		context.TODO(),
		&authenticationapi.TokenReview{
			Spec: authenticationapi.TokenReviewSpec{
				Token: token,
			},
		},
		metav1.CreateOptions{},
	)

	if err != nil {
		return nil, err
	}

	if result.Status.Error != "" {
		return nil, fmt.Errorf("%s", result.Status.Error)
	}

	return &result.Status.User.Username, nil
}

func getUserNameBySelfSubjectReview(kubeclient kubernetes.Interface) (*string, error) {
	result, err := kubeclient.AuthenticationV1().SelfSubjectReviews().Create(
		context.TODO(),
		&authenticationapi.SelfSubjectReview{},
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, err
	}
	if result.Status.UserInfo.Username == "" {
		return nil, fmt.Errorf("SelfSubjectReview returned empty username")
	}
	return &result.Status.UserInfo.Username, nil
}

func getUserNameFromClientCertificate(restConfig *rest.Config) (*string, error) {
	var certPEM []byte
	switch {
	case len(restConfig.CertData) > 0:
		certPEM = restConfig.CertData
	case restConfig.CertFile != "":
		var err error
		certPEM, err = os.ReadFile(restConfig.CertFile)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("no client certificate data or file configured")
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode client certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	if cert.Subject.CommonName == "" {
		return nil, fmt.Errorf("client certificate has empty CommonName")
	}
	userName := cert.Subject.CommonName
	return &userName, nil
}

func createSubjectRulesReviews(namespace string, kubeclient kubernetes.Interface) error {
	sar := &authorizationv1.SelfSubjectRulesReview{
		Spec: authorizationv1.SelfSubjectRulesReviewSpec{
			Namespace: namespace,
		},
	}
	_, err := kubeclient.AuthorizationV1().SelfSubjectRulesReviews().Create(context.TODO(), sar, metav1.CreateOptions{})
	return err
}
