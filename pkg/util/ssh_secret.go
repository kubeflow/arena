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

package util

import (
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

func GenerateRsaKey() (map[string]string, error) {
	bitSize := 2048

	privateKey, err := rsa.GenerateKey(cryptoRand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// id_rsa
	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(&privBlock)

	// id_rsa.pub
	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	data := make(map[string]string)
	data["privateKey"] = base64.StdEncoding.EncodeToString(privateKeyBytes)
	data["publicKey"] = base64.StdEncoding.EncodeToString(publicKeyBytes)
	data["config"] = generateSSHConfig()

	return data, nil
}

func generateSSHConfig() string {
	config := "StrictHostKeyChecking no\nUserKnownHostsFile /dev/null\n"
	return base64.StdEncoding.EncodeToString([]byte(config))
}
