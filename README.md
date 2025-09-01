# <p align="center">gorain 🌧️</p>
> **Get a rainy vibe in your terminal**

### Inspired by rmaake1's rain project

You should check out his awesome creation!
> https://github.com/rmaake1/terminal-rain-lightning

## Code Preview

### Calm Rain
<img src="https://i.ibb.co/jZ5CFrRv/ezgif-2f2e52fc37abfa.gif" alt="calmrain" width="350" height="500">

### Thunderstorm
<img src="https://i.ibb.co/Y4b47Dg8/ezgif-260c1d5c115c6f.gif" alt="thunderstorm" width="350" height="500">

> **Yes, all done with Termux :)**

---
## ⚙️ Features

- **Customizable colors** for raindrops and lightning 🌩️ 
- **Thunderstorm Mode** - toggle for intense rain and epic lightning bursts!
- **Snowfall Mode** - toggle for a calm snowfall scene ❄️
- **Night Mode** - toggle for a starry night sky with a moon! Stars will twinkle faintly.
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
 ```
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

## ⚒️ How to add Go to PATH
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
- Click OK, restart terminal ✅

---
### Once you do that, you'll be able to just type:
 ```bash
 gorain
 ```
from anywhere 😁

## How to Use
After installing, just run:
```bash
gorain
```

### Controls
<table>
  <tr>
    <th>Key</th>
    <th>Action</th>
  </tr>
  <tr>
    <th>t/T</th>
    <th>Toggle Thunderstorm mode ⛈️</th>
  </tr>
  <tr>
    <th>s/S</th>
    <th>Toggle Snowfall mode 🌨️</th>
  </tr>
  <tr>
    <th>n/N</th>
    <th>Toggle Night mode (moon & stars) 🌙✨</th>
  </tr>
  <tr>
    <th>h/H</th>
    <th>Toggle hiding all weather effects (useful for viewing just the night sky)</th>
  </tr>
  <tr>
    <th>q/Q/Esc/Ctrl+C</th>
    <th>Quit the program</th>
  </tr>
</table>

---

## Command-line options
```bash
gorain [OPTIONS]
```

<table>
  <tr>
    <th>Option</th>
    <th>Description</th>
    <th>Default</th>
  </tr>
  <tr>
    <th>--rain-color [COLOR]</th>
    <th>Set the color of the raindrops</th>
    <th>Cyan</th>
  </tr>
  <tr>
    <th>--lightning-color [COLOR]</th>
    <th>Set the color of the lightning</th>
    <th>Yellow</th>
  </tr>
  <tr>
    <th>--snow-color [COLOR]</th>
    <th>Set the color of the snowflakes</th>
    <th>White</th>
  </tr>
  <tr>
    <th>--moon-color [COLOR]</th>
    <th>Set the color of the moon</th>
    <th>Yellow</th>
  </tr>
  <tr>
    <th>--moon-char [CHAR]</th>
    <th>Set the character used for the moon</th>
    <th>#</th>
  </tr>
  <tr>
    <th>--help</th>
    <th>Show help message and exit</th>
    <th>--</th>
  </tr>
</table>

## Available colors:
`black`,`red`,`green`,`yellow`,`blue`,`magenta`,`cyan`,`white`

## Example:
```bash
gorain --rain-color green --lightning-color white
```

Thanks for spending your time reading this :)
