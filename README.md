# Mirrorlist
Simple [Pacman](https://wiki.archlinux.org/index.php/Pacman) mirrorlist generator, written in Go

## Build
Use Go compiler to build `mirrorlist` executable.
```
$ go build
```

## Usage
There is one flag that can be used with `mirrorlist`. Define count of mirrors to be generated by using `-C` flag. It should be integer. The default value is `5`.
```
$ mirrorlist -C 10
```
