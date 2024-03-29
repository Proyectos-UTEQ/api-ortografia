package main

import (
	"Proyectos-UTEQ/api-ortografia/internal/data"
	"Proyectos-UTEQ/api-ortografia/internal/db"
	"Proyectos-UTEQ/api-ortografia/internal/handlers"
	"Proyectos-UTEQ/api-ortografia/internal/services"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"log"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/spf13/viper"
)

func main() {
	config := viper.New()

	// Read environment variables
	config.AutomaticEnv()

	config.SetDefault("APP_PORT", "3000")
	config.SetDefault("PORT", "8080")
	config.SetDefault("APP_ENV", "development")

	// Read the config file
	config.SetConfigName("config")
	config.SetConfigType("env")
	config.AddConfigPath(".")
	config.AddConfigPath("/etc/secrets/")
	// config.AddConfigPath("/workspaces/api-ortografia")

	// Load the config
	err := config.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	// Connect to the database
	database := db.ConnectDB(config)

	migrate := config.GetBool("APP_MIGRATE")
	if migrate {
		// Migrate the schema
		err = database.AutoMigrate(
			&data.User{},
			&data.ResetPassword{},
			&data.Module{},
			&data.Subscription{},
			&data.Class{},
			&data.Matricula{},
			&data.Course{},
			&data.Question{},
			&data.HistoryChat{},
			&data.ChatIssue{},
			&data.AnswerUser{},
			&data.Answer{},
			&data.Questionnaire{},
			&data.TestModule{},
		)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Create fiber app
	app := fiber.New(fiber.Config{
		AppName: "API REST Poliword",
		Prefork: config.GetBool("APP_PREFORK"),
	})

	// configuració de cors
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "",
		AllowCredentials: true,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hola, utiliza postman para probar la API")
	})

	// Create handlers
	userHandler := handlers.NewUserHandler(config)
	authHandler := handlers.NewAuthHandler(config)
	authHandler.ConfigProvider()
	jwtHandler := handlers.NewJWTHandler(config)
	moduleHandler := handlers.NewModuleHandler(config)

	api := app.Group("/api")

	//student := api.Group("/students", jwtHandler.JWTMiddleware, handlers.Authorization("student"))

	statisticsGroup := api.Group("/statistics")
	statisticsGroup.Get("/points-module", moduleHandler.GetPointsStudentsForModule)

	auth := api.Group("/auth")
	// Routes for auth users
	auth.Post("/sign-in", userHandler.HandlerSignin)
	auth.Post("/sign-up", userHandler.HandlerSignup)
	auth.Post("/reset-password", userHandler.HandlerResetPassword) // se encarga de enviar el correo electronico al usuario
	auth.Put("/change-password", userHandler.HandlerChangePassword)
	auth.Put("/change-password/inside", jwtHandler.JWTMiddleware, userHandler.HandlerChangePasswordInside)

	auth.Get("/google", adaptor.HTTPHandlerFunc(authHandler.BeginAuthGoogle))
	auth.Get("/google/callback", adaptor.HTTPHandlerFunc(authHandler.GetAuthCallbackFunction))
	auth.Get("/google/success", adaptor.HTTPHandlerFunc(authHandler.GetAuthSuccessFunction))

	userGroup := api.Group("/users", jwtHandler.JWTMiddleware)
	userGroup.Get("/me", userHandler.HandlerGetUser)
	userGroup.Put("/me", userHandler.HandlerUpdateUser)

	// Adminstración de usuarios
	userGroup.Get("/", handlers.Authorization("admin"), userHandler.GetAllUsers)
	userGroup.Put("/:id/approved", handlers.Authorization("admin"), userHandler.ActiveUser)
	userGroup.Put("/:id/blocked", handlers.Authorization("admin"), userHandler.BlockedUser)

	module := api.Group("/module", jwtHandler.JWTMiddleware) // solo con JWT se tiene acceso.
	module.Put("/:id", handlers.Authorization("teacher", "admin"), moduleHandler.UpdateModule)
	// Lista todos los modulos.
	module.Get("/teacher", handlers.Authorization("teacher", "admin"), moduleHandler.GetModulesForTeacher)
	// Lista todos los modulos.
	module.Get("/", moduleHandler.GetModules)

	// Recupera todos los modulos y ademas indica si el usuario esta suscrito o no.
	module.Get("/with-is-subscribed", moduleHandler.GetModuleWithIsSubscribed)

	module.Post("/subscribe", moduleHandler.Subscribe)
	module.Get("/subscribed", moduleHandler.Subscriptions)

	// Listar todos los estudiantes que estan suscritos a un modulo.
	module.Get("/:id/students", moduleHandler.GetStudents)

	// Routes for modules
	// Crea un modulo.
	module.Post("/", handlers.Authorization("teacher", "admin"), moduleHandler.CreateModuleForTeacher)
	module.Get("/:id", moduleHandler.GetModuleByID) // Recupera un módulo por el ID

	// Rutas para los test de los módulos.
	testModule := module.Group("/:id/test", handlers.Authorization("student"))
	testModule.Post("/", moduleHandler.GenerateTest)
	testModule.Get("/my-tests", moduleHandler.GetMyTestsByModule)
	module.Get("/test/:id", moduleHandler.GetTestByID)
	module.Put("/test/validate-answer/:answer_user_id", handlers.Authorization("student"), moduleHandler.ValidationAnswerForTestModule)
	module.Post("/test/feedback-answer/:answer_user_id", handlers.Authorization("student"), moduleHandler.GetFeedbackAnswerUser)
	module.Put("/test/:id/finish", handlers.Authorization("student"), moduleHandler.FinishTest)

	// Routes for questions
	questionHandler := handlers.NewQuestionHandler(config)
	moduleQuestionGroup := module.Group("/:id/question")
	moduleQuestionGroup.Post("/", questionHandler.RegisterQuestionForModule)
	moduleQuestionGroup.Get("/", questionHandler.GetQuestionsForModule)
	moduleQuestionGroup.Delete("/:idquestion", questionHandler.DeleteQuestion)
	moduleQuestionGroup.Put("/:idquestion", questionHandler.UpdateQuestion)
	moduleQuestionGroup.Get("/activities", questionHandler.GetActivityForModule)
	module.Get("/question/:id", questionHandler.GetQuestionByID)

	// Routes for upload
	upload := api.Group("/uploads")
	uploadHandler := handlers.NewUploadHandler(config)

	// Routes for GPT AI.
	gptHandlers := handlers.NewGPTHandler(config)
	gptGroup := api.Group("/gpt", jwtHandler.JWTMiddleware, handlers.Authorization("admin", "teacher", "student"))
	gptGroup.Post("/generate-question", gptHandlers.GenerateQuestion)
	gptGroup.Post("/generate-response", gptHandlers.GenerateResponse)
	gptGroup.Post("/generate-image", gptHandlers.GenerateImage)

	// Routes for upload files.
	upload.Post("/", jwtHandler.JWTMiddleware, uploadHandler.UploadFiles)
	upload.Static("/", "./uploads")
	upload.Post("/google", jwtHandler.JWTMiddleware, uploadHandler.UploadFileToGoogle)

	classesHandler := handlers.NewClassesHandler(config)
	classesGroup := api.Group("/classes", jwtHandler.JWTMiddleware)
	classesGroup.Post("/", handlers.Authorization("teacher", "admin"), classesHandler.NewClasses)
	classesGroup.Put("/:id", handlers.Authorization("teacher", "admin"), classesHandler.UpdateClassByID)
	classesGroup.Put("/:id/archive", handlers.Authorization("teacher", "admin"), classesHandler.ArchiveClassByID)
	classesGroup.Post("/subscribe", handlers.Authorization("student"), classesHandler.SuscribeClass)
	classesGroup.Delete("/:id/unsubscribe", handlers.Authorization("student"), classesHandler.UnsubscribeClass)
	classesGroup.Get("/subscribed", handlers.Authorization("student"), classesHandler.GetClassesSubscribedByStudent)
	classesGroup.Get("/:id/students", classesHandler.GetStudentsByClass)
	api.Get("/professors/:id/classes", jwtHandler.JWTMiddleware, handlers.Authorization("teacher"), classesHandler.GetClassesByTeacher)
	api.Get("/professors/:id/classes/archived", jwtHandler.JWTMiddleware, handlers.Authorization("teacher"), classesHandler.GetClassesArchivedByTeacher)

	go services.TelegramBot(config)
	err = app.Listen(":" + config.GetString("PORT"))
	if err != nil {
		log.Println(err)
	}
}
