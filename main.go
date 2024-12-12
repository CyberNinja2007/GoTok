package main

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"

	"simple-gotok/db"
	"simple-gotok/db/controllers"
	"simple-gotok/mailer"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env файла:", err)
	}

	// Получаем значения переменных окружения
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	mailUser := os.Getenv("MAIL_USER")
	mailPass := os.Getenv("MAIL_PASS")
	mailHost := os.Getenv("MAIL_HOST")
	mailPort := os.Getenv("MAIL_PORT")
	dbName := os.Getenv("DB_NAME")
	jwtSecret := os.Getenv("JWT_SECRET")

	if dbUser == "" || dbPass == "" || dbHost == "" || dbPort == "" || dbName == "" || jwtSecret == "" {
		fmt.Fprintf(os.Stderr, "Не удалось прочитать конфигурацию приложения. Проверьте конфигурационный файл")
		os.Exit(1)
	}

	connstring := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	pool, err := db.NewPG(context.Background(), connstring)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось подключиться к БД: %v\n", err)
		os.Exit(1)
	}

	router := gin.Default()

	router.POST("/generate", func(c *gin.Context) {
		guid := c.Query("guid")

		if guid == "" {
			c.String(http.StatusBadRequest, fmt.Sprintf("Id пользователя не существует"))
			return
		}

		_, err = controllers.GetUser(pool, context.Background(), guid)
		if err != nil {
			c.String(http.StatusNotFound, fmt.Sprintf("Пользователя с заданным Id не существует"))
			return
		}

		ip := c.ClientIP()

		accessToken, refreshToken, err := createTokens(ip, guid, jwtSecret)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Не удалось создать токены"))
			return
		}

		bytes, err := bcrypt.GenerateFromPassword([]byte(refreshToken), 14)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Не удалось создать токены"))
			return
		}

		_, err = controllers.UpdateRefresh(pool, context.Background(), guid, string(bytes))
		if err != nil {
			c.String(http.StatusNotFound, fmt.Sprintf("Пользователя с заданным Id не существует"))
			return
		}

		c.SetCookie("access", accessToken, 60*60, "/", "", false, true)
		c.SetCookie("refresh", refreshToken, 7*24*60*60, "/", "", false, true)
		c.String(http.StatusCreated, "Успешно")
		return
	})

	router.POST("/refresh", func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh")
		if refreshToken == "" || err != nil {
			c.String(http.StatusBadRequest, "Refresh токена нет")
			return
		}

		claims := UserClaims{}

		validatedRefresh, err := validateToken(jwtSecret, refreshToken, &claims)
		if err != nil {
			c.String(http.StatusBadRequest, "Refresh токен недействителен")
			return
		}

		user, err := controllers.GetUser(pool, context.Background(), claims.Id)
		if err != nil {
			c.String(http.StatusNotFound, fmt.Sprintf("Пользователя с заданным Id не существует"))
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Refresh), []byte(refreshToken))
		if err != nil {
			c.String(http.StatusBadRequest, "Refresh токен недействителен")
			return
		}

		_, err = validatedRefresh.Claims.GetExpirationTime()
		if err != nil {
			c.String(http.StatusBadRequest, "Refresh токен недействителен")
			return
		}

		ip := c.ClientIP()

		if ip != claims.Ip {
			emails := []string{user.Email}
			err = mailer.SendEmail(mailHost, mailPort, mailUser, mailPass, emails, fmt.Sprintf("В ваш аккаунт был произведен вход с нового ip %s", ip))
			if err != nil {
				c.String(http.StatusInternalServerError, "Не удалось создать новый токен")
			}
		}

		accessToken, _, err := createTokens(ip, claims.Id, jwtSecret)
		if err != nil {
			c.String(http.StatusInternalServerError, "Не удалось создать новый токен")
			return
		}

		c.SetCookie("access", accessToken, 60*60, "/", "", false, true)
		c.String(http.StatusCreated, "Успешно")
		return
	})

	router.Run(":8080")
}
