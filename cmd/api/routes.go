// Filename: cmd/api/routes.go
package main

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)
func (app *application) routes () *httprouter.Router{
	//create a new httprouter router instance
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/todo_list", app.listtodo_listHandler)

	router.HandlerFunc(http.MethodPost, "/v1/todo_list", app.createtodo_listHandler)
	router.HandlerFunc(http.MethodGet, "/v1/todo_list/:id", app.showtodo_listHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/todo_list/:id", app.updateTodo_listHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/todo_list/:id", app.deleteTodo_listItemHandler)
	return router
}