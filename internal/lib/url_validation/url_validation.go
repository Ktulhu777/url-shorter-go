package url_validation

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"fmt"
)

var (
	ErrContainsSpace = errors.New("url contains a space")
	ErrNotValid      = errors.New("url is not valid")
	ErrEmpty         = errors.New("url is empty")
)

func IsValidURL(u string) error {
	const fn = "lib.url_validation.IsValidURL"

	if strings.Contains(u, " ") {
		return fmt.Errorf("%s: %w", fn, ErrContainsSpace)
	}

	re := regexp.MustCompile(`[<>#%"{}|\^~\[]`)
	if re.MatchString(u) {
		return fmt.Errorf("%s: %w", fn, ErrNotValid)
	}

	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, ErrNotValid)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("%s: %w", fn, ErrEmpty)
	}

	return nil
}
