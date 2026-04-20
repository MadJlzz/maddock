# Installation

## Agent

Maddock distributes the agent as a static Linux binary. The install
script auto-detects the CPU architecture and fetches the latest release
from GitHub.

```sh
curl -fsSL https://raw.githubusercontent.com/MadJlzz/maddock/main/install.sh | sh
```

Install a specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/MadJlzz/maddock/main/install.sh | sh -s -- v0.1.0
```

Change the installation directory:

```sh
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/MadJlzz/maddock/main/install.sh | sh
```

Supported platforms: `linux/amd64`, `linux/arm64`.

## Server

The server is distributed as a container image on GitHub Container
Registry:

```sh
docker pull ghcr.io/madjlzz/maddock-server:latest
```

Tag schemes available: `latest`, `MAJOR.MINOR`, and exact `MAJOR.MINOR.PATCH`.

## From source

```sh
git clone https://github.com/MadJlzz/maddock
cd maddock
mise run build
```

Binaries are produced under the working directory.
