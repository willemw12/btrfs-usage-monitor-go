package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// BtrfsUsage prints a warning if Btrfs filesystem data usage drops below the free limit percentage.
// Path is the Btrfs filesystem location.
func BtrfsUsage(config *Config, path string, freeLimitPercentage uint64) error {
	cmdOutUsageRaw, err := usageRaw(config, path)
	if err != nil {
		return err
	}
	cmdOutUsageHuman, err := usageHuman(config, path)
	if err != nil {
		return err
	}

	usg := usage{}
	if err := usg.extractUsage(cmdOutUsageRaw, cmdOutUsageHuman); err != nil {
		return err
	}
	fmt.Printf("%s", usg.usageWarning(path, freeLimitPercentage))

	return nil
}

//

type usage struct {
	// Raw Btrfs filesystem data usage in bytes.
	// freeMin uint64
	deviceSize, free uint64

	// Human readable Btrfs filesystem data usage (e.g., 1K 234M 2G).
	// deviceSizeStr string
	freeStr, freeMinStr string
}

// Returns raw Btrfs filesystem data usage in bytes.
func usageRaw(config *Config, path string) ([]byte, error) {
	command := "btrfs"
	args := []string{"filesystem", "usage", "--raw", path}
	if config.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: executing: %s %s\n", command, args)
	}
	cmd := exec.Command(command, args...)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	out, err := cmd.Output()
	stderrStr := stderrBuf.String()
	if !strings.Contains(stderrStr, "run as root") {
		fmt.Fprint(os.Stderr, stderrStr)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, strings.Join(cmd.Args, " "))
		return nil, err
	}

	return out, nil
}

// Returns human-readable filesystem Btrfs data usage.
func usageHuman(config *Config, path string) ([]byte, error) {
	command := "btrfs"
	args := []string{"filesystem", "usage", path}
	if config.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: executing: %s %s\n", command, args)
	}
	cmd := exec.Command(command, args...)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	out, err := cmd.Output()
	stderrStr := stderrBuf.String()
	if !strings.Contains(stderrStr, "run as root") {
		fmt.Fprint(os.Stderr, stderrStr)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, strings.Join(cmd.Args, " "))
		return nil, err
	}

	return out, nil
}

// Collects Btrfs filesystem filesystem data usage.
func (usg *usage) extractUsage(cmdOutUsageRaw []byte, cmdOutUsageHuman []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(cmdOutUsageRaw))
	for scanner.Scan() {
		line := scanner.Text()

		// var value int
		// n, err := fmt.Sscanf(line, "    Device size:%d", &value)
		// if n == 1 && err == nil {
		// 	usg.deviceSize = value
		// }
		//
		// n, err = fmt.Sscanf(line, "    Free (estimated):%d", &value)
		// if n == 1 && err == nil {
		// 	usg.free = value
		// }

		// Input example: Device size:                      370643304448
		re, err := regexp.Compile(`Device size:(.*)`)
		if err != nil {
			return err
		}
		match := re.FindStringSubmatch(line)
		if len(match) == 2 {
			// s, err := strconv.Atoi(strings.TrimSpace(match[1]))
			s, err := strconv.ParseUint(strings.TrimSpace(match[1]), uintBase, uintBitSize)
			if err != nil {
				return err
			}
			// if s == 0 {
			// 	// panic("btrfs: device size is 0")
			// 	return errors.New("btrfs: device size is 0")
			// }

			usg.deviceSize = s
		}

		// Input example: Free (estimated):                 221948088320      (min: 111016611840)
		// re, err = regexp.Compile(`Free \(estimated\):(.*)\(`)
		re, err = regexp.Compile(`Free \(estimated\):(.*)\(min:(.*)\)`)
		if err != nil {
			return err
		}
		match = re.FindStringSubmatch(line)
		if len(match) == 3 {
			// if s, err := strconv.Atoi(strings.TrimSpace(match[1])); err == nil {
			if s, err := strconv.ParseUint(strings.TrimSpace(match[1]), uintBase, uintBitSize); err == nil {
				usg.free = s
			}
			// // if s, err := strconv.Atoi(strings.TrimSpace(match[2])); err == nil {
			// if s, err := strconv.Atoi(strings.TrimSpace(match[2])); err == nil {
			// 	usg.freeMin = s
			// }
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	//

	scanner = bufio.NewScanner(bytes.NewReader(cmdOutUsageHuman))
	for scanner.Scan() {
		line := scanner.Text()

		// if strings.HasPrefix(strings.TrimLeft(line, " \t"), "Device size:") {
		// if strings.HasPrefix(strings.TrimLeft(line, " \t"), "Free (estimated):") {

		// // Input example: Device size:                 345.19GiB
		// re, _ := regexp.Compile(`Device size:(.*)`)
		// match := re.FindStringSubmatch(line)
		// if len(match) == 2 {
		// 	usg.deviceSizeStr = strings.TrimSpace(match[1])
		// }

		// Input example: Free (estimated):            206.71GiB      (min: 103.39GiB)
		// re, _ := regexp.Compile(`Free \(estimated\):(.*)\(`)
		re, _ := regexp.Compile(`Free \(estimated\):(.*)\(min:(.*)\)`)
		match := re.FindStringSubmatch(line)
		if len(match) == 3 {
			usg.freeStr = strings.TrimSpace(match[1])
			usg.freeMinStr = strings.TrimSpace(match[2])
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// if usg.deviceSizeStr == "" || usg.freeStr == "" || usg.freeMinStr == "" {
	if usg.freeStr == "" || usg.freeMinStr == "" {
		// panic("parse error in filesystem data usage: at 'Free (estimated)'")
		return errors.New("parse error in filesystem data usage: at 'Free (estimated)'")
	}

	return nil
}

// Returns a warning if Btrfs filesystem data usage drops below the free limit percentage.
// func (u *usage) ...
func (usg usage) usageWarning(path string, freeLimitPercentage uint64) string {
	// if debug {
	// 	fmt.Fprintf(os.Stderr, "Device size: %s (%d)\n", usg.deviceSizeStr, usg.deviceSize)
	// 	fmt.Fprintf(os.Stderr, "Free : %s (%d)\n", usg.freeStr, usg.free)
	// 	fmt.Fprintf(os.Stderr, "Free min: %s (%d)\n", usg.freeMinStr, usg.freeMin)
	// }

	if usg.deviceSize == 0 {
		// // return ""
		// panic("runtime error: integer divide by zero")
		return fmt.Sprintf("ERROR: %s: device size is 0\n", path)
	}

	if freeLimitPercentage < 0 {
		freeLimitPercentage = 0
	} else if freeLimitPercentage > 100 {
		freeLimitPercentage = 100
	}

	freePercentage := (usg.free * 100) / usg.deviceSize
	if freePercentage < freeLimitPercentage {
		// return fmt.Sprintf("WARNING: %s, %d (min: %d)\n", path, usg.free, usg.freeMin)
		return fmt.Sprintf("WARNING: %s, free: %s (min: %s), %d%% (limit: %d%%)\n",
			path, usg.freeStr, usg.freeMinStr, freePercentage, freeLimitPercentage)
	}

	return ""
}
