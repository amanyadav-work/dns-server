package main

import (
	"dns-go/models"
	"encoding/json"
	"io/ioutil"
	"net"
)

func GetNames() ([]models.Name, error) {

	data, err := ioutil.ReadFile("./names.json")
	if err != nil {
		return nil, err
	}

	var models []models.NameModel

	err = json.Unmarshal(data, &models)
	if err != nil {
		return nil, err
	}

	return To(models), nil
}


// Parser
func To(nameModels []models.NameModel) []models.Name {
	var names = make([]models.Name, 0)
	for _, model := range nameModels {
		names = append(names, models.Name{
			Name:    model.Name,
			Address: net.ParseIP(model.Address),
		})
	}
	return names
}
