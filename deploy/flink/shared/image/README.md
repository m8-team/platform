# Shared Flink runtime image

This image is installation-neutral. It is based on Flink 2.2.1 with Java 17 and
adds only the Kafka SQL connector, the built-in Presto S3 plugin activation, and
Yandex trust anchors.

The two certificate files are the intermediate and root certificates published
in Yandex Cloud's `CA.pem` bundle. Their SHA-256 fingerprints are:

```text
YandexCloudCA.crt         E1:D5:3D:D1:D7:56:6D:0D:C6:91:C9:ED:6F:CA:0C:91:0F:58:B9:5D:4E:D7:F0:A9:58:AC:C7:67:A1:B2:49:37
YandexInternalRootCA.crt  E3:C3:19:26:46:53:6B:C4:FF:AE:6D:D3:43:24:EF:9D:D8:D3:B1:6D:FA:13:A2:4E:13:18:0E:F1:4B:A2:1B:BA
```

Source: <https://storage.yandexcloud.net/cloud-certs/CA.pem>

Rotate the checked-in trust anchors through a reviewed change before certificate
expiry. No installation configuration or credentials belong in this directory.
