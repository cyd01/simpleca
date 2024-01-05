MAKEFILE_NAME:=$(notdir $(lastword $(MAKEFILE_LIST)))
TARGET=$(notdir $(abspath $(lastword $(MAKEFILE_LIST)/..)))

## Define default command (before everything else)
.PHONY: all
all : help
	@:

-include variables.mk
-include .makerc
-include ~/.makerc

## Go definition
ARCH ?=386
CGO ?=0

## Go compiler
GO ?=$(shell which go 2> /dev/null)

## Binary compressor
UPX ?=$(shell which upx 2> /dev/null)

## Curl client
CURL ?=$(shell which curl 2> /dev/null)

## Git command-line tool
GIT ?=$(shell which git 2> /dev/null)

## Docker client
DOCKER ?=$(shell which docker 2> /dev/null)

MAKER=$(shell test -n "$${USER}" && echo $${USER} || { test -n "$${USERNAME}" && echo $${USERNAME} || { which logname > /dev/null 2>&1 && logname > /dev/null 2>&1 && logname || echo Unknown ; } } )

## convinient variable to define current date/time which can be used for build time
NOW=$(shell date --rfc-3339=seconds)
export NOW
DEBUG = printf -- "| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "DEBUG"
INFO = printf -- "\e[92m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "INFO"
WARN = printf -- "\e[93m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "WARN"
ERROR = printf -- "\e[91m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "ERROR"
FATAL = printf -- "\e[31m| $$(date --rfc-3339=seconds) | %-s | %-5.5s | %-s |\e[39m\n" "$(subst $(ROOT),$(PROJECT_NAME),$(shell pwd)/$(MAKEFILE_NAME))" "FATAL"

## Show this help prompt.
.PHONY: help
help:
	@ echo '$(TARGET) builder'
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
	@ (echo '   Name?=Default value?=Description'; echo '   ----?=-------------?=-----------'; (awk -F"\?=" '/^## /{ comment = substr($$0,4) } comment && /^[a-zA-Z][a-zA-Z0-9_-]+[ ]+\?= /{ print "   " $$1 "?=" substr($$2,0,80) "?=" comment }' $(MAKEFILE_LIST) 2>/dev/null | sort -d)) | sed -e 's/\?= /?=/g' | column -t -s '?='
	@ echo ''

## Will display value of variable.
debug/%:
	@test -z "$(wordlist 2,3,$(subst /, ,$*))" && echo '$*=$($*)' || $(MAKE) -C "$(*:$(firstword $(subst /, ,$*))/%=%)" "debug/$(firstword $(subst /, ,$*))"

## Will clean workspace
.PHONY: clean
clean:
	@echo "Cleaning..."
	-rm -f $(TARGET) $(TARGET)-linux $(TARGET).exe $(TARGET)-darwin main_darwin.go main_windows.go main_linux.go resource.syso $(TARGET).exe.manifest versioninfo.rc bindata.go
	-rm -f ./cmd/$(TARGET)/$(TARGET)*
	-@$(DOCKER) image rm -f $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) 2> /dev/null || true
	-@$(DOCKER) container rm -f $(DOCKER_IMAGE) 2> /dev/null || true

.PHONY: purge
purge :
	@echo "Purging..."
	-@docker system prune --force --all --volumes
	-@docker volume prune --force --all

## Initiate go.mod
go.mod : 
	@echo "Initiate go.mod..."
	-$(GO) mod init $(TARGET)

## Update dependencies
.PHONY: update
update: 
	@echo "Update dependencies..."
	@CGO_ENABLED=$(CGO) $(GO) get -u
	@CGO_ENABLED=$(CGO) $(GO) mod tidy
	@$(MAKE) --no-print-directory vendor

## Ensures that the go.mod file matches the source code in the module
.PHONY: tidy
tidy:
	@echo "Executing tidy..."
	@go mod tidy

## Update vendor
.PHONY: vendor
vendor: tidy
	@echo "Generating vendor..."
	@test ! -d vendor || CGO_ENABLED=$(CGO) $(GO) mod vendor

## Run tests
.PHONY: test
test:
	@echo "Run tests suites"
	find . -type f -name '*_test.go' | sed 's#/[^/]*$$##' | sort -u | xargs go test -v

## Compile for Linux target
main_linux.go :
	@echo "Generating main_linux.go..."
	@test -d ./cmd/$(TARGET) && printf '//go:build linux\n// +build linux\n\n//go:generate rm -f resource.syso\npackage $(TARGET)\n' > main_linux.go || printf '//go:build linux\n// +build linux\n\n//go:generate rm -f resource.syso\npackage main\n' > main_linux.go

.PHONY: linux
linux: main_linux.go go.mod bindata.go
	@GOOS=linux GOARCH=$(ARCH) CGO_ENABLED=$(CGO) $(MAKE) --no-print-directory compile
	@mv -f $(TARGET).out $(TARGET)-linux
	@test $$(uname) != "Linux" || mv -f $(TARGET)-linux $(TARGET)
	@rm -f main_linux.go

## Compile for Windows target
main_windows.go :
	@echo "Generating main_windows.go..."
	@test -d ./cmd/$(TARGET) && printf '//go:build windows\n// +build windows\n\n//go:generate goversioninfo -icon=$(TARGET).ico -manifest=$(TARGET).exe.manifest\npackage $(TARGET)\n' > main_windows.go || printf '//go:build windows\n// +build windows\n\n//go:generate goversioninfo -icon=$(TARGET).ico -manifest=$(TARGET).exe.manifest\npackage main\n' > main_windows.go
# .syso file can also be done with rsrc tool (https://github.com/akavel/rsrc)

$(TARGET).exe.manifest :
	@echo "Generating $(TARGET).exe.manifest..."
	@printf '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>\n<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">\n  <assemblyIdentity\n    type="win32"\n    name="$(TARGET)"\n    version="1.0.0.0"\n    processorArchitecture="*"/>\n <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">\n   <security>\n     <requestedPrivileges>\n       <requestedExecutionLevel\n         level="asInvoker"\n         uiAccess="false"/>\n       </requestedPrivileges>\n   </security>\n </trustInfo>\n</assembly>\n' > $(TARGET).exe.manifest

versioninfo.rc :
	@echo "Generating versioninfo.rc..."
	@printf '#define RT_MANIFEST 24\n1 VERSIONINFO\nFILEVERSION     1,0,0,0\nPRODUCTVERSION  1,0,0,0\nFILEFLAGSMASK   0X3FL\nFILEFLAGS       0L\nFILEOS          0X40004L\nFILETYPE        0X1\nFILESUBTYPE     0\nBEGIN\n    BLOCK "StringFileInfo"\n    BEGIN\n        BLOCK "040904B0"\n        BEGIN\n			VALUE "ProductVersion", "v1.0.0.0"\n        END\n    END\n    BLOCK "VarFileInfo"\n    BEGIN\n            VALUE "Translation", 0x0409, 0x04B0\n    END\nEND\n\n1 ICON "$(TARGET).ico"\n\n1 RT_MANIFEST "$(TARGET).exe.manifest"' > versioninfo.rc

.PHONY: windows
windows: main_windows.go $(TARGET).exe.manifest versioninfo.rc go.mod bindata.go
	@which goversioninfo > /dev/null || go get -d github.com/josephspurrier/goversioninfo/cmd/goversioninfo
	@GOOS=windows GOARCH=$(ARCH) CGO_ENABLED=$(CGO) $(MAKE) --no-print-directory compile
	@mv -f $(TARGET).out $(TARGET).exe
	@rm -f main_windows.go

.PHONY: mingw64
mingw64 : windows
	@:

## Compile for MacOS target
main_darwin.go :
	@echo "Generating main_darwin.go..."
	@test -d ./cmd/$(TARGET) && printf '//go:build darwin\n// +build darwin\n\n//go:generate rm -f resource.syso\npackage $(TARGET)\n' > main_darwin.go || printf '//go:build darwin\n// +build darwin\n\n//go:generate rm -f resource.syso\npackage main\n' > main_darwin.go

.PHONY: darwin
darwin: main_darwin.go go.mod bindata.go
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO) $(MAKE) --no-print-directory compile
	@mv -f $(TARGET).out $(TARGET)-darwin
	@rm -f main_darwin.go


.PHONY: macos
macos : darwin
	@:

## Compile for both Linux, MacOS and Windows target
bindata.go :
	@which go-bindata > /dev/null 2>&1 || { echo "Installing go-bindata..." ; sudo apt update > /dev/null 2>&1 ; sudo apt install -y go-bindata > /dev/null 2>&1 ; true ; }
	@grep bindata *.go > /dev/null 2>&1 && { echo "Generating bindata.go..." ;	GOOS=darwin GOARCH=$(ARCH) CGO_ENABLED=$(CGO) $(GO) generate ; } || true

.PHONY: buildall
buildall: linux windows darwin
	@:

## Compile locally
.PHONY: compile
compile: vendor
	@echo "Compiling for $(GOOS)/$(GOARCH)..."
	$(GO) get
	-which goversioninfo > /dev/null && $(GO) generate
	test -d ./cmd/$(TARGET) && DIR=./cmd/$(TARGET) ; $(GO) build -o $(TARGET).out -ldflags="-X main.BuildTime=$$(date '+%Y-%m-%dT%H:%M:%S') -X main.Version=$$(date '+%Y-%m-%dT%H:%M:%S') -X main.Author=$(MAKER) -X main.Revision=$$( { $(GIT) rev-parse HEAD 2>/dev/null || echo Undefined ; } | grep -v HEAD )" $${DIR}

## Compile for current target
.PHONY: build
build:
	@make --no-print-directory $$(uname | sed 's/[_-].*$$//' | tr '[:upper:]' '[:lower:]')

## Compress all targets
.PHONY: compress
compress: upx
	@:

upxone :
	@test -z "$(filter-out $@,$(MAKECMDGOALS))" || for file in $(filter-out $@,$(MAKECMDGOALS)) ; do test ! -f $$file || { $(UPX) -t $$file > /dev/null 2> /dev/null || { echo "Compressing $$file" ; $(UPX) -9qf $$file > /dev/null ; } ; } ; done

.PHONY: upx
upx :
	-@which upx > /dev/null && { test ! -f $(TARGET) || $(MAKE) --quiet --no-print-directory upxone $(TARGET) ; }
	-@which upx > /dev/null && { test ! -f $(TARGET)-linux || $(MAKE) --quiet --no-print-directory upxone $(TARGET)-linux ; }
	-@which upx > /dev/null && { test ! -f $(TARGET).exe || $(MAKE) --quiet --no-print-directory upxone $(TARGET).exe ; }
	-@which upx > /dev/null && { test ! -f $(TARGET)-darwin || $(MAKE) --quiet --no-print-directory upxone $(TARGET)-darwin ; }
	-@which upx > /dev/null && { test ! -f ./cmd/$(TARGET)/$(TARGET) || $(MAKE) --quiet --no-print-directory upxone ./cmd/$(TARGET)/$(TARGET) ; }

## Run binary
.PHONY: run
run:
	@test ! -f $(TARGET) || { echo "Starting binary <$(TARGET)> with parameters <$(RUN_PARAMETERS)>" ; ./$(TARGET) $(RUN_PARAMETERS) ; }

## Make docker image
.PHONY: image
image:
	@test ! -f Dockerfile || echo "Building docker image..."
	@test ! -f Dockerfile || $(DOCKER) build --force-rm --file Dockerfile --tag $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) .

## Build a docker image
.PHONY: docker
docker : image
	@:

## Run a docker image
.PHONY: imagerun
imagerun:
	@echo "Starting docker container <$(TARGET)> on image <$(DOCKER_IMAGE)> with parameters <$(DOCKER_RUN_PARAMETERS)>"
	@$(DOCKER) run --rm --interactive --tty --name $(DOCKER_IMAGE) $(DOCKER_RUN_PARAMETERS) $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: dockerrun
dockerrun :

## Upload binaries to Nexus
.PHONY: push
push:
	@echo "Uploading to Nexus"
	@test -n "$(NEXUS_AUTH_USER)" -a -n "$(NEXUS_AUTH_PASS)" || { echo "Credentials not set !!! See help." ; exit 1 ; }
	test ! -f $(TARGET)-linux || $(CURL) -u "$(NEXUS_AUTH_USER):$(NEXUS_AUTH_PASS)" --request PUT --upload-file $(TARGET)-linux $(NEXUS_URL)/repository/core-raw-releases/$(DOCKER_IMAGE)/$(TARGET)-linux
	test ! -f $(TARGET) || $(CURL) -u "$(NEXUS_AUTH_USER):$(NEXUS_AUTH_PASS)" --request PUT --upload-file $(TARGET) $(NEXUS_URL)/repository/core-raw-releases/$(DOCKER_IMAGE)/$(TARGET)-linux
	test ! -f $(TARGET).exe || $(CURL) -u "$(NEXUS_AUTH_USER):$(NEXUS_AUTH_PASS)" --request PUT --upload-file $(TARGET).exe $(NEXUS_URL)/repository/core-raw-releases/$(DOCKER_IMAGE)/$(TARGET)-windows.exe
	test ! -f $(TARGET)-darwin || $(CURL) -u "$(NEXUS_AUTH_USER):$(NEXUS_AUTH_PASS)" --request PUT --upload-file $(TARGET)-darwin $(NEXUS_URL)/repository/core-raw-releases/$(DOCKER_IMAGE)/$(TARGET)-darwin

## Upload binaries to Gitlab
.PHONY: package
package:
	@echo "Uploading to Gitlab"
	@test -n "$(GITLAB_URL)" || { echo "GITLAB_URL not set !!! See help." ; exit 1 ; }
	@test -n "$(GITLAB_TOKEN)" || { echo "GITLAB_TOKEN not set !!! See help." ; exit 1 ; }
	{ PACKAGE_ID=$$($(CURL) --silent --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" "$(GITLAB_URL)/api/v4/projects/$(GITLAB_PACKAGE)/packages" | jq '.[]|select(.version=="current")|.id') 2> /dev/null && $(CURL) --silent --request DELETE --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" "$(GITLAB_URL)/api/v4/projects/$(GITLAB_PACKAGE)/packages/$${PACKAGE_ID}" > /dev/null ; } || echo
	test ! -f $(TARGET)-linux || curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(TARGET)-linux "$(GITLAB_URL)/api/v4/projects/$(GITLAB_PACKAGE)/packages/generic/$(DOCKER_IMAGE)/current/$(TARGET)-linux" ; echo
	test ! -f $(TARGET) || curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(TARGET) "$(GITLAB_URL)/api/v4/projects/$(GITLAB_PACKAGE)/packages/generic/$(DOCKER_IMAGE)/current/$(TARGET)-linux" ; echo
	test ! -f $(TARGET).exe || curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(TARGET).exe "$(GITLAB_URL)/api/v4/projects/$(GITLAB_PACKAGE)/packages/generic/$(DOCKER_IMAGE)/current/$(TARGET)-windows.exe" ; echo
	test ! -f $(TARGET)-darwin || curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(TARGET)-darwin "$(GITLAB_URL)/api/v4/projects/$(GITLAB_PACKAGE)/packages/generic/$(DOCKER_IMAGE)/current/$(TARGET)-darwin" ; echo

## Prepare environment
.PHONY: preparecmd
preparecmd : 
	@test -d ./cmd/$(TARGET) || { echo "Creating directory ./cmd/$(TARGET)" ; mkdir -p ./cmd/$(TARGET) ; }
	@test -f ./cmd/$(TARGET)/main.go || { echo "Generating ./cmd/$(TARGET)/main.go" ; printf 'package main\nimport (\n    "$(TARGET)"\n	"os"\n)\nfunc main() {\n    $(TARGET).Main(os.Args)\n}\n' > ./cmd/$(TARGET)/main.go ; }
	@test -f ./cmd/$(TARGET)/.gitignore || { echo "Generating ./cmd/$(TARGET)/.gitignore" ; printf '# Binary\n/'$(TARGET)'\n/'$(TARGET)'-*\n/'$(TARGET)'_*\n/'$(TARGET)'.exe\n' > ./cmd/$(TARGET)/.gitignore ; }

.PHONY: prepareroot
prepareroot : 
	@test -f main.go || { echo "Generating ./main.go" ; printf 'package $(TARGET)\nimport (\n    "fmt"\n)\nfunc usage() {\n    fmt.Print(`\nUsage:  ExampleCmd COMMAND [OPTIONS]\n\nA simple tools ...\n\nCommands:\n`)\n}\nfunc Main(Args []string) {\n    if len(Args) <= 1 {\n        usage()\n    } else {\n    }\n}\nvar (\n    BuildTime = "Undefined"\n    Author    = "Unknown"\n    Revision  = "Undefined"\n)\nfunc printInfo() {\n	fmt.Println("Build at", BuildTime, "by", Author, "on revision", Revision)\n}\n' > main.go ; sed -i 's/ExampleCmd/'$(TARGET)'/' main.go ; }

.PHONY: preparedevcontainer
preparedevcontainer :
	@test -d ./.devcontainer/scripts || { echo "Generating ./.devcontainer/scripts" ; mkdir -p ./.devcontainer/scripts ; }
	@test -f ./.devcontainer/scripts/postAttach.sh || { echo "Generating ./.devcontainer/scripts/postAttach.sh" ; printf 'IyEvYmluL2Jhc2gKCmRlYnVnKCkgeyBbIC1uICIke1NDUklQVFNfREVCVUd9IiBdICYmIFsgIiR7U0NSSVBUU19ERUJVR30iICE9ICIwIiBdICYmIHByaW50ZiAtLSAiXGVbMzltfCBERUJVRyB8ICQxXGVbMzltXG4iOyB9CmluZm8oKSB7IHByaW50ZiAtLSAiXGVbOTJtfCBJTkZPIHwgJDFcZVszOW1cbiI7IH0Kd2FybigpIHsgcHJpbnRmIC0tICJcZVs5M218IFdBUk4gfCAkMVxlWzM5bVxuIjsgfQplcnJvcigpIHsgcHJpbnRmIC0tICJcZVs5MW18IEVSUk9SIHwgJDFcZVszOW1cbiI7IH0KCmRlYnVnICJleGVjdXRpbmcgJDAiCgp0ZXN0IC1yIH4vaG9zdC8uZGV2Y29udGFpbmVyX3JjIFwKJiYgeyAKICAgIGRlYnVnICJjaGVja2luZyBmb3Igfi9ob3N0Ly5kZXZjb250YWluZXJfcmMuLi4iCiAgICAuIH4vaG9zdC8uZGV2Y29udGFpbmVyX3JjIFwKICAgICYmIGluZm8gIn4vaG9zdC8uZGV2Y29udGFpbmVyX3JjIHN1Y2Vzc2Z1bGx5IGV4ZWN1dGVkIiBcCiAgICB8fCAgZXJyb3IgIn4vaG9zdC8uZGV2Y29udGFpbmVyX3JjIgp9IFwKfHwgewogICAgd2FybiAibm8gLmRldmNvbnRhaW5lcl9yYyBmaWxlIGZvdW5kIGluIHlvdXIgaG9tZTogY3JlYXRlIG9uZSB0byB0dW5lIHlvdXIgY29uZmlndXJhdGlvbiI7Cn0KCkhPU1RfRklMRVM9JHtIT1NUX0ZJTEVTOi1kZWZhdWx0fQpkZWJ1ZyAiY2hlY2tpbmcgZm9yIEhPU1RfRklMRVMgKCR7SE9TVF9GSUxFU30pLi4uIgpIT1NUX0ZJTEVTX0ZJTkFMPQpmb3IgZiBpbiAke0hPU1RfRklMRVN9OyBkbwogICAgWyAiJHtmfSIgPSAiZGVmYXVsdCIgXSBcCiAgICAgICAgJiYgewogICAgICAgICAgICBbIC16ICIke0hJU1RGSUxFfSIgXSBcCiAgICAgICAgICAgICYmIHdhcm4gIkhJU1RGSUxFIG5vdCBkZWZpbmVkIDogbm90IGV4cG9ydGVkIG9uIHlvdXIgaG9zdCA/IiBcCiAgICAgICAgICAgIHx8ICBIT1NUX0ZJTEVTX0ZJTkFMPSIke0hPU1RfRklMRVNfRklOQUx9ICR7SElTVEZJTEUjJHtIT01FfS99IgogICAgICAgICAgICBIT1NUX0ZJTEVTX0ZJTkFMPSIke0hPU1RfRklMRVNfRklOQUx9IC5uZXRyYyAubWFrZXJjIgogICAgICAgIH0gXAogICAgICAgIHx8IEhPU1RfRklMRVNfRklOQUw9IiR7SE9TVF9GSUxFU19GSU5BTH0gJHtmfSIKZG9uZQpmb3IgZiBpbiAke0hPU1RfRklMRVNfRklOQUx9OyBkbwogICAgaG9zdGY9IiR7SE9NRX0vaG9zdC8ke2YjJHtIT01FfS99IgogICAgWyAtZSAke2hvc3RmfSBdIFwKICAgICAgICAmJiB7CiAgICAgICAgICAgIGNkIH4gJiYgbG4gLXMgLWYgJHtob3N0Zn0gJHtmfSBcCiAgICAgICAgICAgICYmIGluZm8gImxpbmsgZm9yICR7Zn0gc3VjZXNzZnVsbHkgY3JlYXRlZCIgXAogICAgICAgICAgICB8fCBlcnJvciAiZmFpbGVkIHRvIGNyZWF0ZSBsaW5rIGZvciAke2Z9IiAgICAKICAgICAgICB9IFwKICAgICAgICB8fCB7CiAgICAgICAgICAgIHdhcm4gIiR7Zn0gbGluayBub3QgY3JlYXRlZDogaG9zdCBmaWxlICgke2hvc3RmfSkgbm90IGZvdW5kIgogICAgICAgIH0KZG9uZQoKISB0ZXN0IC1yICR7V09SS1NQQUNFfS9wYWNrYWdlLmpzb24gfHwgKGNkICR7V09SS1NQQUNFfSAmJiBucG0gY2xlYW4taW5zdGFsbCkKCmluZm8gIiQwIGV4ZWN1dGVkIgo=' | base64 -d > ./.devcontainer/scripts/postAttach.sh ; }
	@test -f ./.devcontainer/scripts/common.sh || { echo "Generating ./.devcontainer/scripts/common.sh" ; printf 'IyEvYmluL2Jhc2gKc2V0IC1lCkVYSVRfQ09ERT0wCgpbIC16ICIke1NUQVJUVVBfREVCVUd9IiBdICYmIFNUQVJUVVBfREVCVUc9MCB8fCBTVEFSVFVQX0RFQlVHPTEKZnVuY3Rpb24gZGVidWcoKSB7CiAgICBbICRTVEFSVFVQX0RFQlVHIC1lcSAwIF0gfHwgcHJpbnRmIC0tICJcZVszOW0lLTUuNXMgfCAlLXNcZVszOW1cbiIgImRlYnVnIiAiJCoiCn0KZnVuY3Rpb24gZGVidWdmKCkgewogICAgWyAkU1RBUlRVUF9ERUJVRyAtZXEgMCBdIHx8IHsKICAgICAgICBmb3JtYXQ9IiQxIgogICAgICAgIHNoaWZ0CiAgICAgICAgcHJpbnRmIC0tICJcZVszOW0lLTUuNXMgfCAkZm9ybWF0XGVbMzltXG4iICJkZWJ1ZyIgIiQqIgogICAgfQp9CmZ1bmN0aW9uIGluZm8oKSB7CiAgICBwcmludGYgLS0gIlxlWzM5bSUtNS41cyB8ICUtc1xlWzM5bVxuIiAiaW5mbyIgIiQqIgp9CmZ1bmN0aW9uIGluZm9mKCkgewogICAgZm9ybWF0PSIkMSIKICAgIHNoaWZ0CiAgICBwcmludGYgLS0gIlxlWzM5bSUtNS41cyB8ICRmb3JtYXRcZVszOW1cbiIgImluZm8iICIkKiIKfQoKZnVuY3Rpb24gaW5mb2dyZWVuKCkgewogICAgcHJpbnRmIC0tICJcZVs5Mm0lLTUuNXMgfCAlLXNcZVszOW1cbiIgImluZm8iICIkKiIKfQpmdW5jdGlvbiBpbmZvZ3JlZW5mKCkgewogICAgZm9ybWF0PSIkMSIKICAgIHNoaWZ0CiAgICBwcmludGYgLS0gIlxlWzkybSUtNS41cyB8ICRmb3JtYXRcZVszOW1cbiIgImluZm8iICIkKiIKfQoKZnVuY3Rpb24gd2FybigpIHsKICAgIHByaW50ZiAtLSAiXGVbOTNtJS01LjVzIHwgJS1zXGVbMzltXG4iICJ3YXJuIiAiJCoiCn0KZnVuY3Rpb24gd2FybmYoKSB7CiAgICBmb3JtYXQ9IiQxIgogICAgc2hpZnQKICAgIHByaW50ZiAtLSAiXGVbOTNtJS01LjVzIHwgJGZvcm1hdFxlWzM5bVxuIiAid2FybiIgIiQqIgp9CgpmdW5jdGlvbiBlcnJvcigpIHsKICAgIHByaW50ZiAtLSAiXGVbOTFtJS01LjVzIHwgJS1zXGVbMzltXG4iICJlcnJvciIgIiQqIgogICAgRVhJVF9DT0RFPSQoKCRFWElUX0NPREUgKyAxKSkKfQpmdW5jdGlvbiBlcnJvcmYoKSB7CiAgICBmb3JtYXQ9IiQxIgogICAgc2hpZnQKICAgIHByaW50ZiAtLSAiXGVbOTFtJS01LjVzIHwgJGZvcm1hdFxlWzM5bVxuIiAiZXJyb3IiICIkKiIKICAgIEVYSVRfQ09ERT0kKCgkRVhJVF9DT0RFICsgMSkpCn0KZnVuY3Rpb24gZmF0YWwoKSB7CiAgICBwcmludGYgLS0gIlxlWzMxbSUtNS41cyB8ICUtc1xlWzM5bVxuIiAiZmF0YWwiICIkKiIKICAgIEVYSVRfQ09ERT0kKCgkRVhJVF9DT0RFICsgMSkpCiAgICBleGl0ICRFWElUX0NPREUKfQpmdW5jdGlvbiBmYXRhbGYoKSB7CiAgICBmb3JtYXQ9IiQxIgogICAgc2hpZnQKICAgIHByaW50ZiAtLSAiXGVbMzFtJS01LjVzIHwgJGZvcm1hdFxlWzM5bVxuIiAiZmF0YWwiICIkKiIKICAgIEVYSVRfQ09ERT0kKCgkRVhJVF9DT0RFICsgMSkpCiAgICBleGl0ICRFWElUX0NPREUKfQo=' | base64 -d > ./.devcontainer/scripts/common.sh ; }
	@test -f ./.devcontainer/scripts/startup.sh || { echo "Generating ./.devcontainer/scripts/startup.sh" ; printf 'IyEvYmluL2Jhc2gKIyBzYXZlIHN0ZG91dCB0byBhbiBhdmFpbGFibGUgZmlsZSBkZXNjcmlwdG9yCmV4ZWMgNjQ+JjEKIyByZWRpcmVjdCBzdGRvdXQgdG8gc3RkZXJyIGZvciBpdCBtYXkgYnJlYWsgY29tbWFuZCBvdXRwdXQKZXhlYyAxPiYyCgphZGFwdF9wZXJtaXNzaW9ucygpIHsKICAgIGxvY2FsIEJJTkRJUj0kKGRpcm5hbWUgJDApCiAgICBzZXQgLWUKICAgIHNvdXJjZSAke0JJTkRJUn0vY29tbW9uLnNoCgogICAgbG9jYWwgc3RhcnR1cF9kZWJ1Z19wcmVmaXg9ImV4ZWN1dGluZyIKICAgIGxvY2FsIHN0YXJ0dXBfY29tbWFuZF9wcmVmaXg9CgogICAgZGVidWcgImN1cnJlbnQgdXNlciBpcyAkKGlkKSIKICAgIGRlYnVnICJjaGVja2luZyAvdmFyL3J1bi9kb2NrZXIuc29jayBleGlzdGVuY2UuLi4iCiAgICB0ZXN0IC1lIC92YXIvcnVuL2RvY2tlci5zb2NrICYmIHsKICAgICAgICBkZWJ1ZyAiZG9ja2VyIHNvY2tldCBkZXRlY3RlZCwgc3RhcnRpbmcgd2l6YXJkcnkuLi4iCiAgICAgICAgQ1VSUkVOVF9VU0VSPSQoaWQgLS11c2VyIC0tbmFtZSkKICAgICAgICBbICQ/IC1lcSAwIF0gXAogICAgICAgIHx8IHsKICAgICAgICAgICAgZmF0YWwgImZhaWxlZCB0byBnZXQgY3VycmVudCB1c2VyIG5hbWUiCiAgICAgICAgICAgIGV4aXQgMQogICAgICAgIH0KICAgICAgICBkZWJ1ZyAibG9va2luZyBmb3IgZG9ja2VyIHNvY2tldCBnaWQgJHtET0NLRVJfU09DS19HUk9VUF9JRH0uLi4iCiAgICAgICAgRE9DS0VSX1NPQ0tfR1JPVVBfSUQ9JChzdGF0IC1jICclZycgL3Zhci9ydW4vZG9ja2VyLnNvY2spCiAgICAgICAgWyAkPyAtZXEgMCBdIFwKICAgICAgICB8fCB7CiAgICAgICAgICAgIGZhdGFsICJmYWlsZWQgdG8gcmVhZCBkb2NrZXIgc29ja2V0IGdyb3VwIGlkIgogICAgICAgICAgICBleGl0IDEKICAgICAgICB9CiAgICAgICAgZGVidWcgImxvb2tpbmcgZm9yIGRvY2tlciBzb2NrZXQgZ2lkICR7RE9DS0VSX1NPQ0tfR1JPVVBfSUR9Li4uIgogICAgICAgIGdldGVudCBncm91cCAke0RPQ0tFUl9TT0NLX0dST1VQX0lEfSA+IC9kZXYvbnVsbCBcCiAgICAgICAgJiYgewogICAgICAgICAgICBkZWJ1ZyAiYSBncm91cCBmb3IgaWQgJERPQ0tFUl9TT0NLX0dST1VQX0lEIHdhcyBmb3VuZCIKICAgICAgICB9IFwKICAgICAgICB8fCB7CiAgICAgICAgICAgIGRlYnVnICJubyBncm91cCB3aXRoIGlkICR7RE9DS0VSX1NPQ0tfR1JPVVBfSUR9IHdhcyBmb3VuZCwgY3JlYXRpbmcgb25lIgogICAgICAgICAgICBzdWRvIGdyb3VwYWRkIC1nICR7RE9DS0VSX1NPQ0tfR1JPVVBfSUR9IGRvY2tlci1ob3N0IFwKICAgICAgICAgICAgJiYgewogICAgICAgICAgICAgICAgZGVidWcgImRvY2tlci1ob3N0IGdyb3VwIGNyZWF0ZWQgd2l0aCBpZCAkRE9DS0VSX1NPQ0tfR1JPVVBfSUQiCiAgICAgICAgICAgIH0gXAogICAgICAgICAgICB8fCB7CiAgICAgICAgICAgICAgICBmYXRhbCAiZG9ja2VyLWhvc3QgZ3JvdXAgZmFpbGVkIHRvIGJlIGNyZWF0ZWQiCiAgICAgICAgICAgICAgICBleGl0IDEKICAgICAgICAgICAgfQoKICAgICAgICB9CiAgICAgICAgRE9DS0VSX1NPQ0tfR1JPVVBfTkFNRT0kKGdldGVudCBncm91cCAke0RPQ0tFUl9TT0NLX0dST1VQX0lEfSB8IGF3ayAtRjogJ3twcmludCAkMTt9JykKICAgICAgICBbICQ/IC1lcSAwIF0gXAogICAgICAgIHx8IHsKICAgICAgICAgICAgZmF0YWwgImZhaWxlZCB0byByZXRyaWV2ZSBkb2NrZXIgc29ja2V0IGdyb3VwIG5hbWUiCiAgICAgICAgICAgIGV4aXQgMQogICAgICAgIH0KICAgICAgICBnZXRlbnQgZ3JvdXAgJHtET0NLRVJfU09DS19HUk9VUF9JRH0gfCBhd2sgLUY6ICd7cHJpbnQgJDR9JyB8IGVncmVwICIoXnwsKSQoaWQgLS11c2VyIC0tbmFtZSkoLHwkKSIgPiAvZGV2L251bGwgXAogICAgICAgICYmIHsKICAgICAgICAgICAgZGVidWcgInVzZXIgYWxyZWFkeSBpbiBkb2NrZXIgc29ja2V0IGdyb3VwICR7RE9DRVJfU09DS19HUk9VUF9OQU1FfSIKICAgICAgICB9IFwKICAgICAgICB8fCB7CiAgICAgICAgICAgIHN1ZG8gdXNlcm1vZCAtLWFwcGVuZCAtLWdyb3VwcyAke0RPQ0tFUl9TT0NLX0dST1VQX05BTUV9ICR7Q1VSUkVOVF9VU0VSfSBcCiAgICAgICAgICAgICYmIHsKICAgICAgICAgICAgICAgIGRlYnVnICJ1c2VyICR7Q1VSUkVOVF9VU0VSfSB3YXMgYWRkZWQgdG8gZ3JvdXAgaWQgJHtET0NLRVJfU09DS19HUk9VUF9OQU1FfSIKICAgICAgICAgICAgICAgIHN0YXJ0dXBfZGVidWdfcHJlZml4PSJhY3F1aXJpbmcgbmV3IHBlcm1pc3Npb25zIGFuZCBydW5uaW5nIgogICAgICAgICAgICAgICAgc3RhcnR1cF9jb21tYW5kX3ByZWZpeD0ic3VkbyAtLXByZXNlcnZlLWVudiAtLXVzZXI9JHtDVVJSRU5UX1VTRVJ9IgogICAgICAgICAgICB9IFwKICAgICAgICAgICAgfHwgewogICAgICAgICAgICAgICAgZmF0YWwgImZhaWxlZCB0byBhZGQgJHtDVVJSRU5UX1VTRVJ9IHRvIGdyb3VwIGlkICR7RE9DS0VSX1NPQ0tfR1JPVVBfTkFNRX0iCiAgICAgICAgICAgICAgICBleGl0IDEKICAgICAgICAgICAgfQogICAgICAgIH0KICAgIH0gXAogICAgfHwgewogICAgICAgIGRlYnVnICJubyBkb2NrZXIgc29jayBmb3VuZCwgc2tpcHBpbmcgd2l6YXJkcnkiCiAgICB9CiAgICBkZWJ1ZyAkKHByaW50ZiAtLSAiJXMgIiAiJHtzdGFydHVwX2RlYnVnX3ByZWZpeH0gJHtzdGFydHVwX2NvbW1hbmRfcHJlZml4fSAiOyBwcmludGYgLS0gJyVxICcgIiRAIikKICAgICMgcmV2ZXJ0IHN0ZG91dCBiYWNrIHRvIG9yaWdpbmFsIGZpbGUgZGVzY3JpcHRvcgogICAgZXhlYyAxPiY2NCA2ND4mLQogICAgIyBzdGFydGluZyBjb21tYW5kCiAgICBleGVjICR7c3RhcnR1cF9jb21tYW5kX3ByZWZpeH0gIiRAIgp9CgphZGFwdF9wZXJtaXNzaW9ucyAiJEAiCg==' | base64 -d > ./.devcontainer/scripts/startup.sh ; }
	@test -f ./.devcontainer/devcontainer.json || { echo "Generating ./.devcontainer/devcontainer.json" ; printf 'Ly8gRm9yIGZvcm1hdCBkZXRhaWxzLCBzZWUgaHR0cHM6Ly9ha2EubXMvZGV2Y29udGFpbmVyLmpzb24uIEZvciBjb25maWcgb3B0aW9ucywgc2VlIHRoZSBSRUFETUUgYXQ6Ci8vIGh0dHBzOi8vZ2l0aHViLmNvbS9taWNyb3NvZnQvdnNjb2RlLWRldi1jb250YWluZXJzL3RyZWUvdjAuMjA5LjQvY29udGFpbmVycy91YnVudHUKewogIC8vIG5pY2VseSBuYW1lIHlvdXIgZGV2IGNvbnRhaW5lciBlbnZpcm9ubWVudCAoc2hvd24gYXQgbGVmdC1oYW5kIGJvdHRvbSBjb3JuZXIgaW4gVlNDb2RlKQogICJuYW1lIjogIiR7Y29udGFpbmVyV29ya3NwYWNlRm9sZGVyQmFzZW5hbWV9IiwKCiAgLy8gZGlyZWN0bHkgdXNlIG91ciBjaSBiYXNlIGltYWdlCiAgLy8gImltYWdlIjogImlkcC1kb2NrZXItcmVsZWFzZXMucmVwb3MudGVjaGxhYmZkai5pby9jaS9iYXNlaW1hZ2U6dlguWS5aIiwKCiAgLy8geW91IGNhbiBpbnN0ZWFkIGJ1aWxkIGFuIGltYWdlIHVzaW5nIHRoZSBpbmNsdWRlZCBEb2NrZXJmaWxlIHRvIGRvIHNvbWUgdHVuaW5nIHJlcXVpcmVkIGF0IGRldmVsb3BtZW50IG9ubHkKICAvLyBpZiByZXF1aXJlZCwgdHVuZSBEb2NrZXJmaWxlLCBjb21tZW50IGltYWdlIGFuZCB1bmNvbW1lbnQgdGhlIGZvbGxvd2luZyB0byB0cmlnZ2VyIGJ1aWxkIGF0IGRldiBjb250YWluZXIgY3JlYXRpb24KICAiYnVpbGQiOiB7CiAgICAiZG9ja2VyZmlsZSI6ICJEb2NrZXJmaWxlIiwKICAgICJhcmdzIjogewogICAgICAiSE9TVF9VU0VSIjogIiR7bG9jYWxFbnY6VVNFUn0iLAogICAgICAiSE9TVF9IT01FIjogIiR7bG9jYWxFbnY6SE9NRX0iLAogICAgICAiSE9TVF9TSEVMTCI6ICIke2xvY2FsRW52OlNIRUxMfSIsCiAgICAgICJET0NLRVJfUkVMRUFTRVNfSE9TVE5BTUUiOiAiZG9ja2VyLmlvLyIsCiAgICAgICJET0NLRVJfSU1BR0VfTkFNRSI6ICJsaWJyYXJ5L3VidW50dSIsCiAgICAgICJET0NLRVJfSU1BR0VfVEFHIjogImxhdGVzdCIKICAgIH0KICB9LAoKICAvLyBtb3VudCB5b3VyIHdvcmtzcGFjZTogd2UgbW91bnQgaW4gdGhlIGV4YWN0IHNhbWUgZm9sZGVyIGFzIHRoZSBob3N0IG9uZQogICJ3b3Jrc3BhY2VNb3VudCI6ICJzb3VyY2U9JHtsb2NhbFdvcmtzcGFjZUZvbGRlcn0sdGFyZ2V0PSR7bG9jYWxXb3Jrc3BhY2VGb2xkZXJ9LHR5cGU9YmluZCxjb25zaXN0ZW5jeT1jYWNoZWQiLAogIC8vIHlvdSBjYW4gY2hvb3NlIHRvIG1vdW50IHlvdXIgd29ya3NwYWNlIHBhcmVudCBkaXJlY3RvcnkgaW5zdGVhZAogIC8vIndvcmtzcGFjZU1vdW50IjogInNvdXJjZT0ke2xvY2FsV29ya3NwYWNlRm9sZGVyfS8uLix0YXJnZXQ9JHtsb2NhbFdvcmtzcGFjZUZvbGRlcn0vLi4sdHlwZT1iaW5kLGNvbnNpc3RlbmN5PWNhY2hlZCIsCgogIC8vIGRpcmVjdG9yeSB0byB1c2UgYXMgd29ya3NwYWNlIGluc2lkZSBjb250YWluZXI6IHlvdSBjYW4gdHVuZSBpdCBpZiByZXF1aXJlZAogICJ3b3Jrc3BhY2VGb2xkZXIiOiAiJHtsb2NhbFdvcmtzcGFjZUZvbGRlcn0iLAoKICAibW91bnRzIjogWwogICAgInNvdXJjZT0ke2xvY2FsRW52OkhPTUV9LHRhcmdldD0ke2xvY2FsRW52OkhPTUV9L2hvc3QsdHlwZT1iaW5kLGNvbnNpc3RlbmN5PWNhY2hlZCIsCiAgICAic291cmNlPS92YXIvcnVuL2RvY2tlci5zb2NrLHRhcmdldD0vdmFyL3J1bi9kb2NrZXIuc29jayx0eXBlPWJpbmQiCiAgXSwKCiAgLy8gVlNDb2RlIHNldHRpbmdzIGFuZCBleHRlbnNpb25zCiAgImN1c3RvbWl6YXRpb25zIjogewogICAgInZzY29kZSI6IHsKICAgICAgInNldHRpbmdzIjogewogICAgICAgICJlZGl0b3IuaW5zZXJ0U3BhY2VzIjogdHJ1ZSwKICAgICAgICAiZWRpdG9yLnRhYlNpemUiOiAyLAogICAgICAgICJlZGl0b3IuZGVmYXVsdEZvcm1hdHRlciI6ICJlc2JlbnAucHJldHRpZXItdnNjb2RlIiwKICAgICAgICAiZWRpdG9yLmZvcm1hdE9uU2F2ZSI6IHRydWUsCiAgICAgICAgImVkaXRvci5mb3JtYXRPblR5cGUiOiB0cnVlLAogICAgICAgICJlZGl0b3IuZm9ybWF0T25QYXN0ZSI6IHRydWUsCgogICAgICAgICJbZ29dIjogewogICAgICAgICAgImVkaXRvci5kZWZhdWx0Rm9ybWF0dGVyIjogImdvbGFuZy5nbyIsCiAgICAgICAgICAiZWRpdG9yLmluc2VydFNwYWNlcyI6IGZhbHNlLAogICAgICAgICAgImVkaXRvci5jb2RlQWN0aW9uc09uU2F2ZSI6IHsKICAgICAgICAgICAgInNvdXJjZS5vcmdhbml6ZUltcG9ydHMiOiB0cnVlCiAgICAgICAgICB9LAogICAgICAgICAgImVkaXRvci5zdWdnZXN0LnNuaXBwZXRzUHJldmVudFF1aWNrU3VnZ2VzdGlvbnMiOiBmYWxzZQogICAgICAgIH0sCiAgICAgICAgImdvLmxpbnRPblNhdmUiOiAid29ya3NwYWNlIiwKICAgICAgICAiZ28ubGludFRvb2wiOiAic3RhdGljY2hlY2siLAoKICAgICAgICAiW2phdmFzY3JpcHRdIjogewogICAgICAgICAgImVkaXRvci5jb2RlQWN0aW9uc09uU2F2ZSI6IHsKICAgICAgICAgICAgInNvdXJjZS5maXhBbGwuZXNsaW50IjogdHJ1ZQogICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgInRlcm1pbmFsLmludGVncmF0ZWQuYWxsb3dDaG9yZHMiOiBmYWxzZSwKICAgICAgICAidGVybWluYWwuaW50ZWdyYXRlZC5kcmF3Qm9sZFRleHRJbkJyaWdodENvbG9ycyI6IGZhbHNlCiAgICAgIH0sCiAgICAgICJleHRlbnNpb25zIjogWwogICAgICAgICJtcy1henVyZXRvb2xzLnZzY29kZS1kb2NrZXIiLAogICAgICAgICJtcy12c2NvZGUtcmVtb3RlLnJlbW90ZS1zc2giLAogICAgICAgICJiaWVybmVyLm1hcmtkb3duLW1lcm1haWQiLAogICAgICAgICJkYXZpZGFuc29uLnZzY29kZS1tYXJrZG93bmxpbnQiLAogICAgICAgICJkYmFldW1lci52c2NvZGUtZXNsaW50IiwKICAgICAgICAiZXNiZW5wLnByZXR0aWVyLXZzY29kZSIsCiAgICAgICAgImdvbGFuZy5nbyIKICAgICAgXQogICAgfQogIH0sCiAgLy8gcnVuIGN1c3RvbQogICJwb3N0QXR0YWNoQ29tbWFuZCI6ICIvdXNyL2xvY2FsL3NjcmlwdHMvcG9zdEF0dGFjaC5zaCIsCiAgIm92ZXJyaWRlQ29tbWFuZCI6IGZhbHNlLAogIC8vIHJ1biBpbiBuZXR3b3JrIGhvc3QgbW9kZSB0byBhY2NlcyBwdWJsaXNoZWQgcG9ydHMgaW4gbG9jYWwgZGVtbyBlbnZpcm9ubWVudHMKICAicnVuQXJncyI6IFsiLS1uZXR3b3JrIiwgImhvc3QiXSwKCiAgInJlbW90ZUVudiI6IHsKICAgIC8vIGZvcndhcmQgc29tZSBsb2NhbGx5IGRlZmluZWQgZW52aXJvbm1lbnQgdmFyaWFibGVzIHRvIGNvbnRhaW5lcgogICAgIkhJU1RGSUxFIjogIiR7bG9jYWxFbnY6SElTVEZJTEV9IiwKICAgICJISVNUU0laRSI6ICIke2xvY2FsRW52OkhJU1RTSVpFfSIsCiAgICAvLyBoZWxwZnVsIHZhcmlhYmxlIHNvIHlvdSBnbyBiYWNrIHRvIHlvdXIgbG9jYWwgd29ya3NwYWNlIGZvbGRlcgogICAgIldPUktTUEFDRSI6ICIke2xvY2FsV29ya3NwYWNlRm9sZGVyfSIsCiAgICAiSVFfVVNFUk5BTUUiOiAiJHtsb2NhbEVudjpJUV9VU0VSTkFNRX0iLAogICAgIklRX1RPS0VOIjogIiR7bG9jYWxFbnY6SVFfVE9LRU59IiwKICAgIC8vCiAgICAvLyBzZXQgdG8gMSB0byBkZWJ1ZyBwb3N0QXR0YWNoLnNoCiAgICAiU0NSSVBUU19ERUJVRyI6ICIwIiwKICAgIC8vCiAgICAvLyBkZWZhdWx0IHZhbHVlIGZvciBDR09fRU5BQkxFRCBpcyAwIGluIGJhc2VpbWFnZQogICAgLy8gc28gaXQgeW91IHJlYWxseSB3YW50IHRvIG92ZXJ3cml0ZSBpdCwgZG8gaXQgaGVyZQogICAgLy8gIkNHT19FTkFCTEVEIjogIjEiLAogICAgLy8KICAgIC8vIEhPU1RfRklMRVMgaXMgdXNlZCBieSBwb3N0QXR0YWNoIHRvIGNyZWF0ZSBzeW1saW5rcyB0byB5b3VyIGhvc3QgaG9tZSBmaWxlcwogICAgLy8gZGVmYXVsdCBtZWFucyBISVNURklMRSAoZS5nLiBiYXNoX2hpc3RvcnkpIGFuZCAubmV0cmMKICAgIC8vIHlvdSBtYXkgc2F5ICJkZWZhdWx0IGFub3RoZXJmaWxlIiB0byBhbHNvIHN5bWxpbmsgdGhhdCBvdGhlciBmaWxlCiAgICAiSE9TVF9GSUxFUyI6ICJkZWZhdWx0IgogIH0sCgogIC8vIHVzZXIgdG8gc3RhcnQgY29udGFpbmVyIHdpdGgKICAiY29udGFpbmVyVXNlciI6ICIke2xvY2FsRW52OlVTRVJ9IiwKICAvLyB1c2VyIHRvIGV4ZWMgaW5zaWRlIHRoZSBjb250YWluZXIKICAvLyB1YnVudHUgaWYgdXNpbmcgYmFzZSBpbWFnZQogIC8vInJlbW90ZVVzZXIiOiAidWJ1bnR1IiwKICAvLyB5b3VyIGxvY2FsIHVzZXIgbmFtZSBpZiB1c2luZyBhIGN1c3RvbSBidWlsdCBpbWFnZSB3aXRoIHlvdXIgdXNlciBpbiBpdAogICJyZW1vdGVVc2VyIjogIiR7bG9jYWxFbnY6VVNFUn0iLAoKICAvLyB1cGRhdGUgcmVtb3RlIHVzZXIgdWlkIDogdHJ1ZSBieSBkZWZhdWx0CiAgLy8gYWRkZWQgaGVyZSB0byByZW1pbmQgeW91IHJlbW90ZVVzZXIgaWRzIG11c3QgYmUgYWRqdXN0ZWQKICAidXBkYXRlUmVtb3RlVXNlclVJRCI6IHRydWUsCgogIC8vIGh0dHBzOi8vZ2l0aHViLmNvbS9kZXZjb250YWluZXJzL2ZlYXR1cmVzCiAgImZlYXR1cmVzIjogewogICAgLy8gdGhpcyBvbmUgaXMgcmVxdWlyZWQgZm9yIGRvY2tlciBvcGVyYXRpb25zCiAgICAvLyBpdCB3aWxsIGFsc28gYWRkIGRvY2tlciBWU0NvZGUgZXh0ZW5zaW9uCiAgICAvLyBodHRwczovL2dpdGh1Yi5jb20vZGV2Y29udGFpbmVycy9mZWF0dXJlcy90cmVlL21haW4vc3JjL2RvY2tlci1vdXRzaWRlLW9mLWRvY2tlcgogICAgLy8gImdoY3IuaW8vZGV2Y29udGFpbmVycy9mZWF0dXJlcy9kb2NrZXItb3V0c2lkZS1vZi1kb2NrZXI6MS4wLjkiOiB7fQogIH0KfQo=' | base64 -d > ./.devcontainer/devcontainer.json ; }
	@test -f ./.devcontainer/Dockerfile || { echo "Generating ./.devcontainer/Dockerfile" ; printf 'IyBTZWUgaGVyZSBmb3IgaW1hZ2UgY29udGVudHM6IGh0dHBzOi8vZ2l0aHViLmNvbS9taWNyb3NvZnQvdnNjb2RlLWRldi1jb250YWluZXJzL3RyZWUvdjAuMjA5LjUvY29udGFpbmVycy91YnVudHUvLmRldmNvbnRhaW5lci9iYXNlLkRvY2tlcmZpbGUKQVJHICBET0NLRVJfUkVMRUFTRVNfSE9TVE5BTUUKQVJHICBET0NLRVJfSU1BR0VfTkFNRT0ke0RPQ0tFUl9JTUFHRV9OQU1FOi0idWJ1bnR1In0KQVJHICBET0NLRVJfSU1BR0VfVEFHPSR7RE9DS0VSX0lNQUdFX1RBRzotImxhdGVzdCJ9CkZST00gJHtET0NLRVJfUkVMRUFTRVNfSE9TVE5BTUV9JHtET0NLRVJfSU1BR0VfTkFNRX06JHtET0NLRVJfSU1BR0VfVEFHfQoKIyBzd2l0Y2ggdG8gcm9vdApVU0VSIHJvb3QKCiMgcGVyZm9ybSBhIGxhc3QgdXBncmFkZSBhbmQgaW5zdGFsbCBjdXN0b20gcGFja2FnZXM6IGlmIHRoZXJlIGFyZSB0d28gbWFueSB1cGRhdGVzLCB0aW1lIHRvIHZlcnNpb24geW91ciBiYXNlIGltYWdlIHRvIGF2b2lkIGl0ICEKUlVOICBhcHQtZ2V0IHVwZGF0ZSAmJiBleHBvcnQgREVCSUFOX0ZST05URU5EPW5vbmludGVyYWN0aXZlICBcCiAgICAgICYmIGFwdC1nZXQgZnVsbC11cGdyYWRlIC0teWVzIC0tbm8taW5zdGFsbC1yZWNvbW1lbmRzICAgXAogICAgICAmJiBhcHQtZ2V0IC15IGluc3RhbGwgLS1uby1pbnN0YWxsLXJlY29tbWVuZHMgICAgICAgICAgIFwKICAgICAgICAgYnNkbWFpbnV0aWxzICAgICAgICAgIFwKICAgICAgICAgY2EtY2VydGlmaWNhdGVzICAgICAgIFwKICAgICAgICAgY3VybCAgICAgICAgICAgICAgICAgIFwKICAgICAgICAgZG9ja2VyLmlvICAgICAgICAgICAgIFwKICAgICAgICAgZG9ja2VyLWNvbXBvc2UgICAgICAgIFwKICAgICAgICAgZG9ja2VyLWJ1aWxkeCAgICAgICAgIFwKICAgICAgICAgZ2l0ICAgICAgICAgICAgICAgICAgIFwKICAgICAgICAgbWFrZSAgICAgICAgICAgICAgICAgIFwKICAgICAgICAgb3BlbnNzaC1jbGllbnQgICAgICAgIFwKICAgICAgICAgb3BlbnNzaC1zZXJ2ZXIgICAgICAgIFwKICAgICAgICAgc3VkbyAgICAgICAgICAgICAgICAgIFwKICAgICAgICAgdXB4LXVjbCAgICAgICAgICAgICAgIFwKICAgICAgICAgd2dldCAgICAgICAgICAgICAgICAgIFwKICAgICAgICAgbHNvZiBcCiAgICAgICAgIG5hbm8gXAogICAgICAgICBuZXQtdG9vbHMgXAogICAgICAgICBzdWRvIFwKICAgICAgICAgdGlnIFwKICAgICAgICAgdmltIFwKICAgICAgICAgenNoIFwKICAgICAgJiYgYXB0LWdldCBhdXRvcmVtb3ZlIC15ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICBcCiAgICAgICYmIGFwdC1nZXQgY2xlYW4gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgXAogICAgICAmJiBybSAtcmYgL3RtcC8qIC92YXIvdG1wLyogICAgICAgICAgICAgICAgICAgICAgICAgICAgIFwKICAgICAgJiYgcm0gLXJmIC92YXIvbGliL2FwdC9saXN0cy8qIC92YXIvY2FjaGUvYXB0LyoKCiMgSW5zdGFsbCBHbwpSVU4gIGN1cmwgLWZrc1NMIC0tbm8tcHJvZ3Jlc3MtbWV0ZXIgLW8gL3RtcC9nby1saW51eC1hbWQ2NC50YXIuZ3ogImh0dHBzOi8vZ28uZGV2L2RsLyQoY3VybCAtZmtzU0wgJ2h0dHBzOi8vZ28uZGV2L1ZFUlNJT04/bT10ZXh0JyB8IGhlYWQgLTEpLmxpbnV4LWFtZDY0LnRhci5neiIgXAogICAgICAgJiYgcm0gLXJmIC91c3IvbG9jYWwvZ28gXAogICAgICAgJiYgY2QgL3Vzci9sb2NhbCAmJiB0YXIgenhmIC90bXAvZ28tbGludXgtYW1kNjQudGFyLmd6ICYmIGNkIC0gPiAvZGV2L251bGwgXAogICAgICAgJiYgY2QgL3Vzci9sb2NhbC9iaW4gJiYgZm9yIGZpbGUgaW4gL3Vzci9sb2NhbC9nby9iaW4vKjsgZG8gbG4gLXMgLWYgJGZpbGU7IGRvbmUgJiYgY2QgLSA+IC9kZXYvbnVsbCBcCiAgICAgICAmJiBybSAtZiAvdG1wL2dvLWxpbnV4LWFtZDY0LnRhci5neiBcCiAgICAgICAmJiBnbyB2ZXJzaW9uID4mMgoKIyBjcmVhdGUgZGVmYXVsdCB1c2VyCkFSRwlERUZBVUxUX1VTRVI9dWJ1bnR1CkVOViAgICAgREVGQVVMVF9VU0VSPSR7REVGQVVMVF9VU0VSfQpSVU4gICAgIGdldGVudCBwYXNzd2QgJHtERUZBVUxUX1VTRVJ9ICYmIHVzZXJtb2QgLS1ncm91cHMgcm9vdCAke0RFRkFVTFRfVVNFUn0gXAogICAgICAgIHx8IHVzZXJhZGQgLS11c2VyLWdyb3VwIC0tZ3JvdXBzIHJvb3QgLS1jcmVhdGUtaG9tZSAtLXNoZWxsIC9iaW4vYmFzaCAke0RFRkFVTFRfVVNFUn0KCiMgY2hhbmdlIGRlZmF1bHQgdXNlciB0byBiZSBob3N0IHVzZXIgbmFtZQpBUkcgIEhPU1RfVVNFUgpBUkcgIEhPU1RfSE9NRQpBUkcgIEhPU1RfU0hFTEwKUlVOICB1c2VybW9kIC0tbG9naW4gJHtIT1NUX1VTRVJ9IC0taG9tZSAke0hPU1RfSE9NRX0gLS1tb3ZlLWhvbWUgLS1zaGVsbCAke0hPU1RfU0hFTEx9ICR7REVGQVVMVF9VU0VSfSBcCiAgICAgJiYgZ3JvdXBtb2QgLS1uZXctbmFtZSAke0hPU1RfVVNFUn0gJHtERUZBVUxUX1VTRVJ9ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgXAogICAgICYmIHVzZXJtb2QgLS1ncm91cHMgcm9vdCAke0hPU1RfVVNFUn0gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIFwKICAgICAmJiB1c2VybW9kIC0tZ3JvdXBzIHN1ZG8gJHtIT1NUX1VTRVJ9ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICBcCiAgICAgJiYgZWNobyAiJHtIT1NUX1VTRVJ9IEFMTD0oQUxMKSBOT1BBU1NXRDpBTEwiID4+IC9ldGMvc3Vkb2Vycy5kLzEwLXVzZXIKRU5WICBERUZBVUxUX1VTRVI9JHtIT1NUX1VTRVJ9CgojIGNvcHkgc2NyaXB0cwpDT1BZIC4vc2NyaXB0cy8gL3Vzci9sb2NhbC9zY3JpcHRzLwpSVU4gIGNobW9kIHVnbytyeCAvdXNyL2xvY2FsL3NjcmlwdHMvKi5zaAoKIyBiYWNrIHRvIGRlZmF1bHQgdXNlcgpVU0VSICR7REVGQVVMVF9VU0VSfQpSVU4gIHRvdWNoIC9ob21lLyR7REVGQVVMVF9VU0VSfS8uc3Vkb19hc19hZG1pbl9zdWNjZXNzZnVsCgojIGFkZCBoZXJlIFZTQ29kZSBnbyBtb2R1bGVzIHJlcXVpcmVkIGJ5IHlvdXIgcHJvamVjdCAKUlVOIGdvIGluc3RhbGwgaG9ubmVmLmNvL2dvL3Rvb2xzL2NtZC9zdGF0aWNjaGVja0AyMDIzLjEuNiBcCiAgICAgJiYgZ28gY2xlYW4gLWNhY2hlICYmIGdvIGNsZWFuIC1tb2RjYWNoZQoKIyBlbnN1cmUgd2Ugc2xlZXAgZm9yZXZlciBvdGhlcndpc2UgZGV2IGNvbnRhaW5lciB3aWxsIGV4aXQgdXBvbiBjcmVhdGlvbgpDTUQgWyAic2xlZXAiLCAiaW5maW5pdHkiIF0KRU5UUllQT0lOVCBbICIvdXNyL2xvY2FsL3NjcmlwdHMvc3RhcnR1cC5zaCIgXQo=' | base64 -d > ./.devcontainer/Dockerfile ; }

.PHONY: prepare
prepare:
	@test -f variables.mk || { echo "Generating variables.mk file" ; printf '## Binary\nTARGET = '$$( pwd | sed 's#^.*/##' )'\n\n## Docker target image\nDOCKER_IMAGE ?= $$(TARGET)\n\n## Docker target image tag\nDOCKER_IMAGE_TAG ?= latest\n\n## Docker run additional parameters\nDOCKER_RUN_PARAMETERS ?=\n\n## Gitlab URL\nGITLAB_URL ?=\n\n## Gitlab access token\nGITLAB_TOKEN ?=\n\n## Gitlab destination package\nGITLAB_PACKAGE ?= $$(DOCKER_IMAGE)\n\n## Nexus URL\nNEXUS_URL ?=\n\n## Nexus credentials user\nNEXUS_AUTH_USER ?=\n\n## Nexus credentials pass\nNEXUS_AUTH_PASS ?=\n\n' > variables.mk ; echo "You must rerun <make prepare>" ; exit 1 ; }
	@test -f go.mod || { echo "Generating go.mod file for module $(TARGET)" ; go mod init $(TARGET) ; }
	@test -d ./bin || { echo "Generating directory ./bin" ; mkdir -p ./bin ; }
	@test -d ./internal || { echo "Generating directory ./internal" ; mkdir -p ./internal ; }
	@test -d ./pkg || { echo "Generating directory ./pkg" ; mkdir -p ./pkg ; }
	@test -d ./tools || { echo "Generating directory ./tools" ; mkdir -p ./tools ; }
	@test -f main.go || { $(MAKE) --no-print-directory prepareroot ; }
	@test -f ./cmd/$(TARGET)/main.go || { grep "package[[:blank:]]*main" main.go > /dev/null || $(MAKE) --no-print-directory preparecmd ; }
	@test -f $(TARGET).ico || { echo "Generating ./$(TARGET).ico" ; printf 'AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAD19fWf9fX1//X19f/29vb/9fX1//b29v/b29v/ubm5/7m5uf/a2tr/9fX1//b29v/19fX/9vb2//b29v/29vaf9vb2n/b29v/29vb/9fX1/52dnf8vLy//Nzc3/19fX/9fX1//Nzc3/y4uLv+cnJz/9fX1//b29v/29vb/9fX1n/b29p/29vb/7e3t/01NTf9WVlb/3d3d//7+/v/+/v7//v7+//7+/v/f39//WVlZ/0tLS//s7Oz/9vb2//X19Z/19fWf9fX1/09PT/+EhIT//v7+//7+/v/+/v7//v7+//7+/v/+/v7//v7+//7+/v+Hh4f/S0tL//X19f/29vaf9vb2n6CgoP9UVFT//v7+//7+/v/+/v7//v7+//7+/v/+/v7//v7+//7+/v/+/v7//v7+/1dXV/+dnZ3/9fX1n/X19Z8yMjL/29vb//7+/v/+/v7/z8/P/6Kiov/+/v7/+/v7/4KCgv/t7e3//v7+//7+/v/d3d3/MDAw//b29p/S0tKfMjIy//7+/v/+/v7//v7+//n5+f82Njb/zc3N/4uLi/9kZGT//v7+//7+/v/+/v7//v7+/zU1Nf/Ozs6fnp6en1paWv/+/v7//v7+//7+/v/+/v7/1dXV/zc3N/8qKir/8fHx//7+/v/+/v7//v7+//7+/v9dXV3/l5eXn5ycnJ9aWlr//v7+//7+/v/+/v7//v7+//T09P8qKir/UVFR//7+/v/+/v7//v7+//7+/v/+/v7/XV1d/5qamp/T09OfMjIy//7+/v/+/v7//v7+//7+/v9sbGz/iYmJ/2BgYP+jo6P//v7+//7+/v/+/v7//v7+/zU1Nf/Ozs6f9vb2nzMzM//a2tr//v7+//7+/v/Z2dn/WFhY//v7+//t7e3/U1NT//Pz8//+/v7//v7+/9zc3P8wMDD/9fX1n/X19Z+ioqL/UlJS//7+/v/+/v7//v7+//7+/v/+/v7//v7+//7+/v/+/v7//v7+//7+/v9VVVX/n5+f//b29p/29vaf9fX1/1FRUf+BgYH//v7+//7+/v/+/v7//v7+//7+/v/+/v7//v7+//7+/v+EhIT/Tk5O//X19f/19fWf9fX1n/b29v/u7u7/UFBQ/1NTU//b29v//v7+//7+/v/+/v7//v7+/9zc3P9VVVX/Tk5O/+3t7f/29vb/9vb2n/X19Z/29vb/9vb2//X19f+goKD/MTEx/zQ0NP9cXFz/XFxc/zQ0NP8wMDD/n5+f//X19f/29vb/9vb2//b29p/29vaf9vb2//b29v/19fX/9vb2//X19f/e3t7/vLy8/7u7u//d3d3/9vb2//X19f/29vb/9fX1//X19f/19fWfAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==' | base64 -d > $(TARGET).ico ; }
	@test -f ./.gitignore || { echo "Generating ./.gitignore" ; printf '# Binary\n/'$(TARGET)'\n/'$(TARGET)'-*\n/'$(TARGET)'_*\n/'$(TARGET)'.exe\n/'$(TARGET)'.exe.manifest\n# Builder\n/.makerc\n# OS specific\n/versioninfo.rc\n/main_darwin.go\n/main_linux.go\n/main_windows.go\n/resource.syso\n# Other\n' > ./.gitignore ; }
	@test -f README.md || { echo "Generating ./README.md" ; printf '# '$(TARGET) > README.md ; }
	@test -d ./.devcontainer || $(MAKE) --no-print-directory preparedevcontainer

prepare/% :
	test -z "$(wordlist 2,3,$(subst /, ,$*))" && { int='$*' ; test -d ./cmd/$$int || mkdir -p ./cmd/$$int ; test -d ./internal/$$int || mkdir -p ./internal/$$int ; test -f ./internal/$$int/main.go || printf 'package '$${int}'\nimport (\n    "$(TARGET)/flags"\n)\nvar (\n    f = flags.NewFlag("$(TARGET)")\n)\nfunc usage() {\n    fmt.Print(`\nUsage:  $(TARGET) key COMMAND\n`)\n}\nfunc Main(args []string) {\n}\n' > ./internal/$$int/main.go ; }

## Local installation
.PHONY: install
install:
	@test ! -d ~/bin || { file=$(TARGET)-$$(uname | sed 's/[_-].*$$//' | tr '[:upper:]' '[:lower:]') ; test -f $$file || file=$(TARGET) ; test ! -f $$file || { echo "Copying to ~/bin/$(TARGET)" ; cp $$file ~/bin/$(TARGET) ; } ; }

## SCP remote installation
install/%:
	@test -z "$(wordlist 2,3,$(subst /, ,$*))" && { dest='$*' ; file=$(TARGET)-$$(uname | sed 's/[_-].*$$//' | tr '[:upper:]' '[:lower:]') ; test -f $$file || file=$(TARGET) ; test ! -f $$file || { echo "Sending $$file to $$dest..." ; scp $$file $$dest:$(TARGET) ; } ; }

## Git Management
.PHONY: fixup
fixup :
	@which git > /dev/null 2>&1 && { $(GIT) add . ; $(GIT) commit --fixup=HEAD ; GIT_EDITOR=true $(GIT) rebase --interactive --autosquash HEAD~2 ; $(GIT) push --force ; }

.PHONY: pull
pull :
	@which git > /dev/null 2>&1 && { $(GIT) fetch --force ; $(GIT) fetch --all --prune --prune-tags ; $(GIT) rebase ; $(GIT) pull --all ; }

### Exemple de targets avec /

## Print service logs : (replace % with service name)
.PHONY: log
log :
	@test -z "$(filter-out $@,$(MAKECMDGOALS))" && $(MAKE) --no-print-directory availableservices || for service in $(filter-out $@,$(MAKECMDGOALS)) ; do $(MAKE) --no-print-directory log/$$service ; done

log/%:
	@-test -z "$(wordlist 2,3,$(subst /, ,$*))" && { service='$*' ; echo "Service = $$service" ; }

.PHONY: availableservices
availableservices :
	@echo "Available services: 1 2 3"

1 2 3 :
	@:


# Default target
%:
	@:
