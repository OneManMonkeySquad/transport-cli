package meta_hives

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

type PhpMetaHive struct {
	address string
}

func NewPhp(address string) (*PhpMetaHive, error) {
	return &PhpMetaHive{
		address: address,
	}, nil
}

func (p *PhpMetaHive) Tags() ([]Tag, error) {
	resp, err := http.Get(p.address + "/api?action=get_tags")
	if err != nil {
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tags []Tag
	err = json.Unmarshal(respBytes, &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (p *PhpMetaHive) FindTagByName(name string) (*Tag, error) {
	resp, err := http.Get(p.address + "/api?action=find_tag_by_name&name=" + name)
	if err != nil {
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(respBytes) == 0 {
		return nil, nil
	}

	id, err := uuid.ParseBytes(respBytes)
	if err != nil {
		return nil, err
	}

	return &Tag{
		Name: name,
		Id:   id,
	}, nil
}

func (p *PhpMetaHive) UpdateTag(name string, newId uuid.UUID) error {
	_, err := http.Get(p.address + "/api?action=update_tag&name=" + name + "&new_id=" + newId.String())
	return err
}

func (p *PhpMetaHive) FindEntry(id uuid.UUID) (uuid.UUID, error) {
	resp, err := http.Get(p.address + "/api?action=find_entry&id=" + id.String())
	if err != nil {
		return uuid.Nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, err
	}

	if len(respBytes) == 0 {
		return uuid.Nil, nil
	}

	base_id, err := uuid.ParseBytes(respBytes)
	if err != nil {
		return uuid.Nil, err
	}

	return base_id, nil
}

func (p *PhpMetaHive) AddEntry(id uuid.UUID, baseId uuid.UUID) error {
	_, err := http.Get(p.address + "/api?action=add_entry&id=" + id.String() + "&base_id=" + baseId.String())
	return err
}

func (p *PhpMetaHive) Close() {
}
