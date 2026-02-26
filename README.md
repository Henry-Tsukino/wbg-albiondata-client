# Albion Data Client

A distributed network monitoring client for the [Albion Online Data](https://www.albion-online-data.com/) project.

**This is a fork of [ao-data/albiondata-client](https://github.com/ao-data/albiondata-client). For the official project and upstream repository, please visit the original.**

## Overview

The Albion Data Client monitors local network traffic, identifies UDP packets containing Albion Online game data, and transmits the information to a central NATS message server for distribution to subscribers. This enables community-driven data collection for market analytics, event tracking, and game state monitoring.

## Features

- Real-time network traffic monitoring via libpcap
- UDP packet filtering and analysis for Albion Online protocol
- Integration with NATS distributed messaging
- Cross-platform support (Windows, macOS, Linux)
- System tray integration for Windows and macOS
- Minimal resource footprint
- Docker support

## Legal Notice

The following statement from SBI Games regarding network packet monitoring:

> Our position is quite simple. As long as you just look and analyze we are ok with it. The moment you modify or manipulate something or somehow interfere with our services we will react (e.g. perma-ban, take legal action, whatever).

— MadDave, Technical Lead for Albion Online

Reference: https://forum.albiononline.com/index.php/Thread/51604-Is-it-allowed-to-scan-your-internet-trafic-and-pick-up-logs/?postID=512670#post512670

Users must comply with Albion Online Terms of Service when using this application.

## Installation

Official releases are available at: https://github.com/ao-data/albiondata-client/releases

### Windows

1. Download the latest Windows release from the official releases page
2. Run the installer or executable
3. Configure the application using `config.yaml`
4. Start the client (optionally set to run at startup via system tray)

### macOS

#### Using Finder
1. Download `albiondata-client-amd64-mac.zip` from [Releases](https://github.com/ao-data/albiondata-client/releases)
2. Unzip the file
3. Double-click `run.command` (will request password for network permissions)

#### Using Terminal
1. Download `update-darwin-amd64.gz` from [Releases](https://github.com/ao-data/albiondata-client/releases)
2. Unzip: `gunzip update-darwin-amd64.gz`
3. Make executable: `chmod +x albiondata-client`
4. Run: `./albiondata-client`

### Linux (Debian/Ubuntu)

1. Create directory: `mkdir -p ~/.local/bin`
2. Download binary: `curl -L https://github.com/ao-data/albiondata-client/releases/latest/download/update-linux-amd64.gz -o - | gzip -d > ~/.local/bin/albiondata-client`
3. Make executable: `chmod u+x ~/.local/bin/albiondata-client`
4. Install libpcap: `sudo apt install libpcap-dev`
5. Grant capabilities: `sudo setcap cap_net_raw,cap_net_admin=eip ~/.local/bin/albiondata-client`

## Building from Source

### Requirements

- Go 1.16 or later
- Platform-specific dependencies (libpcap on Unix-like systems)

### Build Instructions

Build scripts are available for each platform:

```bash
# Windows
scripts/build-windows.sh

# macOS
scripts/build-darwin.sh

# Linux
scripts/build-linux.sh
```

Go modules are automatically downloaded during the build process.

## Configuration

Create `config.yaml` based on the provided `config.yaml.example`. The configuration file controls:

- Server connection settings
- Network interface selection
- NATS message broker configuration
- Application behavior

## Docker

The project includes Docker support with `Dockerfile` for containerized deployment.

## Development

### Code Style

Check code formatting: `scripts/validate-fmt.sh`

Auto-format code: `scripts/fmt.sh`

Follow standard Go conventions and best practices.

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on contributing to this project.

### Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Security

For security-related inquiries, see [SECURITY.md](SECURITY.md).

## Related Projects

- [albiondata-deduper-dotNet](https://github.com/ao-data/albiondata-deduper-dotNet)
- [albiondata-sql-dotNet](https://github.com/ao-data/albiondata-sql-dotNet)
- [albiondata-api-dotNet](https://github.com/ao-data/albiondata-api-dotNet)
- [AlbionData.Models](https://github.com/ao-data/albiondata-models-dotNet)
- [albion-data-website](https://github.com/ao-data/albion-data-website)

## Community

The Albion Online community discusses this project on the Albion Online Fansites Discord server:
- Channels: #proj-albiondata or #developers
- Invite: https://discord.gg/TjWdq24

For official project discussions, visit https://www.albion-online-data.com/

## Project Information

- **Official Project**: https://www.albion-online-data.com/
- **Original Repository**: https://github.com/ao-data/albiondata-client
- **Release Statistics**: https://tooomm.github.io/github-release-stats/?username=ao-data&repository=albiondata-client

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for the full text.

## Acknowledgments

### Original Project Maintainers and Contributors

- [Regner](https://github.com/Regner) — Original developer
- [pcdummy](https://github.com/pcdummy) — Original developer
- [Ultraporing](https://github.com/Ultraporing) — Original developer
- [broderickhyman](https://github.com/broderickhyman) — Maintainer and project funding
- [Stanx](https://github.com/phendryx) — Primary maintainer and ongoing project funding
- [Walkynn](https://github.com/walkeralencar) — Long-time contributor

All contributors and community members who have supported this project are appreciated.
