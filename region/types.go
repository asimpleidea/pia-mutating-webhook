package main

// TODO: groups

type Region struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Country     string    `json:"country" yaml:"country"`
	AutoRegion  bool      `json:"auto_region" yaml:"autoRegion"`
	DNS         string    `json:"dns" yaml:"dns"`
	PortForward bool      `json:"port_forward" yaml:"portForward"`
	Geo         bool      `json:"geo" yaml:"geo"`
	Offline     bool      `json:"offline" yaml:"offline"`
	Servers     []*Server `json:"servers" yaml:"servers"`
}

type ServersList struct {
	IkeV2      []*Server `json:"ikev2,omitempty" yaml:"ikev2,omitempty"`
	Meta       []*Server `json:"meta,omitempty" yaml:"meta,omitempty"`
	OpenVPNTCP []*Server `json:"ovpntcp,omitempty" yaml:"ovpntcp,omitempty"`
	OpenVPNUDP []*Server `json:"ovpnudp,omitempty" yaml:"ovpnudp,omitempty"`
	WireGuard  []*Server `json:"wg,omitempty" yaml:"wg,omitempty"`
}

type Server struct {
	IP  string `json:"ip" yaml:"ip"`
	CN  string `json:"cn" yaml:"cn"`
	VAN string `json:"van" yaml:"van,omitempty"`
}
