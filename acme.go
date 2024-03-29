package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"log"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
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

func getCertificate(caDirURL, domain, mail, dnsProviderName, dnsResolver string) (*certificate.Resource, error) {
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

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(
			dnsResolver != "",
			dns01.AddRecursiveNameservers(dns01.ParseNameservers([]string{dnsResolver})),
		),
	)
	if err != nil {
		return nil, err
	}

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
