package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"misc.sahilsasane.net/internal/data"
)

type envelope map[string]interface{}

func (app *application) readIDparam(r *http.Request) string {
	params := httprouter.ParamsFromContext(r.Context())

	id := params.ByName("id")

	return id
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSONtype for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) getUpdatedTree(session *data.Session) (*data.Tree, error) {
	tree, err := app.models.Trees.GetById(session.ChannelId)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}
	newTree := &data.Tree{}
	treeStructure := app.getTreeStructure(session, tree)
	if session.IsRoot {
		newTree = &data.Tree{
			ChannelId:     session.ChannelId,
			Root:          session.ID.Hex(),
			TreeStructure: treeStructure,
		}
	} else {
		newTree = &data.Tree{
			ChannelId:     session.ChannelId,
			TreeStructure: treeStructure,
		}
	}
	return newTree, nil
}

func (app *application) getTreeStructure(session *data.Session, tree *data.Tree) map[string]interface{} {
	if session.IsRoot {
		treeStruct := map[string]interface{}{
			"root":     session.ID.Hex(),
			"children": []interface{}{},
		}
		return treeStruct
	} else {
		existingTree := tree.TreeStructure
		parentId := session.ParentId

		return map[string]interface{}{}
	}
}
