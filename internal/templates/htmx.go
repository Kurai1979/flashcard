package templates

import "github.com/a-h/templ"

// oobAttrs marks a fragment for an htmx out-of-band swap. htmx matches the
// element by its id and replaces it, independently of the response's main
// hx-target — which is how one POST can both reset a form and refresh a list.
//
// It returns nil rather than an empty value when oob is false: templ omits the
// attribute entirely for a nil map, whereas a present hx-swap-oob="" would make
// htmx try (and fail) to swap the element.
func oobAttrs(oob bool) templ.Attributes {
	if !oob {
		return nil
	}
	return templ.Attributes{"hx-swap-oob": "true"}
}
