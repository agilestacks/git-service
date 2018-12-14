package api

import "strings"

func seeErrors(sep string, maybeErrors ...error) string {
	if sep == "" {
		sep = ", "
	}
	errs := make([]string, 0, len(maybeErrors))
	for _, err := range maybeErrors {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return "(no errors)"
	}
	return strings.Join(errs, sep)
}

func seeErrors2(maybeErrors ...error) string {
	return seeErrors("", maybeErrors...)
}
