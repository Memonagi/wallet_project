package xrservice

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
)

type Service struct {
	metrics *metrics
}

func New() *Service {
	return &Service{
		metrics: newMetric(),
	}
}

const round = 100

func (s *Service) GetRate(request models.XRRequest) (float64, error) {
	timeStart := time.Now()
	defer func() {
		s.metrics.externalRequestDuration.WithLabelValues("get_rate").Observe(time.Since(timeStart).Seconds())
	}()

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
