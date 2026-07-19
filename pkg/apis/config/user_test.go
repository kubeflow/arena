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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/client-go/rest"
)

func mustGenerateClientCertPEM(t *testing.T, commonName string) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestGetUserNameFromClientCertificateCertData(t *testing.T) {
	const want = "204360775885168282"
	cfg := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			CertData: mustGenerateClientCertPEM(t, want),
		},
	}
	got, err := getUserNameFromClientCertificate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}

func TestGetUserNameFromClientCertificateCertFile(t *testing.T) {
	const want = "ram-user-from-file"
	dir := t.TempDir()
	path := filepath.Join(dir, "client.crt")
	if err := os.WriteFile(path, mustGenerateClientCertPEM(t, want), 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	cfg := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			CertFile: path,
		},
	}
	got, err := getUserNameFromClientCertificate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}

func TestGetUserNameFromClientCertificateMissing(t *testing.T) {
	if _, err := getUserNameFromClientCertificate(&rest.Config{}); err == nil {
		t.Fatal("expected error for missing certificate")
	}
}

func TestGetUserNameFromClientCertificateEmptyCN(t *testing.T) {
	cfg := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			CertData: mustGenerateClientCertPEM(t, ""),
		},
	}
	if _, err := getUserNameFromClientCertificate(cfg); err == nil {
		t.Fatal("expected error for empty CommonName")
	}
}
