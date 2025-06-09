package mapprovider

import (
	"log"
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

// replace {serverpart:<item1>,<item2>,<item3>} randomly with item1, item2, or item3
func ReplaceServerPart(template string) string {
	re := regexp.MustCompile(`\{serverpart:([^}]+)\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		// extract "1,2,3"
		values := re.FindStringSubmatch(match)
		log.Printf("Match found: %s, Values: %v", match, values)
		if len(values) != 2 {
			return match // fallback
		}
		options := strings.Split(values[1], ",")
		selected := options[rand.Intn(len(options))]
		return selected
	})
}

func TestServerPartReplacement(t *testing.T) {
	template := "https://khms{serverpart:1}.google.com/kh/v=979?x={x}&y={y}&z={z}"
	for range 10 {
		result := ReplaceServerPart(template)
		t.Log(result)
	}
}
