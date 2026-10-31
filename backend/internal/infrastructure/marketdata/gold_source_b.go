package marketdata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const GoldSourceBKey = "btmh"

const (
	BTMHRingProductID = "btmh-jewelry-ring-buy"
	BTMH24KProductID  = "btmh-jewelry-24k-buy"
)

type GoldSourceBProvider struct {
	*HTTPJSONProvider
}

func NewGoldSourceBProvider(endpoints ...string) *GoldSourceBProvider {
	endpoint := DefaultBTMHEndpoint
	if len(endpoints) > 0 && strings.TrimSpace(endpoints[0]) != "" {
		endpoint = strings.TrimSpace(endpoints[0])
	}
	provider, err := NewHTTPJSONProvider(GoldSourceBKey, endpoint, nil, decodeBTMHHTML)
	if err != nil {
		provider = &HTTPJSONProvider{ProviderKey: GoldSourceBKey, Endpoint: endpoint, Decode: decodeBTMHHTML}
	}
	provider.Headers = map[string]string{
		"User-Agent": "MyPocket-market-price/1.0",
		"Accept":     "text/html,application/xhtml+xml",
	}
	return &GoldSourceBProvider{HTTPJSONProvider: provider}
}

func decodeBTMHHTML(body []byte) ([]ProviderQuote, error) {
	html := string(body)
	text := cleanHTMLText(html)
	observedAt, err := parseGoldTimestamp(text, `(?i)Cập nhật lúc\s+([0-9]{4}-[0-9]{2}-[0-9]{2}\s+[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]+)?)`)
	if err != nil {
		return nil, err
	}
	rows := []struct {
		marker string
		id     string
	}{
		{marker: "Trang sức - Nhẫn tròn BTMH", id: BTMHRingProductID},
		{marker: "Vàng trang sức 24K (999.9)", id: BTMH24KProductID},
	}
	quotes := make([]ProviderQuote, 0, len(rows))
	missing := make([]string, 0)
	for _, row := range rows {
		buy, ok := extractBTMHBuy(html, row.marker)
		if !ok {
			missing = append(missing, row.id)
			continue
		}
		quotes = append(quotes, ProviderQuote{
			ExternalProductID: row.id,
			Price:             buy,
			Unit:              goldUnit,
			Currency:          goldCurrency,
			ObservedAt:        observedAt,
		})
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: BTMH missing %s", ErrProviderPartial, strings.Join(missing, ", "))
	}
	return quotes, nil
}

var btmhPricePattern = regexp.MustCompile(`\b\d{7,9}\b`)

func extractBTMHBuy(html, marker string) (int64, bool) {
	index := strings.LastIndex(html, marker)
	if index < 0 {
		return 0, false
	}
	tail := html[index+len(marker):]
	if len(tail) > 600 {
		tail = tail[:600]
	}
	match := btmhPricePattern.FindString(tail)
	if match == "" {
		return 0, false
	}
	price, err := strconv.ParseInt(match, 10, 64)
	return price, err == nil && price > 0
}
