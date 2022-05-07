// Print a warning if the Btrfs filesystem data usage drops below a free limit percentage.
//
// Usage:
//
//     # btrfs-usage-monitor /mnt/btrfs 10
//     WARNING /mnt/btrfs free: 752.58GiB (min: 681.47GiB), 9% (limit: 10%)
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const exitErrorCode = 1

const uintBase = 10
const uintBitSize = 64

var (
	program = filepath.Base(os.Args[0])
)

func main() {
	if err := handleCmd(); err != nil {
		// log.Fatal(err)
		fmt.Fprintf(os.Stderr, "%v\n", err)
		flag.Usage()
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(exitErrorCode)
	}
}

// func handleCmd() (err error) {
func handleCmd() error {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Print a warning if Btrfs filesystem data usage drops below the free limit percentage")
		fmt.Fprintf(os.Stderr, "\nUsage of %s: [FLAGS] [PATH] [FREE_LIMIT_PERCENTAGE]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nArguments:")
		fmt.Fprintln(os.Stderr, "  PATH\n    \tpath to a subvolume or folder on the Btrfs filesystem")
		fmt.Fprintln(os.Stderr, "  FREE_LIMIT_PERCENTAGE\n    \tmaximum free data usage in percentage")
	}

	debugFlag := flag.Bool("debug", false, "print debug information")
	helpFlag := flag.Bool("help", false, "show this help message and exit")
	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return nil
	}

	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		return errors.New("error in arguments")
	}

	config := &Config{Debug: *debugFlag}
	path := flag.Args()[0]
	// freeLimitPercentage, err := strconv.Atoi(args[1])
	freeLimitPercentage, err := strconv.ParseUint(args[1], uintBase, uintBitSize)
	if err != nil {
		return err
	}

	if err := BtrfsUsage(config, path, freeLimitPercentage); err != nil {
		return err
	}

	return nil
}
