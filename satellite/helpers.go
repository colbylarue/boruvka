package satellite

import (
	"log"
	"math"
	"strconv"
	"strings"
)

// Constants
const RADIUS_EARTH_METERS = 6371000
const TWOPI float64 = math.Pi * 2.0
const DEG2RAD float64 = math.Pi / 180.0
const RAD2DEG float64 = 180.0 / math.Pi
const XPDOTP float64 = 1440.0 / (2.0 * math.Pi)

// Holds latitude and Longitude in either degrees or radians
type LatLong struct {
	Latitude, Longitude float64
}

// Holds latitude and Longitude in either degrees or radians
type LatLongAlt struct {
	Latitude, Longitude, Altitude float64
}

// Holds X, Y, Z position
type Vector3 struct {
	X, Y, Z float64
}

//func (v *Vector3) toString() string {
//	return fmt.Sprintln("x: ", v.X, "\ny: ", v.Y, "\nz: ", v.Z)
//}

// Holds an azimuth, elevation and range
type LookAngles struct {
	Az, El, Rg float64
}

// Parses a two line element dataset into a Satellite struct
func ParseTLE(line1, line2 string, gravConst Gravity) (sat Satellite) {
	sat.Line1 = line1
	sat.Line2 = line2

	sat.Error = 0
	sat.whichconst = getGravConst(gravConst)

	// LINE 1 BEGIN
	sat.satnum = parseInt(strings.TrimSpace(line1[2:7]))
	sat.epochyr = parseInt(line1[18:20])
	sat.epochdays = parseFloat(line1[20:32])

	// These three can be negative / positive
	sat.ndot = parseFloat(strings.Replace(line1[33:43], " ", "", 2))
	sat.nddot = parseFloat(strings.Replace(line1[44:45]+"."+line1[45:50]+"e"+line1[50:52], " ", "", 2))
	sat.bstar = parseFloat(strings.Replace(line1[53:54]+"."+line1[54:59]+"e"+line1[59:61], " ", "", 2))
	// LINE 1 END

	// LINE 2 BEGIN
	sat.inclo = parseFloat(strings.Replace(line2[8:16], " ", "", 2))
	sat.nodeo = parseFloat(strings.Replace(line2[17:25], " ", "", 2))
	sat.ecco = parseFloat("." + line2[26:33])
	sat.argpo = parseFloat(strings.Replace(line2[34:42], " ", "", 2))
	sat.mo = parseFloat(strings.Replace(line2[43:51], " ", "", 2))
	sat.no = parseFloat(strings.Replace(line2[52:63], " ", "", 2))
	// LINE 2 END
	return
}

// Converts a two line element data set into a Satellite struct and runs sgp4init
func TLEToSat(line1, line2 string, gravConst Gravity) Satellite {
	//sat := Satellite{Line1: line1, Line2: line2}
	sat := ParseTLE(line1, line2, gravConst)

	opsmode := "i"

	sat.no = sat.no / XPDOTP
	sat.ndot = sat.ndot / (XPDOTP * 1440.0)
	sat.nddot = sat.nddot / (XPDOTP * 1440.0 * 1440)

	sat.inclo = sat.inclo * DEG2RAD
	sat.nodeo = sat.nodeo * DEG2RAD
	sat.argpo = sat.argpo * DEG2RAD
	sat.mo = sat.mo * DEG2RAD

	var year int64 = 0
	if sat.epochyr < 57 {
		year = sat.epochyr + 2000
	} else {
		year = sat.epochyr + 1900
	}

	mon, day, hr, min, sec := days2mdhms(year, sat.epochdays)

	sat.jdsatepoch = JDay(int(year), int(mon), int(day), int(hr), int(min), int(sec))

	sgp4init(&opsmode, sat.jdsatepoch-2433281.5, &sat)

	return sat
}

// Parses a string into a float64 value.
func parseFloat(strIn string) (ret float64) {
	ret, err := strconv.ParseFloat(strIn, 64)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

// Parses a string into a int64 value.
func parseInt(strIn string) (ret int64) {
	ret, err := strconv.ParseInt(strIn, 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func CalculateEarthOcclusion(p1, p2 LatLongAlt) bool {
	wgs84a := 6378137.0
	wgs84b := 6378137.0
	wgs84c := 6356752.314245
	//convert to eci:
	var pos1 = LLAToECI(p1, JDay(2022, 1, 1, 0, 0, 0))
	var coords1 = ECIToECEF(pos1, ThetaG_JD(JDay(2022, 1, 1, 0, 0, 0)))
	var pos2 = LLAToECI(p2, JDay(2022, 1, 1, 0, 0, 0))
	var coords2 = ECIToECEF(pos2, ThetaG_JD(JDay(2022, 1, 1, 0, 0, 0)))

	// make it a unit vector
	coords2.X = coords2.X / math.Abs(math.Sqrt(coords2.X*coords2.X+coords2.Y*coords2.Y+coords2.Z*coords2.Z))
	coords2.Y = coords2.Y / math.Abs(math.Sqrt(coords2.X*coords2.X+coords2.Y*coords2.Y+coords2.Z*coords2.Z))
	coords2.Y = coords2.Y / math.Abs(math.Sqrt(coords2.X*coords2.X+coords2.Y*coords2.Y+coords2.Z*coords2.Z))
	//calc intermediate terms

	value := -math.Pow(wgs84a, 2)*math.Pow(wgs84b, 2)*coords1.X*coords1.Z -
		math.Pow(wgs84a, 2)*math.Pow(wgs84b, 2)*coords2.Y*coords1.Y -
		math.Pow(wgs84b, 2)*math.Pow(wgs84c, 2)*coords2.X*coords1.X
	radical := math.Pow(wgs84a, 2)*math.Pow(wgs84b, 2)*math.Pow(coords2.Z, 2) +
		math.Pow(wgs84a, 2)*math.Pow(wgs84c, 2)*math.Pow(coords2.Y, 2) - math.Pow(wgs84a, 2)*math.Pow(coords2.Y, 2)*math.Pow(coords1.Z, 2) +
		2*math.Pow(wgs84a, 2)*coords2.Y*coords2.Z*coords1.Y*coords1.Z - math.Pow(wgs84a, 2)*math.Pow(coords2.Z, 2)*math.Pow(coords1.Y, 2) +
		math.Pow(wgs84b, 2)*math.Pow(wgs84c, 2)*math.Pow(coords2.X, 2) - math.Pow(wgs84b, 2)*math.Pow(coords2.X, 2)*math.Pow(coords1.Z, 2) +
		2*math.Pow(wgs84b, 2)*coords2.X*coords2.Z*coords1.X*coords1.Z - math.Pow(wgs84b, 2)*math.Pow(coords2.Z, 2)*math.Pow(coords1.X, 2) -
		math.Pow(wgs84c, 2)*math.Pow(coords2.X, 2)*math.Pow(coords1.Y, 2) + 2*math.Pow(wgs84c, 2)*coords2.X*coords2.Y*coords1.X*coords1.Y -
		math.Pow(wgs84c, 2)*math.Pow(coords2.Y, 2)*math.Pow(coords1.X, 2)
	magnitude := math.Pow(wgs84a, 2)*math.Pow(wgs84b, 2)*math.Pow(coords2.Z, 2) + math.Pow(wgs84a, 2)*math.Pow(wgs84c, 2)*math.Pow(coords2.Y, 2) +
		math.Pow(wgs84b, 2)*math.Pow(wgs84c, 2)*math.Pow(coords2.X, 2)

	if radical < 0 {
		// the line of sight vector does not point toward the earth
		return false
	}
	d := (value - wgs84a*wgs84b*wgs84c*math.Sqrt(radical)) / magnitude
	return (d >= 0)
}

func CalculateDistanceFromTwoLLA(p1, p2 LatLongAlt) float64 {
	var pos1 = LLAToECI(p1, JDay(2022, 1, 1, 0, 0, 0))
	var coords1 = ECIToECEF(pos1, ThetaG_JD(JDay(2022, 1, 1, 0, 0, 0)))
	var pos2 = LLAToECI(p2, JDay(2022, 1, 1, 0, 0, 0))
	var coords2 = ECIToECEF(pos2, ThetaG_JD(JDay(2022, 1, 1, 0, 0, 0)))
	var calc = (coords2.X-coords1.X)*(coords2.X-coords1.X) + (coords2.Y-coords1.Y)*(coords2.Y-coords1.Y) + (coords2.Z-coords1.Z)*(coords2.Z-coords1.Z)
	var dist = math.Sqrt(math.Abs(calc))
	return dist
}
