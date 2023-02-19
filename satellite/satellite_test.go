package satellite

import (
	"testing"
)

func TestSimpleSatellite(t *testing.T) {
	t.Run("Simple Satellite Test", func(t *testing.T) {
		ss := SimpleSatellite{Name: "Test 1",
			Tle1: "1 00900U 64063C   22160.52204282  .00000408  00000+0  42495-3 0  9992",
			Tle2: "2 00900  90.1760  40.7701 0029467  47.9267  23.7177 13.73809888869573"}
		InitSat(&ss)

		if ss.Lla.Latitude == 0 {
			t.Errorf("Expected LLa.Lat non-zero; but got %f", ss.Lla.Latitude)
		}
	})
}

func TestECIToLLA(t *testing.T) {
	t.Run("ECI to LLA test", func(t *testing.T) {
		var pos = Vector3{X: 0, Y: 0, Z: 0}
		var pos1 = Vector3{X: -6070000, Y: -1280000, Z: 660000}
		var new_pos = ECIToLLA(pos, GSTimeFromDate(2022, 1, 1, 0, 0, 0))
		var new_pos1 = ECIToLLA(pos1, GSTimeFromDate(2010, 1, 17, 10, 20, 36))

		if new_pos.Latitude != 0.0 {
			t.Errorf("Expected lat to be 0; but got %f", new_pos.Latitude)
		}
		if new_pos.Longitude != -100.63004894472304 {
			t.Errorf("Expected lon to be -100.630049; but got %f", new_pos.Longitude)
		}
		if new_pos.Altitude != -6378.137 {
			t.Errorf("Expected alt to be -6378.137; but got %f", new_pos.Altitude)
		}
		if new_pos1.Latitude != 6.072992234351723 {
			t.Errorf("Expected lat to be 6.072992234351723; but got %f", new_pos.Latitude)
		}
		if new_pos1.Longitude != -79.97508687282334 {
			t.Errorf("Expected lon to be -79.97508687282334; but got %f", new_pos.Longitude)
		}
		if new_pos1.Altitude != 6232123.524570515 {
			t.Errorf("Expected alt to be 6232123.524570515; but got %f", new_pos.Altitude)
		}
	})
}

func TestLLAtoECEF(t *testing.T) {
	t.Run("LLA 2 ECEF Test", func(t *testing.T) {
		lla1 := LatLongAlt{Latitude: 0, Longitude: 0, Altitude: 0}
		lla2 := LatLongAlt{Latitude: 1, Longitude: 1, Altitude: 500}

		var x, y, z = LLAToECEF(lla1.Latitude*DEG2RAD, lla1.Longitude*DEG2RAD, lla1.Altitude)

		var a, b, c = LLAToECEF(lla2.Latitude*DEG2RAD, lla2.Longitude*DEG2RAD, lla2.Altitude)

		if x != 6378137.00 {
			t.Errorf("Expected x to be 6378137.00m; but got %f", x)
		}
		if y != 0.0 {
			t.Errorf("Expected y to be 0; but got %f", y)
		}
		if z != 0.0 {
			t.Errorf("Expected z to be 0; but got %f", z)
		}
		if a != 6376700.6539387885 {
			t.Errorf("Expected x to be 6376700.6539387885m; but got %f", a)
		}
		if b != 111305.72394230908 {
			t.Errorf("Expected y to be 111km; but got %f", b)
		}
		if c != 110577.50102778527 {
			t.Errorf("Expected z to be 111km; but got %f", c)
		}
	})
}

func TestCalculateDistanceFromTwoLLA(t *testing.T) {
	t.Run("LLA 2 ECEF Test", func(t *testing.T) {
		lla1 := LatLongAlt{Latitude: 0, Longitude: 0, Altitude: 0}
		lla2 := LatLongAlt{Latitude: 1, Longitude: 1, Altitude: 0}

		var d_m = CalculateDistanceFromTwoLLA(lla1, lla2)

		if d_m != 156.8955857061789 {
			t.Errorf("Expected x to be 156.895586m; but got %f", d_m)
		}
	})
}

func TestCalculateEarthOcclusion(t *testing.T) {
	t.Run("LLA 2 ECEF Test", func(t *testing.T) {
		lla1 := LatLongAlt{Latitude: 0, Longitude: 0, Altitude: 0}
		lla2 := LatLongAlt{Latitude: 1, Longitude: 1, Altitude: 0}

		var d_m = CalculateDistanceFromTwoLLA(lla1, lla2)

		if d_m != 156.8955857061789 {
			t.Errorf("Expected x to be 156.895586m; but got %f", d_m)
		}
	})
}

func TestParseTLE(t *testing.T) {
	t.Run("Parse TLE Test", func(t *testing.T) {
		// ISS#25544
		sat := ParseTLE("1 25544U 98067A   08264.51782528 -.00002182  00000-0 -11606-4 0  2927", "2 25544  51.6416 247.4627 0006703 130.5360 325.0288 15.72125391563537", "wgs84")
		if sat.satnum != 25544 {
			t.Errorf("Expected %d; but got %d", 25544, sat.satnum)
		}
		if sat.epochyr != 8 {
			t.Errorf("Expected %d; but got %d", 8, sat.epochyr)
		}
		if sat.epochdays != 264.51782528 {
			t.Errorf("Expected %f; but got %f", 264.51782528, sat.epochdays)
		}
		if sat.ndot != -2.182e-05 {
			t.Errorf("Expected %f; but got %f", -2.182e-05, sat.ndot)
		}
		if sat.nddot != 0.0 {
			t.Errorf("Expected %f; but got %f", 0.0, sat.ndot)
		}
		if sat.bstar != -1.1606e-05 {
			t.Errorf("Expected %f; but got %f", -2.182e-05, sat.ndot)
		}
		if sat.inclo != 51.6416 {
			t.Errorf("Expected %f; but got %f", 51.6416, sat.inclo)
		}
		if sat.nodeo != 247.4627 {
			t.Errorf("Expected %f; but got %f", 247.4627, sat.nodeo)
		}
		if sat.ecco != 0.0006703 {
			t.Errorf("Expected %f; but got %f", 0.0006703, sat.ecco)
		}
		if sat.argpo != 130.536 {
			t.Errorf("Expected %f; but got %f", 130.536, sat.argpo)
		}
		if sat.mo != 325.0288 {
			t.Errorf("Expected %f; but got %f", 325.0288, sat.mo)
		}
		if sat.no != 15.72125391 {
			t.Errorf("Expected %f; but got %f", 15.72125391, sat.no)
		}

	})
	t.Run("Simple Satellite Test", func(t *testing.T) {
		// NOAA 19#33591
		sat := ParseTLE("1 33591U 09005A   16163.48990228  .00000077  00000-0  66998-4 0  9990", "2 33591  99.0394 120.2160 0013054 232.8317 127.1662 14.12079902378332", "wgs84")

		if sat.satnum != 33591 {
			t.Errorf("Expected %d; but got %d", 33591, sat.satnum)
		}
		if sat.epochyr != 16 {
			t.Errorf("Expected %d; but got %d", 16, sat.epochyr)
		}
		if sat.epochdays != 163.48990228 {
			t.Errorf("Expected %f; but got %f", 163.48990228, sat.epochdays)
		}
		if sat.ndot != 7.7e-7 {
			t.Errorf("Expected %f; but got %f", 7.7e-7, sat.ndot)
		}
		if sat.nddot != 0.0 {
			t.Errorf("Expected %f; but got %f", 0.0, sat.ndot)
		}
		if sat.bstar != .66998e-4 {
			t.Errorf("Expected %f; but got %f", .66998e-4, sat.ndot)
		}
		if sat.inclo != 99.0394 {
			t.Errorf("Expected %f; but got %f", 99.0394, sat.inclo)
		}
		if sat.nodeo != 120.216 {
			t.Errorf("Expected %f; but got %f", 120.216, sat.nodeo)
		}
		if sat.ecco != 0.0013054 {
			t.Errorf("Expected %f; but got %f", 0.0013054, sat.ecco)
		}
		if sat.argpo != 232.8317 {
			t.Errorf("Expected %f; but got %f", 232.8317, sat.argpo)
		}
		if sat.mo != 127.1662 {
			t.Errorf("Expected %f; but got %f", 127.1662, sat.mo)
		}
		if sat.no != 14.12079902 {
			t.Errorf("Expected %f; but got %f", 14.12079902, sat.no)
		}

	})
}
