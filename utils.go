package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// New function to check if a site exists in the sites.yaml file
func checkSiteExists(siteName string) error {
	var sites []Site

	_, err := os.Stat("data/sites.yaml")
	if err != nil {
		return fmt.Errorf("the site does not exist")
	}

	file, err := ioutil.ReadFile("data/sites.yaml")
	if err != nil {
		return fmt.Errorf("the site does not exist")
	}

	err = yaml.Unmarshal(file, &sites)
	if err != nil {
		return fmt.Errorf("the site does not exist")
	}

	for _, site := range sites {
		if site.Name == siteName {
			return nil
		}
	}

	return fmt.Errorf("the site does not exist")
}

func checkConsecutiveCharacters(s string) bool {
	s = strings.ToLower(s) // convert string to all lowercase for case-insensitivity
	count := 1
	for i := 1; i < len(s); i++ {
		if s[i-1] == s[i] {
			count++
			if count >= 5 {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}
