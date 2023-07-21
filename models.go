package main

import (
	"encoding/xml"
	"fmt"
)

type XMLTable struct {
	Name string `xml:"name,attr" json:"name"`
	Tag  XMLTag `xml:"tag"`
}

type XMLTag struct {
	Name     string    `xml:"name,attr"`
	Type     string    `xml:"type,attr"`
	Writable bool      `xml:"writable,attr"`
	Desc     []XMLDesc `xml:"desc"`
}

type XMLDesc struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

type JSONTag struct {
	Writable    bool              `json:"writable"`
	Path        string            `json:"path"`
	Group       string            `json:"group"`
	Type        string            `json:"type"`
	Description map[string]string `json:"description"`
}

func xmlTableDataToJSONTag(tableXMLData string) (JSONTag, error) {
	var table XMLTable
	err := xml.Unmarshal([]byte(tableXMLData), &table)
	if err != nil {
		return JSONTag{}, err
	}

	tag := JSONTag{
		Writable:    table.Tag.Writable,
		Path:        fmt.Sprintf("%s:%s", table.Name, table.Tag.Name),
		Group:       table.Name,
		Type:        table.Tag.Type,
		Description: make(map[string]string),
	}

	for _, v := range table.Tag.Desc {
		tag.Description[v.Lang] = v.Value
	}

	return tag, nil
}
