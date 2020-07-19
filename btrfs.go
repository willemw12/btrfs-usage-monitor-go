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

// BtrfsUsage prints a warning if Btrfs data usage drops below the free limit percentage.
func BtrfsUsage(path string, freeLimitPercentage uint64) error {
	cmdOutUsageRaw, err := runBtrfsUsageRaw(path)
	if err != nil {
		return err
	}
	cmdOutUsageHuman, err := runBtrfsUsageHuman(path)
	if err != nil {
		return err
	}

	usg := usage{}
	if err := usg.extractBtrfsUsageData(cmdOutUsageRaw, cmdOutUsageHuman); err != nil {
		return err
	}
	fmt.Printf("%s", usg.getUsageWarning(path, freeLimitPercentage))

	return nil
}

//

type usage struct {
	// Raw Btrfs data usage in bytes.
	//freeMin uint64
	deviceSize, free uint64

	// Human readable Btrfs data usage (e.g., 1K 234M 2G).
	// deviceSizeStr string
	freeStr, freeMinStr string
}

// Returns raw Btrfs data usage in bytes.
func runBtrfsUsageRaw(path string) ([]byte, error) {
	cmd := exec.Command("btrfs", "filesystem", "usage", "--raw", path)
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

// Returns human-readable Btrfs data usage.
func runBtrfsUsageHuman(path string) ([]byte, error) {
	cmd := exec.Command("btrfs", "filesystem", "usage", path)
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

// Collects and stores Btrfs usage data.
func (usg *usage) extractBtrfsUsageData(cmdOutUsageRaw []byte, cmdOutUsageHuman []byte) error {
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

	// if usg.deviceSize == 0 || usg.free == 0 || usg.freeMin == 0 {
	if usg.deviceSize == 0 || usg.free == 0 {
		// panic("parse error")
		return errors.New("usage data parse error")
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
		// panic("parse error")
		return errors.New("usage data parse error")
	}

	return nil
}

// Returns a warning if Btrfs data usage drops below the free limit percentage.
// func (u *usage) ...
func (usg usage) getUsageWarning(path string, freeLimitPercentage uint64) string {
	// if debug {
	// 	fmt.Fprintf(os.Stderr, "Device size: %s (%d)\n", usg.deviceSizeStr, usg.deviceSize)
	// 	fmt.Fprintf(os.Stderr, "Free : %s (%d)\n", usg.freeStr, usg.free)
	// 	fmt.Fprintf(os.Stderr, "Free min: %s (%d)\n", usg.freeMinStr, usg.freeMin)
	// }

	if usg.deviceSize == 0 {
		// return ""
		panic("runtime error: integer divide by zero")
	}

	freePercentage := (usg.free * 100) / usg.deviceSize
	if freePercentage < freeLimitPercentage {
		// return fmt.Sprintf("WARNING %s: %d (min: %d)\n", path, usg.free, usg.freeMin)
		return fmt.Sprintf("WARNING %s free: %s (min: %s), %d%% (limit: %d%%)\n",
			path, usg.freeStr, usg.freeMinStr, freePercentage, freeLimitPercentage)
	}

	return ""
}
