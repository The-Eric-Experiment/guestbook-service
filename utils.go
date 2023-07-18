package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// New function to check if a site exists in the sites.yaml file
func checkSiteExists(siteName string) error {
	var sites []Site

	_, err := os.Stat("data/sites.yaml")
	if err != nil {
		return fmt.Errorf("The site does not exist")
	}

	file, err := ioutil.ReadFile("data/sites.yaml")
	if err != nil {
		return fmt.Errorf("The site does not exist")
	}

	err = yaml.Unmarshal(file, &sites)
	if err != nil {
		return fmt.Errorf("The site does not exist")
	}

	for _, site := range sites {
		if site.Name == siteName {
			return nil
		}
	}

	return fmt.Errorf("The site does not exist")
}
