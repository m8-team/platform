FROM scratch
COPY --chmod=0555 event-reconciler /event-reconciler
COPY --chmod=0444 ca-certificates.crt /ca-certificates.crt
ENV SSL_CERT_FILE=/ca-certificates.crt
USER 65532:65532
ENTRYPOINT ["/event-reconciler"]
