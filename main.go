package main

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/patcharp/golib/v2/crypto"
	"github.com/patcharp/golib/v2/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"lesson1/config"
	"lesson1/model"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// 1 - env
// 2 - file config
// 2.1 - yml
// 2.2 - toml

//{
//	"cid": "123456",
//	"name": "alice"
//}

type Account struct {
	// Public - Can access by any package
	Cid       string `json:"cid"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`

	// Private - Cannot access this property by other package such, json, yaml
	lastName string `json:"last_name"`
}

var (
	Version string
	dbCtx   *gorm.DB
	appCfg  *config.Config
)

func main() {
	// Read configuration
	var err error
	cfgFile := util.GetEnv("CONFIG_FILE", "config.yml")
	appCfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Println("[ERR] Read configuration file error -:", err)
		return
	}

	// Database connection
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		appCfg.DatabaseCfg.Username,
		appCfg.DatabaseCfg.Password,
		appCfg.DatabaseCfg.Host,
		appCfg.DatabaseCfg.Port,
		appCfg.DatabaseCfg.DBName,
	)
	dbCtx, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("[ERR] Cannot connect to database:", err)
		return
	}

	// Auto migrate schema
	if err := dbCtx.AutoMigrate(&model.Account{}); err != nil {
		fmt.Println("Migrate table Account error -:", err)
		return
	}
	if err := dbCtx.AutoMigrate(&model.Login{}); err != nil {
		fmt.Println("Migrate table Login error -:", err)
		return
	}

	// New Fiber App
	app := fiber.New()

	// Route define
	// Home
	app.Get("/", getHomeHandler)

	apiGroup := app.Group("/api")
	// Register
	apiGroup.Post("/register", registerAccountHandler)
	// Login
	apiGroup.Post("/login", loginHandler)
	// get account
	apiGroup.Get("/account", getAccountHandler)
	// get account name
	apiGroup.Get("/account/:id", getAccountByIdHandler)

	// Start server
	listen := fmt.Sprintf("%s:%s", appCfg.Server.Listen, appCfg.Server.Port)
	if err := app.Listen(listen); err != nil {
		fmt.Println("[ERR] start server error -:", err)
	}
}

func getHomeHandler(ctx *fiber.Ctx) error {
	// Don't worry, for test only
	return ctx.JSON(map[string]interface{}{
		"version": Version,
		"data": Account{
			Cid:       "123456",
			Name:      "Alice",
			FirstName: "Alice",
			lastName:  "Bob",
		},
	})
}

func registerAccountHandler(ctx *fiber.Ctx) error {
	//	{
	//		"name": "",
	//		"first_name": "",
	//		"cid": "",
	//		"username": "",
	//		"password": ""
	//	}
	type RegisterAccount struct {
		Name      string `json:"name"`
		FirstName string `json:"first_name"`
		Cid       string `json:"cid"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}
	var regAccount RegisterAccount
	if err := ctx.BodyParser(&regAccount); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "invalid json body",
		})
	}

	// Validate information
	if match, _ := regexp.MatchString(`^[a-z0-9]{4,}$`, regAccount.Username); !match {
		// not match
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "invalid username",
		})
	}
	if len(regAccount.Password) < 8 {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "invalid password",
		})
	}

	// Prepare
	account := model.Account{
		Uid:       uuid.New().String(),
		Name:      regAccount.Name,
		Firstname: regAccount.FirstName,
		Cid:       regAccount.Cid,
		Age:       0,
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(regAccount.Password), 10)
	if err != nil {
		fmt.Println("[ERR] hash password error -:", err)
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "register new account error",
		})
	}
	login := model.Login{
		AccountUid: account.Uid,
		Username:   regAccount.Username,
		Password:   string(hashedPwd),
	}

	// Insert database with transaction
	tx := dbCtx.Begin()
	// insert account
	if err := tx.Create(&account).Error; err != nil {
		tx.Rollback()
		fmt.Println("[ERR] insert account failed -:", err)
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "register new account failed",
		})
	}
	// insert login
	if err := tx.Create(&login).Error; err != nil {
		tx.Rollback()
		fmt.Println("[ERR] insert login failed -:", err)
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "register new account failed",
		})
	}

	// Don't forget to commit transaction
	// This may cause of too many connection database error
	tx.Commit()

	// Response
	return ctx.Status(http.StatusOK).JSON(map[string]interface{}{
		"message":     "success",
		"account_uid": account.Uid,
	})
}

func loginHandler(ctx *fiber.Ctx) error {
	type LoginCredential struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var credential LoginCredential
	if err := ctx.BodyParser(&credential); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "invalid json body",
		})
	}

	//fmt.Println("Username", credential.Username, "Password", credential.Password)
	// Validate user and password
	// username container ^[a-z0-9]{4,}$
	// password not care

	if match, _ := regexp.MatchString(`^[a-z0-9]{4,}$`, credential.Username); !match {
		// not match
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "invalid username",
		})
	}

	if len(credential.Password) < 8 {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "invalid password",
		})
	}

	var login model.Login
	// SELECT * FROM login WHERE username="abc" LIMIT 1;
	if err := dbCtx.Where("username=?", credential.Username).Take(&login).Error; err != nil {
		// If error happen
		// 1 - record not found
		// 2 - database error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User not found
			return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
				"error": "request user not found",
			})
		}
		// system error
		fmt.Println("Get login error -:", err)
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "something went wrong",
		})
	}

	//
	// p@ssw0rd ---> Hash function bcrypt (10)
	if err := bcrypt.CompareHashAndPassword([]byte(login.Password), []byte(credential.Password)); err != nil {
		// login failed
		return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
			"error": "incorrect password",
		})
	}

	// AES{uid}server_secret
	token, err := crypto.AESEncrypt([]byte(login.AccountUid), appCfg.Secret.Key)
	if err != nil {
		fmt.Println("[ERR] generate token error -:", err)
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "generate token error",
		})
	}

	return ctx.Status(http.StatusOK).JSON(map[string]interface{}{
		"message": "OK",
		"token":   token,
	})
}

// Authorization: Bearer <token>
func getAccountHandler(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")

	// check bearer prefix token
	if !strings.HasPrefix(token, "Bearer ") {
		return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
			"error": "invalid token type",
		})
	}

	// Trim Bearer from raw token
	token = strings.TrimPrefix(token, "Bearer ")

	uid, err := crypto.AESDecrypt(token, appCfg.Secret.Key)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
			"error": "invalid token",
		})
	}

	var account model.Account
	if err := dbCtx.Where("uid=?", string(uid)).Take(&account).Error; err != nil {
		// If error happen
		// 1 - record not found
		// 2 - database error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User not found
			return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
				"error": "request user not found",
			})
		}
		// system error
		fmt.Println("Get account error -:", err)
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "something went wrong",
		})
	}

	// Data Transfer Object (dto)
	// map -- ok
	// struct --> arrange
	type RespAccount struct {
		Uid       string `json:"uid"`
		Name      string `json:"name"`
		FirstName string `json:"first_name"`
		Age       int    `json:"age"`
		CreatedAt string `json:"created_at"`
	}

	return ctx.Status(http.StatusOK).JSON(map[string]interface{}{
		"message": "OK",
		// Map
		//"data": map[string]interface{}{
		//	"uid":        account.Uid,
		//	"name":       account.Name,
		//	"first_name": account.Firstname,
		//	"age":        account.Age,
		//	"created_at": account.CreatedAt.Format(time.DateTime),
		//},
		// Struct
		"data": RespAccount{
			Uid:       account.Uid,
			Name:      account.Name,
			FirstName: account.Firstname,
			Age:       account.Age,
			CreatedAt: account.CreatedAt.Format(time.DateTime),
		},
	})
}

func getAccountByIdHandler(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	fmt.Println("id", id)

	// TODO: Try yourself

	return ctx.Status(http.StatusOK).JSON(map[string]interface{}{
		"message": "OK",
	})
}
