# mirrorlist

mirrorlist is a simple [pacman](https://wiki.archlinux.org/index.php/Pacman) mirror list generator.

## Install

You can install mirrorlist using `go install` command.

```shell
go install github.com/enindu/mirrorlist
```

## Usage

There are 4 flags in mirrorlist.

- `-http`: Use only HTTP mirrors to generate. This can not use with -https flag.
- `-https`: Use only HTTPS mirrors to generate. This can not use with -http flag.
- `-count`: Count of mirrors to generate. Default value is 5.
- `-pings`: Pings per a mirror. Higher pings means precise results, but high execution time. Default value is 5.
- `-output`: Store mirrors in a file. This truncate any existing file.
- `-verbose`: Display warnings and informations in terminal.

Here're some examples of how to use mirrorlist.

```shell
mirrorlist                    # Run mirrorlist with default options.
mirrorlist -http              # Use only HTTP mirrors.
mirrorlist -https             # Use only HTTPS mirrors.
mirrorlist -count 10          # Generate 10 mirrors.
mirrorlist -pings 10          # Ping a mirror 10 times.
mirrorlist -output mirrorlist # Store mirrors in /etc/pacman.d/mirrorlist file.
                              # You probably need to run this with sudo.
                              # Do not forget to backup current mirrorlist file before run this command.
mirrorlist -verbose           # Display warnings and information in terminal.
```
