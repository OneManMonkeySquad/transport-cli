package main

import (
	"errors"
	"fmt"
)

func tags(persistence Backend) error {
	db, err := downloadDatabase(persistence)
	if err != nil {
		return errors.New("patch database not found")
	}

	for _, tag := range db.Tags {
		fmt.Println(tag.Name)
	}

	return nil
}
