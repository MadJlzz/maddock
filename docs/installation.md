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

## Control plane

The control plane is distributed as a container image on GitHub
Container Registry:

```sh
docker pull ghcr.io/madjlzz/maddock-controlplane:latest
```

Tag schemes available: `latest`, `MAJOR.MINOR`, and exact `MAJOR.MINOR.PATCH`.

## From source

```sh
git clone https://github.com/MadJlzz/maddock
cd maddock
mise run build
```

Binaries are produced under the working directory.

## Bootstrap (TLS)

The control plane pushes to agents over mutual TLS, so before the first
push you must set up a CA and certificates. This is a one-time step on
the control plane plus a per-host step.

1. **Initialize the control plane.** Creates the CA and the control
   plane's own certificate in the state directory.

   ```sh
   maddock-controlplane init --state-dir /var/lib/maddock-controlplane
   ```

2. **Issue an agent certificate per host.** The hostname must match the
   `hostname` field used for that target in the push config.

   ```sh
   maddock-controlplane cert issue --hostname web-1 --output ./web-1/
   ```

3. **Copy the certificates to the host** and start the agent with them:

   ```sh
   scp ./web-1/{ca.crt,web-1.crt,web-1.key} web-1:/etc/maddock/
   # on web-1:
   maddock-agent serve --ca-cert /etc/maddock/ca.crt \
     --cert /etc/maddock/web-1.crt --key /etc/maddock/web-1.key
   ```

Then `maddock-controlplane push` connects over mTLS automatically. See
the [control plane CLI reference](cli/controlplane.md) for details.

> Manual certificate distribution is a stopgap; an automated agent join
> flow is planned. See [the roadmap](https://github.com/MadJlzz/maddock/blob/main/PLAN.md).
