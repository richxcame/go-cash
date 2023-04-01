package main

import (
	"context"
	"fmt"
	"gocash/pkg/arrs"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

type CashBody struct {
	APIKey  string  `json:"api_key" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Contact string  `json:"contact" binding:"required"`
	Detail  string  `json:"detail"`
	Note    string  `json:"note"`
}

type RangeBody struct {
	APIKey string `json:"api_key" binding:"required"`
	Detail string `json:"detail"`
	Note   string `json:"note"`
}

type RangeBodyResponse struct {
	Detail    string    `json:"detail"`
	Note      string    `json:"note"`
	Client    string    `json:"client"`
	CreatedAt time.Time `json:"created_at"`
}

type CashBodyResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	Amount    float64   `json:"amount" binding:"required"`
	Detail    string    `json:"detail"`
	Note      string    `json:"note"`
	Client    string    `json:"client"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
}

var JWT_SECRET []byte

func init() {
	JWT_SECRET = []byte(os.Getenv("JWT_SECRET"))
}

func main() {
	// Database instance
	db := db.CreateDB()
	defer db.Close()

	r := gin.Default()

	r.POST("/cashes", func(ctx *gin.Context) {
		// Get request body
		var body CashBody
		if err := ctx.BindJSON(&body); err != nil {
			logger.Errorf("request body wrong %v", err)
			ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "Request body invalid",
			})
			return
		}

		// Find the client with the given key
		var client string
		err := db.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
		if err != nil {
			logger.Errorf("api key search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Client hasn't been found",
			})
			return
		}

		// Insert request to database
		_uuid := uuid.New().String()
		sqlStatement := `
		INSERT INTO cashes (uuid, created_at, updated_at, client, contact, amount, detail, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = db.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Contact, body.Amount, body.Detail, body.Note)
		if err != nil {
			logger.Errorf("database save error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Couldn't save into cashes",
			})
			return
		}

		// Send success result
		ctx.JSON(201, gin.H{
			"message": "Successfully saved into database",
			"uuid":    _uuid,
		})
	})

	r.POST("/ranges", func(ctx *gin.Context) {
		// Get request body
		var body RangeBody
		if err := ctx.BindJSON(&body); err != nil {
			logger.Errorf("request body wrong %v", err)
			ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "Request body invalid",
			})
			return
		}

		// Find the client with the given key
		var client string
		err := db.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
		if err != nil {
			logger.Errorf("api key search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Client hasn't been found",
			})
			return
		}

		// Insert request to database
		_uuid := uuid.New().String()
		sqlStatement := `
		INSERT INTO ranges (uuid, created_at, updated_at, client, detail, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = db.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Detail, body.Note)
		if err != nil {
			logger.Errorf("database save error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Couldn't save into cashes",
			})
			return
		}

		// Send success result
		ctx.JSON(201, gin.H{
			"message": "Successfully saved into database",
			"uuid":    _uuid,
		})
	})

	r.GET("/ranges", func(ctx *gin.Context) {
		offsetQuery := ctx.DefaultQuery("offset", "0")
		limitQuery := ctx.DefaultQuery("limit", "20")
		offset, err := strconv.Atoi(offsetQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "offset value must be convertable to integer",
			})
			return
		}
		limit, err := strconv.Atoi(limitQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "limit value must be convertable to integer",
			})
			return
		}

		var rangeBodies []RangeBodyResponse
		sqlStatement := `SELECT r.detail, r.note, r.client, r.created_at, r.updated_at FROM ranges r ORDER BY r.created_at DESC OFFSET $1 LIMIT $2;`
		rows, err := db.Query(context.Background(), sqlStatement, offset, limit)
		if err != nil {
			logger.Errorf("ranges search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Something went wrong",
			})
			return
		}
		for rows.Next() {
			var rangeBody RangeBodyResponse
			err := rows.Scan(&rangeBody.Detail, &rangeBody.Note, &rangeBody.Client, &rangeBody.CreatedAt)
			if err != nil {
				logger.Errorf("Scan error %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err,
					"message": "Scan error",
				})
				return
			}
			rangeBodies = append(rangeBodies, rangeBody)
		}
		ctx.JSON(200, gin.H{
			"message": "Successfully get ranges",
			"ranges":  rangeBodies,
		})
	})

	// TODO: need fix
	r.GET("/reports", func(ctx *gin.Context) {
		var cashBody CashBodyResponse
		var cashBodies []CashBodyResponse
		sqlStatement := `
		SELECT
		MAX(client),
		MAX(contact),
		SUM(amount),
		MAX(detail),
		MAX(created_at)
	FROM
		cashes
	WHERE
		created_at > $1
		AND created_at < $2
	GROUP BY
		contact
	ORDER BY
		MAX(created_at) DESC;`
		rows, err := db.Query(context.Background(), sqlStatement, "2023-03-19", "2023-03-20")
		if err != nil {
			logger.Errorf("cashes search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Something went wrong",
			})
			return
		}
		for rows.Next() {
			err := rows.Scan(&cashBody.Client, &cashBody.Contact, &cashBody.Amount, &cashBody.Detail, &cashBody.CreatedAt)
			if err != nil {
				logger.Errorf("Scan error %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err,
					"message": "Scan error",
				})
				return
			}
			cashBodies = append(cashBodies, cashBody)
		}
		ctx.JSON(200, gin.H{
			"message": "Successfully get cashes",
			"ranges":  cashBodies,
		})
	})

	// /cashes
	// Filters: amount, detail, note, client, contact as array
	// Pagination: offset, limit with defaults respectively 0, 50
	r.GET("/cashes", Auth(), func(ctx *gin.Context) {
		offsetQuery := ctx.DefaultQuery("offset", "0")
		limitQuery := ctx.DefaultQuery("limit", "50")
		offset, err := strconv.Atoi(offsetQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "offset value must be convertable to integer",
			})
			return
		}
		limit, err := strconv.Atoi(limitQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "limit value must be convertable to integer",
			})
			return
		}

		urlQueries := ctx.Request.URL.Query()
		index := 0
		var values []interface{}
		var queries []string
		for k, v := range urlQueries {
			if arrs.Contains([]string{"uuid", "client", "contact", "amount", "detail", "note"}, k) {
				str := ""
				for _, v := range v {
					str += v + "|"
				}
				str = strings.TrimSuffix(str, "|")
				str += ""
				values = append(values, str)
				index++

				queries = append(queries, fmt.Sprintf("%s ~* $", k)+strconv.Itoa(index))
			}
		}
		valuesWithPagination := append(values, offset, limit)

		sqlStatement := `SELECT c.uuid, c.amount, c.contact, c.client, c.detail, c.note, c.created_at FROM cashes c`
		sqlFilters := ""
		if len(queries) > 0 {
			sqlFilters += " WHERE "
			sqlFilters += strings.Join(queries, " AND ")
		}
		sqlStatement += sqlFilters
		sqlStatement += " ORDER BY created_at DESC "
		sqlStatement += fmt.Sprintf(" offset $%v limit $%v", index+1, index+2)
		rows, err := db.Query(context.Background(), sqlStatement, valuesWithPagination...)
		if err != nil {
			logger.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Couldn't search from cashes",
			})
			return
		}
		defer rows.Close()

		cashes := make([]CashBodyResponse, 0)
		for rows.Next() {
			var cash CashBodyResponse
			err := rows.Scan(&cash.UUID, &cash.Amount, &cash.Contact, &cash.Client, &cash.Detail, &cash.Note, &cash.CreatedAt)
			if err != nil {
				logger.Errorf("Scan error %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err,
					"message": "Scan error",
				})
				return
			}
			cashes = append(cashes, cash)
		}

		totalCashes := 0
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM cashes"+sqlFilters, values...).Scan(&totalCashes)
		if err != nil {
			logger.Errorf("cash count error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Couldn't count total number of transactions",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"cashes": cashes,
			"total":  totalCashes,
		})
	})

	r.POST("/login", func(ctx *gin.Context) {
		// Get body from the request
		var user User
		if err := ctx.BindJSON(&user); err != nil {
			logger.Error(err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "Couln't parse the request to user",
			})
			return
		}

		// Find the user with given data from database
		var dUser User
		err := db.QueryRow(context.Background(), "select username, password from users where username = $1", user.Username).Scan(&dUser.Username, &dUser.Password)
		if err != nil {
			logger.Error(err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "Couldn't find user",
			})
			return
		}

		// Compare passwords
		if err := bcrypt.CompareHashAndPassword([]byte(dUser.Password), []byte(user.Password)); err != nil {
			logger.Error(err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "wrong_password",
				"message": "Invalid password",
			})
			return
		}

		// Generate new token
		tokens, err := GenerateJWT(user.Username)
		if err != nil {
			logger.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Coulnd't create token",
			})
			return
		}

		// Send success response
		ctx.JSON(http.StatusOK, gin.H{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		})
	})

	r.POST("/token", func(c *gin.Context) {
		claims := &Claims{}

		// Get refresh token from request body
		token := Tokens{}
		if err := c.BindJSON(&token); err != nil {
			logger.Errorf("couldn't bind token body %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "Couldn't parse the request body",
			})
			return
		}

		// Validate jwt token
		tkn, err := jwt.ParseWithClaims(token.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})
		if err != nil {
			logger.Errorf("token didn't parse %v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "Couldn't parse token",
			})
			return
		}

		// Check expire time of the given token
		if claims.ExpiresAt.Unix() < time.Now().Local().Unix() {
			logger.Errorf("refresh_token is expired")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "token expired",
				"message": "Token is expired",
			})
			return
		}

		// Again validate token
		if !tkn.Valid {
			logger.Errorf("invalid token")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "invalid token",
				"message": "The token is invalid",
			})
			return
		}

		// Create refresh token
		tokens, err := RefreshToken(claims)
		if err != nil {
			logger.Errorf("couldn't create refresh token")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Couldn't create refresh token",
			})
			return
		}

		tokens.RefreshToken = token.RefreshToken
		c.JSON(http.StatusOK, tokens)
	})

	r.Run()
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Password  string    `json:"password"`
}

type Claims struct {
	User User `json:"user"`
	jwt.RegisteredClaims
}

// Authentication middleware
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := &Claims{}
		var token string

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "token_required",
				"message": "Auth token is required",
			})
			return
		}
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) > 1 {
			token = splitToken[1]
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "token_wrong",
				"message": "Invalid token",
			})
			return
		}
		tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "Couldn't parse token",
			})
			return
		}

		if claims.ExpiresAt.Unix() < time.Now().Local().Unix() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "token_expired",
				"message": "Token expired",
			})
			return
		}

		if !tkn.Valid {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_token",
				"message": "Invalid token",
			})
			return
		}
		c.Next()
	}
}

// GenerateJWT creates access and refresh tokens with user's username
func GenerateJWT(username string) (token Tokens, err error) {
	// Create access token
	accessTokenExp := time.Now().Add(3 * time.Hour)
	accessClaims := &Claims{
		User: User{
			Username: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: accessTokenExp},
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	token.AccessToken, err = accessToken.SignedString(JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	// Create refresh token
	refreshTokenExp := time.Now().Add(24 * time.Hour * 30)
	refreshClaims := &Claims{
		User: User{
			Username: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: refreshTokenExp},
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	token.RefreshToken, err = refreshToken.SignedString(JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	return token, nil
}

func RefreshToken(claims *Claims) (token Tokens, err error) {
	accessTokenTimeOut, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_TIMEOUT"))
	if err != nil {
		return Tokens{}, err
	}
	expirationTime := time.Now().Add(time.Duration(accessTokenTimeOut) * time.Second)

	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.AccessToken, err = accessToken.SignedString(JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	refreshTokenTimeOut, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_TIMEOUT"))
	if err != nil {
		return Tokens{}, err
	}
	expirationTime = time.Now().Add(time.Duration(refreshTokenTimeOut) * time.Second)

	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token.RefreshToken, err = refreshToken.SignedString(JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	return token, nil
}
