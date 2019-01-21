package main

import (
	"bufio"
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
		mirrors        []string
		executionTimes []float64
	)

	start := time.Now()
	response, errors := http.Get("https://www.archlinux.org/mirrorlist/all")

	if errors != nil {
		log.Fatal("Error:", errors)
	}

	bytes, errors := ioutil.ReadAll(response.Body)

	if errors != nil {
		log.Fatal("Error:", errors)
	}

	response.Body.Close()

	body := strings.TrimSpace(string(bytes))
	scanner := bufio.NewScanner(strings.NewReader(body))

	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "##") && scanner.Text() != "" {
			mirrors = append(mirrors, strings.Replace(strings.Replace(scanner.Text(), "#Server = ", "", -1), "/$repo/os/$arch", "", -1))
		}
	}

	sortedMirrors := make(map[float64]string)

	for _, mirror := range mirrors {
		start := time.Now()
		_, errors := http.Get(mirror)
		end := time.Now().Sub(start).Seconds()

		if errors != nil || end >= 1 {
			continue
		}

		if len(sortedMirrors) == 3 {
			break
		}

		if end < 1 {
			sortedMirrors[end] = mirror
		}
	}

	for key := range sortedMirrors {
		executionTimes = append(executionTimes, key)
	}

	sort.Float64s(executionTimes)

	for _, executionTime := range executionTimes {
		fmt.Printf("Server = %s/$repo/os/$arch # %f\n", sortedMirrors[executionTime], executionTime)
	}

	end := time.Now().Sub(start).Seconds()

	fmt.Println("# Total execution time:", end)
}
