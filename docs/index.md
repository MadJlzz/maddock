---
layout: home

hero:
  name: "Maddock"
  text: "Infrastructure as code for Linux"
  tagline: Converge hosts to a desired state with YAML manifests over gRPC.
  actions:
    - theme: brand
      text: Get Started
      link: /installation
    - theme: alt
      text: View on GitHub
      link: https://github.com/MadJlzz/maddock

features:
  - title: Declarative
    details: Describe packages, files, and services in YAML. Maddock handles idempotent convergence.
  - title: Push-based
    details: Central server orchestrates agent hosts over gRPC with streamed per-resource reports.
  - title: Single static binary
    details: Agent ships as a Linux binary with no runtime dependencies. One curl to install.
---
