# oax(OpenAI eXecutor)

💥 CLI for OpenAI's ChatGPT.

*For basic use cases.*
![gif](https://res.cloudinary.com/dkerzyk09/image/upload/v1679860503/tools/oax/oax-chat.gif)


*Resuming chat sessions from a previous point.*
![gif](https://res.cloudinary.com/dkerzyk09/image/upload/v1679860850/tools/oax/chat-resume.gif)

## Installation

```bash
brew tap shuntaka9576/tap
brew install shuntaka9576/tap/oax
```

## Requirements

* Vim/Neovim

## Quick Start


Open profile (`~/.config/oax/profiles.toml`).
```bash
oax config --profiles
```

Specify the OpenAI API API key in the apiKey field (replace `sk-xxxx`).
```toml
[[profiles]]
name = "personal"
apiKey = "sk-xxxx" # <--
default = true
```

Open setting (`~/.config/oax/settings.toml`).
```bash
oax config --settings
```

Specify edit to lunch editor. `vim` or `nvim`.
```toml
[setting]
  editor = "vim" # <--
```

Start ChatGPT. Default model `gpt-3.5-turbo`. Open ChatGPT request file (`~/.config/oax/chat-log/2006-01-02_15-04-05.toml`).
```bash
oax chat
```

Specify sent content to ChatGPT. exit with wq.
```toml
[[messages]]
  sequentialNumber = 0
  role = "user"
  content = '''
# Remove this comment and specify content to send to OpenAI API; otherwise, nothing is sent.
'''
```

Streaming response is returned from ChatGPT.
```bash
$ oax chat
Hello! How can I assist you today?

continue (y/n)?: n
saved: ~/.config/oax/chat-log/2023-03-26_17-01-54.toml
```

Files can be resumed from the middle of the process by specifying the full path of the file.
```bash
oax chat -m "gpt-3.5-turbo" -f "~/.config/oax/chat-log/2023-03-26_15-11-04.toml"
```


## Configuration

|File Path|Description|Open Command
|---|---|---|
|`~/.config/oax/settings.toml`|Specify command assist information for oax.|`oax config --settings`
|`~/.config/oax/profiles.toml`|Specify information required for API connection.|`oax config --profiles`

### Setting

|Option|Description|Required|Default|
|---|---|---|---|
|editor|Integrated editor|true|`vim`|
|chatLogDir|Directory for saving chat logs|false|`~/.config/oax/chat-log`|

e.g.
```toml
[setting]
  editor = "nvim"
  chatLogDir = "~/.config/oax/chat-log"
```

### Profiles

|Option|Description|Required|Default|
|---|---|---|---|
|name|Profile name|true|`vim`|
|apiKey|OpenAI API key|true|`~/.config/oax/chat-log`|
|default|Set the default profile configuration (API key) to be used.|false. Please ensure that the "default" option is set for at least one Profile.|`true`|


e.g.
```toml
[[profiles]]
  name = "me"
  apiKey= "sk-xxxx"
  default = true

[[profiles]]
  name = "org"
  organizationId = ""
```


## Troubleshooting

### Reset

```bash
rm -rf ~/.config/oax
```
