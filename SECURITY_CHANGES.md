# Security review: changes and remaining work

simpleca is intended for home lab use. Finding numbers come from the code review.

## Changes on the `security-hardening` branch

### 1. CA private key could be downloaded

- `web -dir` now defaults to empty, and no static files are served unless it is set.
- `web` refuses to start if the CA key or TLS key is inside the served directory. Symbolic links are resolved first.

### 4. Certificate lifetime

- Certificates are limited to 1095 days (about three years) by `ca.MaxCertDays`, enforced in `CASign`. This covers `/sign`, `/crt` and `ca sign`.
- The default validity for `/sign`, `/crt` and `ca sign` is now 1095 days, previously 3650.
- `/sign` and `/crt` return 400 for a validity below 1 day or above the limit.
- `web -nbdays` and `acme -days` outside 1 to 1095 stop the server at startup.
- ACME applies the same limit through its startup check.
- CA certificates are not limited. One auto-created by `web` now always lasts 3650 days instead of following `-nbdays`.

### 5. SHA-1 signatures

- `CASign` no longer copies the signature algorithm from the CSR. It is chosen from the CA key.
- As a side effect, ECDSA CSRs can now be signed by an RSA CA.

### 6. Private key file modes

- Private key files are written as 0600, including when an existing file is replaced.

### 9. Overload limits

- `web` and `acme` have read, write and idle timeouts, and a 1 MiB request body limit (413 in the `web` API).
- The `web` API generates RSA keys of at most 4096 bits, or the `-size` value if larger.
- Key generation and passphrase handling run in at most one slot per CPU. Extra requests get 503 with `Retry-After`.
- The iteration count of an encrypted key supplied to the API is capped at 10,000,000.
- ACME keeps the latest 10,000 orders and access to them is now locked.
- As a side effect, `/order/-1` now returns 404 instead of panicking.

### 10. Weak keys and encryption

- RSA generation requires at least 2048 bits, on the CLI and in the API.
- Signing rejects RSA keys under 2048 or over 16384 bits and curves weaker than P-256, for `web`, `ca sign` and ACME.
- Passphrase-protected keys are written as PKCS#8 `ENCRYPTED PRIVATE KEY`: PBKDF2-HMAC-SHA256 with 600,000 iterations and AES-256-CBC. OpenSSL can read them.
- Keys in the old encrypted PEM format can still be loaded.

### Supporting changes

- `go.mod` moves from Go 1.18 to 1.24, which `crypto/pbkdf2` needs. This also drops the old runtime security defaults.
- Non-constant `Fprintf` calls became `Fprint`, because `go vet` rejects them at this Go version.
- `make run` no longer passes `-dir .`, which would now refuse to start.
- README and `swagger.yaml` are updated.
- Unit tests were added in `internal/key` and `internal/web`.

## Upgrade notes

- Existing key files keep their mode until rewritten: run `chmod 600 ca.key`.
- Anyone relying on the file server from the current directory must pass `-dir` and keep keys outside it.
- Re-encrypt old RSA keys with `simpleca key crypt -passphrase OLD -outpass NEW -out new.key old.key`.
- Scripts requesting more than 1095 days, or RSA keys under 2048 bits, will now get an error.

## Left to do

### Deliberately deferred

| Finding | Item |
|---|---|
| 2, 3 | ACME does not verify JWS signatures, nonces, challenges, CSR signatures, or that a CSR matches its order. Accepted for home lab use; add a runtime toggle to enforce validation later. |
| 4 | The `web` API has no authentication or name policy, so anyone who can reach it can get a certificate for any name. Accepted for home lab use; a runtime toggle to enforce this could be added. |

### Security

| Finding | Item |
|---|---|
| 7 | Passphrases are sent in query strings and logged. `-ca-pass` is visible in `ps`. |
| 8 | With `-ssl` and no `-key`, the CA key is also the TLS key. |

### Correctness

| Finding | Item |
|---|---|
| 11 | CSR and CA creation hardcode RSA signature algorithms, so an ECDSA key still fails in `/csr` and `csr create`. |
| 12 | Plain PKCS#8 EC keys and legacy-encrypted EC keys cannot be loaded. `key crypt` does not support EC keys. |
| 13 | `web` keeps running without serving if the port is in use. |
| 14 | `r.ParseForm()` consumes form-encoded bodies, so curl's default content type fails. |
| 16 | An invalid IP returns 500 instead of 400. |
| 17 | `cert self -days` is ignored: always 180 days. |
| 18 | `csr read` panics on a CSR with no extensions. |
| 19 | `cert read` prints the Subject CN as the Issuer CN. |
| 21 | ACME URLs come from the `Host` header and always use `https`. |
| 22 | ACME advertises `revoke-cert` and `key-change`, which are not implemented. |
| 23 | `entrypoint.sh` does not export `SIMPLECA_*`, so settings such as `ISSUER_CERT_URL` are ignored. `web` also reads `nbdays`, not `days`. |
| 24 | The Docker healthcheck uses port 80 but the server only listens on 443. |
| 25 | A non-string config value panics. |
| 26 | The CA serial is hardcoded to 2019. |

### Minor

- `ca sign` ignores errors from `WriteCertFile`.
- `/crt` returns private keys without the `no-store` header.
- `LoadCertFile` indexes `certs[0]` before checking the error.
- `go vet` reports malformed struct tags in `internal/acme/acme.go` and an unbuffered signal channel in `internal/web/main.go`.
