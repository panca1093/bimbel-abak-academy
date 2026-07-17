// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Province struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type City struct {
	ID         string `json:"id"`
	ProvinceID string `json:"province_id"`
	Name       string `json:"name"`
}

type District struct {
	ID     string `json:"id"`
	CityID string `json:"regency_id"`
	Name   string `json:"name"`
}

const baseURL = "https://raw.githubusercontent.com/emsifa/api-wilayah-indonesia/master/static/api"

var client = &http.Client{Timeout: 60 * time.Second}

func fetch(url string, v interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func main() {
	fmt.Fprintln(os.Stderr, "Fetching provinces...")
	var provinces []Province
	if err := fetch(baseURL+"/provinces.json", &provinces); err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching provinces: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Got %d provinces\n", len(provinces))

	// Fetch all cities
	var allCities []City
	for i, p := range provinces {
		fmt.Fprintf(os.Stderr, "Fetching cities for %s (%d/%d)...\n", p.Name, i+1, len(provinces))
		var cities []City
		if err := fetch(fmt.Sprintf("%s/regencies/%s.json", baseURL, p.ID), &cities); err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching cities for province %s: %v\n", p.ID, err)
			os.Exit(1)
		}
		allCities = append(allCities, cities...)
	}
	fmt.Fprintf(os.Stderr, "Got %d cities total\n", len(allCities))

	// Fetch all districts
	var allDistricts []District
	for i, c := range allCities {
		fmt.Fprintf(os.Stderr, "Fetching districts for %s (%d/%d)...\n", c.Name, i+1, len(allCities))
		var districts []District
		if err := fetch(fmt.Sprintf("%s/districts/%s.json", baseURL, c.ID), &districts); err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching districts for city %s: %v\n", c.ID, err)
			os.Exit(1)
		}
		allDistricts = append(allDistricts, districts...)
	}
	fmt.Fprintf(os.Stderr, "Got %d districts total\n", len(allDistricts))

	// Sort for deterministic output
	sort.Slice(provinces, func(i, j int) bool { return provinces[i].ID < provinces[j].ID })
	sort.Slice(allCities, func(i, j int) bool { return allCities[i].ID < allCities[j].ID })
	sort.Slice(allDistricts, func(i, j int) bool { return allDistricts[i].ID < allDistricts[j].ID })

	// Escape single quotes
	esc := func(s string) string { return strings.ReplaceAll(s, "'", "''") }

	// Generate up.sql
	up, _ := os.Create(os.Args[1])
	defer up.Close()

	fmt.Fprintln(up, "-- 0029_seed_province_city_district.up.sql")
	fmt.Fprintln(up, "-- Auto-generated from emsifa/api-wilayah-indonesia static data")
	fmt.Fprintln(up, "-- Generated at:", time.Now().Format(time.RFC3339))
	fmt.Fprintln(up)
	fmt.Fprintln(up, "CREATE TABLE IF NOT EXISTS province (")
	fmt.Fprintln(up, "    id TEXT PRIMARY KEY,")
	fmt.Fprintln(up, "    name TEXT NOT NULL")
	fmt.Fprintln(up, ");")
	fmt.Fprintln(up)
	fmt.Fprintln(up, "CREATE TABLE IF NOT EXISTS city (")
	fmt.Fprintln(up, "    id TEXT PRIMARY KEY,")
	fmt.Fprintln(up, "    province_id TEXT NOT NULL REFERENCES province(id),")
	fmt.Fprintln(up, "    name TEXT NOT NULL")
	fmt.Fprintln(up, ");")
	fmt.Fprintln(up)
	fmt.Fprintln(up, "CREATE TABLE IF NOT EXISTS district (")
	fmt.Fprintln(up, "    id TEXT PRIMARY KEY,")
	fmt.Fprintln(up, "    city_id TEXT NOT NULL REFERENCES city(id),")
	fmt.Fprintln(up, "    name TEXT NOT NULL")
	fmt.Fprintln(up, ");")

	// Insert provinces
	fmt.Fprintf(up, "\n-- %d provinces\n", len(provinces))
	for i, p := range provinces {
		if i%50 == 0 && i > 0 {
			fmt.Fprintf(up, "\n")
		}
		fmt.Fprintf(up, "INSERT INTO province (id, name) VALUES ('%s', '%s');\n", p.ID, esc(p.Name))
	}

	// Insert cities in batches
	fmt.Fprintf(up, "\n-- %d cities/regencies\n", len(allCities))
	for i, c := range allCities {
		if i%50 == 0 && i > 0 {
			fmt.Fprintf(up, "\n")
		}
		fmt.Fprintf(up, "INSERT INTO city (id, province_id, name) VALUES ('%s', '%s', '%s');\n", c.ID, c.ProvinceID, esc(c.Name))
	}

	// Insert districts in batches
	fmt.Fprintf(up, "\n-- %d districts (kecamatan)\n", len(allDistricts))
	for i, d := range allDistricts {
		if i%50 == 0 && i > 0 {
			fmt.Fprintf(up, "\n")
		}
		fmt.Fprintf(up, "INSERT INTO district (id, city_id, name) VALUES ('%s', '%s', '%s');\n", d.ID, d.CityID, esc(d.Name))
	}

	// Generate down.sql
	down, _ := os.Create(os.Args[2])
	defer down.Close()

	fmt.Fprintln(down, "-- 0029_seed_province_city_district.down.sql")
	fmt.Fprintln(down)
	fmt.Fprintln(down, "DROP TABLE IF EXISTS district;")
	fmt.Fprintln(down, "DROP TABLE IF EXISTS city;")
	fmt.Fprintln(down, "DROP TABLE IF EXISTS province;")

	fmt.Fprintf(os.Stderr, "Done.\n")
}
