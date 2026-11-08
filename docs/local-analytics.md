# Local analytics: YTsaurus, CHYT, and DataLens

The local analytics stack is isolated from the main development compose file in
`docker-compose.analytics.yml`. It starts a single-node YTsaurus cluster, a
single-instance CHYT clique named `ch_public`, and the official DataLens stack.

This setup is intended only for local development. It uses the default DataLens
development credentials and has no production hardening or fault tolerance.

## Requirements

- Docker Compose v2 with support for `include`
- at least 12 GB of memory available to Docker (8 GB for YTsaurus and 4 GB
  for DataLens are the upstream minimums)
- on Apple Silicon, Docker Desktop with Rosetta enabled

YTsaurus recommends x86_64 Linux. The compose file therefore defaults its
YTsaurus and CHYT containers to `linux/amd64`; emulation on Apple Silicon is not
guaranteed by upstream.

## Start and stop

```bash
make analytics:up
make analytics:status
make analytics:logs
make analytics:down
```

To remove YTsaurus and DataLens data as well:

```bash
make analytics:reset
```

Startup can take several minutes on the first run because the images are large
and CHYT has to upload its binaries and schedule a YTsaurus operation.

## Endpoints

| Component | URL | Credentials |
| --- | --- | --- |
| YTsaurus HTTP proxy | <http://localhost:8000> | none |
| YTsaurus UI | <http://localhost:8001> | none |
| DataLens | <http://localhost:18081> | `admin` / `admin` |

## Connect DataLens to CHYT

Create a **YTsaurus CHYT** connection in DataLens with:

- host: `yt-backend`
- port: `80`
- clique alias: `ch_public`
- HTTPS: disabled
- token: empty

`localhost` must not be used as the host: DataLens runs in a container and
reaches YTsaurus over the compose network.

Check CHYT directly from the host:

```bash
curl --location-trusted \
  'http://localhost:8000/chyt?chyt.clique_alias=ch_public' \
  --data-binary 'SELECT 1'
```

## Versions and ports

Override defaults when needed:

```bash
CHYT_VERSION=2.18.6 YT_PROXY_PORT=18000 YT_UI_PORT=18001 \
  make analytics:up
```

DataLens is pinned in `deploy/local/datalens/docker-compose.yaml` to upstream
commit `1cebb39b37e3ccd879f43ba17d96d98fb90c85ab`. Its local overrides are in
`deploy/local/datalens/.env`.
