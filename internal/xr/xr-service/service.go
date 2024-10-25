package xrservice

import (
	"fmt"
	"math"
	"strings"

	"github.com/Memonagi/wallet_project/internal/models"
)

type Service struct {
	Metrics *metrics
}

func New() *Service {
	return &Service{
		Metrics: newMetrics(),
	}
}

const round = 100

func (s *Service) GetRate(request models.XRRequest) (float64, error) {
	exchangeRates := map[string]float64{
		"USD": 1.5, //nolint:mnd
		"EUR": 1.6, //nolint:mnd
		"RUB": 1,
		"JPY": 0.8, //nolint:mnd
		"CNY": 1.2, //nolint:mnd
		"CAD": 1.3, //nolint:mnd
		"AUD": 1.1, //nolint:mnd
	}

	fromRate, fromExist := exchangeRates[strings.ToUpper(request.FromCurrency)]
	toRate, toExist := exchangeRates[strings.ToUpper(request.ToCurrency)]

	if !fromExist || !toExist {
		return 0, fmt.Errorf("currency not found in map: %w", models.ErrWrongCurrency)
	}

	rate := toRate / fromRate

	roundedRate := math.Ceil(rate*round) / round

	return roundedRate, nil
}
