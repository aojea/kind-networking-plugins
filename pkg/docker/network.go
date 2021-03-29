package docker

import (
	"fmt"
	"net"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"sigs.k8s.io/kind/pkg/exec"
)

// CreateNetwork create a docker network with the passed parameters
func CreateNetwork(name, subnet string, masquerade bool) error {
	args := []string{"network", "create", "-d=bridge"}
	// set the interface name, if not set it defaults to "br-" + id[:12]
	args = append(args, "-o", fmt.Sprintf("com.docker.network.bridge.name=%s", "br-"+name[:12]))
	// enable docker iptables rules to masquerade network traffic
	args = append(args, "-o", fmt.Sprintf("com.docker.network.bridge.enable_ip_masquerade=%t", masquerade))
	// configure the subnet and the gateway provided
	args = append(args, "--subnet", subnet)
	// and only allocate ips for the containers for the first 32 ips /27
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	m := net.CIDRMask(27, 32)
	cidr.Mask = m
	args = append(args, "--ip-range", cidr.String(), name)
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
	pid, err := getContainerPid(name)
	if err != nil {
		return nil, err
	}
	ns, err := netns.GetFromPid(pid)
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

func ConnectNetwork(nameOrId, network, ip string) error {
	cmd := exec.Command("docker", "network", "connect",
		"--ip", ip, network, nameOrId)
	return cmd.Run()
}

func ReplaceGateway(name, gateway string) error {
	gw := net.ParseIP(gateway)
	// TODO: support IPv6
	if gw.To4() == nil {
		return fmt.Errorf("unsupported IP %s", gateway)
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	pid, err := getContainerPid(name)
	if err != nil {
		return err
	}
	ns, err := netns.GetFromPid(pid)
	if err != nil {
		return err
	}
	defer ns.Close()
	// Swith to the docker namespace to get the container interfaces
	if err := netns.Set(ns); err != nil {
		return err
	}

	defaultRoute := &netlink.Route{
		Dst: nil,
		Gw:  gw,
	}
	return netlink.RouteReplace(defaultRoute)
}

func getContainerId(name string) (string, error) {
	cmd := exec.Command("docker", "inspect",
		"--format", `{{ .Id }}`, name)
	lines, err := exec.OutputLines(cmd)
	if err != nil || len(lines) != 1 {
		return "", errors.Wrapf(err, "error trying to get container %s id", name)
	}
	return lines[0], nil
}

func getContainerPid(name string) (int, error) {
	cmd := exec.Command("docker", "inspect",
		"--format", `{{ .State.Pid }}`, name)
	lines, err := exec.OutputLines(cmd)
	if err != nil || len(lines) != 1 {
		return 0, errors.Wrapf(err, "error trying to get container %s id", name)
	}
	pid, err := strconv.Atoi(lines[0])
	if err != nil {
		return 0, err
	}
	return pid, nil
}
