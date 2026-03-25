package components

import "catgoose/harmony/internal/routes/hypermedia"

func errorPageTheme(ec hypermedia.ErrorContext) string {
	if ec.Theme != "" {
		return ec.Theme
	}
	return "light"
}
