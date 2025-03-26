package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activations", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	router.HandlerFunc(http.MethodGet, "/v1/channels/:id", app.getChannelHandler)
	router.HandlerFunc(http.MethodGet, "/v1/channels/:id/sessions", app.getAllChannelSessionsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/channels/", app.createChannelHandler)

	router.HandlerFunc(http.MethodPost, "/v1/sessions/", app.createSessionHandler)
	router.HandlerFunc(http.MethodGet, "/v1/sessions/:id", app.getSessionHandler)
	router.HandlerFunc(http.MethodPost, "/v1/sessions/copy", app.copySessionHandler)
	router.HandlerFunc(http.MethodPut, "/v1/sessions/:id", app.appendContextHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/sessions/:id", app.deleteSessionHandler)
	router.HandlerFunc(http.MethodGet, "/v1/sessions/:id/messages", app.getAllSessionMessagesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/sessions/message", app.sendSessionMessageHandler)

	return router
}
