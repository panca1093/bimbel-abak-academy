package model

// Province is a top-level administrative region (provinsi).
// ID uses the official Kemendagri code.
type Province struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// City is a second-level administrative region (kota/kabupaten).
// ID uses the official Kemendagri code; ProvinceID references Province.
type City struct {
	ID         string `json:"id"`
	ProvinceID string `json:"province_id"`
	Name       string `json:"name"`
}

// District is a third-level administrative region (kecamatan).
// ID uses the official Kemendagri code; CityID references City.
type District struct {
	ID     string `json:"id"`
	CityID string `json:"city_id"`
	Name   string `json:"name"`
}
