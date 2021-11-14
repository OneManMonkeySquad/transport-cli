package main

import "fmt"

func tags(persistence Backend) {
	db, err := downloadDatabase(persistence)
	if err != nil {
		fmt.Println("Patch database not found.")
		return
	}

	for _, tag := range db.Tags {
		fmt.Println(tag.Name)
	}
}
