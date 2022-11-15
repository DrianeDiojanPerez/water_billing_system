// Filename: cmd/api/entries.go
package main

import (
	"errors"
	"fmt"
	"net/http"

	"water.biling.system.driane.perez.net/internal/data"
	"water.biling.system.driane.perez.net/internal/validator"
)

// create entires hander for the POST /v1/entries endpoint
func (app *application) createwaterbill_listHandler(w http.ResponseWriter, r *http.Request) {
	//our target decode destination
	var todo_listtodolistdata struct {
		Waterbill   string   `json:"waterbill"`
		Description string   `json:"description"`
		Notes       string   `json:"notes"`
		Category    string   `json:"category"`
		Priority    string   `json:"priority"`
		Status      []string `json:"status"`
	}
	err := app.readJSON(w, r, &todo_listtodolistdata)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//copyung the values
	entries := &data.Todo_list{
		Waterbill:   todo_listtodolistdata.Waterbill,
		Description: todo_listtodolistdata.Description,
		Notes:       todo_listtodolistdata.Notes,
		Category:    todo_listtodolistdata.Category,
		Priority:    todo_listtodolistdata.Priority,
		Status:      todo_listtodolistdata.Status,
	}

	//initialize a new validator instance
	v := validator.New()

	//check the map to determine if there were any validation errors
	if data.ValidateEntires(v, entries); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//create a todo_list
	err = app.models.Todo_list.Insert(entries)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	//creates a location header for newly created resource/todo_list
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/waterbill/%d", entries.ID))
	//write the JSON response with 201 - created status code with a the body
	//being the school todolistdata and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"waterrbill": entries}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// create showentires hander for the GET /v1/entries/:id endpoint
func (app *application) showwaterbill_listHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	//fetch the specific todolist
	todolistdata_todolist, err := app.models.Todo_list.Get(id)
	//handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//write the todolistdata returned by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"waterbill": todolistdata_todolist}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) updatewaterbill_listHandler(w http.ResponseWriter, r *http.Request) {
	// This method does a partial replacement
	// Get the id for the todo_list item that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the original record from the todolistdatabase
	todolist, err := app.models.Todo_list.Get(id)
	// hadles error
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an input struct to hold todolistdata read in from the client
	// We update the input struct to use pointers because pointers have a
	// default value of nil
	// if a field remains nil then we know that the client did not update it
	//create an input struct to hold the todo_list data
	var todolistdata struct {
		Waterbill   *string  `json:"waterbill"`
		Description *string  `json:"description"`
		Notes       *string  `json:"notes"`
		Category    *string  `json:"category"`
		Priority    *string  `json:"priority"`
		Status      []string `json:"status"`
	}

	//Initalize a new json.Decoder instance
	err = app.readJSON(w, r, &todolistdata)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for updates
	if todolistdata.Waterbill != nil {
		todolist.Waterbill = *todolistdata.Waterbill
	}
	if todolistdata.Description != nil {
		todolist.Description = *todolistdata.Description
	}
	if todolistdata.Notes != nil {
		todolist.Notes = *todolistdata.Notes
	}
	if todolistdata.Category != nil {
		todolist.Category = *todolistdata.Category
	}
	if todolistdata.Priority != nil {
		todolist.Priority = *todolistdata.Priority
	}
	if todolistdata.Status != nil {
		todolist.Status = todolistdata.Status
	}

	// Perform Validation on the updated Todo_list item. If validation fails then
	// we send a 422 - unprocessable entity response to the client
	// initialize a new Validator instance
	v := validator.New()

	//Check the map to determine if there were any validation errors
	if data.ValidateEntires(v, todolist); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Pass the update todo record to the Update() method
	err = app.models.Todo_list.Update(todolist)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"todo_list": todolist}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
// The deleteTodo_listItemHandler() allows the user to delete a todo_list item from the databse by using the ID
func (app *application) deletewaterbill_listItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the todo item from the database. Send a 404 Not Found status code to the
	// client if there is no matching record
	err = app.models.Todo_list.Delete(id)
	// Error handling
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return 200 Status OK to the client with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "todo item successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
// The listtodo_listHandler() allows the client to see a listing of todo items
// based on a set criteria
func (app *application) waterbill_listHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold our query parameter
	var input struct {
		Waterbill string
		Priority  string
		Status    []string
		data.Filters
	}
	// Initialize a validator
	v := validator.New()
	// Get the URL values map
	qs := r.URL.Query()
	// use the helper methods to extract values
	input.Waterbill = app.readString(qs, "waterbill", "")
	input.Priority = app.readString(qs, "priority", "")
	input.Status = app.readCSV(qs, "status", []string{})
	// Get the page information using the read int method
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Get the sort information
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Specify the allowed sort values
	input.Filters.SortList = []string{"id", "waterbill", "priority", "-id", "-waterbill", "-priority"}
	// Check for validation errors
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get a listing of all todo items
	todo_list, metadata, err := app.models.Todo_list.GetAll(input.Waterbill, input.Priority, input.Status, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containing all the todo_list items
	err = app.writeJSON(w, http.StatusOK, envelope{"waterbill": todo_list, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}