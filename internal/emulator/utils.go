package emulator

import (
	"bytes"
	"context"

	"github.com/IKarasev/bcatt/views"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func renderTempl(ctx echo.Context, cmp templ.Component) error {
	return cmp.Render(ctx.Request().Context(), ctx.Response())
}

func renderViewToBytes(ctx context.Context, cmp templ.Component) []byte {
	b := new(bytes.Buffer)
	cmp.Render(ctx, b)
	return b.Bytes()
}

func renderLogRow(ctx context.Context, i int, s string) []byte {
	b := new(bytes.Buffer)
	_ = views.LogRow(i, s).Render(ctx, b)
	return b.Bytes()
}
