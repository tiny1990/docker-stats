FROM alpine:3.6

COPY dp-docker-stats /

VOLUME ["/var/run/docker.sock"]

ENTRYPOINT ["/dp-docker-stats"]