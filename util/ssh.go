package util

import (
	"crypto/rand"
	"crypto/rsa"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	SshKeySize = 4096
)

func GenerateSsh(outputDirectory string) (privateKey *rsa.PrivateKey, publicKeyString string, err error) {
	log.Infof("ssh: generating %dbit rsa key", SshKeySize)
	privateKey, err = rsa.GenerateKey(rand.Reader, SshKeySize)
	if err != nil {
		return nil, "", err
	}

	// TODO: write to outputDirectory

	publicKey := privateKey.PublicKey

	sshPublicKey, err := ssh.NewPublicKey(&publicKey)
	if err != nil {
		return nil, "", err
	}
	authorizedKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)
	authorizedKey := string(authorizedKeyBytes)

	return privateKey, authorizedKey, nil
}

func (s *SshProperties) OpenSshPublicKey() (string, error) {
	privateKey, err := PemToPrivateKey(s.PrivateKeyPem)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.PublicKey

	sshPublicKey, err := ssh.NewPublicKey(&publicKey)
	if err != nil {
		return "", err
	}
	authorizedKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)
	authorizedKey := string(authorizedKeyBytes)
	return authorizedKey, nil
}
