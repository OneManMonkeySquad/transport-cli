package main

import (
	"fmt"
)

func tags(metaHive MetaHive) error {
	tags, err := metaHive.Tags()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}
