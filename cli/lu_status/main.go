// lu_status displays current state of a Lustre client
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.intel.com/hpdd/lustre/fs"
	"github.intel.com/hpdd/lustre/status"
)

func mdcStatus(c *status.LustreClient, mdc string) {
	path := c.ClientPath("mdc", mdc)
	imp, err := status.ReadImport(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  %s: %s %s\n", mdc, imp.Connection.CurrentConnection, imp.State)
}

func oscStatus(c *status.LustreClient, osc string) string {
	path := c.ClientPath("osc", osc)
	imp, err := status.ReadImport(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  %s: %s %s %v\n", osc, imp.Connection.CurrentConnection, imp.State,
		imp.Averages.MegabytesPerSec)
	return ""
}

func readFile(path string) (string, error) {
	line, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(line)), nil
}

func clientStatus(c *status.LustreClient) string {
	fmt.Println(c)

	fmt.Println(" lmv:")
	for _, mdc := range c.LMVTargets() {
		mdcStatus(c, mdc)
	}

	fmt.Println(" lov:")
	for _, osc := range c.LOVTargets() {
		oscStatus(c, osc)
	}

	return ""
}

var mountPath string

func init() {
	flag.StringVar(&mountPath, "mnt", "", "Lustre mount point.")
}

/*
Calculate stddev from lustre stat files:
write_bytes               39 samples [bytes] 276272 1048576 38268528 38735913463296
                          count                              sum       sum_squares
One reference I saw but can't find again said it was this:
mean = sum/count
stddev = sqrt((sum_squares/count) - (mean**2))

Another reference says this:
stdev = sqrt((sum_squares - sum**2)/(count*count-1))

From: http://mathcentral.uregina.ca/QQ/database/QQ.09.02/carlos1.html
 (Note reference to a better method from Knuth)

*/

func main() {
	flag.Parse()
	if mountPath == "" {
		log.Fatal("missing -mnt paramter")
	}

	c, err := fs.MountID(mountPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	clientStatus(c)
}
