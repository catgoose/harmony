package components

import "github.com/catgoose/linkwell"

func errorPageTheme(ec linkwell.ErrorContext) string {
	if ec.Theme != "" {
		return ec.Theme
	}
	return "light"
}
