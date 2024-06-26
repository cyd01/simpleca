# OpenSSL

## Command lines

```bash
domain=${domain:-localhost}

# Generate CA key and certificate
openssl genrsa -out easyca.key 4096
openssl req -x509 -sha512 -days 3650 \
  -new -noenc \
  -key easyca.key \
  -outform PEM -out easyca.crt \
  -addext "basicConstraints = critical, CA:TRUE" \
  -addext "subjectKeyIdentifier = hash" \
  -addext "authorityKeyIdentifier = keyid:always, issuer" \
  -addext "keyUsage = critical, keyCertSign, cRLSign, digitalSignature" \
  -subj "/C=FR/ST=France/L=Paris/O=Orga/OU=Unit/CN=EasyCA"

# Remove/Add CA certificate from/to system truststore
test ! -f /usr/local/share/ca-certificates/demo.crt || { sudo rm -f /usr/local/share/ca-certificates/demo.crt && sudo update-ca-certificates -f ; }
sudo cp easyca.crt /usr/local/share/ca-certificates/demo.crt && sudo update-ca-certificates
# The CA trust store as generated by update-ca-certificates is available at the following locations:
# - As a single file (PEM bundle) in /etc/ssl/certs/ca-certificates.crt
# - As an OpenSSL compatible certificate directory in /etc/ssl/certs

# Generate a private key and a self-signed certificate
openssl req -x509 -sha512 -days 3650 \
  -newkey rsa:4096 -nodes -keyform PEM -keyout ${domain}.key \
  -outform PEM -out ${domain}.crt \
  -subj "/C=FR/ST=France/L=Paris/O=Orga/OU=Unit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"
# for multiple alternate names: subjectAltName = DNS:echo-ssl,DNS:mock-ssl

# Generate a self-signed certificate from an existing private key
openssl req -x509 -new -sha512 -days 3650 \
  -key ${domain}.key \
  -outform PEM -out ${domain}.crt \
  -subj "/C=FR/ST=France/L=Paris/O=Orga/OU=Unit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"

### Generate a private key and a certificate signing request
openssl req -new \
  -newkey rsa:4096 -nodes -keyform PEM -keyout ${domain}.key \
  -outform PEM -out ${domain}.csr \
  -subj "/C=FR/ST=France/L=Paris/O=Orga/OU=Unit/CN=${domain}" \
  -addext "basicConstraints = CA:FALSE" \
  -addext "subjectKeyIdentifier= hash" \
  -addext "keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"
# or
  openssl req -new \
  -newkey rsa:4096 -nodes -keyform PEM -keyout ${domain}.key \
  -outform PEM -out ${domain}.csr \
  -subj "/C=FR/ST=France/L=Paris/O=Orga/OU=Unit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"
  
# Generate a certificate signing request from an existing key
openssl req -new \
  -key ${domain}.key \
  -outform PEM -out ${domain}.csr \
  -subj "/C=FR/ST=France/L=Paris/O=Orga/OU=Unit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"
  
# Generate a certificate signing request from existing certificate and private key  
openssl x509 \
  -signkey ${domain}.key \
  -in ${domain}.crt \
  -x509toreq -out ${domain}.csr

# Generate a self-signed certificate from an existing private key and CSR
openssl x509 -req -sha512 -days 3650 \
  -signkey ${domain}.key \
  -in ${domain}.csr \
  -outform PEM -out ${domain}.crt

# Extract public key from a private key
openssl rsa -pubout \
  -in ${domain}.key \
  -out ${domain}.pub

# Sign a certificate signing request against a CA
openssl x509 -req -days 3650 -sha512 \
  -CA easyca.crt -CAkey easyca.key -CAcreateserial \
  -in ${domain}.csr \
  -copy_extensions copy \
  -outform PEM -out ${domain}.crt


# Verify private key
openssl rsa -check -in ${domain}.key

# Read certificate signing request
openssl req -text -noout -verify -in ${domain}.csr

# Read certificate
openssl x509 -text -noout -in ${domain}.crt

# Encrypt a private key
openssl rsa -des3 \
  -in ${domain}.key \
  -outform PEM -out ${domain}.enckey -passout pass:password

# Decrypt a private key
openssl rsa \
  -in ${domain}.enckey -passin pass:password \
  -out ${domain}.key

# Convert PEM certificate and private key to PKCS12
openssl pkcs12 \
  -inkey ${domain}.key \
  -in ${domain}.crt \
  -export -out ${domain}.p12

# Convert PKCS12 to PEM
openssl pkcs12 \
  -in ${domain}.p12 \
  -nodes -out ${domain}.combined.crt
  # ouput file will contain all items
```
> https://www.digitalocean.com/community/tutorials/openssl-essentials-working-with-ssl-certificates-private-keys-and-csrs  
> https://www.golinuxcloud.com/add-x509-extensions-to-certificate-openssl/

## Makefile

```Makefile
MAKEFILE_NAME:=$(notdir $(lastword $(MAKEFILE_LIST)))
TARGET=$(notdir $(abspath $(lastword $(MAKEFILE_LIST)/..)))
MAKETARGETS=$(filter-out $@,$(MAKECMDGOALS))

include variables.mk
-include .makerc
-include ~/.makerc

## Default Domain name
DOMAIN ?= localhost

## Default private key size
SIZE ?= 4096

## Default CA days duration
CANBDAYS ?= 3650

## Default certificate days duration
NBDAYS ?= 3650

## Country code
C ?= FR

## State
ST ?= France

## Location
L ?= Paris

## Organization
O ?= Orga

## Organization Unit
OU ?= Unit

## CA common name
CACN ?= EasyCA

## Alternative names (space separator)
ALTNAME ?= localhost

## IP addresses (space separator)
IP ?= 

## Openssl path
OPENSSL ?= $(shell which openssl 2> /dev/null)

MAKER=$(shell test -n "$${USER}" && echo $${USER} || { test -n "$${USERNAME}" && echo $${USERNAME} || { which logname > /dev/null 2>&1 && logname || echo Unknown ; } } )

## convinient variable to define current date/time which can be used for build time
NOW=$(shell date --rfc-3339=seconds)
export NOW
DEBUG = printf -- "| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "DEBUG"
INFO = printf -- "\e[92m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "INFO"
WARN = printf -- "\e[93m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "WARN"
ERROR = printf -- "\e[91m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "ERROR"
FATAL = printf -- "\e[31m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "FATAL"

## Show this help prompt.
.PHONY: all
all : help
	@:

## Show this help prompt.
.PHONY: help
help:
	@ echo '  Easy server certificate generator'
	@ echo
	@ echo '  Usage:'
	@ echo ''
	@ echo '    [flags...] make <target>'
	@ echo ''
	@ echo '  Targets:'
	@ echo ''
	@ (echo '   Name:Description'; echo '   ----:-----------'; (awk -F: '/^## /{ comment = substr($$0,4) } comment && /^[a-zA-Z][a-zA-Z ]*[^ ]+:/{ print "   " substr($$1,0,80) ":" comment }' $(MAKEFILE_LIST) | sort -d)) | column -t -s ':'
	@ echo ''
	@ echo '  Flags:'
	@ echo ''
	@ (echo '   Name?=Default value?=Description'; echo '   ----?=-------------?=-----------'; (awk -F"\?=" '/^## /{ comment = substr($$0,4) } comment && /^[a-zA-Z][a-zA-Z0-9_-]*[ ]+\?= /{ print "   " $$1 "?=" substr($$2,0,80) "?=" comment }' $(MAKEFILE_LIST) 2>/dev/null | sort -d)) | sed -e 's/\?= /?=/g' | column -t -s '?='
	@ echo ''

## Will display value of variable.
debug/%:
	@test -z "$(wordlist 2,3,$(subst /, ,$*))" && echo '$*=$($*)' || $(MAKE) -C "$(*:$(firstword $(subst /, ,$*))/%=%)" "debug/$(firstword $(subst /, ,$*))"

## Make key and certificate for certificates authority (ca.key and ca.crt)
.PHONY: ca
ca: ca.crt

.PHONY: ac
ac : ca

ca.key :
	@echo "Generating private key <ca.key> for certificates authority $(CACN)"
	@$(OPENSSL) genrsa -out ca.key $(SIZE)

ca.crt : ca.key
	@echo "Generating certificate <ca.crt> for certificates authority $(CACN)"
	@$(OPENSSL) req -x509 -sha512 -days $(CANBDAYS) \
	-new -noenc \
	-key ca.key \
	-outform PEM -out ca.crt \
	-addext "basicConstraints = critical, CA:TRUE" \
	-addext "subjectKeyIdentifier = hash" \
	-addext "authorityKeyIdentifier = keyid:always, issuer" \
	-addext "keyUsage = critical, keyCertSign, cRLSign, digitalSignature" \
	-subj "/C=$(C)/ST=$(ST)/L=$(L)/O=$(O)/OU=$(OU)/CN=$(CACN)"

## Read certificate (readcrt <filename>)
.PHONY: readcrt
readcrt:
	@$(MAKE) --no-print-directory readcrt/$(filter-out $@,$(MAKECMDGOALS))

.PHONY: readcrt/%
readcrt/% :
	@test -z "$(wordlist 2,3,$(subst /, ,$*))" && { echo "Reading certificate <$*>" ; $(OPENSSL) x509 -text -noout -in $* ; }

## Generate a private key and a certificate signing request for a domain (req <domain>)
.PHONY: req
req:
	@DOMAIN=$(filter-out $@,$(MAKECMDGOALS)) $(MAKE) --no-print-directory csr

.PHONY: domainkey
domainkey :
	@echo "Generating private key <$(DOMAIN).key> for domain $(DOMAIN)"
	@$(OPENSSL) genrsa -out $(DOMAIN).key $(SIZE)

.PHONY: csr
csr :
	@test -f $(DOMAIN).key || $(MAKE) --no-print-directory domainkey
	@echo "Generating certificate signing request <domain.csr> for domain $(DOMAIN)"
	@for ip in $(IP);do IPS="$${IPS},IP:$${ip}";done;for name in $(ALTNAME);do NAMES="$${NAMES},DNS:$${name}";done;$(OPENSSL) req -new \
	-key $(DOMAIN).key \
	-outform PEM -out $(DOMAIN).csr \
	-subj "/C=$(C)/ST=$(ST)/L=$(L)/O=$(O)/OU=$(OU)/CN=$(DOMAIN)" \
	-addext "subjectAltName = IP:127.0.0.1$${IPS},DNS:$(DOMAIN)$${NAMES}"

## Read certificate signing request (readcsr <filename>)
.PHONY: readcsr
readcsr:
	@$(MAKE) --no-print-directory readcsr/$(filter-out $@,$(MAKECMDGOALS))

.PHONY: readcsr/%
readcsr/% :
	@test -z "$(wordlist 2,3,$(subst /, ,$*))" && { echo "Reading certificate signing request <$*>" ; $(OPENSSL) req -text -noout -verify -in $* ; }

## Sign a certificate signing request (sign <domain>)
.PHONY: sign
sign: ca
	@test -f $(filter-out $@,$(MAKECMDGOALS)).csr || $(MAKE) --no-print-directory req $(filter-out $@,$(MAKECMDGOALS))
	@echo "Signing certificate signing request"
	@$(OPENSSL) x509 -req -days $(NBDAYS) -sha512 \
	-CA ca.crt -CAkey ca.key -CAcreateserial \
	-in $(filter-out $@,$(MAKECMDGOALS)).csr \
	-copy_extensions copy \
	-outform PEM -out $(filter-out $@,$(MAKECMDGOALS)).crt


## Clean all or clean domain (clean <domain>)
.PHONY: clean

clean:
	@echo "Cleaning..."
	-@$(MAKE) --no-print-directory clean/$(filter-out $@,$(MAKECMDGOALS))
	-@rm -f ca.key ca.crt domain.key domain.crt

.PHONY: clean/%
clean/% :
	-@test -z "$(wordlist 2,3,$(subst /, ,$*))" && { rm -f $*.key $*.csr $*.crt ; }


# Default target
%:
	@:
```
