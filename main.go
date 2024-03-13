// This file is part of Mirrorlist.
// Copyright (C) 2024 Enindu Alahapperuma
//
// Mirrorlist is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// Mirrorlist is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with
// Mirrorlist. If not, see <https://www.gnu.org/licenses/>.

// Mirrorlist is a simple [pacman] mirror list generator.
//
// Usage:
//
//	mirrorlist [flags]
//
// The flags are:
//
//	-mirror-list-timeout
//		Request timeout to send and receive response from mirror list URL.
//	-mirror-timeout
//		Request timeout to send and receive response from mirror URL.
//	-http-only
//		Use only HTTP mirrors to generate mirror list. This can not use with
//		-https-only flag.
//	-https-only
//		Use only HTTPS mirrors to generate mirror list. This can not use with
//		-http-only flag.
//	-count
//		Count of mirrors to generate.
//	-pings
//		Pings per a mirror. Higher pings means precise results, but high
//		execution time.
//	-output
//		Store mirrors in a file. This truncate any existing file.
//	-verbose
//		Display warning messages in command line.
//
// [pacman]: https://wiki.archlinux.org/index.php/Pacman
package main

import (
	"bufio"
	"cmp"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/enindu/palette"
)

const (
	allLink   string = "https://archlinux.org/mirrorlist/all"
	httpLink  string = "https://archlinux.org/mirrorlist/all/http"
	httpsLink string = "https://archlinux.org/mirrorlist/all/https"
)

var wait *sync.WaitGroup = &sync.WaitGroup{}

var (
	info *palette.Printer = palette.NewPrinterInfo()
	warn *palette.Printer = palette.NewPrinterWarn()
	erro *palette.Printer = palette.NewPrinterErro()
)

var (
	mirrorListTimeout *time.Duration = flag.Duration("mirror-list-timeout", 10*time.Second, "Request timeout to send and receive response from mirror list URL.")
	mirrorTimeout     *time.Duration = flag.Duration("mirror-timeout", 5*time.Second, "Request timeout to send and receive response from mirror URL.")
	httpOnly          *bool          = flag.Bool("http-only", false, "Use only HTTP mirrors to generate mirror list. This can not use with -https-only flag.")
	httpsOnly         *bool          = flag.Bool("https-only", false, "Use only HTTPS mirrors to generate mirror list. This can not use with -http-only flag.")
	count             *int           = flag.Int("count", 5, "Count of mirrors to generate.")
	pings             *int           = flag.Int("pings", 5, "Pings per a mirror. Higher pings means precise results, but high execution time.")
	output            *string        = flag.String("output", "", "Store mirrors in a file. This truncate any existing file.")
	verbose           *bool          = flag.Bool("verbose", false, "Display warning messages in command line.")
)

func main() {
	// Parse flags.
	flag.Parse()

	// Check if both -http-only and https-only flags used.
	if *httpOnly && *httpsOnly {
		erro.Print("can not use both -http-only and -https-only flags\n")
		return
	}

	// Create mirror list link.
	link := ""
	switch {
	case *httpOnly:
		link = httpLink
	case *httpsOnly:
		link = httpsLink
	default:
		link = allLink
	}

	// Create mirror list HTTP client.
	client := &http.Client{
		Timeout: *mirrorListTimeout,
	}

	// Get response from mirror list link.
	response, err := client.Get(link)
	if err != nil {
		erro.Print("could not get response from %s\n", link)
		return
	}
	defer response.Body.Close()

	// Create mirror links.
	links := []string{}
	regex := regexp.MustCompile(`#Server\s*\=\s*(.+?)\/\$repo\/os\/\$arch`)
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		link := scanner.Text()
		matches := regex.FindStringSubmatch(link)
		if len(matches) < 1 {
			continue
		}
		links = append(links, matches[1])
	}

	// Create mirror HTTP client.
	client.Timeout = *mirrorTimeout

	// Create mirrors channel.
	length := len(links)
	channel := make(chan *mirror, length)

	// Define execution beginning time.
	begin := time.Now()

	// Ping mirror links.
	wait.Add(length)
	for _, v := range links {
		go ping(v, client, channel)
	}
	wait.Wait()
	close(channel)

	// Define execution ending time.
	end := time.Since(begin).Seconds()

	// Check if mirrors count less than requested count.
	if len(channel) < *count {
		erro.Print("could not get %d mirror(s)\n", *count)
		return
	}

	// Sort mirrors by time.
	mirrors := []mirror{}
	for v := range channel {
		mirrors = append(mirrors, *v)
	}
	slices.SortFunc(mirrors, func(x mirror, y mirror) int {
		return cmp.Compare(x.time, y.time)
	})

	// Get output.
	writer := os.Stdout
	if *output != "" {
		file, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			erro.Print("could not create %s\n", *output)
			return
		}
		defer file.Close()
		writer = file
	}
	for _, v := range mirrors[:*count] {
		fmt.Fprintf(writer, "# %f\n", v.time)
		fmt.Fprintf(writer, "Server = %s/$repo/os/$arch\n", v.link)
	}
	fmt.Fprintf(writer, "# Generated by github.com/enindu/mirrorlist\n")

	// Print information messages.
	info.Print("mirror list is generated\n")
	info.Print("executed in %.2f seconds\n", end)
}

func ping(l string, c *http.Client, m chan *mirror) {
	// Defer wait done.
	defer wait.Done()

	// Ping mirror link.
	end := time.Duration(0)
	for i := 0; i < *pings; i++ {
		begin := time.Now()
		response, err := c.Get(l)
		if err != nil {
			if *verbose {
				warn.Print("could not get response from %s\n", l)
			}
			return
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			if *verbose {
				warn.Print("got %d status code from %s\n", response.StatusCode, l)
			}
			return
		}
		end = end + time.Since(begin)
	}
	if end <= 0 {
		if *verbose {
			warn.Print("%s is not responding\n", l)
		}
		return
	}

	// Send mirror to mirrors channel.
	m <- &mirror{
		link: l,
		time: end.Seconds() / float64(*pings),
	}
}
