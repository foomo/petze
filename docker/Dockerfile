FROM scratch


# timezone setup
ENV ZONEINFO /zoneinfo.zip
ADD files/zoneinfo.zip /zoneinfo.zip

# install ca root certificates for outgoing https calls
# https://curl.haxx.se/docs/caextract.html
# http://blog.codeship.com/building-minimal-docker-containers-for-go-applications/
COPY files/cacert.pem /etc/ssl/certs/ca-certificates.crt

COPY petze /usr/sbin/petze

# Start petze and search the config in "/etc/petzconf"
ENTRYPOINT ["/usr/sbin/petze", "/etc/petzconf"]
