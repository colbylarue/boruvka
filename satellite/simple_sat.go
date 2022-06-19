//IN PROGRESS

package satellite

// satellite data generated from https://www.celestrak.com/NORAD/elements/table.php?GROUP=active&FORMAT=tle
// TODO: need to check copyright or PR

//type SatelliteGraph struct
// Note the Satellite Graph will have N^2 connections worst cast (technically line of sight narrows it down) (Look at subgraphs)
type SimpleSatellite struct {
	Name string
	Ole1 string
	Ole2 string
}

//lint:ignore U1000 Ignore unused function
func (s *SimpleSatellite) toString() [3]string {
	return [3]string{s.Name, s.Ole1, s.Ole2}
}
