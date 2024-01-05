FROM    golang AS build
COPY    . /go/simpleca
ARG     GOOS=${GOOS:-linux}
ARG     GOARCH=${GOARCH:-386}
ARG     CGO_ENABLED=${CGO_ENABLED:-0}
RUN     cd /go/simpleca && go mod download && go mod tidy && go build ./cmd/simpleca

FROM    ubuntu:${UBUNTU_TAG:-jammy}
COPY    --from=build /go/simpleca/simpleca /usr/local/bin/simpleca
ARG     DEBIAN_FRONTEND=noninteractive
RUN     apt update && apt install -y ca-certificates wget
COPY    entrypoint.sh /entrypoint.sh
RUN     chmod +x /usr/local/bin/simpleca /entrypoint.sh
RUN     mkdir /data /web

VOLUME  /ca

WORKDIR /web

EXPOSE  443

HEALTHCHECK --interval=10m --timeout=3s CMD wget -O - http://localhost/alive
ENTRYPOINT [ "/entrypoint.sh" ]
