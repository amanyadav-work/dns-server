
# DNS GO — Code Overview

This project is a minimal DNS server written in Go, designed to answer only A record (IPv4 address) queries for a small set of domains. It is intentionally simple, making it easy to understand the DNS protocol and Go networking basics.

## How It Works

**Supported Record Types:**
- Only DNS A records (IPv4) are supported. All other record types (e.g., AAAA, MX) will return no answer.

**Supported Domains:**
- The list of domains and their IPv4 addresses is defined in `names.json`. Example:
  - google.com → 3.1.3.7
  - acint.net → 192.168.0.102
  - yadavaman.duckdns.org → 100.125.140.68

**Port:**
- The server listens on UDP port **8282**.

## File Roles

- `main.go`: Entry point. Sets up the UDP server, handles incoming DNS requests, parses queries, and sends responses.
- `dblookup.go`: Loads the domain-to-IP mapping from `names.json` and provides lookup logic.
- `models/dns.go`: Contains Go structs for DNS headers, resource records, and domain models.
- `names.json`: The configuration file listing supported domains and their IPv4 addresses.

## Request/Response Flow

1. **DNS Query Sent:**
  - A client (e.g., using `dig`) sends a DNS query for an A record to UDP port 8282.

2. **Server Receives Query:**
  - `main.go` reads the UDP packet, parses the DNS header and question section.

3. **Domain Lookup:**
  - The server checks if the query is for an A record (`TypeA`) and if the domain exists in `names.json`.
  - If both match, it prepares a DNS response with the mapped IPv4 address.
  - If not, it returns an empty answer section.

4. **Response Sent:**
  - The server serializes the DNS response and sends it back to the client over UDP.

## Example Query and Flow

Suppose you run:

```sh
dig @127.0.0.1 -p 8282 google.com A
```

**What happens:**
- The server receives the query for `google.com` A record on port 8282.
- It finds `google.com` in `names.json` and responds with 3.1.3.7.
- The dig output will show this IP in the ANSWER section.

If you query for an unsupported record type or unknown domain, the ANSWER section will be empty.

## Summary

- Only A records for domains in `names.json` are answered.
- The server is intentionally simple and easy to read.
- Great for learning about DNS and Go networking.

---
MIT License. Free to use, modify, and share.

## Project Structure
- `main.go` — Entry point, starts the DNS server
- `dblookup.go` — Handles DNS record lookups
- `models/` — Data models (e.g., DNS record structs)
- `names.json` — Example DNS records or configuration
- `tmp/` — Temporary files or runtime data

## Why Use This Project?
- **Simplicity:** The codebase is small and easy to understand.
- **Learning:** Great for learning about DNS and Go networking.
- **Customization:** Easily add your own DNS logic or records.

## License
MIT License. Free to use, modify, and share.
