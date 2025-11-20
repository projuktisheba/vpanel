package run

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// RunBashScript executes a bash script with the provided arguments.
// It captures Stderr separately to return meaningful error messages.
func RunBashScript(scriptPath string, args []string) error {
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script file not found: %s", scriptPath)
	}

	// Prepare the command: bash script.sh arg1 arg2 ...
	// We prepend the scriptPath to the args slice for the command construction
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("bash", cmdArgs...)

	// Capture Stdout and Stderr separately
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("Executing: %s %v\n", scriptPath, args)

	// Run the command
	err := cmd.Run()

	// Print standard output regardless of error (useful for logs)
	if stdout.Len() > 0 {
		fmt.Println("--- Script Output ---")
		fmt.Println(stdout.String())
	}

	// Handle errors
	if err != nil {
		// If the script wrote to stderr, include that in the error message
		if stderr.Len() > 0 {
			return fmt.Errorf("script failed: %v\nError Output:\n%s", err, stderr.String())
		}
		// Fallback if no stderr output but exit code was non-zero
		return fmt.Errorf("script failed with %v", err)
	}

	return nil
}

func main() {
	// --- Usage Example ---

	// 1. Create a dummy script for demonstration
	scriptContent := `#!/bin/bash
echo "Processing data..."
if [ "$1" == "fail" ]; then
    echo "Critical Error: Invalid argument provided!" >&2
    exit 1
fi
echo "Success: Processed argument '$1'"
`
	scriptName := "test_script.sh"
	os.WriteFile(scriptName, []byte(scriptContent), 0755)
	defer os.Remove(scriptName) // Cleanup after running

	// 2. Run with valid arguments
	fmt.Println(">>> Test 1: Valid Run")
	err := RunBashScript(scriptName, []string{"hello"})
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// 3. Run with invalid arguments (triggering error)
	fmt.Println("\n>>> Test 2: Error Run")
	err = RunBashScript(scriptName, []string{"fail"})
	if err != nil {
		log.Printf("Error Caught: %v", err)
	}
}
