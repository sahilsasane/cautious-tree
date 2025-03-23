package main

import (
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"misc.sahilsasane.net/internal/data"
	"misc.sahilsasane.net/internal/validator"
)

func (app *application) getChannelHandler(w http.ResponseWriter, r *http.Request) {

	id := app.readIDparam(r)

	v := validator.New()

	channel, err := app.models.Channel.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("channel", "not found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"channel": channel}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getAllChannelSessionsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ChannelId string `json:"channel_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	channel, err := app.models.Channel.GetById(input.ChannelId)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("channel", "not found")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	sesssions := []string{}
	for _, sessionID := range channel.Sessions {
		sesssions = append(sesssions, sessionID.Hex())
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sessions": sesssions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createChannelHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserId string `json:"user_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	newChannelId := primitive.NewObjectID()

	tree := &data.Tree{
		ChannelId: newChannelId.Hex(),
		Root:      "",
	}

	treeId, err := app.models.Trees.Insert(tree)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	treeObjID, err := primitive.ObjectIDFromHex(treeId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	channel := &data.Channel{
		ID:       newChannelId,
		UserId:   input.UserId,
		Sessions: []primitive.ObjectID{},
		Tree:     treeObjID,
	}

	v := validator.New()

	channelId, err := app.models.Channel.Insert(channel)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrCannotInsert):
			v.AddError("channel", "not able to insert")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"message": "Created channel successfully", "channel_id": channelId}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
