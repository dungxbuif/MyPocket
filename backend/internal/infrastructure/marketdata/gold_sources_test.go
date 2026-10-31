package marketdata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestGoldSourceAProviderMapsBTMCBuybackPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`<html><body>
<h2>VÀNG MIẾNG VRTL BẢO TÍN <br/>MINH CHÂU</h2><div>14.180</div><div>14.580</div>
<h2>NHẪN TRÒN TRƠN BẢO TÍN<br/> MINH CHÂU</h2><div>14.170</div><div>14.570</div>
<span>Cập nhật lúc 31/10/2026 13:06</span>
</body></html>`))
	}))
	defer server.Close()

	quotes, err := NewGoldSourceAProvider(server.URL).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(quotes) != 2 {
		t.Fatalf("quotes = %d, want 2", len(quotes))
	}
	if quotes[0].ExternalProductID != BTMCVRTLProductID || quotes[0].Price != 14180000 {
		t.Fatalf("first quote = %#v", quotes[0])
	}
	if quotes[1].ExternalProductID != BTMCRingProductID || quotes[1].Price != 14170000 {
		t.Fatalf("second quote = %#v", quotes[1])
	}
	if quotes[0].Unit != goldUnit || quotes[0].Currency != goldCurrency {
		t.Fatalf("quote identity = %#v", quotes[0])
	}
	want := time.Date(2026, 10, 31, 6, 6, 0, 0, time.UTC)
	if !quotes[0].ObservedAt.Equal(want) {
		t.Fatalf("ObservedAt = %s, want %s", quotes[0].ObservedAt, want)
	}
}

func TestGoldSourceBProviderMapsBTMHBuybackPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`<html><body>
<div>Cập nhật lúc 2026-10-31 18:14:58.0</div>
<script>"Trang sức - Nhẫn tròn BTMH",13280000,[470,527],13760000,"2026-10-31 08:17:25.0"
"Vàng trang sức 24K (999.9)",13180000,[476,542],13660000,"2026-10-31 08:17:27.0"</script>
</body></html>`))
	}))
	defer server.Close()

	quotes, err := NewGoldSourceBProvider(server.URL).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(quotes) != 2 {
		t.Fatalf("quotes = %d, want 2", len(quotes))
	}
	if quotes[0].ExternalProductID != BTMHRingProductID || quotes[0].Price != 13280000 {
		t.Fatalf("first quote = %#v", quotes[0])
	}
	if quotes[1].ExternalProductID != BTMH24KProductID || quotes[1].Price != 13180000 {
		t.Fatalf("second quote = %#v", quotes[1])
	}
	if quotes[0].ObservedAt.IsZero() || quotes[0].ObservedAt.Location() != time.UTC {
		t.Fatalf("ObservedAt = %#v, want UTC", quotes[0].ObservedAt)
	}
}

func TestGoldSourceProviderFailsClosedOnPartialResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`<div>VÀNG MIẾNG VRTL BẢO TÍN MINH CHÂU</div><div>14.180</div><div>Cập nhật lúc 31/10/2026 13:06</div>`))
	}))
	defer server.Close()

	_, err := NewGoldSourceAProvider(server.URL).Fetch(context.Background())
	if err == nil || !errors.Is(err, ErrProviderPartial) {
		t.Fatalf("Fetch() error = %v, want ErrProviderPartial", err)
	}
}

func TestGoldSourcesReadOnlyLiveContract(t *testing.T) {
	if os.Getenv("RUN_GOLD_SOURCE_LIVE") != "1" {
		t.Skip("set RUN_GOLD_SOURCE_LIVE=1 for the bounded external GET check")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, provider := range DefaultGoldProviders() {
		quotes, err := provider.Fetch(ctx)
		if err != nil {
			t.Fatalf("%s live fetch failed: %v", provider.Key(), err)
		}
		if len(quotes) != 2 {
			t.Fatalf("%s returned %d quotes, want 2", provider.Key(), len(quotes))
		}
		for _, quote := range quotes {
			if quote.Price <= 0 || quote.Unit != goldUnit || quote.Currency != goldCurrency || quote.ObservedAt.IsZero() || quote.ObservedAt.After(time.Now().UTC().Add(5*time.Minute)) {
				t.Fatalf("%s returned invalid quote %#v", provider.Key(), quote)
			}
		}
	}
}
