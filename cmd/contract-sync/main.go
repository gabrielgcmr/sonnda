// cmd/contract-sync/main.go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type contractLock struct {
	source  string
	asset   string
	version string
	sha256  string
}

func main() {
	lockPath := flag.String("lock", "contracts.lock", "contract lock file")
	outputPath := flag.String("output", "dist/openapi.yaml", "download destination")
	verifyOnly := flag.Bool("verify", false, "verify the existing bundle without downloading")
	version := flag.String("version", "", "override locked version; requires -sha256")
	checksum := flag.String("sha256", "", "override locked SHA-256; requires -version")
	flag.Parse()

	lock, err := readLock(*lockPath)
	if err == nil {
		err = lock.override(*version, *checksum)
	}
	if err != nil {
		fail(err)
	}

	if *verifyOnly {
		data, readErr := os.ReadFile(*outputPath)
		if readErr != nil {
			fail(fmt.Errorf("read cached contract: %w", readErr))
		}
		if err := verifyChecksum(data, lock.sha256); err != nil {
			fail(err)
		}
		fmt.Printf("contract verified: %s@%s\n", lock.asset, lock.version)
		return
	}

	response, err := http.Get(lock.downloadURL())
	if err != nil {
		fail(fmt.Errorf("download contract: %w", err))
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fail(fmt.Errorf("download contract: %s", response.Status))
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		fail(fmt.Errorf("read contract download: %w", err))
	}
	if err := verifyChecksum(data, lock.sha256); err != nil {
		fail(err)
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(fmt.Errorf("create contract cache directory: %w", err))
	}
	if err := os.WriteFile(*outputPath, data, 0o644); err != nil {
		fail(fmt.Errorf("write contract cache: %w", err))
	}
	fmt.Printf("contract synced: %s@%s\n", lock.asset, lock.version)
}

func readLock(path string) (contractLock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contractLock{}, fmt.Errorf("read contract lock: %w", err)
	}
	values := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			return contractLock{}, fmt.Errorf("invalid contract lock line %q", line)
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	lock := contractLock{source: values["source"], asset: values["asset"], version: values["version"], sha256: values["sha256"]}
	if lock.source == "" || lock.asset == "" || lock.version == "" || lock.sha256 == "" {
		return contractLock{}, errors.New("contract lock requires source, asset, version and sha256")
	}
	return lock, nil
}

func (l *contractLock) override(version, checksum string) error {
	if (version == "") != (checksum == "") {
		return errors.New("-version and -sha256 must be provided together")
	}
	if version != "" {
		l.version = version
		l.sha256 = checksum
	}
	return nil
}

func (l contractLock) downloadURL() string {
	return strings.TrimRight(l.source, "/") + "/" + l.version + "/" + l.asset
}

func verifyChecksum(data []byte, expected string) error {
	actual := sha256.Sum256(data)
	if strings.EqualFold(hex.EncodeToString(actual[:]), strings.TrimSpace(expected)) {
		return nil
	}
	return fmt.Errorf("contract checksum mismatch: got %x", actual)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
