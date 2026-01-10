package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Check if file exists and prompt user for file deletion
// Return true if file does not exist or if file was deleted.
// Return false if file found but did not want to delete
func PromptFileDeletion(file string) bool {
	if PathExists(file) {
		var s string
		r := bufio.NewReader(os.Stdin)
		fmt.Printf("File %s already exists\nDo you want to remove existing file? [y/N] ", file)
		s, _ = r.ReadString('\n')
		answer := strings.TrimSpace(strings.ToLower(s))
		fmt.Println(strings.ToLower(answer))
		if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
			err := os.Remove(file)
			if err != nil {
				fmt.Printf("Error deleting file %s\n", file)
			}
			return true
		} else {
			return false
		}
	}
	return true
}
