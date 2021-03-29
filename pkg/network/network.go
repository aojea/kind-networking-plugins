package network

import (
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
