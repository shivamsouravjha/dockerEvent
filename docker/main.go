package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

var cancel context.CancelFunc

func main() {
	username := os.Getenv("SUDO_USER")
	fmt.Println("username ", username)
	ctx := NewCtx()
	keployAlias := "docker-compose -f /app/docker-compose.yml up"

	cmd := exec.CommandContext(ctx, "sh", "-c", keployAlias)
	// if username != "" {
	// 	// print all environment variables
	// 	// Run the command as the user who invoked sudo to preserve the user environment variables and PATH
	// 	cmd = exec.CommandContext(ctx, "sudo", "-E", "-u", os.Getenv("SUDO_USER"), "env", "PATH="+os.Getenv("PATH"), "sh", "-c", keployAlias)
	// }

	// Set the cancel function for the command
	cmd.Cancel = func() error {

		return InterruptProcessTree(cmd, cmd.Process.Pid, syscall.SIGTERM)
	}
	// wait after sending the interrupt signal, before sending the kill signal
	cmd.WaitDelay = 100 * time.Second

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Set the output of the command
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		fmt.Println("errro in start", err.Error())
	}
	fmt.Println("the pid", cmd.Process.Pid)
	go func() {
		time.Sleep(10 * time.Second)
		fmt.Println("the chils pids" + fmt.Sprint(FindChildPIDs(cmd.Process.Pid)))
	}()
	err = cmd.Wait()
	if err != nil {
		fmt.Println("errror in wait", err.Error())
	}
}

// InterruptProcessTree interrupts an entire process tree using the given signal
func InterruptProcessTree(cmd *exec.Cmd, ppid int, sig syscall.Signal) error {
	// Find all descendant PIDs of the given PID & then signal them.
	// Any shell doesn't signal its children when it receives a signal.
	// Children may have their own process groups, so we need to signal them separately.
	children, err := FindChildPIDs(ppid)
	if err != nil {
		return err
	}

	children = append(children, ppid)

	sort.Slice(children, func(i, j int) bool { return children[i] > children[j] })

	for _, pid := range children {
		if cmd.ProcessState == nil {
			err := syscall.Kill(pid, sig)
			if err != nil {
				fmt.Println("failed to send signal to process", zap.Int("pid", pid), zap.Error(err))
			}
			// time.Sleep(250 * time.Millisecond)
		}
	}

	return nil
}

// findChildPIDs takes a parent PID and returns a slice of all descendant PIDs.
func FindChildPIDs(parentPID int) ([]int, error) {
	var childPIDs []int

	// Recursive helper function to find all descendants of a given PID.
	var findDescendants func(int)
	findDescendants = func(pid int) {
		procDirs, err := os.ReadDir("/proc")
		if err != nil {
			return
		}

		for _, procDir := range procDirs {
			if !procDir.IsDir() {
				continue
			}

			childPid, err := strconv.Atoi(procDir.Name())
			if err != nil {
				continue
			}

			statusPath := filepath.Join("/proc", procDir.Name(), "status")
			statusBytes, err := os.ReadFile(statusPath)
			if err != nil {
				continue
			}

			status := string(statusBytes)
			for _, line := range strings.Split(status, "\n") {
				if strings.HasPrefix(line, "PPid:") {
					fields := strings.Fields(line)
					if len(fields) == 2 {
						ppid, err := strconv.Atoi(fields[1])
						if err != nil {
							break
						}
						if ppid == pid {
							childPIDs = append(childPIDs, childPid)
							findDescendants(childPid)
						}
					}
					break
				}
			}
		}
	}

	// Start the recursion with the initial parent PID.
	findDescendants(parentPID)

	return childPIDs, nil
}

func NewCtx() context.Context {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())

	SetCancel(cancel)
	// Set up a channel to listen for signals
	sigs := make(chan os.Signal, 1)
	// os.Interrupt is more portable than syscall.SIGINT
	// there is no equivalent for syscall.SIGTERM in os.Signal
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine that will cancel the context when a signal is received
	go func() {
		<-sigs
		fmt.Println("Signal received, canceling context...")
		cancel()
	}()

	return ctx
}

func SetCancel(c context.CancelFunc) {
	cancel = c
}
