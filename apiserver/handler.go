package apiserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r SignupRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

type ApiResponse[T any] struct {
	Data    *T     `json:"data"`
	Message string `json:"message,omitempty"`
}

func (s *ApiServer) signupHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[SignupRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		existingUser, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if existingUser != nil {
			return NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("user exists: %v", existingUser))
		}

		if _, err = s.store.Users.CreateUser(r.Context(), req.Email, req.Password); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode[ApiResponse[struct{}]](ApiResponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})

}


type SigninRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func (r SigninRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

func (s *ApiServer) signinHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[SigninRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err := user.ComparePassword(req.Password); err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}
	})
}