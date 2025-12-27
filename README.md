

# 📡 PacketChat — LAN UDP Terminal Chat

PacketChat is a **terminal-based LAN chat application** written in Go.
It allows devices on the same local network to chat **without a central server**, automatically discovering other users on the network.

PacketChat currently uses **UDP broadcast** for sending messages and **pcap-based packet capture** for receiving them.
This design allows correct behavior even when running **multiple instances on the same machine**.

The UI is built with **tcell**, providing a clean, interactive terminal interface with color-coded messages.

This project is intended as a **networking and systems experiment**, not a secure production chat application.

---

## ✨ Key Features

* 🌐 **Serverless LAN chat** — No central server required
* 🔍 **Automatic user discovery** — No configuration needed
* 🕵️ **Pseudonymous users** — Choose any nickname, no login
* 💬 **Real-time messaging**
* 🎨 **Color-coded messages** — Self, remote users, and system messages
* 🖥️ **Multiple instances supported** — Even on the same PC
* 🧠 **Self-message filtering** — No duplicate messages
* 🔌 **Network interface selection** — Explicitly choose your LAN interface
* ❗ **More features planned (see below)**

---

## 🚀 How It Works (Current Version)

* Messages are sent using **UDP broadcast**
* Incoming packets are captured using **pcap**
* Self-sent packets are ignored by checking the **source UDP port**
* No server, no authentication, no persistent state

pcap is currently used to ensure reliable packet handling and correct behavior in edge cases.

---

## 📦 Requirements

* Go **1.20+**
* libpcap / WinPcap / Npcap
* Network interface that supports broadcast

### Linux

```bash
sudo apt install libpcap-dev
```

### Windows

* Install **Npcap**
* Run the terminal as **Administrator**

---

## 🔧 Installation

```bash
git clone https://github.com/yourusername/packetchat.git
cd packetchat
go mod tidy
```

---

## ▶️ Usage

```bash
go run . <username>
```

Example:

```bash
go run . Slat
```

You will be prompted to select a network interface.

---

## 💻 Commands

| Command | Description             |
| ------- | ----------------------- |
| `/help` | Show available commands |
| `/exit` | Exit the chat           |
| `/h`    | Alias for help          |
| `/e`    | Alias for exit          |

---

## ⚠️ Notes & Limitations

* Intended for **local networks only**
* **No encryption** (packets are visible on the LAN. Will be encrypted soon!)
* **No authentication** (any device can send packets)
* Requires elevated permissions due to packet capture (will be changes soon)

---

## 🛠️ Planned Features / TODO
---

### 🔹 Security & Networking

* [ ] 🔐 **Message encryption**
* [ ] Replace **pcap** with standard UDP receive mode
* [ ] Packet validation
* [ ] Better input sanitization

---

### 🔹 Core Features

* [ ] **JOIN / LEAVE** packet types for user presence
* [ ] `/online` command to list online users
* [ ] `/whoami` command
* [ ] Message **timestamps**
* [ ] **Unique username colors**
* [ ] Rich text formatting:
  * **Bold**
  * *Italic*
  * ~~Strikethrough~~
  * Custom colors

---

### 🔹 Configuration & UX

* [ ] Config file support:
  * Username color
  * Auto-select network interface
  * Default commands
  * Message formatting preferences
* [ ] Improved UI design
* [ ] Optional in-app **TUI settings editor**

---
### 🔹 Cross-Platform & Testing

* [ ] Test on **Linux, Windows, macOS**
* [ ] Ensure multiple instances behave correctly
* [ ] Handle packet loss and network edge cases

---

### 🔹 Advanced Ideas

* [ ] `/msg <user>` — private LAN messages
* [ ] File transfer over LAN
* [ ] Emoji support
* [ ] Message queueing (deliver messages when user joins)

---

## 🔄 Future Direction

pcap is currently used for packet capture, but **will be replaced in future updates** with a simpler UDP receive mechanism.
The goal is to:

* Remove admin/root requirements
* Simplify setup
* Improve portability

Encryption and additional protocol features are **planned and coming soon**.

---

## 🧾 License

MIT License
Free to use, modify, and learn from.

--- 
## 🤝 Contributing PacketChat is open source! Feel free to **push your changes** and **contribute ideas** to the project. Every improvement is welcome! 🚀 
---


## 🙌 Built With

* Go
* gopacket
* pcap
* tcell
