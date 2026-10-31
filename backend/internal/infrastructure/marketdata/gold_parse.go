package marketdata

import (
	"html"
	"regexp"
	"strings"
)

func cleanHTMLText(value string) string {
	withoutExecutable := regexp.MustCompile(`(?is)<script[\s\S]*?</script>|<style[\s\S]*?</style>|<svg[\s\S]*?</svg>`).ReplaceAllString(value, " ")
	withoutTags := regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(withoutExecutable, " ")
	return strings.Join(strings.Fields(html.UnescapeString(withoutTags)), " ")
}
