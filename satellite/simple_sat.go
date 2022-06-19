package satellite

// satellite data generated from https://www.celestrak.com/NORAD/elements/table.php?GROUP=active&FORMAT=tle
// TODO: need to check copyright or PR

//type SimpleSatellite struct
//SimpleSatellite is intended to abstract away all of the Satellite orbital calculations. Contains LLA
type SimpleSatellite struct {
	Name string
	Ole1 string
	Ole2 string
	Lla  LatLongAlt
}

//lint:ignore U1000 Ignore unused function
func (s *SimpleSatellite) ToString() [3]string {
	return [3]string{s.Name, s.Ole1, s.Ole2}
}

//func initSat pulls the TLE data to generate a LLA position and sets the LLA variable.
func InitSat(s *SimpleSatellite) {
	temp_sat := TLEToSat(s.Ole1, s.Ole2, GravityWGS84)
	pos, _ := Propagate(temp_sat, 2022, 6, 1, 0, 0, 0)
	s.Lla = ECIToLLA(pos, GSTimeFromDate(2022, 1, 1, 0, 0, 0))
}
