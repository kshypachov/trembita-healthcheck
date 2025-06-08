# Trembita Security Server Healthcheck Utility

This Go-based utility performs a full-stack **healthcheck** of the [Trembita 2.0](https://www.egov.org.ua/) Security Server (ШБО). It checks whether the proxy component is fully operational by simulating a real encrypted and signed SOAP transaction to the internal `getSecurityServerHealthData` meta-service.

## 📦 Description

The healthcheck works in several steps:

1. **HTTP GET** to `http://127.0.0.1/listSecurityServerClients` to retrieve the local Security Server client metadata (e.g., `xRoadInstance`, `memberClass`, and `memberCode`).
2. **Parse the XML** response and extract a client of type `MEMBER`.
3. **Build a valid SOAP request** to the meta-service `getSecurityServerHealthData`.
4. **Send a POST HTTPS request** to `https://127.0.0.1/` using client TLS authentication.
5. **Check the response body** for a `<om:getSecurityServerHealthDataResponse>` tag.

If the tag is found, the server is healthy and considered fully operational.

> **Why this works**: The request simulates a real transaction. If successful, it confirms that the security server:
> - Is running
> - Can access certificates and keys
> - Can parse and decrypt encrypted requests
> - Can process SOAP calls

## ⚙️ Usage

```bash
go run main.go --log-level=info
```

### Options

| Flag         | Description                                |
|--------------|--------------------------------------------|
| `--log-level`| Logging verbosity: `fatal`, `error`, `warn`, `info` (default: `info`) |

## 🔐 TLS Requirements

This utility uses mutual TLS to access the Security Server. You must:

1. Generate a certificate/key pair for the client.
2. Place them in the following paths:

```bash
/etc/certs/tls.crt
/etc/certs/tls.key
```

3. Register `tls.crt` as a **trusted client certificate** using the Web UI of the Security Server.

> 🚨 **Note:** Trembita 2.0 requires that all meta-service requests must:
> - Use **HTTPS**
> - Be authenticated with a **client certificate**

## 🛠 Dependencies

- Go 1.18+ (tested)
- [Logrus](https://github.com/sirupsen/logrus) logger (`go get github.com/sirupsen/logrus`)

## ✅ Exit Codes

| Code | Meaning                     |
|------|-----------------------------|
| `0`  | Healthcheck successful      |
| `1`  | General failure             |

## 📄 License

MIT or other, if applicable.
