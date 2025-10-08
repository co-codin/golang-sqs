package apiserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	Email    string `json:"email"`
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

type SigninResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *ApiServer) signinHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[SigninRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := user.ComparePassword(req.Password); err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		tokenPair, err := s.JwtManager.GenerateTokenPair(user.Id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshTokens.DeleteUserTokens(r.Context(), user.Id)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshTokens.Create(r.Context(), user.Id, tokenPair.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(ApiResponse[SigninResponse]{
			Data: &SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r TokenRefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}

	return nil
}

func (s *ApiServer) tokenRefreshHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[TokenRefreshRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		currentRefreshToken, err := s.JwtManager.Parse(req.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userIdStr, err := currentRefreshToken.Claims.GetSubject()
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		currentRefreshTokenRecord, err := s.store.RefreshTokens.ByPrimaryKey(r.Context(), userId, currentRefreshToken)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusUnauthorized
			}
			return NewErrWithStatus(status, err)
		}

		if currentRefreshTokenRecord.ExpiresAt.Before(time.Now()) {
			return NewErrWithStatus(http.StatusUnauthorized, errors.New("refresh token expired"))
		}

		tokenPair, err := s.JwtManager.GenerateTokenPair(userId)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if _, err = s.store.RefreshTokens.DeleteUserTokens(r.Context(), userId); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if _, err = s.store.RefreshTokens.Create(r.Context(), userId, tokenPair.RefreshToken); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(ApiResponse[TokenRefreshResponse]{
			Data: &TokenRefreshResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type CreateReportRequest struct {
	ReportType string `json:"report_type"`
}

func (r CreateReportRequest) Validate() error {
	if r.ReportType == "" {
		return errors.New("report_type is required")
	}
	return nil
}

func (s *ApiServer) createReportHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[CreateReportRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, ok := UserFromContext(r.Context())
		if !ok {
			return NewErrWithStatus(http.StatusUnauthorized, errors.New("user not found in context"))
		}

		report, err := s.store.ReportStore.Create(r.Context(), user.Id, req.ReportType)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		
	})
}
