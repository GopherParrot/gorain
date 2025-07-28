# <p align="center">gorain 🌧️</p>
> **Get a rainy vibe in your terminal**

### Inspired my rmaake1's rain project

You should check out his awesome creation! to
> https://github.com/rmaake1/terminal-rain-lightning

- **Customizable colors** for raindrops and lightning 🌩️ 
- **Thunderstorm Mode** - toggle for intense rain and epic lightning bursts!
- **Responsive design** - works on various terminal sizes and devices (e.g. Termux, etc)
- **Lightweight & fast** - pure Go, no external dependencies
- **Vibes included** - cuz why not :)

---

## 🚀 Installation
> Requires **Go 1.20+** - make sure you have Go installed:
```bash
go version
```
If not, run:

### 1. Linux 🐧
- Debian / Ubuntu
 ```bash
 sudo apt update
 sudo apt install golang -y
 ```
- Arch / Manjaro
 ```bash
 sudo pacman -S go
 ```
- Termux (Android)
 ```bash
 pkg update
 pkg install golang
 ```

---

### 2. macOS 🍎 
- With Homebrew
 ```bash
 brew install go
 ```
📦 **Manual Installer**
1. Visit > https://go.dev/DL
2. Download the `.pkg` file
3. Run the installer and follow instructions
✅ Then open Terminal:
 ```bash
 go version
 ```

---

### 3. Windows 🪟
📦 **Using Installer**
1. Go to > https://go.dev/DL
2. Download the `.msi` file
3. Run it and follow the instructions
4. **Restart** Command Prompt / PowerShell
✅ Then test:
 ```bash
 go version
 ```

---

## ⚒️ Common Troubleshooting
**Go not found?**
Make sure Go's `bin` folder is in your system `PATH`
### Example (Linux/Termux):
 ```bash
 export PATH=$PATH:$HOME/go/bin

---

### Windows:
- Open "System Environment Variables"
- Edit `PATH` and add: `C:\Go\bin`

---

## ✅ Verify Installation

After setup, run:
 ```bash
 go version
 ```
You should see something like:
 ```bash
 go version go1.21.0 linux/amd64
 ```

---

## 💡 Tip
Use `go env` to see all Go paths and config:
 ```bash
 go env
 ```
---

## Now to the fun part :)

### ⚒️ Using `go install` (recommended)
if you have Go installed, you can install **gorain** directly from the terminal:
 ```bash
 go install github.com/GopherParrot/gorain@latest
 ```
Then run it with:
 ```bash
 gorain
 ```

---

To be able to run your Go program from **anywhere in the terminal**,you need to make sure this folder is in your system's `PATH`:
 ```bash
 $GOPATH/bin
 ```
Or if you never change your `GOPATH`, it's usually:
 ```bash
 $HOME/go/bin
 ```

## ⚒️ How to Add to PATH
### 1. Linux /macOS / Termux:
Add this line to your shell config (like .bashrc, .zshrc, or .profile):
 ```bash
 export PATH="$HOME/go/bin:$PATH"
 ```
Then run:
 ```bash
 source ~/.bashrc # or .zshrc, depending on your shell
 ```

---

### 2. Windows:
- Search for "**Environment Variables**" in Start menu
- Click "**Edit environment variables for your account**"
- Under **User variables**, find or create a `PATH` variable
- Add this path:
 ```bash
 C:\Users\<YourUsername>\go\bin
 ```
> ## IMPORTANT: replace <YourUsername> with your real username
- Click OK, restart terminal ✅

---
### Once you do that, you'll be able to just type:
 ```bash
 gorain
 ```
from anywhere 😁