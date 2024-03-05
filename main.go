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
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	allMirrorList   string = "https://www.archlinux.org/mirrorlist/all"
	httpMirrorList  string = "https://archlinux.org/mirrorlist/all/http"
	httpsMirrorList string = "https://archlinux.org/mirrorlist/all/https"
)

func main() {
	// Create flags.
	httpFlag := flag.Bool("http", false, "Use only HTTP mirrors to generate")
	httpsFlag := flag.Bool("https", false, "Use only HTTPS mirrors to generate")
	countFlag := flag.Int("count", 5, "Count of mirrors to generate")
	pingsFlag := flag.Int("pings", 5, "Pings per a mirror. Higher pings means precise results, but high execution time.")
	flag.Parse()

	// Create mirrors list URL string.
	mirrorsListURLString := allMirrorList
	if *httpFlag {
		mirrorsListURLString = httpMirrorList
	}
	if *httpsFlag {
		mirrorsListURLString = httpsMirrorList
	}

	// Define execution start time.
	start := time.Now()

	// Create mirrors list URL. If error occurs while creating URL, return error.
	mirrorsListURL, mirrorsListURLFault := url.Parse(mirrorsListURLString)
	if mirrorsListURLFault != nil {
		fmt.Printf("%v\n", mirrorsListURLFault)
		return
	}

	// Get response from mirrors list URL. If error occurs while getting response,
	// return error.
	mirrorsListURLString = mirrorsListURL.String()
	mirrorsListResponse, mirrorsListResponseFault := http.Get(mirrorsListURLString)
	if mirrorsListResponseFault != nil {
		fmt.Printf("%v\n", mirrorsListResponseFault)
		return
	}
	defer mirrorsListResponse.Body.Close()

	// Create mirror URLs.
	mirrorURLStrings := []string{}
	mirrorsScanner := bufio.NewScanner(mirrorsListResponse.Body)
	for mirrorsScanner.Scan() {
		line := mirrorsScanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "##") {
			continue
		}
		line = strings.TrimPrefix(line, "#Server = ")
		line = strings.TrimSuffix(line, "/$repo/os/$arch")
		mirrorURL, mirrorURLFault := url.Parse(line)
		if mirrorURLFault != nil {
			continue
		}
		mirrorURLString := mirrorURL.String()
		mirrorURLStrings = append(mirrorURLStrings, mirrorURLString)
	}

	// Get mirrors.
	mirrorURLStringsLength := len(mirrorURLStrings)
	mirrorsChannel := make(chan mirror, mirrorURLStringsLength)
	wait := sync.WaitGroup{}
	wait.Add(mirrorURLStringsLength)
	for _, mirrorURLString := range mirrorURLStrings {
		go ping(mirrorsChannel, mirrorURLString, *pingsFlag, &wait)
	}
	wait.Wait()
	close(mirrorsChannel)
	mirrors := []mirror{}
	for mirror := range mirrorsChannel {
		mirrors = append(mirrors, mirror)
	}

	// Sort mirrors by duration.
	slices.SortFunc(mirrors, func(x mirror, y mirror) int {
		return cmp.Compare(x.duration, x.duration)
	})

	// Print mirrors as in /etc/pacman.d/mirrorlist.
	for _, item := range mirrors[:*countFlag] {
		fmt.Printf("Server = %s/$repo/os/$arch # %f\n", item.url, item.duration)
	}

	// Define execution end time.
	end := time.Since(start).Seconds()

	// Print information messages.
	fmt.Printf("## Executed in %.2f seconds\n", end)
	fmt.Printf("## Generated by mirrorlist\n")
}

func ping(mirrorsChannel chan mirror, mirrorURLString string, pings int, wait *sync.WaitGroup) {
	// Defer wait done.
	defer wait.Done()

	// Send requests pings times to mirror URL and get total execution time. If
	// execution time equals to 0, which means URL is not responding, return.
	end := time.Duration(0)
	for i := 0; i < pings; i++ {
		start := time.Now()
		mirrorResponse, fault := http.Get(mirrorURLString)
		if fault != nil {
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
		url:      mirrorURLString,
		duration: end.Seconds() / float64(pings),
	}
}
