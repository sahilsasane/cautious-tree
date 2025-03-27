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

	if input.ParentId != "" {
		parentSession, err := app.models.Sessions.GetById(input.ParentId)
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
		session.Messages = parentSession.Messages
	}

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
	id := app.readIDparam(r)
	var input struct {
		SrcSessionId string `json:"src_session_id"`
	}

	v := validator.New()

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	session, err := app.models.Sessions.GetById(input.SrcSessionId)
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

	newSession := &data.Session{
		Messages: session.Messages,
	}

	err = app.models.Sessions.Update(id, newSession)
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

	app.sessionMutex.Lock()
	delete(app.activeSessions, id)
	app.sessionMutex.Unlock()

	err = app.writeJSON(w, http.StatusAccepted, envelope{"result": "Session " + id + " appended context from" + session.ID.Hex()}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
func (app *application) deleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	id := app.readIDparam(r)

	v := validator.New()
	err := app.models.Sessions.Delete(id)
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
	err = app.writeJSON(w, http.StatusAccepted, envelope{"result": "Session " + id + " deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
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

	// Create message for DB storage
	message := &data.Message{
		SessionId: input.SessionId,
		Data:      input.Data,
	}

	// Insert user message into database
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

	// Get or create the chat session
	app.sessionMutex.RLock()
	chatSession, exists := app.activeSessions[input.SessionId]
	app.sessionMutex.RUnlock()
	if !exists {
		// Session not in cache, need to build it from DB
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

		// Get previous messages
		previousMessages, err := app.models.Messages.GetAllMesssageById(session.Messages)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Create new chat session with history
		chatSession = llm.NewChatSession(app.geminiClient)
		for _, msg := range previousMessages {
			if msg.Data.Role == "user" && len(msg.Data.Parts) > 0 {
				if textValue, ok := msg.Data.Parts[0]["text"]; ok {
					chatSession.AddUserMessage(textValue)
				}
			} else if msg.Data.Role == "model" && len(msg.Data.Parts) > 0 {
				if textValue, ok := msg.Data.Parts[0]["text"]; ok {
					chatSession.AddModelMessage(textValue)
				}
			}
		}

		// Store in cache
		app.sessionMutex.Lock()
		app.activeSessions[input.SessionId] = chatSession
		app.sessionMutex.Unlock()
	}

	// Extract user message text
	userMessageText := ""
	if len(input.Data.Parts) > 0 {
		if text, ok := input.Data.Parts[0]["text"]; ok {
			userMessageText = text
		}
	}

	// Add current message to the session
	chatSession.AddUserMessage(userMessageText)

	// Get AI response using entire conversation history
	aiResponse, err := chatSession.GetGeminiResponse(chatSession.Messages)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Add AI response to chat history
	chatSession.AddModelMessage(aiResponse)

	// Create AI message for DB storage
	aiMessage := &data.Message{
		SessionId: input.SessionId,
		Data: struct {
			Role  string              "json:\"role\""
			Parts []map[string]string "json:\"parts\""
		}{
			Role: "model",
			Parts: []map[string]string{
				{"text": aiResponse},
			},
		},
	}

	// Insert AI message into database
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

	// Update session with new message IDs
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
