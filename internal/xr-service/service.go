package xrservice

import (
	"fmt"
	"math"
	"strings"

	"github.com/Memonagi/wallet_project/internal/models"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

const round = 100

func (s *Service) GetRate(request models.XRRequest) (float64, error) {
	fromRate, fromExist := models.ExchangeRates[strings.ToUpper(request.FromCurrency)]
	toRate, toExist := models.ExchangeRates[strings.ToUpper(request.ToCurrency)]

	if !fromExist || !toExist {
		return 0, fmt.Errorf("currency not found in map: %w", models.ErrWrongCurrency)
	}

	rate := toRate / fromRate

	roundedRate := math.Ceil(rate*round) / round

	return roundedRate, nil
}
