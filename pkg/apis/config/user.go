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
	"fmt"
	"os"

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
		return nil, fmt.Errorf(result.Status.Error)
	}

	return &result.Status.User.Username, nil
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
