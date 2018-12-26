package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"log"

	"github.com/xenolf/lego/certcrypto"
	"github.com/xenolf/lego/certificate"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/lego"
	"github.com/xenolf/lego/providers/dns"
	"github.com/xenolf/lego/registration"
)

type acmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u acmeUser) GetEmail() string {
	return u.Email
}

func (u acmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u acmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func getCertificate(caDirURL string, domain string, mail string, dnsProviderName string) (*certificate.Resource, error) {
	const rsaKeySize = 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, err
	}
	myUser := acmeUser{
		Email: mail,
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)
	config.CADirURL = caDirURL
	config.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}

	var provider challenge.Provider
	switch dnsProviderName {
	case "manual":
		provider, err = dns01.NewDNSProviderManual()
	default:
		provider, err = dns.NewDNSChallengeProviderByName(dnsProviderName)
	}
	if err != nil {
		return nil, err
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return nil, err
	}
	client.Challenge.Exclude([]challenge.Type{challenge.HTTP01, challenge.TLSALPN01})

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
