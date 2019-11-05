package steamclient

import (
	"flag"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	steamDir = flag.String("steam-dir", "~/.steam/", "Path to steam installation")
)

func IsRunning() bool {
	pidFile, err := homedir.Expand(filepath.Join(*steamDir, "steam.pid"))
	if err != nil {
		log.Fatalf("Error expanding home dir: %s", err.Error())
		return false
	}

	pidContent, err := ioutil.ReadFile(pidFile)
	if err != nil {
		log.Fatalf("Error reading PID file, assuming steam is not running: %s", err.Error())
		return false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidContent)))
	if err != nil {
		log.Fatalf("Error converting PID, content was '%s': %s", string(pidContent), err.Error())
		return false
	}

	_, err = os.Stat(fmt.Sprintf("/proc/%d/stat", pid))
	return !os.IsNotExist(err)
}

func HasCompatibilityTool(name string) bool {
	file, err := homedir.Expand(filepath.Join(*steamDir, "compatibilitytools.d", name))
	if err != nil {
		log.Fatalf("Error expanding home dir: %s", err.Error())
		return false
	}

	_, err = os.Stat(file)
	return err == nil || os.IsExist(err)
}

func CompatibilityToolPath() string {
	dir, err := homedir.Expand(filepath.Join(*steamDir, "compatibilitytools.d"))
	if err != nil {
		log.Fatalf("Error expanding home dir: %s", err.Error())
	}

	return dir
}

func CreateCompatibilityToolFile(name string, mode int64) (*os.File, error) {
	file, err := homedir.Expand(filepath.Join(*steamDir, "compatibilitytools.d", name))
	if err != nil {
		log.Fatalf("Error expanding home dir: %s", err.Error())
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return nil, err
	}

	return os.OpenFile(file, os.O_CREATE|os.O_RDWR, os.FileMode(mode))
}

func Shutdown() {
	file, err := homedir.Expand(filepath.Join(*steamDir, "steam.sh"))
	if err != nil {
		log.Fatalf("Error expanding home dir: %s", err.Error())
	}

	if err := exec.Command(file, "-shutdown").Run(); err != nil {
		log.Fatalf("Unable to shutdown steam: %s", err.Error())
	}

	time.Sleep(time.Second)
}
