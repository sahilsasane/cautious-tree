package main

import (
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"misc.sahilsasane.net/internal/data"
	"misc.sahilsasane.net/internal/llm"
	"misc.sahilsasane.net/internal/validator"
)

func (app *application) createSessionHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ChannelId string `json:"channel_id"`
		IsRoot    bool   `json:"is_root"`
		ParentId  string `json:"parent_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	session := &data.Session{
		ChannelId: input.ChannelId,
		Messages:  []primitive.ObjectID{},
		IsRoot:    input.IsRoot,
		ParentId:  input.ParentId,
	}

	v := validator.New()
	sessionId, err := app.models.Sessions.Insert(session)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrCannotInsert):
			v.AddError("session", "not able to insert data")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	objID, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	channel := &data.Channel{
		Sessions: []primitive.ObjectID{objID},
	}

	err = app.models.Channel.Update(input.ChannelId, channel)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	session.ID, err = primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// get updated tree structure
	newTree, err := app.getUpdatedTree(session)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// update tree structure
	err = app.models.Trees.Update(input.ChannelId, newTree)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"message": "Created session successfully", "session_id": sessionId}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
func (app *application) getSessionHandler(w http.ResponseWriter, r *http.Request) {
	id := app.readIDparam(r)

	v := validator.New()

	session, err := app.models.Sessions.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("session", "not found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"session": session}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) copySessionHandler(w http.ResponseWriter, r *http.Request) {

}
func (app *application) appendContextHandler(w http.ResponseWriter, r *http.Request) {

}
func (app *application) deleteSessionHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *application) sendSessionMessageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionId string `json:"session_id"`
		Data      struct {
			Role  string              `json:"role"`
			Parts []map[string]string `json:"parts"`
		} `json:"data"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	message := &data.Message{
		SessionId: input.SessionId,
		Data:      input.Data,
	}

	// fmt.Printf("\n\n%+v\n\n", message)

	userMessageId, err := app.models.Messages.Insert(message)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrCannotInsert):
			v.AddError("message", "cannot send message")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	userMessageObjId, err := primitive.ObjectIDFromHex(userMessageId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	aiResponse, err := llm.GetGeminiResponse(&llm.Data{
		Role:  input.Data.Role,
		Parts: input.Data.Parts,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// fmt.Print("\n\n", aiResponse, "\n\n")

	aiMessage := &data.Message{
		SessionId: input.SessionId,
		Data: struct {
			Role  string              "json:\"role\""
			Parts []map[string]string "json:\"parts\""
		}{
			Role: "model",
			Parts: []map[string]string{
				map[string]string{
					"text": aiResponse,
				},
			},
		},
	}

	aiMessageId, err := app.models.Messages.Insert(aiMessage)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrCannotInsert):
			v.AddError("message", "cannot send message")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	aiMessageObjId, err := primitive.ObjectIDFromHex(aiMessageId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	session, err := app.models.Sessions.GetById(input.SessionId)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("session", "not found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	session.Messages = append(session.Messages, userMessageObjId, aiMessageObjId)

	err = app.models.Sessions.Update(input.SessionId, session)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"message": aiResponse}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getAllSessionMessagesHandler(w http.ResponseWriter, r *http.Request) {
	id := app.readIDparam(r)

	v := validator.New()
	session, err := app.models.Sessions.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("session", "not found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	messages, err := app.models.Messages.GetAllMesssageById(session.Messages)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("session", "not found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"messages": messages}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
