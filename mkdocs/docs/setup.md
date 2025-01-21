## Quick Start - Release Binary

0. Download a release

1. copy it to your desired location

2. Jump to 'Configuration'

## Quick Start - I'm'a Developer!

0. `git clone git@github.com:ArchiMoebius/hexer.git && cd hexer`

1. `make`

2. copy ./dist/temp/*flavor* to your desired location

3. Jump to 'Configuration'

## Configuration

You can use a combination of CLI Parameters, environment variables, .hexer.yaml, and default settings to get up and running. The precedence for application of state is as follows:

```bash
	1: CLI Parameters
	2: environment
	3: .hexer.yaml
	4: defaults
```

### Quickstart

The simplest way to deploy Hexer is by using the 'local' mode - no SSH server is setup, and upon execution you are dropped into a TUI.

See the following example `.hexer.yaml` configuration file.

```yaml
storage: ./data
ssh-user: user
ssh-key: >
  -----BEGIN OPENSSH PRIVATE KEY-----
  b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
  QyNTUxOQAAACBcWl5/iii6jfsE0zArXJWUj+X/LjlPTgwM1dW5oigb1AAAAJAhG/DKIRvw
  ygAAAAtzc2gtZWQyNTUxOQAAACBcWl5/iii6jfsE0zArXJWUj+X/LjlPTgwM1dW5oigb1A
  AAAEBTqrQNOTC/u+ETrWCEOzRbwwEi7nxlHkNgZS4C2IBa9FxaXn+KKLqN+wTTMCtclZSP
  5f8uOU9ODAzV1bmiKBvUAAAACnVzZXJAY29tcHkBAgM=
  -----END OPENSSH PRIVATE KEY-----
local: true
ip: 127.0.0.1
port: 2222
database_ip: 127.0.0.1
database_port: 8090
hostkey: >
  -----BEGIN RSA PRIVATE KEY-----
  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
  -----END RSA PRIVATE KEY-----
theme: # base from https://ethanschoonover.com/solarized/
  style:
    header_text:
      foreground:
        light: "#859900"
        dark: "#859900"
    status:
      foreground:
        light: "#859900"
        dark: "#859900"
    status_header:
      foreground:
        light: "#586E75"
        dark: "#93A1A1"
    title_style:
      foreground:
        light: "#859900"
        dark: "#859900"
      background:
        light: "#586E75"
        dark: "#93A1A1"
    error_header_text:
      foreground:
        light: "#C24543"
        dark: "#C24543"
    tab_style:
      inactive_highlight_color:
        border_foreground:
          light: "#457A9F"
          dark: "#457A9F"
      active_highlight_color:
        border_foreground:
          light: "#5960D7"
          dark: "#5960D7"
    list_default_item_style:
      normal_title:
        foreground:
          light: "#2380C2"
          dark: "#2380C2"
      normal_description:
        foreground:
          light: "#457A9F"
          dark: "#457A9F"
      selected_title:
        foreground:
          light: "#585FD8"
          dark: "#585FD8"
        border_foreground:
          light: "#5960D7"
          dark: "#5960D7"
      selected_description:
        foreground:
          light: "#6C71C4"
          dark: "#6C71C4"
      dimmed_title:
        foreground:
          light: "#AB4BB0"
          dark: "#AB4BB0"
      dimmed_description:
        foreground:
          light: "#8B3D8F"
          dark: "#8B3D8F"
```