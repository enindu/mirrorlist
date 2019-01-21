# Mirrorlist
Simple [Pacman](https://wiki.archlinux.org/index.php/Pacman) mirrorlist generator, written in Go

## Build
```
$ cd mirrorlist/
$ go build
```

## Usage
```
mirrorlist [OPTION...]

-c int
    Count of mirrors (Default: 3)
-m float
    Maximum response time (In seconds) of a mirror (Default: 1)
-u string
    Mirrorlist URL (Default: "https://www.archlinux.org/mirrorlist/all")
```

## Examples
```
$ mirrorlist
$ mirrorlist -m 10
$ mirrorlist -c 10 -m 5 -u https://www.archlinux.org/mirrorlist/?country=AU
```