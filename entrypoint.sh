#!/bin/sh

test -d /ca || mkdir -p /ca
test -d /cfg || mkdir -p /cfg
test -d /web || mkdir -p /web

SIMPLECA_DAYS=${DAYS:-3650}
SIMPLECA_SIZE=${SIZE:-2048}
SIMPLECA_C=${C:-FR}
SIMPLECA_ST=${ST:-France}
SIMPLECA_L=${L:-Paris}
SIMPLECA_O=${O:-Orga}
SIMPLECA_OU=${OU:-Unit}

if [ ! -f /ca/ca.key -o ! -f /ca/ca.crt ] ; then
  echo "Generating Certificate Authority"
  /usr/local/bin/simpleca ca create \
    -key /ca/ca.key \
    -out /ca/ca.crt \
    -CN ${CN:-EasyCA}
fi

echo "Generating Certificate Authority web server certificate"
simpleca key create -out /cfg/simpleca.key
simpleca csr create -key /cfg/simpleca.key -out /cfg/simpleca.csr -CN ${HOSTNAME:-localhost} -alt-names localhost -ips 127.0.0.1
simpleca ca sign -ca-key /ca/ca.key -ca-cert /ca/ca.crt -out /cfg/simpleca.crt /cfg/simpleca.csr

echo "Starting Certificate Authority web server"
exec /usr/local/bin/simpleca web -dir /web -ca-key /ca/ca.key -ca-cert /ca/ca.crt -ssl -port 443 -key /cfg/simpleca.key -cert /cfg/simpleca.crt
