package marketdata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const GoldSourceAKey = "btmc"

const (
	BTMCVRTLProductID = "btmc-vrtl-buy"
	BTMCRingProductID = "btmc-ring-buy"
)

type GoldSourceAProvider struct {
	*HTTPJSONProvider
}

func NewGoldSourceAProvider(endpoints ...string) *GoldSourceAProvider {
	endpoint := DefaultBTMCEndpoint
	if len(endpoints) > 0 && strings.TrimSpace(endpoints[0]) != "" {
		endpoint = strings.TrimSpace(endpoints[0])
	}
	provider, err := NewHTTPJSONProvider(GoldSourceAKey, endpoint, nil, decodeBTMCHTML)
	if err != nil {
		provider = &HTTPJSONProvider{ProviderKey: GoldSourceAKey, Endpoint: endpoint, Decode: decodeBTMCHTML}
	}
	provider.Headers = map[string]string{
		"User-Agent": "MyPocket-market-price/1.0",
		"Accept":     "text/html,application/xhtml+xml",
	}
	return &GoldSourceAProvider{HTTPJSONProvider: provider}
}

func decodeBTMCHTML(body []byte) ([]ProviderQuote, error) {
	text := cleanHTMLText(string(body))
	observedAt, err := parseGoldTimestamp(text, `(?i)Cập nhật lúc\s+([0-9]{1,2}/[0-9]{1,2}/[0-9]{4}\s+[0-9]{1,2}:[0-9]{2})`)
	if err != nil {
		return nil, err
	}
	rows := []struct {
		marker string
		id     string
	}{
		{marker: "VÀNG MIẾNG VRTL BẢO TÍN MINH CHÂU", id: BTMCVRTLProductID},
		{marker: "NHẪN TRÒN TRƠN BẢO TÍN MINH CHÂU", id: BTMCRingProductID},
	}
	quotes := make([]ProviderQuote, 0, len(rows))
	missing := make([]string, 0)
	for _, row := range rows {
		values := extractBTMCValues(text, row.marker)
		if len(values) != 2 {
			missing = append(missing, row.id)
			continue
		}
		price, priceErr := parseBTMCPrice(values[0])
		if priceErr != nil {
			missing = append(missing, row.id)
			continue
		}
		quotes = append(quotes, ProviderQuote{
			ExternalProductID: row.id,
			Price:             price,
			Unit:              goldUnit,
			Currency:          goldCurrency,
			ObservedAt:        observedAt,
		})
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: BTMC missing %s", ErrProviderPartial, strings.Join(missing, ", "))
	}
	return quotes, nil
}

const (
	goldUnit     = "chỉ"
	goldCurrency = "VND"
)

var btmcNumberPattern = regexp.MustCompile(`\b\d{2}\.\d{3}\b|\b\d{5}\b`)

func extractBTMCValues(text, marker string) []string {
	index := strings.Index(text, marker)
	if index < 0 {
		return nil
	}
	window := text[index:]
	if len(window) > 1200 {
		window = window[:1200]
	}
	matches := btmcNumberPattern.FindAllString(window, -1)
	if len(matches) < 2 {
		return nil
	}
	return matches[:2]
}

func parseBTMCPrice(value string) (int64, error) {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, value)
	if digits == "" {
		return 0, fmt.Errorf("invalid BTMC price %q", value)
	}
	n, err := strconv.ParseInt(digits, 10, 64)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid BTMC price %q", value)
	}
	// BTMC HTML labels these values as 1 = 1,000 VND. Keep the quote in
	// canonical VND/chỉ for the market catalog; this is a unit normalization,
	// not a cross-product conversion.
	if n < 1_000_000 {
		n *= 1_000
	}
	return n, nil
}

func parseGoldTimestamp(text, pattern string) (time.Time, error) {
	matches := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("missing source timestamp")
	}
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return time.Time{}, err
	}
	value := strings.TrimSpace(matches[1])
	for _, layout := range []string{"2/1/2006 15:04", "2006-01-02 15:04:05.999", "2006-01-02 15:04:05"} {
		if parsed, parseErr := time.ParseInLocation(layout, value, location); parseErr == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid source timestamp %q", value)
}
