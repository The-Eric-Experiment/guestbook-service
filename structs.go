package main

type Site struct {
    Name string `yaml:"name"`
}

type Sites struct {
    Sites []Site `yaml:"sites"`
}
