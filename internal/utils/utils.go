package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func PathExists(path string) bool {
	normalizedPath := strings.TrimSpace(path)
	_, err := os.Stat(normalizedPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// Verify that dependency is available on users machine
func CheckCommandAvailable(command string, remedy string) error {
	cmd := exec.Command("which", command)
	response, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to verify %s: %s\n%s\n", command, err, remedy)
	} else if !PathExists(string(response)) {
		return fmt.Errorf("%s is not available. %s", command, remedy)
	}
	return nil
}

func RemoveNil(errArr []error) []error {
	var arr []error
	for _, e := range errArr {
		if e != nil {
			arr = append(arr, e)
		}
	}
	return arr
}

func OutputPath(isShow bool, show string, season int) string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err) // or return error if you prefer
	}

	base := filepath.Join(home, "Videos", show)

	if isShow {
		return filepath.Join(base, "Season_"+strconv.Itoa(season))
	}

	return base
}

func CreateDir(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
		} else {
			fmt.Println("Directory created:", dirPath)
		}
	}
}

func TerminateProcess(cmd *exec.Cmd, timeout time.Duration) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Step 1: Try graceful shutdown
	if runtime.GOOS == "windows" {
		_ = cmd.Process.Kill() // Windows has no SIGTERM equivalent
	} else {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}

	// Step 2: Wait or force kill
	select {
	case <-time.After(timeout):
		fmt.Println("Graceful shutdown timed out, killing process")
		_ = cmd.Process.Kill()
		<-done
	case <-done:
		fmt.Println("Process exited cleanly")
	}
}
