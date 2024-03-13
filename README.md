# mirrorlist

Mirrorlist is a simple [pacman](https://wiki.archlinux.org/index.php/Pacman) mirror list generator.

## Install

You can install Mirrorlist using `go install` command.

```shell
go install github.com/enindu/mirrorlist
```

## Usage

There are 9 flags for `mirrorlist` command.

- `-h`: Display help message.
- `-mirror-list-timeout`: Request timeout to send and receive response from mirror list URL.
- `-mirror-timeout`: Request timeout to send and receive response from mirror URL.
- `-http-only`: Use only HTTP mirrors to generate mirror list. This can not use with -https-only flag.
- `-https-only`: Use only HTTPS mirrors to generate mirror list. This can not use with -http-only flag.
- `-count`: Count of mirrors to generate.
- `-pings`: Pings per a mirror. Higher pings means precise results, but high execution time.
- `-output`: Store mirrors in a file. This truncate any existing file.
- `-verbose`: Display warning messages in command line.

You can see more information using `go doc` command.

```shell
go doc mirrorlist
```
