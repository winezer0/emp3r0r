# emp3r0r

An advanced post-exploitation framework designed for Linux/Windows environments

<https://github.com/user-attachments/assets/5a918e49-2073-44c6-b93a-0f9a5dd36e0c>

[View Screenshots and Videos](./Screenshots.md)

---

## Getting Started

### Installation

```shell
curl -sSL https://raw.githubusercontent.com/jm33-m0/emp3r0r/refs/heads/v3/install.sh | bash
```

### Server Configuration

```shell
emp3r0r server --c2-hosts '192.168.200.3' --port 12345 --operators 3
```

This command initiates emp3r0r with:

- HTTP2/TLS agent listener on a random port, with valid hostname `192.168.200.3` in TLS server certificate
- WireGuard operator service on port `12345`
- Operator mTLS server on `wg_ip:12346` (operators share the same certificate, but have different WireGuard profiles)
- 3 pre-registered operator slots

The server will display all necessary configuration details for operator setup.

### Operator Connection

```shell
emp3r0r client --c2-host '192.168.200.3' --c2-port 12345 --server-wg-ip 'C2 Server IP (WG)' --server-wg-key 'C2 Public Key' --operator-wg-ip 'IP ADDRESS'
```

This command:

- Establishes a connection between the emp3r0r operator and the remote server at `192.168.200.3:12345`
- Configures WireGuard connectivity between operator and server using the parameters provided in server output
- Supports multiple operators (each requiring unique WireGuard profiles)
- Note: If connection stalls after entering the operator's private key, verify that keys/IPs match correctly

### Agent Payload Generation

Use the `generate` command from within the emp3r0r shell interface.

## Important Notes

- Breaking changes are typically documented in release logs. Cross-version compatibility is not guaranteed due to ongoing feature development and bug fixes.
- If you encounter issues, try removing `~/.emp3r0r` directory and starting fresh.
- The wiki may not reflect all features in `v3`. Refer to command-line help for the most current information. Community contributions to the wiki are welcome.

## Project Background

emp3r0r was initially developed as a research project for implementing Linux adversary techniques alongside original ideas. It has evolved into a comprehensive framework addressing the need for advanced post-exploitation capabilities specifically targeting Linux environments.

**What distinguishes emp3r0r** is its position as one of the first C2 frameworks purpose-built for Linux targets while providing seamless integration with external tools. The comprehensive [feature list](#features) demonstrates its versatility.

For extensibility, emp3r0r offers complete [python3 support](https://github.com/jm33-m0/emp3r0r/wiki/Write-modules-for-emp3r0r#python) via the [`vaccine`](./core/modules/vaccine) module (15MB total), including essential packages like `Impacket`, `Requests`, and `MySQL`. The framework supports diverse module formats including `bash`, `powershell`, `python`, `dll`, `so`, and `exe`.

While emp3r0r is primarily designed as a single-operator framework, the SSH C2-Relay feature enables separation between C2 server and operator functions.

---

## Features

- **Advanced Command-Line Interface**
  - Built on [console](https://github.com/reeflective/console) and [cobra](https://github.com/spf13/cobra) frameworks
  - Comprehensive auto-completion with syntax highlighting
  - Multi-tasking capabilities through [tmux](https://github.com/tmux/tmux) integration
  - Secure operator-server architecture using WireGuard and mTLS

- **Operational Security**
  - Dynamic `argv` manipulation for process listing obfuscation
  - File and PID concealment through Glibc hijacking (via `patcher` in `get_persistence`)
  - [**Bring Your Own Shell**](https://github.com/jm33-m0/emp3r0r/blob/master/core/modules/elvish/config.json) functionality supporting [`elvish`](https://elv.sh) and other interactive programs through custom modules

- **Secure Communications**
  - HTTP2/TLS-based command and control
  - [**UTLS**](https://github.com/refraction-networking/utls) implementation to defeat [**JA3**](https://github.com/salesforce/ja3) fingerprinting
  - KCP-based fast multiplexed UDP tunneling
  - Support for external proxying via [**TOR** and **CDN**s](https://github.com/jm33-m0/emp3r0r/raw/master/img/c2transports.png)
  - [**C2 relaying**](https://github.com/jm33-m0/emp3r0r/wiki/C2-Relay) through SSH

- **Memory Forensics Capabilities**
  - Cross-platform memory dumping
  - Windows mini-dump extraction compatible with [pypykatz](https://github.com/skelsec/pypykatz)

- **Flexible Payload Delivery**
  - Multi-stage delivery for both Linux and Windows targets
  - [HTTP Listener with AES encryption and compression](https://github.com/jm33-m0/emp3r0r/wiki/Listener)
  - Platform-specific payloads: [**DLL agent**](https://github.com/jm33-m0/emp3r0r/wiki/DLL-Agent), [**Shellcode agent**](https://github.com/jm33-m0/emp3r0r/wiki/Shellcode-Agent-for-Windows) (Windows), and [**Shared Library stager**](https://github.com/jm33-m0/emp3r0r/wiki/Shared-Library-Stager-for-Linux) (Linux)

- **Network Traversal**
  - Automatic agent bridging via **Shadowsocks proxy chain** for internal network access
  - Reverse proxy capabilities through SSH and KCP tunneling
  - [**External target access**](https://github.com/jm33-m0/emp3r0r/wiki/Getting-started#bring-agents-to-c2) for endpoints unreachable by direct connection

- **Operational Efficiency**
  - Parallel command execution for uninterrupted workflow
  - Modular architecture supporting [custom extensions](https://github.com/jm33-m0/emp3r0r/wiki/Write-modules-for-emp3r0r)
  - In-memory execution for `bash`, `powershell`, `python`, and ELF binaries via [**`CGO` ELF loader**](https://jm33.me/offensive-cgo-an-elf-loader.html)

- **Enhanced Shell Experience**
  - SSH integration with PTY support
  - Windows compatibility with standard SSH clients

- **Additional Capabilities**
  - [Bettercap](https://github.com/bettercap/bettercap) integration
  - Multiple persistence mechanisms
  - Comprehensive post-exploitation toolset
  - [**OpenSSH credential harvesting**](https://jm33.me/sshd-injection-and-password-harvesting.html)
  - Advanced [Process](https://jm33.me/emp3r0r-injection.html) and [Shellcode](https://jm33.me/process-injection-on-linux.html) injection
  - ELF binary patching for persistent access
  - Bidirectional port mapping (TCP/UDP)
  - Agent-side Socks5 proxy with UDP support
  - Privilege escalation tools and suggestions
  - System information collection
  - File management with integrity verification and compression
  - **SFTP** integration for convenient remote file access
  - Log sanitization utilities
  - Screenshot functionality
  - Anti-analysis capabilities
  - Network connectivity verification
