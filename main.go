package main

import (
	"context"
	"fmt"
	"gocash/config"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"gocash/utils/arrs"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
	UUID        uuid.UUID  `json:"uuid"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Client      string     `json:"client"`
	Detail      string     `json:"detail"`
	Note        string     `json:"note"`
	TotalAmount float64    `json:"total_amount"`
	Currencies  Currencies `json:"currencies"`
}

type Currencies struct {
	One        Currency `json:"one"`
	Five       Currency `json:"five"`
	Ten        Currency `json:"ten"`
	Twenty     Currency `json:"twenty"`
	Fifty      Currency `json:"fifty"`
	OneHundred Currency `json:"one_hundred"`
}

type Currency struct {
	TotalAmount float64 `json:"total_amount"`
	Amount      uint    `json:"amount"`
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

func init() {
	config.InitGlobals()
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

	r.GET("/ranges", Auth(), func(ctx *gin.Context) {
		offset, limit := Paginate(ctx)

		// Find ranges
		var rangeBodies []RangeBodyResponse
		sqlStatement := `SELECT r.uuid, r.created_at, r.updated_at, r.client, r.detail, r.note FROM ranges r ORDER BY r.created_at DESC OFFSET $1 LIMIT $2;`
		rows, err := db.Query(context.Background(), sqlStatement, offset, limit)
		if err != nil {
			logger.Errorf("ranges search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Something went wrong",
			})
			return
		}
		for rows.Next() {
			var rangeBody RangeBodyResponse
			err := rows.Scan(&rangeBody.UUID, &rangeBody.CreatedAt, &rangeBody.UpdatedAt, &rangeBody.Client, &rangeBody.Detail, &rangeBody.Note)
			if err != nil {
				logger.Errorf("Scan error %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err.Error(),
					"message": "Scan error",
				})
				return
			}
			rangeBodies = append(rangeBodies, rangeBody)
		}

		resultRanges := make([]RangeBodyResponse, 0)
		for _, v := range rangeBodies {
			var rangeBody RangeBodyResponse
			err := db.QueryRow(context.Background(), "SELECT r.created_at, r.client FROM ranges r  WHERE created_at < $1 AND client=$2 ORDER BY created_at DESC limit 1", v.CreatedAt, v.Client).Scan(&rangeBody.CreatedAt, &rangeBody.Client)
			if err != nil && err != pgx.ErrNoRows {
				logger.Errorf("last range not found: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err.Error(),
					"message": "Couldn't find searching range",
				})
				return
			} else if err == pgx.ErrNoRows {
				rangeBody.CreatedAt = time.Date(2001, 12, 28, 0, 0, 0, 0, time.Now().Location())
			}

			// Search sum of cashes between the two ranges
			var totalAmount *float64
			err = db.QueryRow(context.Background(), "SELECT SUM(amount) FROM cashes where created_at >= $1 AND created_at <= $2 AND client=$3", rangeBody.CreatedAt, v.CreatedAt, v.Client).Scan(&totalAmount)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err.Error(),
					"message": "Couldn't find total amount of the range",
				})
			}
			if totalAmount == nil {
				defaultTotalAmount := float64(0)
				totalAmount = &defaultTotalAmount
			}

			// Search sum and amount of one currency cashes between the two ranges
			var currencies Currencies
			for _, vCurrency := range []uint{1, 5, 10, 20, 50, 100} {
				var currency Currency
				rowOne, err := db.Query(context.Background(), "SELECT SUM(amount),COUNT(amount) FROM cashes where created_at >= $1 AND created_at <= $2 AND client=$3 AND amount = $4 GROUP BY amount", rangeBody.CreatedAt, v.CreatedAt, v.Client, vCurrency)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"error":   err.Error(),
						"message": "Couldn't find total amount of the range",
					})
				}

				for rowOne.Next() {
					err := rowOne.Scan(&currency.TotalAmount, &currency.Amount)
					if err != nil {
						logger.Errorf("Scan error %v", err)
						ctx.JSON(http.StatusInternalServerError, gin.H{
							"error":   err.Error(),
							"message": "Scan error",
						})
						return
					}
				}

				switch vCurrency {
				case 1:
					currencies.One = currency
				case 5:
					currencies.Five = currency
				case 10:
					currencies.Ten = currency
				case 20:
					currencies.Twenty = currency
				case 50:
					currencies.Fifty = currency
				case 100:
					currencies.OneHundred = currency
				}
			}

			resultRanges = append(resultRanges, RangeBodyResponse{
				UUID:        v.UUID,
				CreatedAt:   v.CreatedAt,
				UpdatedAt:   v.UpdatedAt,
				Client:      v.Client,
				Note:        v.Note,
				Detail:      v.Detail,
				TotalAmount: *totalAmount,
				Currencies:  currencies,
			})
		}

		// Find total count of ranges
		totalRanges := 0
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM ranges").Scan(&totalRanges)
		if err != nil {
			logger.Errorf("couldn't find total count of ranges: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Couldn't find total number of ranges",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"ranges": resultRanges,
			"total":  totalRanges,
		})
	})

	// /cashes
	// Filters: amount, detail, note, client, contact as array
	// Pagination: offset, limit with defaults respectively 0, 50
	r.GET("/cashes", Auth(), func(ctx *gin.Context) {
		offset, limit := Paginate(ctx)

		urlQueries := ctx.Request.URL.Query()
		index := 0
		var values []interface{}
		var queries []string
		for k, v := range urlQueries {
			if arrs.Contains([]string{"uuid", "client", "contact", "amount", "detail", "note"}, k) {
				if k == "contact" {
					for kcon, vcon := range v {
						if strings.Contains(vcon, " ") {
							str, _ := url.QueryUnescape(strings.Split(vcon, " ")[1])
							v[kcon] = str
						}
					}
				}
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
			return config.GlobalConfig.JWT_SECRET, nil
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

	r.GET("/cashes/:uuid", Auth(), func(ctx *gin.Context) {
		// Get UUID from URL param
		uuid, ok := ctx.Params.Get("uuid")
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "uuid is required",
				"message": "Coulnd't find UUID",
			})
			return
		}

		// Find the cash with given UUID
		var cash CashBodyResponse
		err := db.QueryRow(context.Background(), "SELECT uuid, created_at, client, contact, amount, detail, note FROM cashes where uuid = $1", uuid).Scan(&cash.UUID, &cash.CreatedAt, &cash.Client, &cash.Contact, &cash.Amount, &cash.Detail, &cash.Note)
		if err != nil {
			logger.Error(err)
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   err.Error(),
				"message": "Couldn't find the cash details",
			})
			return
		}
		ctx.JSON(200, gin.H{
			"cash": cash,
		})
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
			return config.GlobalConfig.JWT_SECRET, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
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
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
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
	os.Getenv("ACCESS_TOKEN_TIMEOUT")
	accessTokenExp := time.Now().Add(time.Duration(config.GlobalConfig.ACCESS_TOKEN_TIMEOUT) * time.Second)
	accessClaims := &Claims{
		User: User{
			Username: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: accessTokenExp},
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	token.AccessToken, err = accessToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	// Create refresh token
	refreshTokenExp := time.Now().Add(time.Duration(config.GlobalConfig.REFRESH_TOKEN_TIMEOUT) * time.Minute)
	refreshClaims := &Claims{
		User: User{
			Username: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: refreshTokenExp},
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	token.RefreshToken, err = refreshToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	return token, nil
}

func RefreshToken(claims *Claims) (token Tokens, err error) {
	expirationTime := time.Now().Add(time.Duration(config.GlobalConfig.ACCESS_TOKEN_TIMEOUT) * time.Second)

	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.AccessToken, err = accessToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	expirationTime = time.Now().Add(time.Duration(config.GlobalConfig.REFRESH_TOKEN_TIMEOUT) * time.Second)

	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token.RefreshToken, err = refreshToken.SignedString(config.GlobalConfig.JWT_SECRET)
	if err != nil {
		return Tokens{}, err
	}

	return token, nil
}

func Paginate(ctx *gin.Context) (offset, limit int) {
	offset, limit = 0, 20
	// Prepare pagination details
	offsetQuery := ctx.DefaultQuery("offset", "0")
	limitQuery := ctx.DefaultQuery("limit", "20")

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		offset = 0
	}
	limit, err = strconv.Atoi(limitQuery)
	if err != nil {
		limit = 50
	}
	return offset, limit

}
