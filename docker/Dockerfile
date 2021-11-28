ARG IMAGE=alpine
ARG TAG=3.14

FROM ${IMAGE}:${TAG} AS build

ENV USER=acidns
ENV UID=10001

RUN apk --no-cache add ca-certificates \
    && update-ca-certificates \
    && adduser \
   --disabled-password \
   --gecos "" \
   --home "/nonexistent" \
   --shell "/sbin/nologin" \
   --no-create-home \
   --uid "${UID}" \
   "${USER}"

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

COPY aci-dns-manager /aci-dns-manager

USER ${USER}:${USER}

ENTRYPOINT [ "/aci-dns-manager" ]
