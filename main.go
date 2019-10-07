package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"html/template"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gobuffalo/packr/v2"
)

type Lease struct {
	Hostname string
	MAC      string
	IPv4     string
	Expire   time.Time
}

func main() {
	box := packr.New("templates", "./templates")

	var leasesPath = flag.String("in", "/var/lib/kea/kea-leases4.csv", "Path to your leases csv file")
	var outputPath = flag.String("out", "output.html", "Path to the rendered HTML file")
	flag.Parse()

	csvFile, err := os.Open(*leasesPath)
	defer csvFile.Close()

	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))

	// Read the first line, we are not interested in the CSV header
	reader.Read()

	currentTime := int64(time.Now().Unix())

	var leases []Lease
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		i, _ := strconv.ParseInt(line[4], 10, 64)
		if i < currentTime {
			// Lease already expired
			continue
		}

		l := Lease{
			IPv4:     line[0],
			MAC:      line[1],
			Expire:   time.Unix(i, 0),
			Hostname: line[8],
		}
		leases = append(leases, l)
	}

	templateString, _ := box.FindString("template.html")
	t, _ := template.New("email").Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(*outputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	t.Execute(out, leases)

}
