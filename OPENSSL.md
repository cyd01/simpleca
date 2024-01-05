# OpenSSL

```bash
domain=${domain:-localhost}

# Generate a private key and a self-signed certificate
openssl req -x509 -sha256 -days 3650 \
  -newkey rsa:4096 -nodes -keyform PEM -keyout ${domain}.key \
  -outform PEM -out ${domain}.crt \
  -subj "/C=FR/ST=France/L=Paris/O=MyOrg/OU=MyUnit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"
# for multiple alternate names: subjectAltName = DNS:echo-ssl,DNS:mock-ssl

# Generate a self-signed certificate from an existing private key
openssl req -x509 -new -sha256 -days 3650 \
  -key ${domain}.key \
  -outform PEM -out ${domain}.crt \
  -subj "/C=FR/ST=France/L=Paris/O=MyOrg/OU=MyUnit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"

### Generate a private key and a certificate signing request
openssl req -new \
  -newkey rsa:4096 -nodes -keyform PEM -keyout ${domain}.key \
  -outform PEM -out ${domain}.csr \
  -subj "/C=FR/ST=France/L=Paris/O=MyOrg/OU=MyUnit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"
# or
  openssl req -new \
  -newkey rsa:4096 -nodes -keyform PEM -keyout ${domain}.key \
  -outform PEM -out ${domain}.csr \
  -subj "/C=FR/ST=France/L=Paris/O=MyOrg/OU=MyUnit/CN=${domain}"
  
# Generate a certificate signing request from an existing key
openssl req -new \
  -key ${domain}.key \
  -outform PEM -out ${domain}.csr \
  -subj "/C=FR/ST=France/L=Paris/O=MyOrg/OU=MyUnit/CN=${domain}" \
  -addext "subjectAltName = IP:127.0.0.1,DNS:localhost,DNS:${domain}"
  
# Generate a certificate signing request from existing certificate and private key  
openssl x509 \
  -signkey ${domain}.key \
  -in ${domain}.crt \
  -x509toreq -out ${domain}.csr

# Generate a self-signed certificate from an existing private key and CSR
openssl x509 -req -sha256 -days 3650 \
  -signkey ${domain}.key \
  -in ${domain}.csr \
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


> https://www.digitalocean.com/community/tutorials/openssl-essentials-working-with-ssl-certificates-private-keys-and-csrs```