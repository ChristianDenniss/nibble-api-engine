package service

const geohashBase32 = "0123456789bcdefghjkmnpqrstuvwxyz"

// EncodeGeohash returns a standard geohash for lat/lng at the given character precision (1–12).
func EncodeGeohash(lat, lng float64, precision int) string {
	if precision < 1 {
		precision = 1
	}
	if precision > 12 {
		precision = 12
	}

	var latMin, latMax = -90.0, 90.0
	var lngMin, lngMax = -180.0, 180.0
	var hash []byte
	var bit, ch, even int

	for len(hash) < precision {
		if even == 0 {
			mid := (lngMin + lngMax) / 2
			if lng >= mid {
				ch |= 1 << (4 - bit)
				lngMin = mid
			} else {
				lngMax = mid
			}
		} else {
			mid := (latMin + latMax) / 2
			if lat >= mid {
				ch |= 1 << (4 - bit)
				latMin = mid
			} else {
				latMax = mid
			}
		}
		even = 1 - even
		if bit < 4 {
			bit++
			continue
		}
		hash = append(hash, geohashBase32[ch])
		bit, ch = 0, 0
	}
	return string(hash)
}
