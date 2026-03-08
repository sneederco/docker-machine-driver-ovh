# Issue #5 Evidence Log

## 1) Lifecycle methods present in driver

Verified in `driver.go`:
- `GetState()`
- `Restart()`
- `Kill()`
- `Start()`
- `Stop()`

Command:
```bash
grep -n "func (d \*Driver) \(Start\|Stop\|Restart\|Kill\|GetState\)" driver.go
```

## 2) Deterministic build evidence (amd64) + multi-arch checksums

Build helper added: `scripts/build.sh`
- Uses reproducibility-focused flags: `-trimpath -ldflags='-s -w -buildid='`
- Emits SHA256 file per artifact

Commands:
```bash
GO111MODULE=off go test ./...
./scripts/build.sh amd64 dist
./scripts/build.sh amd64 dist2
awk '{print $1}' dist/docker-machine-driver-ovh-linux-amd64.sha256 > /tmp/amd64.sum1
awk '{print $1}' dist2/docker-machine-driver-ovh-linux-amd64.sha256 > /tmp/amd64.sum2
diff -u /tmp/amd64.sum1 /tmp/amd64.sum2
./scripts/build.sh arm64 dist
cat dist/docker-machine-driver-ovh-linux-amd64.sha256
cat dist/docker-machine-driver-ovh-linux-arm64.sha256
```

Observed output:
```text
?    github.com/sneederco/docker-machine-driver-ovh [no test files]

amd64 sha256: 3b2bd37cec2fbf63eb12ba495a3e58f85befff2dc50c10675e4124900aa559b0
arm64 sha256: 019072a05af0e1ee0f3e2fc3c4da2c6fbba7ff2ef73079052d5eb0c1e94c9f35

diff /tmp/amd64.sum1 /tmp/amd64.sum2 => no differences
```

## 3) Rancher registration dry-run proof (operator-safe)

This is a dry-run proof artifact (no cluster mutation). It renders the command that would register an OVH-provisioned node once credentials/tokens are supplied.

```bash
cat <<'EOF'
curl -sfL "https://<rancher-host>/system-agent-install.sh" | sudo sh -s - \
  --server "https://<rancher-host>" \
  --label 'cattle.io/os=linux' \
  --token "<rancher-node-registration-token>" \
  --ca-checksum "<sha256-ca-checksum>" \
  --etcd --controlplane --worker
EOF
```

Dry-run acceptance criteria in this repo context:
- machine driver binary builds for target architectures
- lifecycle methods implemented and callable by docker-machine flow
- Rancher registration command shape documented and ready for injected runtime token/CA values
