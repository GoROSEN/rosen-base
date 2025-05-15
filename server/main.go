package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	logging "log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GoROSEN/rosen-apiserver/core/auth"
	"github.com/GoROSEN/rosen-apiserver/core/blockchain"
	"github.com/GoROSEN/rosen-apiserver/core/common"
	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/GoROSEN/rosen-apiserver/core/cronjob"
	"github.com/GoROSEN/rosen-apiserver/core/user"

	"github.com/GoROSEN/rosen-apiserver/business/rosen"
	"github.com/GoROSEN/rosen-apiserver/core/notification"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/google/martian/log"
	"github.com/oschwald/geoip2-golang"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func migratedb(db *gorm.DB) {

	config := config.GetConfig()
	log.Infof("migrating database...\n")
	user.MigrateDB(db)
	notification.MigrateDB(db)

	if config.ActiveModules == "*" || strings.Index(config.ActiveModules, "rosen") >= 0 {
		rosen.MigrateDB(db)
	}

	log.Infof("database migrated\n")
}

func main() {

	app := &cli.App{
		Name:  "ark",
		Usage: "a management system for lamb",
		Commands: []*cli.Command{
			{
				Name: "serve",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "migratedb",
						Value: "no",
						Usage: "set 'yes' to migrate models to database",
					},
					&cli.StringFlag{
						Name:  "config",
						Value: "",
						Usage: "file path name to config file. e.g. config.yaml",
					},
				},
				Usage:  "serve restful api services",
				Action: ark,
			},
			{
				Name: "update-user-wallets",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Value: "",
						Usage: "file path name to config file. e.g. config.yaml",
					},
				},
				Usage:  "update user wallets with tokens in config file",
				Action: updateUserWallets,
			},
		},
	}

	log.SetLevel(log.Error)
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("%v", err)
	}
}

func dbFromCli(c *cli.Context, config *config.Config) *gorm.DB {
	log.Infof("open database %v: %v\n", config.Db.Driver, config.Db.ConnStr)
	var db *gorm.DB
	var err error
	if config.Db.Driver == "mysql" {
		db, err = gorm.Open(mysql.Open(config.Db.ConnStr), &gorm.Config{})
		if err != nil {
			log.Errorf("%v", err)
		}
	} else if config.Db.Driver == "postgres" {
		db, err = gorm.Open(postgres.Open(config.Db.ConnStr), &gorm.Config{})
		if err != nil {
			log.Errorf("%v", err)
		}
	} else {
		panic("Error: db driver not supported")
	}

	schema.RegisterSerializer("json", common.JSONSerializer{})
	if c.String("migratedb") == "yes" || c.String("migratedb") == "true" || c.String("migratedb") == "1" {
		migratedb(db)
	}
	return db
}

func configFromCli(c *cli.Context) *config.Config {

	config := config.GetConfig()
	if c.String("config") != "" {
		// load from file
		log.Infof("loading config from %v\n", c.String("config"))
		if config.LoadFromFile(c.String("config")) != nil {
			panic("cannot load config")
		}
	} else {
		// load from env
		log.Infof("loading config from environment")
		if config.LoadFromEnv() != nil {
			panic("cannot load config")
		}
	}

	return config
}

func updateUserWallets(c *cli.Context) error {

	log.Infof("updateUserWallets start")
	log.SetLevel(log.Info)
	cfg := configFromCli(c)
	db := dbFromCli(c, cfg)
	dbSQL, _ := db.DB()
	if dbSQL != nil {
		defer dbSQL.Close()
	}

	log.Infof("updateUserWallets creating crud service")
	srv := common.NewCrudService(db)
	if srv == nil {
		log.Errorf("cannot init crud service")
		return errors.New("cannot init crud service")
	}

	var wallets []rosen.Wallet
	if err := srv.ListAllWhere(&wallets, "pri_key != ''"); err != nil {
		log.Errorf("%v", err)
		return err
	}
	log.Infof("updateUserWallets retrieved %v wallets", len(wallets))

	userWallets := make(map[uint][]*rosen.Wallet) // userId -> wallets
	for i := range wallets {
		w := &wallets[i]
		_, exists := userWallets[w.OwnerID]
		if !exists {
			userWallets[w.OwnerID] = make([]*rosen.Wallet, 0, 4)
		}
		userWallets[w.OwnerID] = append(userWallets[w.OwnerID], w)
	}
	log.Infof("updateUserWallets collected %v users", len(userWallets))

	var solc blockchain.BlockChainAccess
	for oId, ws := range userWallets {
		missing_tokens := map[*config.BlockchainConfig]*config.BlockChainTokenConfig{}
		for i := range cfg.Rosen.Chains {
			chain := &cfg.Rosen.Chains[i]
			for _, tk := range chain.Tokens {
				tkfound := false
				for _, w := range ws {
					if w.Token == tk.Name {
						tkfound = true
						break
					}
				}
				if !tkfound {
					missing_tokens[chain] = tk
				}
			}
			if solc == nil && chain.Name == "solana" {
				solc, _ = blockchain.NewSolanaChainAccess(chain)
			}
		}
		if len(missing_tokens) == 0 {
			continue
		}
		log.Infof("updateUserWallets user %v missing %v tokens", oId, len(missing_tokens))
		// 补
		var gethW *rosen.Wallet
		var solW *rosen.Wallet
		for i := range ws {
			if len(ws[i].Token) == 0 {
				if ws[i].Chain == "solana" {
					solW = ws[i]
				} else if gethW == nil {
					gethW = ws[i]
				}
			}
		}
		if solW == nil || gethW == nil {
			log.Errorf("user %v missing solana wallet or geth wallet!", oId)
			continue
		}
		for chain, token := range missing_tokens {
			log.Infof("updateUserWallets creating %v token on %v chain for user %v", token.Name, chain.Name, oId)
			w := rosen.Wallet{}
			w.OwnerID = oId
			w.Chain = chain.Name
			w.Token = token.Name
			w.ContractAddress = token.ContractAddress
			if chain.Name == "solana" {
				if solc == nil {
					continue
				}
				acc, _ := solc.FindTokenAccount(token.ContractAddress, solW.Address)
				w.Address = acc
				w.PubKey = acc
				w.PriKey = solW.PriKey
				w.PassPhrase = solW.PassPhrase
				w.Cipher = solW.Cipher
			} else {
				w.Address = gethW.Address
				w.PubKey = gethW.PubKey
				w.PriKey = gethW.PriKey
				w.PassPhrase = gethW.PassPhrase
				w.Cipher = gethW.Cipher
			}
			if err := srv.CreateModel(&w); err != nil {
				log.Errorf("updateUserWallets save user wallet failed: %v", err)
			} else {
				log.Infof("updateUserWallets save user wallet successfully")
			}
		}
	}

	return nil
}

func ark(c *cli.Context) error {

	config := configFromCli(c)

	// setup logger
	switch strings.ToUpper(config.Logging.Level) {
	case "DEBUG":
		log.SetLevel(log.Debug)
	case "INFO":
		log.SetLevel(log.Info)
	case "ERROR":
		log.SetLevel(log.Error)
	}

	if len(config.Logging.File) > 0 {
		logFile, err := os.OpenFile(config.Logging.File, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		mw := io.MultiWriter(os.Stdout, logFile)
		logging.SetOutput(mw)
		defer logFile.Close()
	}

	db := dbFromCli(c, config)
	dbSQL, _ := db.DB()
	if dbSQL != nil {
		defer dbSQL.Close()
	}

	log.Infof("current dialector of db is %v", db.Dialector.Name())

	log.Infof("open redis %v:%v @ %v", config.Redis.Host, config.Redis.Port, config.Redis.DB)
	rs := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password, // no password set
		DB:       config.Redis.DB,       // use default DB
	})
	defer rs.Close()
	if _, err := rs.Ping().Result(); err != nil {
		log.Errorf("cannot open redis: %v", err)
	}

	r := gin.Default()

	// setup session
	store := cookie.NewStore([]byte(config.Web.SessionToken))
	//store, _ := sredis.NewStore(config.Redis.DB, "tcp", config.Redis.Host+":"+config.Redis.Port, config.Redis.Password, []byte(config.Web.SessionToken))
	r.Use(sessions.Sessions("s", store))

	// setup cors
	if config.Cors.Enable {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowOrigins = config.Cors.AllowOrigins
		corsCfg.AllowHeaders = strings.Split(config.Cors.AllowHeaders, ",")
		corsCfg.AllowCredentials = true
		corsCfg.ExposeHeaders = []string{"Content-Length"}
		log.Infof("CORS: allowing origins %v\n", corsCfg.AllowOrigins)
		r.Use(cors.New(corsCfg))
	}

	// setup geoip
	geoPath := config.Geoip.DB
	geodb, err := geoip2.Open(geoPath)
	if err != nil {
		log.Errorf("unable load geodb %v, geoip will not work", geoPath)
		geodb = nil
	} else {
		defer geodb.Close()
	}

	// init controllers
	r.GET("/api/open/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// set auth interceptor to all
	log.Infof("initializing auth controller...")
	auth.NewController(r, rs, db)
	userAuthInterceptor := auth.NewInterceptor(rs, db)
	r.Use(userAuthInterceptor.AuthInterceptor)
	r.Use(userAuthInterceptor.ApiKeyAuthInterceptor)

	log.Infof("initializing user controller...")
	user.NewController(r, db)

	if user.InitBootstrapUser(db) {
		log.Infof("Default user is created.")
		log.Infof("You can login with admin / adminpass for the first time")
	}

	if config.ActiveModules == "*" || strings.Contains(config.ActiveModules, "rosen") {
		log.Infof("initializing rosen controller...")
		rosen.NewController(r, db, rs, geodb)
	}

	if config.EnableCronJob {
		log.Infof("initializing cronjob...")
		cronjob.GetScheduler().Start()
	}

	log.Infof("start listening")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r.Handler(),
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		quit := make(chan os.Signal, 1)
		// kill (no param) default send syscall.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Infof("shutting down the server...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Errorf("Server Shutdown:", err)
		}
		close(idleConnsClosed)
	}()

	// service connections
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Errorf("listen: %s\n", err)
	}

	<-idleConnsClosed

	log.Infof("Server exiting")

	return nil
}
