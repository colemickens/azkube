package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	ValidityDuration = time.Hour * 24 * 365 * 2
	PkiKeySize       = 4096
)

// TODO(colemickens): could surely refactor/dedupe the two functions below, similar to x509's CreateCertificate
// TODO(Colemickens): potential options, duration, alt subj names, etc

func CreateKubeCertificates(masterFQDN string, extraFQDNs []string) (*PkiKeyCertPair, *PkiKeyCertPair, *PkiKeyCertPair, error) {
	log.Info("pki: generating certificate authority")
	caCertificate, caPrivateKey, err := createCertificate("ca", nil, nil, false, "")
	if err != nil {
		return nil, nil, nil, err
	}
	log.Info("pki: generating apiserver server certificate")
	apiserverCertificate, apiserverPrivateKey, err := createCertificate("apiserver", caCertificate, caPrivateKey, true, masterFQDN, extraFQDNs...)
	if err != nil {
		return nil, nil, nil, err
	}
	log.Info("pki: generating client certificate")
	clientCertificate, clientPrivateKey, err := createCertificate("client", caCertificate, caPrivateKey, false, "")
	if err != nil {
		return nil, nil, nil, err
	}

	return &PkiKeyCertPair{CertificatePem: string(CertificateToPem(caCertificate.Raw)), PrivateKeyPem: string(PrivateKeyToPem(caPrivateKey))},
		&PkiKeyCertPair{CertificatePem: string(CertificateToPem(apiserverCertificate.Raw)), PrivateKeyPem: string(PrivateKeyToPem(apiserverPrivateKey))},
		&PkiKeyCertPair{CertificatePem: string(CertificateToPem(clientCertificate.Raw)), PrivateKeyPem: string(PrivateKeyToPem(clientPrivateKey))}, nil
}

func createCertificate(commonName string, caCertificate *x509.Certificate, caPrivateKey *rsa.PrivateKey, isServer bool, FQDN string, extraFQDNs ...string) (*x509.Certificate, *rsa.PrivateKey, error) {
	var err error

	isCA := (caCertificate == nil)

	now := time.Now()

	template := x509.Certificate{
		Subject:   pkix.Name{CommonName: commonName},
		NotBefore: now,
		NotAfter:  now.Add(ValidityDuration),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
		template.IsCA = isCA
	} else if isServer {
		// TODO; this doesn't go here, but need to unblock
		extraFQDNs = append(extraFQDNs, FQDN)

		template.DNSNames = extraFQDNs
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	} else {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}

	snMax := new(big.Int).Lsh(big.NewInt(1), 128)
	template.SerialNumber, err = rand.Int(rand.Reader, snMax)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, PkiKeySize)

	var privateKeyToUse *rsa.PrivateKey
	var certificateToUse *x509.Certificate
	if !isCA {
		privateKeyToUse = caPrivateKey
		certificateToUse = caCertificate
	} else {
		privateKeyToUse = privateKey
		certificateToUse = &template
	}

	certDerBytes, err := x509.CreateCertificate(rand.Reader, &template, certificateToUse, &privateKey.PublicKey, privateKeyToUse)
	if err != nil {
		return nil, nil, err
	}

	certificate, err := x509.ParseCertificate(certDerBytes)
	if err != nil {
		return nil, nil, err
	}

	return certificate, privateKey, nil
}
