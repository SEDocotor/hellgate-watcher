# Hellgate Watcher

A Discord bot that monitors and reports recent 2v2 and 5v5 Hellgate battles from Albion Online servers.

# Upgrades

- **Replaced JSON storage with a real database**, allowing for a huge increase in throughput.
- **Switched backend to GO to utilize goroutines allowing for a significant performance increase

## Features

- **Automatic Battle Reporting:** Periodically fetches new Hellgate battles on the specified Albion Online servers (Europe, Americas, and Asia).
- **Image Generation:** Generates images of battle reports showing teams, their gear, and the outcome.
- **Server and Mode specific:** Allows setting up different channels for different servers and Hellgate modes (2v2 or 5v5).
- **Configurable:** Most settings, such as the battle check interval and image generation settings, can be configured.

## Commands

The bot uses a slash command to set the channel for battle reports:

- `/setchannel <server> <mode> <channel>`: Sets the channel for Hellgate battle reports.
  - **server:** The Albion Online server to get reports from (`Europe`, `Americas`, or `Asia`).
  - **mode:** The Hellgate mode (`2v2` or `5v5`).
  - **channel:** The Discord channel where the reports will be sent.

This command requires administrator permissions.

## Setup and Installation

### Prerequisites

- Node.js 20+
- Go 1.21+

### 1. Clone the repository

```bash
git clone https://github.com/SEDocotor/hellgate-watcher.git
cd hellgate-watcher
