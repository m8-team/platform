#!/usr/bin/env bash
# Run one disruption against a disposable Docker fixture created by this script.
set -euo pipefail

mode=${1:-restart}
case "$mode" in
  restart|failover) ;;
  *) echo "usage: $0 [restart|failover]" >&2; exit 2 ;;
esac

scratch=$(mktemp -d)
name="m8-event-disruption-$(basename "$scratch" | tr -cd '[:alnum:]' | tr '[:upper:]' '[:lower:]')"
image=ydbplatform/local-ydb:26.3.1.25
test_pid=
cleanup() {
  code=$?
  trap - EXIT
  if [[ -n "$test_pid" ]]; then
    kill "$test_pid" 2>/dev/null || true
    wait "$test_pid" 2>/dev/null || true
  fi
  if (( code != 0 )); then
    [[ ! -f "$scratch/test.log" ]] || cat "$scratch/test.log" >&2
    docker logs --tail=30 "$name" >&2 2>/dev/null || true
  fi
  docker rm -f "$name" >/dev/null 2>&1 || true
  rm -f "$scratch/ready" "$scratch/ready.tmp" "$scratch/resume" "$scratch/test.log" "$scratch/event.test"
  rmdir "$scratch"
  exit "$code"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

wait_for_local_ydb() {
  for ((i=0; i<45; i++)); do
    if docker exec "$name" /health_check >/dev/null 2>&1; then
      return
    fi
    sleep 1
  done
  echo "disposable YDB did not become healthy" >&2
  return 1
}

if [[ "$mode" == restart ]]; then
  port=${M8_EVENT_FIXTURE_GRPC_PORT:-32236}
  docker run -d --hostname localhost --name "$name" -p "127.0.0.1:$port:$port" \
    -e "GRPC_PORT=$port" "$image" >/dev/null
  wait_for_local_ydb
  M8_EVENT_TEST_DISRUPTION_DIR="$scratch" M8_EVENT_TEST_DISRUPTION_MODE=restart \
    M8_EVENT_TEST_YDB_REQUIRED=1 M8_EVENT_TEST_YDB_DSN="grpc://127.0.0.1:$port/local" \
    M8_EVENT_TEST_YDB_PREFIX=/local \
    go test -count=1 -v -timeout=4m -run '^TestDisruptionYDB$' ./internal/platform/event >"$scratch/test.log" 2>&1 &
  test_pid=$!
else
  # The official Docker entrypoint's fixed ports collide for multi-node setups.
  # Run its deployment helper with distinct ports in the container namespace.
  docker run -d --hostname localhost --name "$name" --entrypoint bash \
    -e YDB_TINY_MODE=true -e YDB_GRPC_ENABLE_TLS=false -e YDB_PDISKS_DIRECTORY=/ydb_data/pdisks \
    "$image" -c 'mkdir -p /ydb_data/pdisks && /local_ydb deploy --ydb-working-dir /ydb_data --ydb-binary-path /ydbd --erasure mirror-3-dc && exec sleep infinity' >/dev/null
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o "$scratch/event.test" ./internal/platform/event
  for ((i=0; i<120; i++)); do
    if docker exec "$name" test -s /ydb_data/ydb_recipe.json; then
      break
    fi
    sleep 1
  done
  port=$(docker exec "$name" python3 -c 'import json; print(json.load(open("/ydb_data/ydb_recipe.json"))["nodes"]["1"]["grpc_port"])')
  docker cp "$scratch/event.test" "$name":/tmp/event.test
  docker exec "$name" mkdir /tmp/m8-event-control
  docker exec \
    -e M8_EVENT_TEST_DISRUPTION_DIR=/tmp/m8-event-control -e M8_EVENT_TEST_DISRUPTION_MODE=failover \
    -e M8_EVENT_TEST_YDB_REQUIRED=1 -e "M8_EVENT_TEST_YDB_DSN=grpc://127.0.0.1:$port/local" \
    -e M8_EVENT_TEST_YDB_PREFIX=/local \
    "$name" /tmp/event.test -test.v -test.timeout=4m -test.run '^TestDisruptionYDB$' >"$scratch/test.log" 2>&1 &
  test_pid=$!
fi

ready=
for ((i=0; i<90; i++)); do
  if [[ "$mode" == restart && -f "$scratch/ready" ]]; then
    ready=$(cat "$scratch/ready")
    break
  fi
  if [[ "$mode" == failover ]] && ready=$(docker exec "$name" cat /tmp/m8-event-control/ready 2>/dev/null); then
    break
  fi
  kill -0 "$test_pid"
  sleep 1
done
[[ "$ready" =~ ^[1-9][0-9]*$ ]]

if [[ "$mode" == restart ]]; then
  docker kill --signal=KILL "$name"
  docker start "$name"
  wait_for_local_ydb
  touch "$scratch/resume"
else
  # Resolve the tablet's owner to the PID in this fixture and verify the
  # command line before delivering a signal to the disposable YDB process.
  docker exec "$name" python3 -c '
import json, os, signal, sys, time
node = sys.argv[1]
pid = json.load(open("/ydb_data/ydb_recipe.json"))["nodes"][node]["pid"]
args = open(f"/proc/{pid}/cmdline", "rb").read().split(b"\0")
assert args[:2] == [b"/ydbd", b"server"] and f"--node={node}".encode() in args
os.kill(pid, signal.SIGKILL)
for _ in range(50):
    if not os.path.exists(f"/proc/{pid}/status") or "State:\tZ" in open(f"/proc/{pid}/status").read():
        break
    time.sleep(0.1)
else:
    sys.exit("tablet node did not stop")
' "$ready"
  docker exec "$name" touch /tmp/m8-event-control/resume
fi
wait "$test_pid"
test_pid=
cat "$scratch/test.log"
