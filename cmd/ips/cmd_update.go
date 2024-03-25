package ips

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/robfig/cron"
)

func AutoUpdate() {
	c := cron.New()

	// Update the database every day
	c.AddFunc("0 0 0 * * *", UpdateDatabase)

	c.Start()

	UpdateDatabase()

}

func UpdateDatabase() {
	//update city.free.ipdb
	manager.Download("city.free.ipdb", "")
	//update qqwry.dat
	manager.Download("qqwry.dat", "")
	//update zxipv6wry.db
	manager.Download("zxipv6wry.db", "")
	//update GeoLite2-City.mmdb
	manager.Download("GeoLite2-City.mmdb", "")
	//update ip2region.xdb
	manager.Download("ip2region.xdb", "")

	//get the latest version of dbip-city-lite.mmdb

	// filter dbip url
	resp, err := http.Get("https://db-ip.com/db/download/ip-to-city-lite")
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	re := regexp.MustCompile(`https://download.db-ip.com/free/dbip-city-lite-[0-9]{4}-[0-9]{2}.mmdb.gz`)
	matches := re.FindAllString(string(body), -1)

	if len(matches) == 0 {
		fmt.Println("No matching URL found.")
		return
	}
	url := matches[0]
	manager.Download("dbip-city-lite.mmdb", url)

	//replace string "city-lite" with "asn-lite" to get the latest version of dbip-asn-lite.mmdb
	url = regexp.MustCompile(`city-lite`).ReplaceAllString(url, "asn-lite")
	manager.Download("dbip-asn-lite.mmdb", url)

}
