# Security Policy

## Supported Versions

`shine` is pre-1.0. Security fixes target the latest release and the `main` branch.

| Version | Supported |
| --- | --- |
| 0.1.x | Yes |

## Reporting a Vulnerability

Please do not open a public issue for a vulnerability.

Use GitHub private vulnerability reporting if available for this repository, or contact the maintainer through the repository owner profile.

Include:

- affected version or commit
- operating system
- reproduction steps
- expected impact
- any suggested fix or mitigation

## Security Scope

Relevant issues include:

- unsafe archive or install behavior
- checksum verification problems
- command execution risks
- file traversal or unexpected local file access
- terminal escape handling that could mislead users

General Markdown rendering bugs should usually be filed as normal issues unless they have security impact.
