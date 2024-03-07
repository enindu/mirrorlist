// This file is part of mirrorlist.
// Copyright (C) 2024 Enindu Alahapperuma
//
// mirrorlist is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// mirrorlist is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with
// mirrorlist. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"cmp"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/enindu/palette"
)

const (
	allMirrorListUrl   string = "https://archlinux.org/mirrorlist/all"
	httpMirrorListUrl  string = "https://archlinux.org/mirrorlist/all/http"
	httpsMirrorListUrl string = "https://archlinux.org/mirrorlist/all/https"
)

var wait *sync.WaitGroup = &sync.WaitGroup{}

var (
	info *palette.Printer = palette.NewPrinterInfo()
	warn *palette.Printer = palette.NewPrinterWarn()
	erro *palette.Printer = palette.NewPrinterErro()
)

var (
	mirrorListTimeout *time.Duration = flag.Duration("mirror-list-timeout", 10*time.Second, "Mirror list request timeout to send and receive response.")
	mirrorTimeout     *time.Duration = flag.Duration("mirror-timeout", 5*time.Second, "Mirror request timeout to send and receive response.")
	httpOnly          *bool          = flag.Bool("http-only", false, "Use only HTTP mirrors to generate. This can not use with -https-only flag.")
	httpsOnly         *bool          = flag.Bool("https-only", false, "Use only HTTPS mirrors to generate. This can not use with -http-only flag.")
	count             *int           = flag.Int("count", 5, "Count of mirrors to generate.")
	pings             *int           = flag.Int("pings", 5, "Pings per a mirror. Higher pings means precise results, but high execution time.")
	output            *string        = flag.String("output", "", "Store mirrors in a file. This truncate any existing file.")
	verbose           *bool          = flag.Bool("verbose", false, "Display warnings and informations in terminal.")
)

func main() {
	// Parse flags.
	flag.Parse()

	// Check if both -http-only and -https-only flags used.
	if *httpOnly && *httpsOnly {
		erro.Write("ERRO could not run mirrorlist, because both -http-only and -https-only flags are used.\n")
		return
	}

	// Create mirror list URL.
	mirrorListUrl := allMirrorListUrl
	if *httpOnly {
		mirrorListUrl = httpMirrorListUrl
	}
	if *httpsOnly {
		mirrorListUrl = httpsMirrorListUrl
	}

	// Create mirror list HTTP client.
	mirrorListClient := &http.Client{
		Timeout: *mirrorListTimeout,
	}

	// Get response from mirror list URL.
	mirrorListResponse, fault := mirrorListClient.Get(mirrorListUrl)
	if fault != nil {
		erro.Write("ERRO could not run mirrorlist, because %s is not responding.\n", mirrorListUrl)
		return
	}
	defer mirrorListResponse.Body.Close()

	// Create mirror URLs.
	mirrorUrls := []string{}
	mirrorListScanner := bufio.NewScanner(mirrorListResponse.Body)
	for mirrorListScanner.Scan() {
		// Create mirror URL.
		mirrorUrl := mirrorListScanner.Text()
		mirrorUrl = strings.TrimSpace(mirrorUrl)
		mirrorUrl = strings.TrimPrefix(mirrorUrl, "#Server = ")
		mirrorUrl = strings.TrimSuffix(mirrorUrl, "/$repo/os/$arch")
		if mirrorUrl == "" {
			continue
		}
		if strings.HasPrefix(mirrorUrl, "##") {
			continue
		}

		// Parse mirror URL.
		parseMirrorUrl, fault := url.Parse(mirrorUrl)
		if fault != nil {
			if *verbose {
				warn.Write("WARN could not parse %s\n", mirrorUrl)
			}
			continue
		}
		mirrorUrl = parseMirrorUrl.String()

		// Store mirror URL.
		mirrorUrls = append(mirrorUrls, mirrorUrl)
	}

	// Create mirror HTTP client.
	mirrorClient := &http.Client{
		Timeout: *mirrorTimeout,
	}

	// Create mirrors channel.
	mirrorUrlsLength := len(mirrorUrls)
	mirrorsChannel := make(chan mirror, mirrorUrlsLength)

	// Define execution start time.
	start := time.Now()

	// Ping mirror URLs and store URL and time.
	wait.Add(mirrorUrlsLength)
	for _, mirrorUrl := range mirrorUrls {
		go ping(mirrorUrl, mirrorClient, mirrorsChannel)
	}
	wait.Wait()
	close(mirrorsChannel)

	// Define execution end time.
	end := time.Since(start).Seconds()

	// Check if mirrors less than count.
	if len(mirrorsChannel) < *count {
		erro.Write("ERRO could not run mirrorlist, because unable to get %d mirror(s).\n", *count)
		return
	}

	// Get mirrors.
	mirrors := []mirror{}
	for mirror := range mirrorsChannel {
		mirrors = append(mirrors, mirror)
	}

	// Sort mirrors by time.
	slices.SortFunc(mirrors, func(x mirror, y mirror) int {
		return cmp.Compare(x.time, y.time)
	})

	// Get output.
	if *output == "" {
		// Print mirrors in terminal.
		for _, mirror := range mirrors[:*count] {
			fmt.Printf("# %f\n", mirror.time)
			fmt.Printf("Server = %s/$repo/os/$arch\n", mirror.url)
		}
	} else {
		// Open or create output file.
		outputFile, fault := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY, 0644)
		if fault != nil {
			erro.Write("ERRO could not run mirrorlist, because unable to create %s.\n", *output)
			return
		}
		defer outputFile.Close()

		// Write mirrors to output file.
		for _, mirror := range mirrors[:*count] {
			timeLine := fmt.Sprintf("# %f\n", mirror.time)
			urlLine := fmt.Sprintf("Server = %s/$repo/os/$arch\n", mirror.url)
			outputFile.WriteString(timeLine)
			outputFile.WriteString(urlLine)
		}
	}

	// Print information messages.
	if *verbose {
		info.Write("INFO executed in %.2f seconds.\n", end)
		info.Write("INFO generated by mirrorlist.\n")
	}
}

func ping(mirrorUrl string, mirrorClient *http.Client, mirrorsChannel chan mirror) {
	// Defer wait done.
	defer wait.Done()

	// Ping mirror URL.
	end := time.Duration(0)
	for i := 0; i < *pings; i++ {
		start := time.Now()
		mirrorResponse, fault := mirrorClient.Get(mirrorUrl)
		if fault != nil {
			if *verbose {
				warn.Write("WARN could not get response from %s.\n", mirrorUrl)
			}
			return
		}
		defer mirrorResponse.Body.Close()
		if mirrorResponse.StatusCode != http.StatusOK {
			return
		}
		end = end + time.Since(start)
	}
	if end == 0 {
		return
	}

	// Send mirror to mirrors channel.
	mirrorsChannel <- mirror{
		url:  mirrorUrl,
		time: end.Seconds() / float64(*pings),
	}
}
