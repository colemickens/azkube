package util

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func PemToCertificate(pemString string) (*x509.Certificate, error) {
	pemBytes := []byte(pemString)
	pemBlock, _ := pem.Decode(pemBytes)

	certificate, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return certificate, err
}

func PemToPrivateKey(pemString string) (*rsa.PrivateKey, error) {
	pemBytes := []byte(pemString)
	pemBlock, _ := pem.Decode(pemBytes)

	privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, err
}

func CertificateToPem(derBytes []byte) []byte {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	pemBuffer := bytes.Buffer{}
	pem.Encode(&pemBuffer, pemBlock)

	return pemBuffer.Bytes()
}

func PrivateKeyToPem(privateKey *rsa.PrivateKey) []byte {
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	pemBuffer := bytes.Buffer{}
	pem.Encode(&pemBuffer, pemBlock)

	return pemBuffer.Bytes()
}
