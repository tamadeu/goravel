package controllers

import (
	stdhttp "net/http"
	"net/mail"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
)

type registerRequest struct {
	Name                 string `form:"name" json:"name"`
	Email                string `form:"email" json:"email"`
	Password             string `form:"password" json:"password"`
	PasswordConfirmation string `form:"password_confirmation" json:"password_confirmation"`
}

type loginRequest struct {
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
}

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (r *AuthController) Register(ctx http.Context) http.Response {
	var input registerRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Dados de cadastro inválidos"})
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if message := validateRegistration(input); message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}

	count, err := facades.Orm().Query().Model(&models.User{}).Where("email", input.Email).Count()
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível validar o email"})
	}
	if count > 0 {
		return ctx.Response().Json(stdhttp.StatusConflict, http.Json{"message": "Este email já está cadastrado"})
	}

	hashedPassword, err := facades.Hash().Make(input.Password)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível criar a conta"})
	}

	user := models.User{Name: input.Name, Email: input.Email, Password: hashedPassword}
	if err := facades.Orm().Query().Create(&user); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível criar a conta"})
	}

	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Conta criada, mas não foi possível iniciar a sessão"})
	}

	return authResponse(ctx, stdhttp.StatusCreated, token, user)
}

func (r *AuthController) Login(ctx http.Context) http.Response {
	var input loginRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Credenciais inválidas"})
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	var user models.User
	if err := facades.Orm().Query().Where("email", input.Email).First(&user); err != nil || !facades.Hash().Check(input.Password, user.Password) {
		return ctx.Response().Json(stdhttp.StatusUnauthorized, http.Json{"message": "Email ou senha incorretos"})
	}

	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível iniciar a sessão"})
	}

	return authResponse(ctx, stdhttp.StatusOK, token, user)
}

func (r *AuthController) Me(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return ctx.Response().Json(stdhttp.StatusUnauthorized, http.Json{"message": "Sessão inválida"})
	}

	return ctx.Response().Success().Json(http.Json{"user": user})
}

func (r *AuthController) Logout(ctx http.Context) http.Response {
	if err := facades.Auth(ctx).Logout(); err != nil {
		return ctx.Response().Json(stdhttp.StatusUnauthorized, http.Json{"message": "Sessão inválida"})
	}

	return ctx.Response().NoContent()
}

func validateRegistration(input registerRequest) string {
	if len(input.Name) < 2 {
		return "Informe seu nome"
	}
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email {
		return "Informe um email válido"
	}
	if len(input.Password) < 8 {
		return "A senha deve ter pelo menos 8 caracteres"
	}
	if input.Password != input.PasswordConfirmation {
		return "As senhas não coincidem"
	}

	return ""
}

func authResponse(ctx http.Context, status int, token string, user models.User) http.Response {
	return ctx.Response().Json(status, http.Json{
		"token":      token,
		"token_type": "Bearer",
		"user":       user,
	})
}
