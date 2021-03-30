<!-- markdownlint-disable MD039 MD041 -->
![Build](https://github.com/tisba/fritz-tls/workflows/Go/badge.svg)
[ ![Go Report Card](https://goreportcard.com/badge/github.com/tisba/fritz-tls)](https://goreportcard.com/report/github.com/tisba/fritz-tls)
<!-- markdownlint-enable MD039 MD041 -->

# FRITZ!Box TLS Certificate Installer

This is a little pet project to install TLS certificates into your [FRITZ!Box](https://en.wikipedia.org/wiki/Fritz!Box). I use [Let’s Encrypt](https://letsencrypt.org/) to get free certificates and I got tired using this tedious process to update the certs all the time. So I started to poke at my FRITZ!Box Fon WLAN 7390 and now it is automated!

Although it should work with other versions as well, it is only tested with:

* FRITZ!Box Fon WLAN 7530 (FRITZ!OS: 07.25)
* FRITZ!Box 7490 (FRITZ!OS: 07.21)

In case you want to know how to do that manually, take a look at AVM's [knowledge base article](https://en.avm.de/service/fritzbox/fritzbox-7390/knowledge-base/publication/show/1525_Importing-your-own-certificate-to-the-FRITZ-Box/).

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
fritz-tls --auto-cert --domain fritz.example.com --email letsencrypt@example.com
```

Done :)

You can also provide a certificate bundle (cert + private key) directly so that can be installed:

1. obtain your TLS certificate, e.g. via [Let’s Encrypt](https://letsencrypt.org/).
1. install the newly generated certificate:

```console
fritz-tls --key=./certbot/live/demo.example.com/privkey.pem --fullchain=./certbot/live/demo.example.com/fullchain.pem
```

General options are:

* `--help` to get usage information
* `--host` (default: `http://fritz.box`) to specify how to talk to your FRITZ!Box. If you want to login with username and password, specify the user like this: `--host http://tisba@fritz.box`.
* `--insecure` (optional) to skip TLS verification when talking to `--host` in case it's HTTPS and you currently have a broken or expired TLS certificate.
* `--tls-port` (default: `443`) TLS port of FRITZ!Box. This is used for certificate validation after installing.

Let's Encrypt specific (`--auto-cert`) options are:

* `--domain` the domain you want to have your certificate generated for (if `--host` is not `fritz.box` it will default to the host name in `--host`)
* `--email` your mail address you want to have registered with Let’s Encrypt
* `--save` (optional) to save generated private key and acquired certificate
* `--dns-provider` (default `manual`) to specify one of [lego's](https://github.com/xenolf/lego/tree/master/providers/dns) supported DNS providers. Note that you might have to set environment variables to configure your provider, e.g. `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` and `AWS_HOSTED_ZONE_ID`. I use name servers by AWS/Route53 and [inwx](https://github.com/xenolf/lego/blob/master/providers/dns/inwx/inwx.go), so I have to provide `INWX_USERNAME`, `INWX_PASSWORD`. I'm not sure if there is a overview, so for now you have to consult the [source](https://github.com/xenolf/lego/tree/master/providers/dns).

Options for non `--auto-cert` mode:

* `--bundle` as an alternative for `--key` and `--fullchain`. The bundle where the password-less private key and certificate are both present.

## TODOs and Ideas

These are some things I'd like to to in the future:

* add validation for private keys and certificate before uploading (avoid trying to upload garbage)
* allow password protected private keys (when not provisioned by LE)
* ~~if `--tls-port` is not given, we should try to use `--host` before failing~~
* ~~add homebrew as a release target for goreleaser~~
* ~~ask for `--user` if not provided (may be empty then) and/or add `--pw-only` flag~~
* ~~allow other then DNS-01 Let's Encrypt challenges and make [legos](https://github.com/xenolf/lego) DNS providers available to make things even more automated!~~
* ~~add `--insecure` to ignore invalid TLS certificates when talking to FRITZ!Box~~
* ~~read FRITZ!Box administrator password from environment~~
* ~~add ability to use already combined private keys and certificate files~~
* ~~add basic Let's Encrypt support~~
* ~~improve detection if certificate installation was successful; currently I'm looking for a string in the response. But maybe we can just wait a little bit and make a https request and check if the certificate is actually being used.~~
* ~~implement FRITZ!Box authentication for user name and password~~
* ~~set up Travis and use [GoReleaser](https://github.com/goreleaser/goreleaser) to build and publish builds~~

## Make Release

Releases are done via Github Actions on push of a git tag. To make a release, run

```terminal
git tag va.b.c
git push --tags
```
