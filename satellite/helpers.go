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

func CalculateEarthOcclusion(p1, p2 Vector3) bool {
	// Determine if a line segment defined by P1(x,y,z) and P2(x,y,z)
	// passes through the sphere defined by P3(0,0,0) and radius R
	// returns true if line intersects the sphere at least one point
	// returns false if the line does not intersect the sphere
	p3 := Vector3{0.0, 0.0, 0.0}
	radius := 6378137.0
	// calculate intermediate terms
	a_term := math.Pow((p2.X-p1.X), 2) + math.Pow((p2.Y-p1.Y), 2) + math.Pow((p2.Z-p1.Z), 2)
	b_term := 2.0 * ((p2.X-p1.X)*(p1.X-p3.X) + (p2.Y-p1.Y)*(p1.Y-p3.Y) + (p2.Z-p1.Z)*(p1.Z-p3.Z))
	c_term := (math.Pow(p3.X, 2) + math.Pow(p3.Y, 2) + math.Pow(p3.Z, 2) +
		math.Pow(p1.X, 2) + math.Pow(p1.Y, 2) + math.Pow(p1.Z, 2) -
		2.0*(p3.X*p1.X+p3.Y*p1.Y+p3.Z*p1.Z) - radius*radius)

	discriminant := b_term*b_term - 4.0*a_term*c_term

	if discriminant < 0.0 || a_term == 0 {
		// prevent divide by zero on a_term
		// the line (infinite in both directions) does not intersect the sphere
		return false
	}

	if ((-b_term+math.Sqrt(discriminant))/(2.0*a_term) <= 1.0 &&
		(-b_term+math.Sqrt(discriminant))/(2.0*a_term) >= 0.0) ||
		((-b_term-math.Sqrt(discriminant))/(2.0*a_term) <= 1.0 &&
			(-b_term-math.Sqrt(discriminant))/(2.0*a_term) >= 0.0) {
		return true
	}
	return false
}

func CalculateDistanceFromTwoLLA(p1, p2 LatLongAlt) float64 {
	x, y, z := LLAToECEF(p1.Latitude*DEG2RAD, p1.Longitude*DEG2RAD, p1.Altitude)
	a, b, c := LLAToECEF(p2.Latitude*DEG2RAD, p2.Longitude*DEG2RAD, p2.Altitude)
	var calc = math.Pow((a-x), 2.0) + math.Pow((b-y), 2.0) + math.Pow((c-z), 2.0)
	var dist = math.Sqrt(math.Abs(calc))
	return dist / 1000 // km
}
