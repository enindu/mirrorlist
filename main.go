package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

func main() {
	var (
		mirrors        []string  // Mirrors slice
		executionTimes []float64 // Mirror execution times slice
	)

	// Set program start time
	start := time.Now()

	// Set command line flags
	count := flag.Int("c", 3, "Count of mirrors")
	maxTime := flag.Float64("m", 1, "Maximum response time (In seconds) of a mirror")
	url := flag.String("u", "https://www.archlinux.org/mirrorlist/all", "Mirrorlist URL")

	flag.Parse()

	// Get response from mirrorlist URL
	response, errors := http.Get(*url)

	if errors != nil {
		log.Fatal("Error:", errors)
	}

	// Convert response to bytes
	bytes, errors := ioutil.ReadAll(response.Body)

	if errors != nil {
		log.Fatal("Error:", errors)
	}

	// Close response body
	response.Body.Close()

	// Convert bytes to string
	body := strings.TrimSpace(string(bytes))

	// Scan string
	scanner := bufio.NewScanner(strings.NewReader(body))

	// Loop through scanner
	for scanner.Scan() {
		// Check if there're unwanted lines
		if !strings.HasPrefix(scanner.Text(), "##") && scanner.Text() != "" {
			// Append mirror to mirrors slice
			mirrors = append(mirrors, strings.Replace(strings.Replace(scanner.Text(), "#Server = ", "", -1), "/$repo/os/$arch", "", -1))
		}
	}

	// Define sorted mirrors map
	sortedMirrors := make(map[float64]string)

	// Loop through mirrors slice
	for _, mirror := range mirrors {
		// Set mirror execution start time
		start := time.Now()

		// Get response of mirror
		_, errors := http.Get(mirror)

		// Set mirror execution end time
		end := time.Now().Sub(start).Seconds()

		// Check if there's any error or mirror execution end time exceeds max time flag
		if errors != nil || end >= *maxTime {
			// Continue loop again
			continue
		}

		// Check if sorted mirrors map length equels to count flag
		if len(sortedMirrors) == *count {
			// Break loop
			break
		}

		// Check if mirror execution end time lower than max time flag
		if end < *maxTime {
			// Append mirror to sorted mirrors map, execution end time as key and mirror as value
			sortedMirrors[end] = mirror
		}
	}

	// Loop through sorted mirrors
	for key := range sortedMirrors {
		// Append mirror execution time to execution times slice
		executionTimes = append(executionTimes, key)
	}

	// Sort execution times slice
	sort.Float64s(executionTimes)

	// Loop through execution time slice
	for _, executionTime := range executionTimes {
		// Print mirror that ordered by execution time (Least first)
		fmt.Printf("Server = %s/$repo/os/$arch\n", sortedMirrors[executionTime])
	}

	// Set program end time
	end := time.Now().Sub(start).Seconds()

	// Print program footer notes
	fmt.Println("# Generated by github.com/enindu/mirrorlist")
	fmt.Println("# Total execution time:", end)
}
