package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

type NagiosStatusVal int

// The values with which a Nagios check can exit
const (
	NAGIOS_OK NagiosStatusVal = iota
	NAGIOS_WARNING
	NAGIOS_CRITICAL
	NAGIOS_UNKNOWN
)

// Maps the NagiosStatusVal entries to output strings
var (
	valMessages = []string{
		"OK:",
		"WARNING:",
		"CRITICAL:",
		"UNKNOWN:",
	}
)

// A type representing a Nagios check status. The Value is a the exit code
// expected for the check and the Message is the specific output string.
type NagiosStatus struct {
	Message string
	Value   NagiosStatusVal
}

// Main program loop
func main() {
	// get our numbers
	atime, mtime, ctime, err := statTimes(fileFullPath)
	now := time.Now()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("WORKING")
	fmt.Println(atime.Before(now))
	fmt.Println("Now:   ", now)
	fmt.Println("Atime: ", atime)
	fmt.Println("END WORK")
	fmt.Println(mtime)
	fmt.Println(ctime)
	fmt.Println(warnTime)
	fmt.Println(critTime)
}

// credit to peterSO from http://stackoverflow.com/questions/20875336/how-can-i-get-a-files-ctime-atime-mtime-and-change-them-using-golang
// for the original function below. It is exactly what was necessary
func statTimes(name string) (atime, mtime, ctime time.Time, err error) {
	fi, err := os.Stat(name)
	if err != nil {
		return
	}
	mtime = fi.ModTime()
	stat := fi.Sys().(*syscall.Stat_t)
	atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	return
}

// Init loop to set our variables on startup

// Three primary variables used to define our check
var fileFullPath, timeMode string
var warnTime, critTime time.Duration

func init() {
	var warnTimeString, critTimeString string
	flag.StringVar(&fileFullPath, "f", "", "The full location to the file to be checked")
	flag.StringVar(&timeMode, "t", "ctime", "Which time stat to use for comparison. [atime|mtime|ctime] (access time, modified time, changed time).")
	flag.StringVar(&warnTimeString, "w", "24h", "The max age a file can be before triggering a warning")
	flag.StringVar(&critTimeString, "c", "48h", "The max age a file can be before triggering a critical")

	flag.Parse()
	var err error

	// check that our time arguments are valid
	warnTime, err = time.ParseDuration(warnTimeString)
	if err != nil {
		log.Fatal(err)
	}

	critTime, err = time.ParseDuration(critTimeString)
	if err != nil {
		log.Fatal(err)
	}

}

// Take a bunch of NagiosStatus pointers and find the highest value, then
// combine all the messages. Things win in the order of highest to lowest.
func (status *NagiosStatus) Aggregate(otherStatuses []*NagiosStatus) {
	for _, s := range otherStatuses {
		if status.Value < s.Value {
			status.Value = s.Value
		}

		status.Message += " - " + s.Message
	}
}

// Exit with an UNKNOWN status and appropriate message
func Unknown(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_UNKNOWN})
}

// Exit with an CRITICAL status and appropriate message
func Critical(err error) {
	ExitWithStatus(&NagiosStatus{err.Error(), NAGIOS_CRITICAL})
}

// Exit with an WARNING status and appropriate message
func Warning(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_WARNING})
}

// Exit with an OK status and appropriate message
func Ok(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_OK})
}

// Exit with a particular NagiosStatus
func ExitWithStatus(status *NagiosStatus) {
	fmt.Fprintln(os.Stdout, valMessages[status.Value], status.Message)
	os.Exit(int(status.Value))
}
