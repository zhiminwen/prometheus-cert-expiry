package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func checkCert(certToCheck Cert, certExpiryVector *prometheus.GaugeVec) error {
	log.Printf("checking cert: %s", certToCheck.Name)

	switch certToCheck.Type {
	case "file":
		return checkCertByFile(certToCheck, certExpiryVector)
	case "address":
		return checkCertByAddress(certToCheck, certExpiryVector)
	default:
		return fmt.Errorf("unknown cert type: %s", certToCheck.Type)
	}

	return nil
}

func checkCertByFile(certToCheck Cert, certExpiryVector *prometheus.GaugeVec) error {
	content, err := os.ReadFile(certToCheck.File)

	if err != nil {
		log.Printf("ReadFile failed: %s", err)
		return err
	}

	block, _ := pem.Decode(content)
	if block == nil {
		err := fmt.Errorf("failed to decode PEM block containing public key")
		log.Printf("failed to decode PEM block containing public key")
		return err
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Printf("ParseCertificate failed: %s", err)
		return err
	}

	setExpiryVector(certToCheck, cert, certExpiryVector)
	return nil
}

func checkCertByAddress(certToCheck Cert, certExpiryVector *prometheus.GaugeVec) error {
	conn, err := tls.Dial("tcp", certToCheck.Address, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("Faile to dial: %s", err)
		return err
	}
	err = conn.Handshake()
	if err != nil {
		log.Printf("Faile to handshake: %s", err)
		return err
	}
	for _, cert := range conn.ConnectionState().PeerCertificates {
		setExpiryVector(certToCheck, cert, certExpiryVector)
	}
	return nil
}

func setExpiryVector(certToCheck Cert, cert *x509.Certificate, certExpiryVector *prometheus.GaugeVec) {
	subject := cert.Subject.String()
	issuer := cert.Issuer.String()
	left := cert.NotAfter.Sub(time.Now())
	log.Printf("cert: %s, issuer: %s, left: %.2f", subject, issuer, left.Hours()/24)

	switch certToCheck.Type {
	case "file":
		certExpiryVector.WithLabelValues(certToCheck.Type, certToCheck.Name, certToCheck.File, "", subject, issuer).Set(left.Hours() / 24)
	case "address":
		certExpiryVector.WithLabelValues(certToCheck.Type, certToCheck.Name, "", certToCheck.Address, subject, issuer).Set(left.Hours() / 24)
	}

}
