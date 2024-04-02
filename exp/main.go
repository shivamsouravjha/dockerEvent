package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	nativeDockerClient "github.com/docker/docker/client"
	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	eventFilter := filters.NewArgs(
		filters.KeyValuePair{Key: "type", Value: "container"},
		filters.KeyValuePair{Key: "type", Value: "network"},
		filters.KeyValuePair{Key: "action", Value: "create"},
		filters.KeyValuePair{Key: "action", Value: "connect"},
		filters.KeyValuePair{Key: "action", Value: "start"},
	)

	events, errors := cli.Events(context.Background(), types.EventsOptions{Filters: eventFilter})

	ctx := context.Background()
	d, _ := nativeDockerClient.NewClientWithOpts(nativeDockerClient.FromEnv,
		nativeDockerClient.WithAPIVersionNegotiation())
	for {
		select {
		case event := <-events:
			switch event.Action {
			case "create":
				// Fetch container details by inspecting using container ID to check if container is created
				info, err := d.ContainerInspect(ctx, event.ID)
				if err != nil {
					fmt.Println("unable to inspect container during create", err)
					continue
				}

				// Set Docker Container ID
				inode, err := getInode(info.State.Pid)
				if err != nil {
					fmt.Println("unable to extract inode during create", err)
					continue
				}
				fmt.Println("Inode during create: ", inode)

				n, ok := info.NetworkSettings.Networks["keploy-network"]
				if !ok || n == nil {
					fmt.Println("container network not found during create", zap.Any("containerDetails.NetworkSettings.Networks", info.NetworkSettings.Networks))
					continue
				}
				iPAddress := n.IPAddress
				fmt.Println("IP ADDRESS during create: ", iPAddress)

			case "start":
				// Fetch container details by inspecting using container ID to check if container is created
				info, err := d.ContainerInspect(ctx, event.ID)
				if err != nil {
					fmt.Println("unable to inspect container during start", err)
					continue
				}

				// Set Docker Container ID
				inode, err := getInode(info.State.Pid)
				if err != nil {
					fmt.Println("unable to extract inode during start", err)
					continue
				}
				fmt.Println("Inode during start: ", inode)

				n, ok := info.NetworkSettings.Networks["keploy-network"]
				if !ok || n == nil {
					fmt.Println("container network not found during start", zap.Any("containerDetails.NetworkSettings.Networks", info.NetworkSettings.Networks))
					continue
				}
				iPAddress := n.IPAddress
				fmt.Println("IP ADDRESS during start: ", iPAddress)
			case "connect":
				// Fetch container details by inspecting using container ID to check if container is created
				info, err := d.ContainerInspect(ctx, event.ID)
				if err != nil {
					fmt.Println("unable to inspect container during connect", err)
					continue
				}

				// Set Docker Container ID
				inode, err := getInode(info.State.Pid)
				if err != nil {
					fmt.Println("unable to extract inode during connect", err)
					continue
				}
				fmt.Println("Inode during connect: ", inode)

				n, ok := info.NetworkSettings.Networks["keploy-network"]
				if !ok || n == nil {
					fmt.Println("container network not found during connect", zap.Any("containerDetails.NetworkSettings.Networks", info.NetworkSettings.Networks))
					continue
				}
				iPAddress := n.IPAddress
				fmt.Println("IP ADDRESS during connect: ", iPAddress)
			}
		case err := <-errors:
			fmt.Printf("Error: %s\n", err)

		}
	}
}

func getInode(pid int) (uint64, error) {
	path := filepath.Join("/proc", strconv.Itoa(pid), "ns", "pid")

	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	// Dev := (f.Sys().(*syscall.Stat_t)).Dev
	i := (f.Sys().(*syscall.Stat_t)).Ino
	if i == 0 {
		return 0, fmt.Errorf("failed to get the inode of the process")
	}
	return i, nil
}
