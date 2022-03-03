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
