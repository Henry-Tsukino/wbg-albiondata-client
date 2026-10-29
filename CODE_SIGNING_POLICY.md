# Code signing policy

Free code signing provided by [SignPath.io](https://signpath.io), certificate by [SignPath Foundation](https://signpath.org).

## What is signed

Only the Windows binaries built from the source code of this repository by the GitHub Actions workflow `.github/workflows/build.yml` (triggered by a release tag):

- `WBG-albion-data-client.exe` (the client, also delivered to installed clients as the self-update asset `update-windows-amd64.exe.gz`);
- `installer-WBG-tool.exe` (the installer).

Third-party programs are never signed with this certificate. Npcap, which the client needs for packet capture, is a separate program with its own license and signature; it is not bundled and not installed silently (the installer only offers to open npcap.com).

The binaries carry the product name `WBG-albiondata-client` and the same product version in every file of a release.

## Privacy

The privacy policy of this program is in [PRIVACY.md](PRIVACY.md). It is shown during installation, and the installer has an option to disable all data sending. The program does not transfer any information to other networked systems other than those listed in that policy.

Third-party services involved: the [Albion Data Project](https://www.albion-online-data.com/), [GitHub](https://docs.github.com/site-policy/privacy-policies/github-general-privacy-statement) (update checks) and [Railway](https://railway.com/legal/privacy) (hosting of the WBG server).

## Team roles

| Role | Who | Responsibility |
|---|---|---|
| Authors (trusted committers) | [Henry-Tsukino](https://github.com/Henry-Tsukino) | Write and merge changes in this repository |
| Reviewers | [Henry-Tsukino](https://github.com/Henry-Tsukino) | Review every change that comes from non-committers (pull requests) before it is merged |
| Approvers | [Henry-Tsukino](https://github.com/Henry-Tsukino) | Manually approve every signing request in SignPath |

All team members use multi-factor authentication for GitHub and SignPath. Every signing request is approved manually; builds come only from tagged commits of this repository.
