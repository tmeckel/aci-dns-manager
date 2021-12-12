FROM mcr.microsoft.com/azuredocs/aci-helloworld

ADD https://github.com/just-containers/s6-overlay/releases/download/v2.2.0.3/s6-overlay-amd64-installer /tmp/
RUN chmod +x /tmp/s6-overlay-amd64-installer \
    && /tmp/s6-overlay-amd64-installer / \
    && apk add --no-cache curl \
    && curl -L https://github.com/tmeckel/aci-dns-manager/releases/download/v0.0.1/aci-dns-manager_Linux_x86_64.tar.gz \
    | tar xvzf - aci-dns-manager --directory /usr/bin/ \
    && chmod 555 /usr/bin/aci-dns-manager

COPY update-dns.sh /etc/cont-init.d/update-dns.sh
COPY remove-dns.sh /etc/cont-finish.d/remove-dns.sh

ENTRYPOINT ["/init"]
CMD ["node", "./index.js"]
