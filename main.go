package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/tisba/fritz-tls/fritzbox"
)

var (
	version string
	date    string
	commit  string
)

type configOptions struct { // nolint: maligned
	host          string
	user          string
	adminPassword string
	insecure      bool
	tlsPort       int

	fullchain         string
	privatekey        string
	bundle            string
	certificateBundle io.Reader

	useAcme         bool
	acmeServer      string
	saveCert        bool
	domain          string
	email           string
	dnsProviderName string

	version bool
}

func main() {
	config := setupConfiguration()

	fritz := &fritzbox.FritzBox{
		Host:     config.host,
		User:     config.user,
		Insecure: config.insecure,
		Domain:   config.domain,
		TLSPort:  config.tlsPort,
	}

	// Login into FRITZ!box
	err := fritz.PerformLogin(config.adminPassword)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Login successful!")

	// Have we been ask to get a certificate from Let's Encrypt?
	if config.useAcme {
		// acquire certificate
		cert, err := getCertificate(config.acmeServer, config.domain, config.email, config.dnsProviderName)
		if err != nil {
			log.Fatal(err)
		}

		// save certificate and private key to disk if requested
		if config.saveCert {
			err := ioutil.WriteFile(config.domain+"-key.pem", cert.PrivateKey, 0644)
			if err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile(config.domain+"-cert.pem", cert.Certificate, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}

		config.certificateBundle = io.MultiReader(bytes.NewReader(cert.Certificate), bytes.NewReader(cert.PrivateKey))
	}

	// Upload certificate and private key
	status, response, err := fritz.UploadCertificate(config.certificateBundle)
	if err != nil {
		log.Fatal(err)
	}

	if status {
		log.Println("TLS certificate upload successful!")

		suc, err := fritz.VerifyCertificate()
		if err != nil {
			log.Fatal(err)
		}

		if suc {
			log.Println("TLS certificate installation verified!")
		}
	} else {
		log.Fatalf("TLS certificate upload not successful, check response: %s\n", response)
	}
}

func setupConfiguration() configOptions {
	var config configOptions

	flag.StringVar(&config.host, "host", "http://fritz.box", "FRITZ!Box host")
	flag.StringVar(&config.adminPassword, "password", "", "FRITZ!Box admin password")
	flag.BoolVar(&config.insecure, "insecure", false, "If host is https:// allow insecure/invalid TLS certificates")

	flag.BoolVar(&config.useAcme, "auto-cert", false, "Use Let's Encrypt to obtain the certificate")
	flag.StringVar(&config.acmeServer, "acme-server", "https://acme-v02.api.letsencrypt.org/directory", "Server URL of ACME")
	flag.StringVar(&config.dnsProviderName, "dns-provider", "manual", "name of DNS provider to use")
	flag.BoolVar(&config.saveCert, "save", false, "Save requested certificate and private key to disk")

	flag.StringVar(&config.domain, "domain", "", "Desired FQDN of your FRITZ!Box")
	flag.IntVar(&config.tlsPort, "tls-port", 443, "TLS port used by FRITZ!Box (used for verification)")
	flag.StringVar(&config.email, "email", "", "Mail address to use for registration at Let's Encrypt")

	flag.StringVar(&config.fullchain, "fullchain", "", "path to full certificate chain")
	flag.StringVar(&config.privatekey, "key", "", "path to private key")
	flag.StringVar(&config.bundle, "bundle", "", "path to certificate-private bundle")

	flag.BoolVar(&config.version, "version", false, "Print version and exit")

	flag.Parse()

	url, err := url.Parse(config.host)
	if err != nil {
		log.Fatal(err)
	}

	if config.version {
		log.Printf("fritz-tls %s (%s, %s)", version, date, commit)
		os.Exit(0)
	}

	if config.useAcme {
		if config.acmeServer == "" {
			log.Fatal("--acme-server is required with --auto-cert!")
		}

		if config.domain == "" {
			if url.Hostname() != "fritz.box" {
				config.domain = url.Hostname()
			} else {
				log.Fatal("--domain is required with --auto-cert!")
			}
		}

		if config.email == "" {
			log.Fatal("--email is required with --auto-cert!")
		}

		if config.bundle != "" {
			log.Fatal("--auto-cert, --bundle, --fullchain and --privatekey are mutually exclusive!")
		}
	} else {
		if config.bundle != "" {
			config.certificateBundle = readerFromFile(config.bundle)
		} else {
			if config.fullchain == "" || config.privatekey == "" {
				log.Fatal("--fullchain and --privatekey are both required, unless --bundle is used!")
			}

			config.certificateBundle = io.MultiReader(readerFromFile(config.fullchain), readerFromFile(config.privatekey))
		}
	}

	config.user = url.User.Username()
	url.User = nil
	config.host = url.String()

	if config.tlsPort == 0 && url.Port() != "" {
		config.tlsPort, err = strconv.Atoi(url.Port())
		if err != nil {
			log.Fatal(err)
		}
	}

	if config.adminPassword == "" {
		config.adminPassword = os.Getenv("FRITZTLS_ADMIN_PASS")
	}

	if config.adminPassword == "" {
		if config.user != "" {
			fmt.Printf("FRITZ!Box Admin Password for %s as %s (will be masked): ", config.host, config.user)
		} else {
			fmt.Printf("FRITZ!Box Admin Password for %s (will be masked): ", config.host)
		}

		config.adminPassword = getPasswdFromStdin()
	}

	if config.adminPassword == "" {
		log.Fatal("FRITZ!Box admin password required!")
	}

	return config
}
