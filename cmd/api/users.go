// Filename: cmd/api/users.go

package main

import (
	"errors"
	"net/http"
	"time"

	"water.biling.system.driane.perez.net/internal/data"
	"water.biling.system.driane.perez.net/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	//Hold data from reuest body
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//Parese the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//copy the data to a new struct
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	//generate a password hash
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Perform validation
	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Insert the datain the database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exist")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Generate a token for the newly created user 
	token, err := app.models.Tokens.New(user.ID, 1*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	
	app.background(func() {
		data := map[string]interface{} {
			"activationToken" : token.Plaintext, 
			"userID" : user.ID,
		}
		// Send the email to the new user
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			// log errors
			app.logger.PrintError(err, nil)
		}

	})

	//write a 202 Accepted status
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
//active user when the PUT endpoint is requested
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	//Parse the plaintext activation token
	var input struct {
		TokenPlaintext string `json:"token"`
		}
		err := app.readJSON(w, r, &input)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
		// Perform validation
		v := validator.New()
		if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		//Get the user details of the provided token or give the
		//client feedback about an invalid token

		user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				v.AddError("token", "invalid or expired activation token")
				app.failedValidationResponse(w, r, v.Errors)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		//Update the user status
		user.Activated = true
		//Save the update user's record in our database
		err = app.models.Users.Update(user)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		//Delete the user's token that was used for activation
		err = app.models.Tokens.DeleteAllForUsers(data.ScopeActivation, user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// Send a json response with the updated details 
		err = app.writeJSON(w, http.StatusOK, envelope{"user":user}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	
