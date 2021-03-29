package docker

import (
	"fmt"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"sigs.k8s.io/kind/pkg/exec"
)

// CreateNetwork create a docker network with the passed parameters
func CreateNetwork(name, ipv6Subnet string, mtu int, masquerade bool) error {
	args := []string{"network", "create", "-d=bridge"}

	// set the interface name, if not set it defaults to "br-" + id[:12]
	args = append(args, "-o", fmt.Sprintf("com.docker.network.bridge.name=%s", "br-"+name[:12]))
	// enable docker iptables rules to masquerade network traffic
	args = append(args, "-o", fmt.Sprintf("com.docker.network.bridge.enable_ip_masquerade=%t", masquerade))

	if mtu > 0 {
		args = append(args, "-o", fmt.Sprintf("com.docker.network.driver.mtu=%d", mtu))
	}
	if ipv6Subnet != "" {
		args = append(args, "--ipv6", "--subnet", ipv6Subnet)
	}
	args = append(args, name)
	return exec.Command("docker", args...).Run()
}

// DeleteNetwork delete a docker network
func DeleteNetwork(name string) error {
	return exec.Command("docker", "network", "rm", name).Run()
}

// GetContainerHostIfacesIndex returns the interfaces name on the host of the container
func GetContainerHostIfacesIndex(name string) ([]string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ifacesIdx := []int{}
	ifaces := []string{}

	// Save the current network namespace
	origns, err := netns.Get()
	if err != nil {
		return nil, err
	}
	defer origns.Close()

	// get docker namespace
	ns, err := netns.GetFromDocker(name)
	if err != nil {
		return nil, err
	}
	defer ns.Close()

	// Swith to the docker namespace to get the container interfaces
	if err := netns.Set(ns); err != nil {
		return nil, err
	}
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}
	// we need to obtain the peer id
	// https://unix.stackexchange.com/questions/441876/how-to-find-the-network-namespace-of-a-veth-peer-ifindex
	for _, l := range links {
		// continue if is not a veth interface
		veth, ok := l.(*netlink.Veth)
		if !ok {
			continue
		}
		// l.Attrs().ParentIndex returns the veth peer index too
		// I don't know which method is better to get the peer index
		index, _ := netlink.VethPeerIndex(veth)
		ifacesIdx = append(ifacesIdx, index)
	}

	// Switch back to the original namespace to get the interface name
	netns.Set(origns)
	for _, idx := range ifacesIdx {
		ifName, err := netlink.LinkByIndex(idx)
		if err != nil {
			return nil, err
		}
		ifaces = append(ifaces, ifName.Attrs().Name)

	}
	return ifaces, nil
}

func ListNetwork() ([]string, error) {
	cmd := exec.Command("docker", "network", "list",
		"--format", `{{ .Name }}`)
	return exec.OutputLines(cmd)
}
