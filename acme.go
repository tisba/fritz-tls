package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"log"

	"github.com/xenolf/lego/acme"
)

type acmeUser struct {
	Email        string
	Registration *acme.RegistrationResource
	key          crypto.PrivateKey
}

func (u acmeUser) GetEmail() string {
	return u.Email
}

func (u acmeUser) GetRegistration() *acme.RegistrationResource {
	return u.Registration
}

func (u acmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func getCertificate(caDirURL string, domain string, mail string) (*acme.CertificateResource, error) {
	const rsaKeySize = 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, err
	}
	myUser := acmeUser{
		Email: mail,
		key:   privateKey,
	}

	client, err := acme.NewClient(caDirURL, &myUser, acme.RSA2048)
	if err != nil {
		return nil, err
	}

	_, err = client.Register(true)
	if err != nil {
		log.Fatal(err)
	}

	// configure manual DNS challenge provider
	// and only ask for DNS01 challenge
	manualDNS, err := acme.NewDNSProviderManual()
	if err != nil {
		return nil, err
	}
	client.SetChallengeProvider(acme.DNS01, manualDNS)
	client.ExcludeChallenges([]acme.Challenge{acme.Challenge("http-01"), acme.Challenge("tls-alpn-01")})

	bundle := true
	cert, err := client.ObtainCertificate([]string{domain}, bundle, nil, false)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
