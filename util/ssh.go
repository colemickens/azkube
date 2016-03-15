package util

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	SshKeySize = 4096
)

func CreateSaveSsh(username, outputDirectory string) (privateKey *rsa.PrivateKey, publicKeyString string, err error) {
	privateKey, publicKeyString, err = CreateSsh()
	if err != nil {
		return nil, "", err
	}

	privateKeyPem := PrivateKeyToPem(privateKey)
	err = SaveDeploymentFile(outputDirectory, fmt.Sprintf("%s_rsa", username), string(privateKeyPem), 0600)
	if err != nil {
		return nil, "", err
	}

	return privateKey, publicKeyString, nil
}

func CreateSsh() (privateKey *rsa.PrivateKey, publicKeyString string, err error) {
	log.Debugf("ssh: generating %dbit rsa key", SshKeySize)
	privateKey, err = rsa.GenerateKey(rand.Reader, SshKeySize)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate private key for ssh: %q", err)
	}

	publicKey := privateKey.PublicKey
	sshPublicKey, err := ssh.NewPublicKey(&publicKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create openssh public key string: %q", err)
	}
	authorizedKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)
	authorizedKey := string(authorizedKeyBytes)

	return privateKey, authorizedKey, nil
}
