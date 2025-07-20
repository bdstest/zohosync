# ZohoSync

> Secure, lightweight Zoho WorkDrive sync client for Linux - Your WorkDrive, Everywhere You Work

![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Platform](https://img.shields.io/badge/platform-linux-lightgrey.svg)

## Overview

ZohoSync is a native Linux desktop application that provides seamless synchronization between your local folders and Zoho WorkDrive. Built with Go and Fyne, it offers a modern, lightweight solution for keeping your files in sync across devices.

## Features (Planned)

- 🔐 Secure OAuth 2.0 authentication with PKCE
- 📁 Bidirectional file synchronization
- 🚀 Lightweight and fast
- 🖥️ Native Linux desktop integration
- 📊 Real-time sync status monitoring
- 🔄 Automatic conflict resolution
- 🌐 Proxy support
- 📦 Multiple installation formats (AppImage, Deb, RPM)

## Development Progress

### Completed Phases ✅
- [ ] Phase 1: Foundation & Project Setup
- [ ] Phase 2: OAuth Authentication
- [ ] Phase 3: API Client
- [ ] Phase 4: Local Storage
- [ ] Phase 5: Basic Sync Engine
- [ ] Phase 6: GUI Application
- [ ] Phase 7: Bidirectional Sync
- [ ] Phase 8: System Integration
- [ ] Phase 9: Advanced Features
- [ ] Phase 10: Packaging

### Current Milestone
Working on: Phase 1 - Foundation & Project Setup
Expected completion: January 2024

## Requirements

- Go 1.21 or higher
- Linux (Ubuntu, Debian, Fedora, Arch)
- GTK 4.0+ libraries (for GUI)

## Building from Source

### Prerequisites

```bash
# Ubuntu/Debian
sudo apt-get install golang gcc pkg-config libgtk-4-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install golang gcc pkg-config gtk4-devel webkit2gtk4.0-devel

# Arch
sudo pacman -S go gcc pkg-config gtk4 webkit2gtk
```

### Build Instructions

```bash
# Clone the repository
git clone https://github.com/bdstest/zohosync.git
cd zohosync

# Download dependencies
go mod download

# Build all components
make build

# Or build specific components
make build-gui      # GUI application
make build-cli      # CLI tool
make build-daemon   # Background daemon
```

## Installation

```bash
# Install system-wide
sudo make install

# Or run directly
./zohosync          # GUI application
./zohosync-cli      # CLI tool
./zohosync-daemon   # Background daemon
```

## Usage

### GUI Application
```bash
# Launch the GUI application
zohosync
```

### CLI Tool
```bash
# Login to Zoho WorkDrive
zohosync-cli login

# List remote files
zohosync-cli list

# Manual sync
zohosync-cli sync

# View sync status
zohosync-cli status
```

## Configuration

Configuration file location: `~/.config/zohosync/config.yaml`

Example configuration:
```yaml
app:
  name: ZohoSync
  version: 0.1.0

sync:
  interval: 300  # seconds
  conflict_resolution: newer  # newer, local, remote

folders:
  - local: ~/Documents/Zoho
    remote: /My Folders/Documents
    sync_mode: bidirectional
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Fyne](https://fyne.io/) - Cross-platform GUI framework for Go
- Uses [Zoho WorkDrive API](https://workdrive.zoho.com/apidocs)

---

**Author**: bdstest  
**Status**: Under active development
