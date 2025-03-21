ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.22
ARG RUNTIME=registry.redhat.io/ubi8/ubi:latest@sha256:8bd1b6306f8164de7fb0974031a0f903bd3ab3e6bcab835854d3d9a1a74ea5db

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/opc

COPY . .
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
RUN git config --global --add safe.directory /go/src/github.com/openshift-pipelines/opc
RUN git rev-parse HEAD > /tmp/HEAD
RUN CGO_ENABLED=0 \
    go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat /tmp/HEAD)'" -mod=vendor -tags disable_gcp -v -o /app/opc main.go

FROM $RUNTIME
ARG VERSION=opc-1.17

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /app/opc ${KO_APP}/opc
COPY --from=builder /tmp/HEAD ${KO_DATA_PATH}/HEAD    

LABEL \
      com.redhat.component="openshift-pipelines-opc-rhel8-container" \
      name="openshift-pipelines/pipelines-opc-rhel8" \
      version=$VERSION \
      summary="A CLI for OpenShift Pipeline" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="opc makes it easy to work with Tekton resources in OpenShift Pipelines. It is built on top of tkn and tkn-pac and expands their capablities to the functionality and user-experience that is available on OpenShift." \
      io.k8s.display-name="Red Hat OpenShift Pipelines opc" \
      io.k8s.description="Red Hat OpenShift Pipelines opc" \
      io.openshift.tags="pipelines,tekton,openshift"

RUN groupadd -r -g 65532 nonroot && useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532
      
ENTRYPOINT ["/ko-app/opc"]
