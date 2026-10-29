# Third-party components

- **Go libraries** listed in `go.mod` / `go.sum`, each under its own open source license (e.g. MIT, BSD, Apache-2.0).
- **Npcap** (packet capture driver on Windows): separate program with its own license, installed by the user from npcap.com. Not bundled or redistributed.
- `thirdparty/rcedit` (MIT, see `thirdparty/rcedit/LICENSE`) and `thirdparty/WpdPack` (WinPcap developer pack headers/libraries): kept for reference, **not used by the current build** (`CGO_ENABLED=0`, icons and version info come from go-winres).
