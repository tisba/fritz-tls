# FRITZ!Box TLS Certificate Installer

This is a little pet project to install TLS certificates into your [FRITZ!Box](https://en.wikipedia.org/wiki/Fritz!Box). I use [Let’s Encrypt](https://letsencrypt.org/) to get free certificates and I got tiered using this tedious process to update the certs all the time. So I started to poke at my FRITZ!Box Fon WLAN 7390 and now it is automated!

Although it should work with other versions as well, it is only tested with:

* FRITZ!Box Fon WLAN 7390 (FRITZ!OS: 06.83)
* FRITZ!Box 7490 (FRITZ!OS: 06.90)

In case you want to know how to do that manually, take a look at AVM's [knowledge base article](https://en.avm.de/service/fritzbox/fritzbox-7390/knowledge-base/publication/show/1525_Importing-your-own-certificate-to-the-FRITZ-Box/). 


## Installation

```
go get -u github.com/tisba/fritz-tls
```


## Usage

1. obtain your TLS certificate, e.g. via [Let’s Encrypt](https://letsencrypt.org/).
1. install the newly generated certificate:

```
fritz-tls --key=./certbot/live/demo.example.com/privkey.pem --fullchain=./certbot/live/demo.example.com/fullchain.pem
```

Other options are

* `--help` to get usage information
* `--host` to specify a different host for your FRITZ!Box (default: `http://fritz.box`)


To give you an idea, this is roughly the process I currently use (I prefer the DNS-based challenge):

```
certbot -d demo.example.com --manual --preferred-challenges dns certonly && \
  fritz-tls --key=./live/demo.example.com/privkey.pem --fullchain=./live/demo.example.com/fullchain.pem
```


## TODOs and Ideas

These are some things I'd like to to in the future:

* read FRITZ!Box administrator password from environment
* implement FRITZ!Box authentication for user name and password
* integrate with cerbot's `--deploy-hook`
* add `--insecure` to ignore invalid TLS certificates when talking to FRITZ!Box
* add validation for private keys and certificate before uploading
* improve detection if certificate installation was successful; currently I'm looking for a string in the response. But maybe we can just wait a little bit and make a https request and check if the certificate is actually being used.
* set up Travis and use [GoReleaser](https://github.com/goreleaser/goreleaser) to build and publish builds
* add ability to use already combined private keys and certificate files
* allow password protected private keys
