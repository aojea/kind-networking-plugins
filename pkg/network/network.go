package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

func CreateBridge(name string) error {
	bridge := &netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: name}}
	return netlink.LinkAdd(bridge)
}

func CreateVeth(name1, name2 string) error {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: name1,
		},
		PeerName: name2,
	}
	return netlink.LinkAdd(veth)

}

func AddInterfaceBridge(ifaz, bridge string) error {
	ifLink, err := netlink.LinkByName(ifaz)
	if err != nil {
		return err
	}
	brLink, err := netlink.LinkByName(bridge)
	if err != nil {
		return err
	}
	return netlink.LinkSetMaster(ifLink, brLink)
}

func DeleteInterface(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	return netlink.LinkDel(link)
}

// GetLastIPSubnet obtains the last IP in the range
func GetLastIPSubnet(cidr string) (net.IP, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	ip := ipnet.IP
	mask := ipnet.Mask

	// get the broadcast address
	lastIP := net.IP(make([]byte, len(ip)))
	for i := range ip {
		lastIP[i] = ip[i] | ^mask[i]
	}
	// get the previous IP
	lastIP[len(ip)-1]--

	return lastIP, nil
}
