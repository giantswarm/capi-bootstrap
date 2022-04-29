package ca

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/config"
	"github.com/giantswarm/capi-bootstrap/pkg/templates"
)

func New(_ config.Config) (*Generator, error) {
	return &Generator{}, nil
}

func (l Generator) Generate(_ context.Context, secret templates.TemplateSecret, installation templates.InstallationInputs) (interface{}, error) {
	ca := &x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName: "Kubernetes API",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var caPEM bytes.Buffer
	err = pem.Encode(&caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	caKeyDER, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var caKeyPEM bytes.Buffer
	err = pem.Encode(&caKeyPEM, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: caKeyDER,
	})

	return map[string]string{
		"cert": caPEM.String(),
		"key":  caKeyPEM.String(),
	}, nil
}
