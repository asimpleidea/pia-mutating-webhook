package main

import "time"

// TODO: groups

type Region struct {
	ID          string       `json:"id" yaml:"id"`
	Name        string       `json:"name" yaml:"name"`
	Country     string       `json:"country" yaml:"country"`
	AutoRegion  bool         `json:"auto_region" yaml:"autoRegion"`
	DNS         string       `json:"dns" yaml:"dns"`
	PortForward bool         `json:"port_forward" yaml:"portForward"`
	Geo         bool         `json:"geo" yaml:"geo"`
	Offline     bool         `json:"offline" yaml:"offline"`
	Servers     *ServersList `json:"servers" yaml:"servers"`
}

func (r *Region) Clone() *Region {
	reg := &Region{
		ID:          r.ID,
		Name:        r.Name,
		Country:     r.Country,
		AutoRegion:  r.AutoRegion,
		DNS:         r.DNS,
		PortForward: r.PortForward,
		Geo:         r.Geo,
		Offline:     r.Offline,
		Servers:     &ServersList{},
	}

	// We use wireguard, so for now we don't copy others.
	reg.Servers.WireGuard = []*Server{}
	for _, w := range r.Servers.WireGuard {
		reg.Servers.WireGuard = append(reg.Servers.WireGuard, w.Clone())
	}

	return reg
}

type ServersList struct {
	IkeV2      []*Server `json:"ikev2,omitempty" yaml:"ikev2,omitempty"`
	Meta       []*Server `json:"meta,omitempty" yaml:"meta,omitempty"`
	OpenVPNTCP []*Server `json:"ovpntcp,omitempty" yaml:"ovpntcp,omitempty"`
	OpenVPNUDP []*Server `json:"ovpnudp,omitempty" yaml:"ovpnudp,omitempty"`
	WireGuard  []*Server `json:"wg,omitempty" yaml:"wg,omitempty"`
}

type Server struct {
	Latency *time.Duration `json:"latency" yaml:"latency"`
	IP      string         `json:"ip" yaml:"ip"`
	CN      string         `json:"cn" yaml:"cn"`
	VAN     bool           `json:"van" yaml:"van,omitempty"`
}

func (s *Server) Clone() *Server {
	return &Server{
		IP:  s.IP,
		CN:  s.CN,
		VAN: s.VAN,
		Latency: func() *time.Duration {
			if s.Latency == nil {
				return nil
			}

			l := *s.Latency
			return &l
		}(),
	}
}

type ServersListResponse struct {
	// Groups
	Regions []*Region `json:"regions" yaml:"regions"`
}
