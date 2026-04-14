# mdnsmap

`mdnsmap` is a Golang CLI for **mDNS / DNS-SD asset discovery** on the local link.
It discovers services by querying `_services._dns-sd._udp.local.`, browsing each discovered service type, and resolving instance details (host, port, TXT, IPs, TTL), then applies CIDR and port filters.

> This is not a generic port scanner.

## Why results vary
mDNS is link-local multicast. Visibility depends on your current Layer-2 segment, multicast forwarding, VLAN/gateway policy, and host firewall settings.

## Build

```bash
go build ./cmd/mdnsmap
```

## Usage

```bash
./mdnsmap \
  --cidr 192.168.1.0/24 \
  --cidr fe80::/64 \
  --ports 80,445,5000-5005 \
  --wait 4s \
  --timeout 1500ms \
  --workers 4
```

Flags:
- `--cidr` repeatable and/or comma-separated
- `--ports` single and range syntax
- `--iface` optional interface name
- `--wait` mDNS collection window
- `--timeout` active HTTP enrichment timeout
- `--workers` enrichment concurrency
- `--json` emit JSON
- `--output` write to file
- `--verbose` verbose mode

## Sample text output

```text
services:
5000/tcp qdiscover:
Name=slw-nas
IPv4=192.168.1.2
Hostname=slw-nas.local
TTL=10
accessType=https,accessPort=86,model=TS-X64,displayModel=TS-464C,fwVer=5.2.9,fwBuildNum=20260214

answers:
PTR:
_qdiscover._tcp.local
```

## JSON shape

```json
{
  "services": [
    {
      "instance_name": "slw-nas",
      "service_type": "_qdiscover._tcp.local",
      "service_short": "qdiscover",
      "hostname": "slw-nas.local",
      "port": 5000,
      "ttl": 10,
      "ipv4": ["192.168.1.2"],
      "ipv6": ["fe80::1"],
      "txt": {"model": "TS-464C"},
      "txt_order": ["model"],
      "raw_txt": ["model=TS-464C"],
      "banner": {"path": "/", "status": "200 OK"}
    }
  ],
  "ptr_answers": ["_qdiscover._tcp.local"],
  "wait": 4000000000
}
```
