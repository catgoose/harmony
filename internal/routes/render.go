// setup:feature:demo

package routes

import (
	"bytes"
	"context"

	"catgoose/harmony/internal/shared"

	"github.com/a-h/templ"
)

// renderToString renders a templ component to an HTML string for use as an
// SSE message body or OOB swap fragment. The description shows up in the
// shared context log so render failures can be traced back to a call site.
//
// On render error this returns the empty string. Demo routes treat this as
// "skip the publish" — no demo path needs the typed error and surfacing it
// would just clutter every call site.
func renderToString(description string, cmp templ.Component) string {
	buf := &bytes.Buffer{}
	ctx := shared.WithContextIDAndDescription(context.Background(), shared.GenerateContextID(), description)
	if err := cmp.Render(ctx, buf); err != nil {
		return ""
	}
	return buf.String()
}
