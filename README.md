# SimpleCA: a simple Go certificate authority

## Usage

```bash
$ simpleca

Usage:  simpleca COMMAND [OPTIONS]

A simple tools to manage SSL stuff

Commands:
  acme             Start an ACME certificate authority web server
  ca               Manage certificate authority
  cert             Manage server certificates
  csr              Manage server certificate signing request
  key              Manage keys
  web              Start an automatic certificate authority web server

Run 'simpleca COMMAND -h' for more informations on a command
```

## How to build

Every operations can be achieve with the provided [Makefile](Makefile):

```bash
$ make

  Usage:

    [flags...] make <target>

  Targets:

   Name       Description                                                         
   ----       -----------                                                         
   buildall   Compile for both Linux, MacOS and Windows target                    
   build      Compile for current target                                          
   clean      Will clean workspace                                                
   compile    Compile locally                                                     
   compress   Compress all targets                                                
   darwin     Compile for MacOS target                                            
   debug/%    Will display value of variable.                                     
   help       Show this help prompt.                                              
   image      Make docker image                                                   
   imagerun   Run a docker image                                                  
   install    Local installation                                                  
   install/%  SCP remote installation                                             
   linux      Compile for Linux target                                            
   log/%      Print service logs                                                   (replace % with service name)
   package    Upload binaries to Gitlab                                           
   prepare    Prepare environment                                                 
   push       Upload binaries to Nexus                                            
   run        Run binary                                                          
   test       Run tests                                                           
   tidy       Ensures that the go.mod file matches the source code in the module  
   update     Update dependencies                                                 
   vendor     Update vendor                                                       
   windows    Compile for Windows target    

  Flags:

   Name                      Default value        Description
   ----                      -------------        -----------
   DOCKER_IMAGE_TAG          latest               Docker target image tag
   DOCKER_IMAGE              $(TARGET)            Docker target image
   DOCKER_RUN_PARAMETERS     --publish 8080:80    Docker run additional parameters
   GITLAB_PACKAGE            $(DOCKER_IMAGE)      Gitlab destination package
   TARGET                    simpleca             Binary
```

### Binary build

`make build` or

```bash
go get && go mod tidy && go build
```

### Docker build

`make image` or

```bash
docker build . --file Dockerfile --tag simpleca
```

## How to use command-line mode client

### How to create a certificate authority

```bash
$ simpleca ca create -h

Usage:  simpleca ca create [OPTIONS]

Create a new certificate authority

Options:
  -C string
    	Country name (default FR)
  -CN string
    	Common name (default MyCA)
  -L string
    	Locality (default Paris)
  -O string
    	Organization (default MyOrg)
  -OU string
    	Unit (default MyUnit)
  -ST string
    	State (default France)
  -days int
    	Not valid after days (default 3650)
  -key, -k string
    	Private key file (default -)
  -out, -c string
    	Output file (- for standard output) (default -)
  -passphrase string
    	Private key passphrase
  -size, -s int
    	Private key size (default 2048)
```

Example:

```bash
simpleca ca create -key ca.key -out ca.crt
```

### How to make a private key

```bash
$ ./simpleca key create -h

Usage:  simpleca key create [OPTIONS]

Create a new private key

Options:
  -out string
    	Output file (- for standard output) (default -)
  -passphrase string
    	Private key passphrase
  -size int
    	Private key size (in bits (default 2048)
  -type string
    	Private key type (rsa, or ecdsa) (default rsa)
``` 

Example:

```bash
simpleca key create -out localhost.key
```

### How to make a certificate signing request

```bash
$ simpleca csr create -h

Usage:  simpleca csr create [OPTIONS]

Create a new certificate signing request

Options:
  -C string
    	Country name
  -CN string
    	Common name
  -L string
    	Locality
  -O string
    	Organization
  -OU string
    	Unit
  -ST string
    	State
  -alt-names, -a string
    	Coma separated alternate names list
  -ips, -i string
    	Coma separated IP addresses list
  -key, -k string
    	Server private key file (default -)
  -out string
    	Output file (- for standard output) (default -)
  -passphrase string
    	Server private key passphrase
  -size, -s int
    	Private key size (default 2048)
```

Example:

```bash
simpleca csr create -CN localhost -L Paris -OU MyUnit -alt-names localhost,www.localhost.com,test.localhost.com -ips 127.0.0.1 -key localhost.key -out localhost.csr
```

```log
Loading private key
Generating certificate sign request
```

### How to sign a previousy generated certificate signing request

```bash
$ simpleca ca sign -h

Usage:  simpleca ca sign [OPTIONS] FILENAME

Sign a certificate signing request

Options:
  -ca-cert string
    	Certificate of the certificate authority (default ca.crt)
  -ca-key string
    	Private key of the certificate authority (default ca.key)
  -ca-pass string
    	Private key passphrase of the certificate authority
  -days int
    	Not valid after days (default 3650)
  -out, -c string
    	Output file (- for standard output) (default -)
```

Example:

```bash
simpleca ca sign -ca-cert ca.crt -ca-key ca.key -out localhost.crt localhost.csr
```

### How to read a certificate

```bash
$ simpleca cert read localhost.crt
Certificate:
    Data:
        Version: 3
        Serial Number: 276140968835945637165011235300300158387 (0xcfbed0ea500009a575b0b5ea13e9f5b3)
        Signature Algorithm: SHA512-RSA
        Issuer: C = [FR] , ST = [France]  L = [Paris]  O = [MyOrg]  OU = [MyUnit]
        Validity:
            Not Before:  2022-09-23 17:40:49 +0000 UTC
            Not After :  2032-09-20 17:40:49 +0000 UTC
        Subject: C = [FR] , ST = [France]  L = [Paris]  O = [MyOrg]  OU = [MyUnit]
        Subject Public Key Info:
            Public Key Algorithm: RSA
                RSA Public-Key: (2048 bit)
                Modulus:
                    00:c6:64:af:30:0e:d2:3e:63:67:92:37:23:13:80:
                    4a:2e:5b:50:e8:29:42:95:08:f3:4b:e8:b6:cf:0e:
                    de:35:35:5e:24:9c:de:4e:95:11:91:27:b2:c4:40:
                    95:8b:5e:e6:a0:6d:f7:b7:f4:22:73:4d:c2:85:3f:
                    7d:da:0f:30:bb:a9:e9:56:84:fe:85:db:75:9e:df:
                    a6:3a:16:72:ec:e7:ca:47:a1:d5:58:05:f9:0b:d7:
                    a5:68:90:bc:cc:c1:71:f2:31:2f:06:fe:fe:52:63:
                    89:c8:11:b2:39:75:67:45:7e:5d:b8:ce:dc:d2:f6:
                    e5:88:b3:b4:4a:05:75:c8:f4:27:48:99:f0:55:61:
                    2e:8d:b4:cc:a1:f0:bd:9c:63:3c:fe:5a:c5:36:3b:
                    76:5b:bb:f0:ea:1f:f8:f5:5e:b8:06:e3:d5:82:07:
                    2a:7b:60:cb:cc:ab:8a:9f:35:de:3b:f1:5f:a9:cf:
                    75:42:fa:60:f5:e9:4b:2f:16:89:a0:dd:c7:10:e1:
                    d8:8c:7e:2e:1f:c4:52:d4:7e:d4:b0:72:2f:5c:f2:
                    2b:a1:da:52:fe:bf:f3:f2:fd:15:65:cb:1c:f4:ef:
                    70:52:0e:cf:f9:38:fd:e9:f7:c0:5b:71:43:58:d4:
                    4a:66:8f:b6:91:82:78:2c:a0:29:3a:9f:9e:d5:8b:
                    96:5b:
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            2.5.29.15: critical
            2.5.29.37:
            2.5.29.35:
            2.5.29.17:
            KeyUsage:
                Digital Signature, Cert Sign,
    Signature Algorithm: SHA512-RSA
         56:bf:79:2c:08:6b:ed:3a:c0:a9:73:5e:3d:73:68:77:5e:40:
         62:39:2f:e0:50:98:88:62:b5:01:09:a7:56:12:11:67:8f:b7:
         03:78:df:c3:48:10:b3:6c:21:a5:2e:9e:88:c5:57:b9:27:e6:
         bf:5f:6a:18:4e:78:84:fd:43:32:c1:86:33:4d:78:ef:84:de:
         39:80:aa:5b:0a:f2:ca:a1:17:37:a7:22:dd:51:c1:f6:bd:58:
         ef:d9:2b:6a:00:5c:0d:62:75:f7:2f:31:a0:23:ad:01:49:76:
         c0:ea:79:d2:bb:f1:e3:27:58:a6:19:5b:3b:b4:d8:5d:0c:98:
         6c:11:44:17:a4:c0:6a:24:ae:65:5a:0d:cc:ad:23:ff:d1:d2:
         a9:f8:00:cc:c4:9b:eb:81:f6:c1:a7:1e:14:b6:fe:82:63:07:
         81:8f:33:e9:9d:26:77:ce:ae:d1:f5:a4:0b:e0:ae:27:d7:72:
         2f:a6:86:a5:d4:45:3f:67:b0:46:48:8d:bb:cf:cd:dd:11:4d:
         05:6c:fb:99:4b:7b:92:3d:d5:70:28:35:72:46:8c:ff:ba:5e:
         ed:f8:06:7d:60:16:87:5c:8b:b9:fb:3c:52:f3:b6:c1:67:17:
         c4:03:81:e6:8b:84:dd:e0:66:af:42:c7:25:1b:a7:1e:01:45:
         64:6f:c6:20:
```

## How to use web server mode

All services are described in [swagger file](swagger.yaml).

### How to start

```bash
$ simpleca web -h
Usage of simpleca:
  -ca-cert string
        Certificate of the certificate authority (default ca.crt)
  -ca-key string
        Private key of the certificate authority (default ca.key)
  -ca-pass string
        Private key passphrase of the certificate authority
  -dir string
        Root directory (default .)
  -port string
        Port server (default :80)
```

Example:

```bash
simpleca web -ca-cert ca.crt -ca-key ca.key
```

### How to get the certificate authority liveness status

```bash
curl -s http://127.0.0.1/alive
```

### How to get the certificate authority certificate

```bash
curl -s http://127.0.0.1/ca/ca.crt
```

### How to generate private key

```bash
curl -s http://127.0.0.1/key
```

### How to extract public key from an existing private key

The private key is in `localhost.key` file.

```bash
curl -s -H "Content-type: application/octet-stream" http://127.0.0.1/key/pub --data-binary "@localhost.key"
```

### How to generate a certificate signing request from a private key

The private key is in `localhost.key` file.

```bash
curl -s -H "Content-type: application/octet-stream" http://127.0.0.1/csr --data-binary "@localhost.key"
```

### How to sign a certificate signing request

The certificate signing request is in `localhost.csr` file.

```bash
curl -s -H "Content-type: application/octet-stream" -X POST http://127.0.0.1/sign --data-binary "@localhost.csr"
```

### How to get all-in-one (key, csr, certificate)

```bash
curl -s http://127.0.0.1/crt?CN=localhost
```

### Output types

In all API calls results, are received in `plain/text`. By setting the header `Accept` to `application/json` the result will be send in JSON format.  
Example with private key:

```bash
curl -s -H 'Accept: application/json' http://127.0.0.1/key
```

```json
{
  "key": {
    "priv": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAoxgHjb0QzB904MMZBG5X/yw1On4+5gRifYiNIu6MiXgry0oC\noCIqgIM9tR9s7V9I/ynTh8dCbnTkqT3XR9cTL0eVUPq0cAyVdKNMNZM99HOICtai\nsmnSOY0SY6Dql7fqhZfmQwwWg6uriy/XNgGuOOVq+Tqjrabva8DWYoWKJkLvO38x\nIuLr4333XjomkcfHaFQTF3Rmsbp7tzCMIIM5FA9I+77pFR7IvMZQbK0yb9ZNPz3T\nlaKFsP8SUx7doUmxUglmLlZyttQwd6cOeykoKyIVcpC3T1f42vRz5+aLmza4kPWx\nTHT3YOxKTD7nK8gQUx5Gl1ls0AVlyXBryFb/fQIDAQABAoIBAEDAEoeoT1nrBzkQ\n3AbRVChfwcY2RvyFMyEJrJb4xFzzk9eCy6YNynf5Iu+cyV84XD/JuEyIxIwb5oh2\nn9jKv7geoF5lGmv94vxKtL/0dD1v/MvoyPEyaB8nLezV/y06/GYLl4R48RtXdaSt\n2rB7XFMEakMGj+MqipVqGvNEd0OW3w0d7U1E/GXW+1WOt+zQGozDxObvfs/DcRes\nZGqCrjXXN3cVP1y+0Ta6Dxx/ZmQYeu4X+BYpixh8CrFZNn/fAJd8voKZUlZKIoBw\nDZCXkv4Xsi5rekHXxPA98N2SNt59FG7puPb5iPoLuqOHZ6Aed0rAL8QfIIYitOjH\nreg5IaECgYEAxMyzT6WSluh9ne+0StxbL2c8bd84JjHdg7CrznaBFI8Fgs5WCJd5\nAEuxKcuzFpabFBAa2Lw4J5scpxBY47QyicnfIYqTELXQLLRTvRWp4LIK5VvffQ9u\nc/ApJBGUEEaTnXQyfwPzn+BPVvMoReszZSMGfLQFLXkOeTDdb6+I+mUCgYEA1Cew\nwCw/bAtyyNzOq/H0BgZc57bb4bzkSltV4IPA5zpHDFYJnxxoR68Gq2ZwoUof+B0I\nsDqybPEqAWtUmne58gNQ7OmA4cRI8tMSeqoh5tQBc2zlEny8hJelc1JAxG1c6l7F\nQ4iuDO/f/xLRblIVVrBPQXAvS7br9AyUYAg+0zkCgYBc+aCVNlIE/Z2rKf3xiB2E\nTa+c8OJkGRbgCm2AwzfEcLVX0QeQU4+U9i2i41uehlSJq/oi/vlArOVigDSejxl5\nQ2gpPoCwWlUZabBOGpgBUdwX88moGcxC9elZ4vwinFVNBAJ/Q2yr0ZtqJsKWFcJY\nO63q6Fmx3AlcdBuJklKSiQKBgQCHsIr/nk1CEWBsz6zvlGR6pf8txGqFyoZIeHpI\ninwKZ9+hKDYnKcgYcP1XCsHmpr4jto4kCKatvuEa30bRNNocy7oqjH3958iwZgdf\npQjh1Z7H8FHirRz1wPf09hquhzPyQoLwWq7XX2Rog+SnJqC3PTSzqcjWKDxpbtJH\nSX7FIQKBgGVO06lkxkxjYV6Xj8q40cyEBm/86tIAwDumH7Fywc0pA26ftX/W0wx1\n7XiXQbfC3LIDo76NIjyX6wPJlinV1oq+oC50f17LVVPQEvmjEWJiadWGkKrm8xgu\nphKL8m4b9MAux4EZkoV/pO7Z0RfyVGnoPrEbwmGQGBBPmq4qmz1N\n-----END RSA PRIVATE KEY-----\n"
  }
}
```

## How to start an ACME mock web server

```bash
$ simpleca acme -h
Usage of simpleca:
  -ca-cert string
    	Certificate of the certificate authority (default ca.crt)
  -ca-key string
    	Private key of the certificate authority (default ca.key)
  -ca-pass string
    	Private key passphrase of the certificate authority
  -cert string
    	Certificate of the ACME web server (if ssl enabled)
  -days int
    	Not valid after days
  -key string
    	Private key of the ACME web server (if ssl enabled)
  -port string
    	Port server (default :8080)
  -ssl
    	Enable SSL server mode
```

> This ACME mock server does not make any verification (no challenge). It always respond a valid certificate on each request.

Example:

```bash
simpleca acme -ssl -port 127.0.0.1:1443 -key acme.key -cert acme.crt -ca-key ca.key -ca-cert ca.crt
```

```log
Starting ACME web server on port 127.0.0.1:1443 ...
```

The ACME directory service is accessible at `https://127.0.0.1:1443/directory`

## How to use docker mode

```bash
docker run --interactive --tty --rm --publish 1443:443 --volume $(pwd)/ca.key:/ca/ca.key --volume $(pwd)/ca.crt:/ca/ca.crt --name simpleca simpleca
```

## How to manage AC in system

### Import CA certificate

Assuming the generated CA certificate is `myca.crt`:

```bash
sudo cp myca.crt /usr/local/share/ca-certificates//myca.crt
sudo update-ca-certificates
```

```log
Updating certificates in /etc/ssl/certs...
rehash: warning: skipping ca-certificates.crt,it does not contain exactly one certificate or CRL
1 added, 0 removed; done.
Running hooks in /etc/ca-certificates/update.d...
done.
```

### Reinit CA certificates store

```bash
sudo rm /usr/local/share/ca-certificates/myca.crt
sudo update-ca-certificates -f
```

```log
Clearing symlinks in /etc/ssl/certs...
done.
Updating certificates in /etc/ssl/certs...
rehash: warning: skipping ca-certificates.crt,it does not contain exactly one certificate or CRL
128 added, 0 removed; done.
Running hooks in /etc/ca-certificates/update.d...
done.
```

> Reference: [https://shaneutt.com/blog/golang-ca-and-signed-cert-go/](https://shaneutt.com/blog/golang-ca-and-signed-cert-go/)
