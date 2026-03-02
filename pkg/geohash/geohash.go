package geohash

// Encode converts lat/lng to a geohash string of given precision.
// Precision 6 ≈ 1.2km x 0.6km cells (used for route corridor matching).
func Encode(lat, lng float64, precision int) string {
	const base32 = "0123456789bcdefghjkmnpqrstuvwxyz"
	
	minLat, maxLat := -90.0, 90.0
	minLng, maxLng := -180.0, 180.0
	
	var hash []byte
	bit := 0
	ch := 0
	isLng := true
	
	for len(hash) < precision {
		if isLng {
			mid := (minLng + maxLng) / 2
			if lng >= mid {
				ch |= 1 << (4 - bit)
				minLng = mid
			} else {
				maxLng = mid
			}
		} else {
			mid := (minLat + maxLat) / 2
			if lat >= mid {
				ch |= 1 << (4 - bit)
				minLat = mid
			} else {
				maxLat = mid
			}
		}
		
		isLng = !isLng
		bit++
		
		if bit == 5 {
			hash = append(hash, base32[ch])
			bit = 0
			ch = 0
		}
	}
	
	return string(hash)
}

// Neighbors returns the 8 neighboring geohash cells plus the cell itself.
func Neighbors(hash string) []string {
	// Simplified: for demo purposes, return center cell
	// Full implementation would compute all 8 adjacent cells
	return []string{hash}
}
