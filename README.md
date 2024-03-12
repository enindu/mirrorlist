# mirrorlist

mirrorlist is a simple [pacman](https://wiki.archlinux.org/index.php/Pacman) mirror list generator.

## Install

You can install mirrorlist using `go install` command.

```shell
go install github.com/enindu/mirrorlist
```

## Usage

There are 9 flags in mirrorlist.

- `-h`: Display help message.
- `-mirror-list-timeout`: Mirror list request timeout to send and receive response.
- `-mirror-timeout`: Mirror request timeout to send and receive response.
- `-http-only`: Use only HTTP mirrors to generate. This can not use with -https-only flag.
- `-https-only`: Use only HTTPS mirrors to generate. This can not use with -http-only flag.
- `-count`: Count of mirrors to generate.
- `-pings`: Pings per a mirror. Higher pings means precise results, but high execution time.
- `-output`: Store mirrors in a file. This truncate any existing file.
- `-verbose`: Display warnings and informations in terminal.

You can see more information using `go doc` command.

```shell
go doc mirrorlist
```
