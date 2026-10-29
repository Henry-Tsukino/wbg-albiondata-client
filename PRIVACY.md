# Privacy Policy

_Last updated: 2026-10-29_

WBG Albion Data Client ("the client") is a free, open-source program that runs on your computer while you play Albion Online. It passively reads the game's network traffic (via libpcap/Npcap) and uploads some of the game data it sees. It does not modify the game, inject code, or read game memory. This page lists exactly what leaves your computer, where it goes, and how to turn it off. It is written from the source code of this repository.

## What is sent, and where

| Destination | Data | Tied to your character? |
|---|---|---|
| **Albion Online Data Project** (`albion-online-data.com`, regional servers such as `pow.west` / `pow.east`) | Market orders (item, location, quality, enchantment, price, amount, order type, expiry), market price histories, gold prices. Each upload also carries a random one-time ID, the game server ID and the client version. | No |
| **WBG server** (hosted on Railway, `hermes-production-817b.up.railway.app`, run by the project maintainer) | Zone map data (zone ID, building types, food/silver/reward values, fees, durability, permissions, coordinates and the owner names shown in-game), red-zone ("bandit") event timing, the festivity/event schedule. | No |
| **WBG server** (same) | **Skill data** (skill ID, level, progress, fame) and **market sale/expiry summaries** read from your in-game mail (item, location, amount, price, totals). | **Yes: your character name and character ID are attached** |
| **GitHub** (`github.com`) | Once an hour a released build asks GitHub whether a newer release exists, and downloads the update if so. GitHub sees your IP address and standard request headers (see GitHub's privacy statement). Development builds never do this. | No |

Like any internet connection, every server above can see your IP address. The client uploads no other kinds of data than those listed here.

**What the WBG server is used for and how long data is kept:** **Personal data is not stored; it remains only in temporary logs, which the hosting provider keeps for about 14 days. Sensitive personal data (with the exception of a profile name) is neither collected nor stored at all. Other data from games is kept within messages sent by the Discord bot and in its database; this information is anonymous and non-sensitive - for example, details regarding daily in-game events that are the same for everyone.**

## What stays on your computer

- `wbg-client.log` (plus up to 10 rotated backups) in the program folder. It can contain your character name. It is never uploaded.
- `config.yaml` (your settings) and a small file remembering the dashboard window position.
- A recording file only if you start the client with `-record`.

## Your controls

- **Stop all uploads:** tick the opt-out checkbox on the privacy page of the installer (it writes `DisableUpload: true` to `config.yaml`; for silent installs pass `/DISABLEUPLOAD=1`), untick "Send data to servers" in the dashboard (saved to `config.yaml`), set `DisableUpload: true` in `config.yaml`, or start the client with `-d`. Nothing is then sent to the Albion Online Data Project or the WBG server.
- **Stop market data:** turn off market in the dashboard, start with `-no-market`, or set `DisableMarket: true` in `config.yaml`. This stops market parsing, market uploads and the market sale/expiry summaries. It does **not** stop skill, map, event or festivity uploads to the WBG server; untick "Send data to servers", or use `DisableUpload: true` or `-d` for that.
- **Send data elsewhere:** `-i` sets the public (market/gold) destination, `-p` the WBG server, `-bandit-webhook` the bandit-event URL.
- **Local WebSocket feed:** off by default. Only if you set `EnableWebsockets: true` in `config.yaml` does the client open local port 8099 and stream the same data to the browser origins you list in `AllowedWebsocketHosts`.
- **Uninstall:** use the Windows uninstaller (Settings, Apps), or on Linux delete `/usr/local/bin/WBG-albion-data-client` and `~/.local/share/applications/wbg-albion.desktop`.

## Third parties

The Albion Online Data Project, GitHub and Railway process data on their own terms; please read their privacy policies: [Albion Data Project](https://www.albion-online-data.com/), [GitHub](https://docs.github.com/site-policy/privacy-policies/github-general-privacy-statement), [Railway](https://railway.com/legal/privacy). Packet capture needs Npcap, a separate program with its own license and privacy terms ([npcap.com](https://npcap.com)). It is not bundled; the installer only offers to open its download page.

## Contact and changes

Questions or deletion requests for data on the WBG server: **contact me via GitHub issues link or Discord**. Changes to this policy are published in this file; see the git history.
