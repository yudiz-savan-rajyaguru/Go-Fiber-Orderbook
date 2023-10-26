package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/opinion-trading/config"
	"github.com/opinion-trading/helper/response"
)

type Response struct {
	Code    int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SendResponse(c *fiber.Ctx, response Response) error {
	return c.Status(response.Code).JSON(response)
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	log.Println(err)

	return SendResponse(c, Response{
		Code:    fiber.StatusInternalServerError,
		Message: err.Error(),
		Data:    nil,
	})
}

func ExtractDatabaseName(dsn string) string {
	parts := strings.Split(dsn, "/")
	dbName := strings.Split(parts[1], "?")

	if len(dbName) >= 2 {
		return dbName[0]
	}
	return "unknown"
}

type CustomJwtClaims struct {
	ID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJwt(userId uint) (string, error) {
	claims := &CustomJwtClaims{
		ID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.ConfigEnv.JWT_VALIDITY) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.ConfigEnv.JWT_KEY))

	if err != nil {
		return "", err
	}

	return tokenString, err
}

func VerifyJwt(c *fiber.Ctx) (*CustomJwtClaims, error) {

	tokenString := c.Get("Auth")
	tokenString = strings.Trim(tokenString, "")

	fmt.Println("Token>>>", tokenString)
	claims := &CustomJwtClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.ConfigEnv.JWT_KEY), nil
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("Valid", token.Valid)

	if !token.Valid {
		return nil, errors.New("invalid token has been passed")
	}

	// if claims, ok := token.Claims.(*CustomJwtClaims); ok && token.Valid {
	// 	return claims, nil
	// }

	return claims, nil
}

func BodyValidator(body interface{}) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if err := c.BodyParser(body); err != nil {
			return SendResponse(c, Response{
				Code:    fiber.StatusBadRequest,
				Message: err.Error(),
				Data:    nil,
			})
		}

		validate := validator.New()
		if err := validate.Struct(body); err != nil {
			return SendResponse(c, Response{
				Message: "Validation failed",
				Code:    fiber.StatusBadRequest,
				Data:    strings.Split(err.Error(), "\n"),
			})
		} else {
			return c.Next()
		}
	}
}

func ValidateEnum(body map[string]string) map[string]string {
	m := make(map[string]string)
	for key, val := range body {
		if val == "unknown" {
			m[key] = fmt.Sprintf("%s %s", response.Words["enum"], response.Message["not_match"])
		}
	}
	return m
}

func Marshal(val interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return nil, ErrorHandler(&fiber.Ctx{}, err)
	}
	return jsonData, nil
}

// currently not used
func Validate(ctx *fiber.Ctx, body interface{}) bool {
	err := ctx.BodyParser(body)

	if err != nil {
		SendResponse(ctx, Response{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
			Data:    nil,
		})
		return false
	}

	validate := validator.New()
	err = validate.Struct(body)

	if err != nil {
		var validationErrors = make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			errorReason := err.ActualTag()
			if err.ActualTag() == strings.ToLower(err.Field()) {
				errorReason = "invalid"
			}
			validationErrors[strings.ToLower(err.Field())] = err.Field() + " is " + errorReason

			// fmt.Println("ERROR>>>> ", validationErrors)s

		}

		SendResponse(ctx, Response{
			Code:    fiber.StatusBadRequest,
			Message: "Validation Error!",
			Data:    validationErrors,
		})
		return false
	}

	return true
}
