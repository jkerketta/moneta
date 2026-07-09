package data

import (
	"os"

	"github.com/jkerketta/stocktui/internal/models"
	"gopkg.in/yaml.v3"
)

const PortfolioPath = "portfolio.yaml"

func LoadPortfolio(path string) (*models.Portfolio, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Portfolio{}, nil
		}
		return nil, err
	}
	var p models.Portfolio
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Holdings == nil {
		p.Holdings = []models.Holding{}
	}
	return &p, nil
}

func SavePortfolio(path string, p *models.Portfolio) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
