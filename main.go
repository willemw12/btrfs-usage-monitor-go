// Print a warning if Btrfs data usage drops below the free limit percentage.
//
// Usage:
//
//     # btrfs-usage-monitor /mnt/btrfs 10
//     WARNING /mnt/btrfs free: 752.58GiB (min: 681.47GiB), 9% (limit: 10%)
package main

import (
	"errors"
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
	if err := handleCmd(os.Args[1:]); err != nil {
		// log.Fatal(err)
		fmt.Fprintf(os.Stderr, "%v\n", err)
		printUsage()
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(exitErrorCode)
	}
}

// func handleCmd(args []string) (err error) {
func handleCmd(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return nil
	}
	// if (len(args) > 0 && args[0][:1] == "-") ||
	if args[0][:1] == "-" ||
		(len(args) > 1 && args[1][:1] == "-") ||
		len(args) != 2 {
		return errors.New("error in arguments")
	}

	path := args[0]
	// freeLimitPercentage, err := strconv.Atoi(args[1])
	freeLimitPercentage, err := strconv.ParseUint(args[1], uintBase, uintBitSize)
	if err != nil {
		return err
	}

	if err := BtrfsUsage(path, freeLimitPercentage); err != nil {
		return err
	}

	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Print a warning if data usage exceeds the limit\n")
	fmt.Fprintf(os.Stderr, "usage: %s [-h|--help] [PATH] [FREE_LIMIT_PERCENTAGE]\n", program)
}
