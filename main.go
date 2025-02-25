package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/tisba/fritz-tls/internal/fritzbox"
	"github.com/tisba/fritz-tls/internal/fritzutils"
)

var (
	version string
	date    string
	commit  string
)

type configOptions struct {
	authcheck bool

	host            string
	user            string
	adminPassword   string
	insecure        bool
	verificationURL *url.URL

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
	dnsResolver     string
	forceRenew      bool

	version bool
}

func main() {
	config := setupConfiguration()

	unexpired, validDomain, expiry := fritzutils.CheckCertValidity(config.verificationURL, config.domain, 30*24*time.Hour)
	if unexpired && validDomain {
		log.Printf("Current certificate still valid for %s until %s\n", config.domain, expiry.Format(time.RFC850))

		if !config.forceRenew {
			os.Exit(0)
		}
	} else if (!unexpired || !validDomain) && !config.insecure {
		log.Fatalf("Current certificate for %s is not valid anymore, please use --insecure if you want to proceed.", config.verificationURL)
	}

	setupAdminPassword(&config)

	fritz := &fritzbox.FritzBox{
		Host:            config.host,
		User:            config.user,
		Insecure:        config.insecure,
		Domain:          config.domain,
		VerificationURL: config.verificationURL,
	}

	// Login into FRITZ!box
	err := fritz.PerformLogin(config.adminPassword)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Login successful!")

	if config.authcheck {
		sessionOk, err := fritz.CheckSession()
		if err != nil {
			log.Fatal(err)
		}
		if sessionOk {
			os.Exit(0)
		} else {
			log.Fatal("Login did work, but could not verify session!")
		}
	}

	// Have we been ask to get a certificate from Let's Encrypt?
	if config.useAcme {
		// acquire certificate
		cert, err := getCertificate(config.acmeServer, config.domain, config.email, config.dnsProviderName, config.dnsResolver)
		if err != nil {
			log.Fatal(err)
		}

		// save certificate and private key to disk if requested
		if config.saveCert {
			err := os.WriteFile(config.domain+"-key.pem", cert.PrivateKey, 0644)
			if err != nil {
				log.Fatal(err)
			}

			err = os.WriteFile(config.domain+"-cert.pem", cert.Certificate, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}

		config.certificateBundle = io.MultiReader(bytes.NewReader(cert.Certificate), bytes.NewReader(cert.PrivateKey))
	}

	sessionOk, err := fritz.CheckSession()
	if err != nil {
		log.Fatal(err)
	}

	if !sessionOk {
		log.Println("Session expired, re-authenticating...")
		err := fritz.PerformLogin(config.adminPassword)
		if err != nil {
			log.Fatal(err)
		}
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

func setupAdminPassword(config *configOptions) {
	if config.adminPassword == "" {
		config.adminPassword = os.Getenv("FRITZTLS_ADMIN_PASS")
	}

	if config.adminPassword == "" {
		if config.user != "" {
			fmt.Printf("FRITZ!Box Admin Password for %s as %s (will be masked): ", config.host, config.user)
		} else {
			fmt.Printf("FRITZ!Box Admin Password for %s (will be masked): ", config.host)
		}

		config.adminPassword = fritzutils.GetPasswdFromStdin()
	}

	if config.adminPassword == "" {
		log.Fatal("FRITZ!Box admin password required!")
	}
}

func setupConfiguration() (config configOptions) {
	var manualCert bool
	var verificationHost string

	flag.StringVar(&config.host, "host", "http://fritz.box", "FRITZ!Box host")
	flag.StringVar(&config.adminPassword, "password", "", "FRITZ!Box admin password")
	flag.BoolVar(&config.insecure, "insecure", false, "If host is https:// allow insecure/invalid TLS certificates")

	flag.BoolVar(&manualCert, "manual", false, "Provide certificate manually")

	flag.BoolVar(&config.authcheck, "authcheck", false, "Only check if credentials are valid")

	// ACME-mode
	flag.StringVar(&config.acmeServer, "acme-server", "https://acme-v02.api.letsencrypt.org/directory", "Server URL of ACME")
	flag.StringVar(&config.dnsProviderName, "dns-provider", "manual", "name of DNS provider to use")
	flag.BoolVar(&config.saveCert, "save", false, "Save requested certificate and private key to disk")
	flag.StringVar(&config.domain, "domain", "", "Desired FQDN of your FRITZ!Box")
	flag.StringVar(&config.email, "email", "", "Mail address to use for registration at Let's Encrypt")
	flag.StringVar(&config.dnsResolver, "dns-resolver", "", "Resolver to use for recursive DNS queries, supported format: host:port; defaults to system resolver")
	flag.StringVar(&verificationHost, "verification-url", "", "URL to use for certificate validation (defaults to 'host')")
	flag.BoolVar(&config.forceRenew, "force-renew", false, "Forces renewal even if current certificate is still valid for at least 30 days")

	// manual mode
	flag.StringVar(&config.fullchain, "fullchain", "", "path to full certificate chain")
	flag.StringVar(&config.privatekey, "key", "", "path to private key")
	flag.StringVar(&config.bundle, "bundle", "", "path to certificate-private bundle")

	flag.BoolVar(&config.version, "version", false, "Print version and exit")

	flag.Parse()

	config.useAcme = !manualCert

	url, err := url.Parse(config.host)
	if err != nil {
		log.Fatal(err)
	}

	if config.version {
		if version != "" {
			fmt.Printf("fritz-tls %s (%s, %s)\n", version, date, commit)
		} else {
			fmt.Println("fritz-tls 0.0.0-dev")
		}

		os.Exit(0)
	}

	if config.useAcme {
		if config.acmeServer == "" {
			log.Fatal("--acme-server is required without --manual!")
		}

		if config.domain == "" {
			if url.Hostname() != "fritz.box" {
				config.domain = url.Hostname()
			} else {
				log.Fatal("--domain is required without --manual!")
			}
		}

		if config.bundle != "" {
			log.Fatal("--bundle, --fullchain and --key only work with --manual!")
		}
	} else {
		if config.bundle != "" {
			config.certificateBundle = fritzutils.ReaderFromFile(config.bundle)
		} else {
			if config.fullchain == "" || config.privatekey == "" {
				log.Fatal("--fullchain and --key are both required, unless --bundle is used!")
			}

			config.certificateBundle = io.MultiReader(fritzutils.OpenFileWithNewline(config.fullchain), fritzutils.OpenFileWithNewline(config.privatekey))
		}
	}

	config.user = url.User.Username()
	url.User = nil
	config.host = url.String()

	if verificationHost == "" {
		verificationHost = config.host

		config.verificationURL, err = url.Parse(config.host)
		if err != nil {
			log.Fatal(err)
		}

		if config.verificationURL.Scheme == "http" {
			config.verificationURL.Scheme = "https"
		}
	}

	return config
}
