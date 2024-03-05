# mirrorlist

mirrorlist is a simple [pacman](https://wiki.archlinux.org/index.php/Pacman) mirror list generator.

## Install

You can install mirrorlist using `go install` command.

```
go install github.com/enindu/mirrorlist
```

## Usage

There are 4 flags in mirrorlist.

- `-http`: Use only HTTP mirrors to generate.
- `-https`: Use only HTTPS mirrors to generate.
- `-count`: Count of mirrors to generate. Default value is 5.
- `-pings`: Pings per a mirror. Higher pings means precise results, but high execution time. Default value is 5.

Here're some examples of how to use mirrorlist.

```
mirrorlist
mirrorlist -count 10 -pings 10
mirrorlist -http -count 10 -pings 10
mirrorlist -https -count 10 -pings 10
```
