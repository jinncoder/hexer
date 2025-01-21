# Hexer

![Code Health](https://github.com/archimoebius/hexer/actions/workflows/golang.yml/badge.svg) | ![Release Status](https://github.com/archimoebius/hexer/actions/workflows/goreleaser.yml/badge.svg)


<p>
    <img src="https://raw.githubusercontent.com/ArchiMoebius/hexer/main/mkdocs/docs/images/logo.png" width="250px" height="250px" alt="logo.png"></br>
    <em style="font-size:0.7em"><a href="https://github.com/invoke-ai/InvokeAI" alt="https://github.com/invoke-ai/InvokeAI" target="_blank">Invoke-AI Generated Logo</a></em>
</p>

A light-weight and easy to deploy SSH application which leverages Golang to expose shell sessions to organize HTB sessions.

> 💡 Check the [`documentation`](https://archimoebius.github.io/hexer/) for usage and more information.

## Quickstart

### Setup Golang (version of at least 1.23 required)

Download at least version 1.23 of [`Golang`](https://go.dev/dl/) for example:

```bash
sudo su -
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
```

### Setup Hexer

Get up and running quick like:

```
go run github.com/archimoebius/hexer@latest serve
```

## Install / Setup

To download hexer and then use it - something like the following will do:

```
go install github.com/archimoebius/hexer@latest
hexer serve
```

## Credits

Without the shoulders of giants to stand upon - this project wouldn't exist... Thank you for crafting such great libraries!

* [CharmBracelet](https://github.com/charmbracelet/)
* [Gliderlabs/SSH](https://github.com/gliderlabs/ssh)
* [Litter](https://github.com/sanity-io/litter)
* [Logrus](https://github.com/sirupsen/logrus)
* [Cobra](https://github.com/spf13/cobra)
* [Viper](https://github.com/spf13/viper)

## 🤝 Contributing

Contributions, issues, and feature requests are welcome. Feel free to check issues page if you want to contribute.

## 📝 License

Copyright ©2025 ArchiMoebius.

This project is [GPL](https://raw.githubusercontent.com/ArchiMoebius/hexer/main/LICENSE) licensed.
