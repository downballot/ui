package api

import (
	"errors"
	"log/slog"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/tekkamanendless/httperror"
	"github.com/tekkamanendless/restapiclient"
)

// Do an API request, navigating to the login page if the user is not logged in
// or if her token is invalid.
func Do(ctx app.Context, method string, path string, input any, output any, options ...restapiclient.Option) error {
	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "API wrapper; state", "api-token", apiToken)
	if apiToken == "" {
		slog.InfoContext(ctx.Context, "API wrapper; user is not logged in (no token)")
		ctx.Navigate("/login")
	}

	client := downballotapi.New("/", restapiclient.OptionHeader("Authorization", "Bearer "+apiToken))
	err := client.Do(ctx.Context, method, path, input, output, options...)
	if err != nil {
		if errors.Is(err, httperror.ErrStatusUnauthorized) {
			slog.InfoContext(ctx.Context, "API wrapper; user is not logged in", "err", err)
			ctx.Navigate("/login")
		}
		slog.DebugContext(ctx.Context, "API wrapper: setting api-error", "err", err)
		ctx.SetState("api-error", err.Error()).ExpiresIn(1 * time.Millisecond)
		return err
	}
	return nil
}
