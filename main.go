/*
Copyright (c) 2022 Thomas Cassaert

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/process"
)

// Proc struct
type Proc struct {
	Name      string
	OpenFiles float64
	Pid       int32
}

func main() {

	// Parse flags
	var critical, warning float64
	flag.Float64Var(&critical, "critical", 0.9, "critical treshold")
	flag.Float64Var(&warning, "warning", 0.8, "warning treshold")
	flag.Parse()

	checkOpenFileDescriptors(critical, warning)
}

func checkOpenFileDescriptors(critical, warning float64) {
	// Get all processes running on the system
	processes, _ := process.Processes()
	// Get the max open files per process
	maxOpenFiles := getMaxOpenFiles()
	// Set tresholds based on provided or default critical/warning values
	criticalValue := maxOpenFiles * critical
	warningValue := maxOpenFiles * warning

	// Slices to store processes that exceed tresholds
	var criticals []Proc
	var warnings []Proc
	// Loop over all processes and store the processes that exceed the tresholds in the correct slice
	for _, proc := range processes {
		pid := proc.Pid
		openFiles, _ := proc.OpenFiles()
		openFilesCount := float64(len(openFiles))
		name, _ := proc.Name()
		procStruct := Proc{Name: name, OpenFiles: openFilesCount, Pid: pid}
		if openFilesCount > criticalValue {
			criticals = append(criticals, procStruct)
		} else if openFilesCount > warningValue {
			warnings = append(warnings, procStruct)
		}
	}

	// Check if there are any critical processes
	if len(criticals) > 0 {
		var pids strings.Builder
		// If there is more than 1 critical process, print all IDs
		if len(criticals) > 1 {
			for _, critical := range criticals {
				pids.WriteString(fmt.Sprintf("%s(PID: %d), ", critical.Name, critical.Pid))
			}
			trimmedString := strings.TrimSuffix(pids.String(), ", ")
			msg := fmt.Sprintf("Processes %s have too many open file descriptors.\n", trimmedString)
			setCritical(msg)
		} else {
			// If there's only one process critical, give some more info on the process
			for _, critical := range criticals {
				msg := fmt.Sprintf("Proccess %s with PID %d uses %d/%d open file descriptors. | %s=%d;;;;%d\n", critical.Name, critical.Pid, int(critical.OpenFiles), int(maxOpenFiles), critical.Name, int(critical.OpenFiles), int(maxOpenFiles))
				setCritical(msg)
			}
		}
		// Check if there are any processes in warning
	} else if len(warnings) > 0 {
		var pids strings.Builder
		// If there is more than 1 process in warning, print all IDs
		if len(warnings) > 1 {
			for _, warning := range warnings {
				pids.WriteString(fmt.Sprintf("%s(PID: %d), ", warning.Name, warning.Pid))
			}
			trimmedString := strings.TrimSuffix(pids.String(), ", ")
			msg := fmt.Sprintf("Processes %s have too many open file descriptors.\n", trimmedString)
			setWarning(msg)
		} else {
			// If there's only one process in warning, give some more info on the process
			for _, warning := range warnings {
				msg := fmt.Sprintf("Proccess %s with PID %d uses %d/%d open file descriptors. | %s=%d;;;;%d\n", warning.Name, warning.Pid, int(warning.OpenFiles), int(maxOpenFiles), warning.Name, int(warning.OpenFiles), int(maxOpenFiles))
				setWarning(msg)
			}
		}
		// If all is fine, print OK message
	} else {
		msg := fmt.Sprintf("All processes' open file descriptors are below the maximum of %d.", int(maxOpenFiles))
		setOk(msg)
	}
}

// Mark check as critical
func setCritical(msg string) {
	fmt.Printf("CRITICAL: %s", msg)
	os.Exit(2)
}

// Mark check as in warning
func setWarning(msg string) {
	fmt.Printf("WARNING: %s", msg)
	os.Exit(1)
}

// Mark check as OK
func setOk(msg string) {
	fmt.Printf("OK: %s", msg)
	os.Exit(0)
}

// Get the systems max open file descriptors
func getMaxOpenFiles() float64 {
	var maxOpenFiles syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &maxOpenFiles)
	if err != nil {
		log.Fatal(err)
	}
	softLimit := maxOpenFiles.Cur
	return float64(softLimit)
}
