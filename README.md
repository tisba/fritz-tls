<!-- markdownlint-disable MD039 MD041 -->
![Build](https://github.com/tisba/fritz-tls/workflows/Go/badge.svg)
[ ![Go Report Card](https://goreportcard.com/badge/github.com/tisba/fritz-tls)](https://goreportcard.com/report/github.com/tisba/fritz-tls)
<!-- markdownlint-enable MD039 MD041 -->

# FRITZ!Box TLS Certificate Installer

This is a little pet project to install TLS certificates into your [FRITZ!Box](https://en.wikipedia.org/wiki/Fritz!Box). I use [Let’s Encrypt](https://letsencrypt.org/) to get free certificates and I got tired using this tedious process to update the certs all the time. So I started to poke at my FRITZ!Box Fon WLAN 7390 and now it is automated!

Although it should work with other versions as well, it is only tested with:

* FRITZ!Box Fon WLAN 7530 (FRITZ!OS: 07.59)
* FRITZ!Box 7490 (FRITZ!OS: 07.57)

In case you want to know how to do that manually, take a look at AVM's knowledge base articles:

* [FRITZ!Box Fon WLAN 7530](https://en.avm.de/service/knowledge-base/dok/FRITZ-Box-7530/1525_Importing-your-own-certificate-to-the-FRITZ-Box/)
* [FRITZ!Box 7490](https://en.avm.de/service/knowledge-base/dok/FRITZ-Box-7490/1525_Importing-your-own-certificate-to-the-FRITZ-Box/)
* [FRITZ!Box 7390](https://en.avm.de/service/knowledge-base/dok/FRITZ-Box-7390-int/1525_Importing-your-own-certificate-to-the-FRITZ-Box/)

## Installation

Homebrew:

```console
brew install tisba/taps/fritz-tls
```

Go

```console
go get -u github.com/tisba/fritz-tls
```

## Usage

```console
fritz-tls --domain fritz.example.com
```

Done :)

General options for `fritz-tls` are:

* `--help` to get usage information
* `--host` (default: `http://fritz.box`) to specify how to talk to your FRITZ!Box. If you want to login with username and password, specify the user in the URL: `--host http://tisba@fritz.box:8080`. The default username (which is sometimes randomly generated) can be found under `System` > `FRITZ!Box Users`.
* `--password` (optional, default: '') to specify the user's password. If unspecified, `fritz-tls` will prompt the user instead.
* `--insecure` (optional) to skip TLS verification when talking to `--host` in case it's HTTPS and you currently have a broken or expired TLS certificate, or if your FRITZ!Box has its own self-signed certificate.
* `--verification-url` (optional) to specify what URL to use to check certificate installation. Defaults to `--host`.
* `--authcheck` (optional) to only check if the provided credentials are valid.
* `--version` Print `fritz-tls` version and exit. All other options are ignored.

`fritz-tls` can install any TLS certificate or acquire one using [Let's Encrypt](https://letsencrypt.org).

### Let's Encrypt Mode

By default, Let's Encrypt is used to acquire a certificate, options are:

* `--domain` the domain you want to have your certificate generated for (if `--host` is not `fritz.box`, `--domain` it will default to the host name in `--host`).
* `--email` (optional) your mail address you want to have registered with [Let’s Encrypt expiration service](https://letsencrypt.org/docs/expiration-emails/).
* `--save` (optional) to save generated private key and acquired certificate.
* `--dns-provider` (default `manual`) to specify one of [lego's](https://github.com/xenolf/lego/tree/master/providers/dns) supported DNS providers. Note that you might have to set environment variables to configure your provider, e.g. `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` and `AWS_HOSTED_ZONE_ID`. I use name servers by AWS/Route53 and [inwx](https://github.com/xenolf/lego/blob/master/providers/dns/inwx/inwx.go), so I have to provide `INWX_USERNAME`, `INWX_PASSWORD`. I'm not sure if there is a overview, so for now you have to consult the [source](https://github.com/xenolf/lego/tree/master/providers/dns).
* `--dns-resolver` (optional) to specify the resolver to be used for recursive DNS queries. If not provided, the system default will be used. Supported format is `host:port`.
* `--force-renew` to force a renewal, even if the current certificate is valid for the requested domain and still valid for at least the next 30 days.
* `--acme-server` (Optional, default `https://acme-v02.api.letsencrypt.org/directory`) The server URL of the ACME server. Use `https://acme-staging-v02.api.letsencrypt.org/directory` for Let's Encrypt staging environment.

### Manual Certificate Installation

You can also provide a certificate bundle (cert + private key) directly to `fritz-tls` so they can be installed:

1. obtain your TLS certificate, e.g. via [Let’s Encrypt](https://letsencrypt.org/).
1. install the newly generated certificate:

```console
fritz-tls --key=./certbot/live/demo.example.com/privkey.pem --fullchain=./certbot/live/demo.example.com/fullchain.pem
```

* `--manual` to use a locally stored TLS material. This option is required when using either `--key` and `--fullchain` or `--bundle`.
* `--key` and `--fullchain` to provide the private key and the certificate chain.
* `--bundle` as an alternative for `--key` and `--fullchain`. The bundle where the password-less private key and certificate are both present.

## Renew Automation

You can use cron (on Linux) or launchd (on macOS) to run `fritz-tls` automatically. By default, it will check if the cert is still valid and only renew if the remaining validity is less then 30 days. Check out <https://www.launchd.info> to learn how launchd can be used or use <https://launched.zerowidth.com> to generate a plist file.

## TODOs and Ideas

These are some things I'd like to to in the future:

* add validation for private keys and certificate before uploading (avoid trying to upload garbage)
* allow password protected private keys (when not provisioned by LE)

## Make Release

Releases are done via Github Actions on push of a git tag. To make a release, run

```terminal
git tag va.b.c
git push --tags
```
