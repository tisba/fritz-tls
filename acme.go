package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"log"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/providers/dns/manual"
	"github.com/go-acme/lego/v5/registration"
)

type acmeUser struct {
	Email        string
	Registration *acme.ExtendedAccount
	key          crypto.Signer
}

func (u acmeUser) GetEmail() string {
	return u.Email
}

func (u acmeUser) GetRegistration() *acme.ExtendedAccount {
	return u.Registration
}

func (u acmeUser) GetPrivateKey() crypto.Signer {
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

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
		KeyType: "RSA2048",
	}

	config := lego.NewConfig(&myUser)
	config.CADirURL = caDirURL

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Registration.Register(context.TODO(), registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}

	var provider challenge.Provider
	switch dnsProviderName {
	case "manual":
		provider, err = manual.NewDNSProvider()
	default:
		provider, err = dns.NewDNSChallengeProviderByName(dnsProviderName)
	}
	if err != nil {
		return nil, err
	}

	if dnsResolver != "" {
		opts := &dns01.Options{RecursiveNameservers: []string{dnsResolver}}
		dns01.SetDefaultClient(dns01.NewClient((opts)))
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return nil, err
	}

	cert, err := client.Certificate.Obtain(context.TODO(), request)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
