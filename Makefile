BINARYPATH="bin/mockoracle"
CONFIGPATH="./example/config.toml"

GOOS    := $(if $(GOOS),$(GOOS),$(shell go env GOOS)) 
GOARCH  := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOENV   := GO111MODULE=on CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO      := $(GOENV) go
GOBUILD := $(GO) build

CMDPATH="./cmd/oracle"

build: clean gotool                                                                                                                                                                                               
	$(GOBUILD) -ldflags '$(LDFLAGS)' -o $(BINARYPATH) $(CMDPATH) 

gotool:
	$(GO) mod tidy

clean:
	@if [ -f ${BINARYPATH}  ] ; then rm ${BINARYPATH} ; fi
