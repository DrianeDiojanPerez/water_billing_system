// Filename: cmd/api/Healthcheck.go
package main

import (
	"net/http"
	//"time"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request){

	// Create a map to hold the healthcheck data
	
	data := envelope{
		"Status": "available",
		"System_Information": map[string]string{
			"Enviornment": app.config.env,
		"Version": version,
		},
		
	}
	//simulate a delay
	//time.Sleep(4 * time.Second)
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}