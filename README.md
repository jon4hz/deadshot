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

## Generate abi packages
By using abigen, you can create native golang packages with bindings for a specific abi.  
```bash
go generate ./...
```

## Create a release
To create a new release, run `make release`.  
This will create a new tag based on the last commit so make sure to commit all changes. Make will also push the new tag, which will trigger the build pipeline.  
The pipeline injects the current version and build code which is required by the unlocking system. So make sure to follow the tag annotation vX.X.X 

## Build dependencies
- [go](https://go.dev/)
- [make](https://www.gnu.org/software/make/)
- [abigen](https://github.com/ethereum/go-ethereum)

## Disclaimer
This bot can be used to trade cryptocurrencies. I hereby disclaim any liability and will not be held responsible for any potential losses.
I wrote this bot while I learned go, so there might be bugs and mistakes in the code.

## Acknowledgement
- [charmbracelet](https://github.com/charmbracelet) for their charming cli tools
- [goreleaser](https://github.com/goreleaser) from whom I stole some parts of the pipeline 
- many other great developers who made this bot possible