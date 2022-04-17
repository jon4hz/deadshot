# deadshot
[![testing](https://github.com/jon4hz/deadshot/actions/workflows/testing.yml/badge.svg)](https://github.com/jon4hz/deadshot/actions/workflows/testing.yml)
[![lint](https://github.com/jon4hz/deadshot/actions/workflows/lint.yml/badge.svg)](https://github.com/jon4hz/deadshot/actions/workflows/lint.yml)
[![goreleaser](https://github.com/jon4hz/deadshot/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/jon4hz/deadshot/actions/workflows/goreleaser.yml)

```
      _                _     _           _   
   __| | ___  __ _  __| |___| |__   ___ | |_ 
  / _` |/ _ \/ _` |/ _` / __| '_ \ / _ \| __|
 | (_| |  __/ (_| | (_| \__ \ | | | (_) | |_ 
  \__,_|\___|\__,_|\__,_|___/_| |_|\___/ \__|
                                              
  ( φ_<)︻┻┳══━一 - - - - - - - - - - - - 💥   

```
A terminal based trading bot

## Install
The **newest** release can always be found [here][release].  

[release]: https://github.com/jon4hz/deadshot/releases

### Linux

#### apt
Linux distributions like debian or ubuntu can use apt to install the binary
```bash
echo 'deb [trusted=yes] https://apt.fury.io/jon4hz/ /' | sudo tee /etc/apt/sources.list.d/jon4hz.list
sudo apt update
sudo apt install deadshot
```

#### yum
```bash
echo '[jon4hz]
name=jon4hz
baseurl=https://yum.fury.io/jon4hz/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/jon4hz.repo
sudo yum install deadshot
```

#### deb, rpm and apk packages

Download the `.deb`, `.rpm` or `.apk` package.  

**deb**
```
# dpkg -i <release_name>.deb
```

**rpm**
```
# rpm -i <release_name>.rpm
```

**apk**
```
# apk add --allow-untrusted <release_name>.apk
```

#### manually
1. Download the archive and unzip it
2. Open a terminal in the folder with right click -> "Open Terminal"

The bot can be started with `./deadshot`

### MacOS
#### manually
1. Download the archive and unzip it
2. Open a terminal in the folder with right click -> "Open Terminal"

The bot can be started with `./deadshot`

### Windows
#### manually

1. Download the archive and unzip it
2. Go to the folder with the program in it
3. Open a terminal in the folder with right click -> "Open in Windows Terminal"

The command `.\deadshot.exe` starts the bot.

If you are using Windows 10, start the bot in the [Windows Terminal][winterm] and not Powershell.

[winterm]: https://www.microsoft.com/en-US/p/windows-terminal/9n0dx20hk701


## Build

### Build dependencies
- [go](https://go.dev/)
- [make](https://www.gnu.org/software/make/)
- [abigen](https://github.com/ethereum/go-ethereum)

### Build
Clone the repository
```
$ git clone https://github.com/jon4hz/deadshot && cd deadshot
```

Build the software
```
$ go build -o deadshot cmd/deadshot/main.go
```

## Disclaimer
This bot can be used to trade cryptocurrencies. I hereby disclaim any liability and will not be held responsible for any potential losses.
I wrote this bot while I learned go, so there might be bugs and mistakes in the code.


## Acknowledgement
- [charmbracelet](https://github.com/charmbracelet) for their charming cli tools
- [goreleaser](https://github.com/goreleaser) from whom I stole some parts of the pipeline 
- many other great developers who made this bot possible