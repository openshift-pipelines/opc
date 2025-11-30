ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.24
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:61d5ad475048c2e655cd46d0a55dfeaec182cc3faa6348cb85989a7c9e196483

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/opc
COPY . .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GOEXPERIMENT="strictfipsruntime"
RUN go build -buildvcs=false -mod=vendor -tags disable_gcp,strictfipsruntime  -o /tmp/opc main.go

FROM $RUNTIME
ARG VERSION=opc-main
COPY --from=builder /tmp/opc /usr/bin

RUN microdnf install -y shadow-utils && \
    groupadd -r -g 65532 nonroot && useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

LABEL \
      com.redhat.component="openshift-pipelines-opc-rhel9-container" \
      name="opc" \
      version=$VERSION \
      com.redhat.component="opc" \
      io.k8s.display-name="opc" \
      maintainer="pipelines-extcomm@redhat.com" \
      summary="A CLI for OpenShift Pipeline" \
      description="opc makes it easy to work with Tekton resources in OpenShift Pipelines. It is built on top of tkn and tkn-pac and expands their capablities to the functionality and user-experience that is available on OpenShift." \
      io.k8s.description="opc makes it easy to work with Tekton resources in OpenShift Pipelines. It is built on top of tkn and tkn-pac and expands their capablities to the functionality and user-experience that is available on OpenShift." \
      io.openshift.tags="pipelines,tekton,openshift"
