FROM scratch
COPY --chown=65532:65532 event-reconciler /event-reconciler
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 65532:65532
ENTRYPOINT ["/event-reconciler"]
