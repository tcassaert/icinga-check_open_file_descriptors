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
	"sort"
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
	var processName string
	flag.Float64Var(&critical, "critical", 0.9, "critical treshold")
	flag.Float64Var(&warning, "warning", 0.8, "warning treshold")
	flag.StringVar(&processName, "process", "", "Name of the process to watch")
	flag.Parse()

	checkOpenFileDescriptors(critical, warning, processName)
}

func checkOpenFileDescriptors(critical, warning float64, processName string) {
	// Get all processes running on the system
	processes, _ := process.Processes()
	// Get the max open files per process
	maxOpenFiles := getMaxOpenFiles()
	// Set tresholds based on provided or default critical/warning values
	criticalValue := maxOpenFiles * critical
	warningValue := maxOpenFiles * warning

	// Slice to store criticals
	var criticals []Proc
	// Slice to store the all the Open File Descriptors of processes matching the -process parameter
	var openFilesSpecificProcess []float64
	// Slice to store all the processes matching the -process parameter
	var specificProcesses []Proc
	// Slice to store warnings
	var warnings []Proc

	/*
		   Range over all processes.
			 Check if their name equals the one in the -process parameter.
			 If it does, add it to the specificProcesses slice
	*/
	for _, proc := range processes {
		procName, _ := proc.Name()
		if procName == processName {
			fmt.Println(procName)
			pid := proc.Pid
			openFiles, _ := proc.OpenFiles()
			openFilesCount := float64(len(openFiles))
			name, _ := proc.Name()
			process := Proc{Name: name, OpenFiles: openFilesCount, Pid: pid}
			specificProcesses = append(specificProcesses, process)
		}
	}

	/*
		    Range over the specificProcesses slice
				Add the number of Open File Descriptors of that process to the openFilesSpecificProcess slice
	*/
	for i, _ := range specificProcesses {
		nrOpenFiles := specificProcesses[i].OpenFiles
		openFilesSpecificProcess = append(openFilesSpecificProcess, nrOpenFiles)
		sort.Float64s(openFilesSpecificProcess)
	}

	/*
		   Range over the specificProcesses slice
			 If number of Open File Descriptors for the specific process matches the biggest number in the openFilesSpecificProcess slice,
			 evaluate it's value and exit appropriately
	*/
	for _, specificProc := range specificProcesses {
		openFilesSpecificProc := openFilesSpecificProcess[len(openFilesSpecificProcess)-1]
		if specificProc.OpenFiles == openFilesSpecificProc {
			if specificProc.OpenFiles > criticalValue {
				msg := fmt.Sprintf("Proccess %s with PID %d uses %d/%d open file descriptors. | %s=%d;;;;%d\n", specificProc.Name, specificProc.Pid, int(specificProc.OpenFiles), int(maxOpenFiles), specificProc.Name, int(specificProc.OpenFiles), int(maxOpenFiles))
				setCritical(msg)
			} else if specificProc.OpenFiles > warningValue {
				msg := fmt.Sprintf("Proccess %s with PID %d uses %d/%d open file descriptors. | %s=%d;;;;%d\n", specificProc.Name, specificProc.Pid, int(specificProc.OpenFiles), int(maxOpenFiles), specificProc.Name, int(specificProc.OpenFiles), int(maxOpenFiles))
				setWarning(msg)
			} else {
				msg := fmt.Sprintf("Proccess %s with PID %d uses %d/%d open file descriptors. | %s=%d;;;;%d\n", specificProc.Name, specificProc.Pid, int(specificProc.OpenFiles), int(maxOpenFiles), specificProc.Name, int(specificProc.OpenFiles), int(maxOpenFiles))
				setOk(msg)
			}
		}
	}

	/*
		    Only getting here if there's no -process parameter passed
				Range over processes and add them to the correct slice (warnings or criticals)
	*/
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
