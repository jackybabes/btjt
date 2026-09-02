# btjt

A small BitTorrent command-line downloader written in Go.

## Usage

```
go build -o btjt .
./btjt [-o dir] [-maxpeers n] [-quiet] <torrent-file>
```

- `-o dir` output directory (default `.`)
- `-maxpeers n` cap on peers to use (default 40)
- `-quiet` suppress the verbose per-peer log

Example:

```
./btjt -o ~/Downloads test_torrents/http.torrent
```

It announces to the HTTP trackers, connects to the returned peers, downloads
every piece (verifying each against its SHA-1 hash), and writes the result to
disk — single-file and multi-file torrents both work.

## What works

- Bencode decoding, infohash, piece hashes
- HTTP tracker announce, compact peer list
- UDP tracker announce (BEP 15)
- Peer handshake, bitfield / have tracking
- Parallel per-peer download loops with a small request pipeline
- Per-piece SHA-1 verification, retry on mismatch
- Writing pieces to single- or multi-file layouts

## Not implemented

- Non-compact (dictionary) tracker responses
- IPv6 peers (and IPv6 UDP tracker announce)
- Seeding / uploading, resume, DHT, magnet links
- Rarest-first piece selection (pieces are picked in order)

## Notes

Some tracker responses omit the compact peer list when IPv6 peers are present;
those are currently skipped.
