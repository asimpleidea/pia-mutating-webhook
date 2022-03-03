package main

type byLowerLatency []*ServerLatency

func (ll byLowerLatency) Len() int {
	return len(ll)
}

func (ll byLowerLatency) Less(i, j int) bool {
	iserv, jserv := ll[i], ll[j]
	if iserv.Latency == nil || jserv.Latency == nil {
		// Don't sort
		return false
	}

	return *iserv.Latency < *jserv.Latency
}

func (ll byLowerLatency) Swap(i, j int) {
	ll[i], ll[j] = ll[j], ll[i]
}

type byGreaterLatency []*ServerLatency

func (ll byGreaterLatency) Len() int {
	return len(ll)
}

func (ll byGreaterLatency) Less(i, j int) bool {
	iserv, jserv := ll[i], ll[j]
	if iserv.Latency == nil || jserv.Latency == nil {
		// Don't sort
		return false
	}

	return *iserv.Latency > *jserv.Latency
}

func (ll byGreaterLatency) Swap(i, j int) {
	ll[i], ll[j] = ll[j], ll[i]
}

type byLowerRegionName []*ServerLatency

func (rn byLowerRegionName) Len() int {
	return len(rn)
}

func (rn byLowerRegionName) Less(i, j int) bool {
	iserv, jserv := rn[i], rn[j]

	if iserv.Region.Name < jserv.Region.Name {
		return true
	}

	if iserv.Region.Name == jserv.Region.Name {
		return *iserv.Latency > *jserv.Latency
	}

	return false
}

func (rn byLowerRegionName) Swap(i, j int) {
	rn[i], rn[j] = rn[j], rn[i]
}

type byGreaterRegionName []*ServerLatency

func (rn byGreaterRegionName) Len() int {
	return len(rn)
}

func (rn byGreaterRegionName) Less(i, j int) bool {
	iserv, jserv := rn[i], rn[j]

	if iserv.Region.Name > jserv.Region.Name {
		return true
	}

	if iserv.Region.Name == jserv.Region.Name {
		return *iserv.Latency > *jserv.Latency
	}

	return false
}

func (rn byGreaterRegionName) Swap(i, j int) {
	rn[i], rn[j] = rn[j], rn[i]
}
