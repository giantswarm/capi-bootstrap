FROM ubuntu AS binaries

WORKDIR /binaries

RUN apt-get update \
    && apt-get install -y ca-certificates curl jq
COPY build.sh .
RUN --mount=type=secret,id=github_token ./build.sh

FROM ubuntu

COPY --from=binaries /binaries/* /usr/bin/
COPY ./capi-bootstrap /usr/bin/capi-bootstrap

ENTRYPOINT ["capi-bootstrap"]
