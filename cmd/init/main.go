package main

import "github.com/wcgcyx/ecl/cliq"

func main() {
	db, _, err := cliq.OpenCliqDB()
	if err != nil {
		panic(err)
	}
	db.Close()
}
