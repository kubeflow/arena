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
	bitSize := 1024

	privateKey, err := rsa.GenerateKey(cryptoRand.Reader, bitSize)
	if err != nil {
		//klog.Errorf("rsa generateKey err: %v", err)
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
		//klog.Errorf("ssh newPublicKey err: %v", err)
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
