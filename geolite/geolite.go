package geolite

import (
	"errors"
	"github.com/oschwald/geoip2-golang"
	"net"
	"os"
	"path/filepath"
)

func init() {

}

type Location struct {
	CountryISO  string
	CountryName string
	CityName    string
}

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func Get(ipAsString string) (location Location, err error) {
	var errInternal error
	var db *geoip2.Reader
	db, errInternal = geoip2.Open(CurrentExePath() + "/geo.mmdb")
	if errInternal != nil {
		err = errInternal
		return
	}
	defer db.Close()
	ip := net.ParseIP(ipAsString)
	if ip == nil {
		err = errors.New("invalid IP:" + ipAsString)
		return
	}

	var country *geoip2.Country
	country, errInternal = db.Country(ip)
	if errInternal != nil {
		err = errInternal
		return
	}
	if country == nil {
		err = errors.New("no country object")
		return
	}
	location.CountryISO = country.Country.IsoCode
	if country.Country.Names != nil {
		location.CountryName, _ = country.Country.Names["en"]
	}

	var city *geoip2.City

	city, errInternal = db.City(ip)
	if errInternal != nil {
		err = nil // not important
		return
	}

	if city == nil {
		err = nil // not important
		return
	}

	if city.City.Names != nil {
		location.CityName, _ = city.City.Names["en"]
	}

	return
}
