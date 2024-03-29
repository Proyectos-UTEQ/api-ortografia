package handlers

import (
	"Proyectos-UTEQ/api-ortografia/internal/data"
	"Proyectos-UTEQ/api-ortografia/internal/utils"
	"Proyectos-UTEQ/api-ortografia/pkg/types"
	"bufio"
	"fmt"
	"github.com/blackestwhite/gopenai"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

type ModuleHandler struct {
	config *viper.Viper
}

// NewModuleHandler crea un nuevo handler de modules.
func NewModuleHandler(config *viper.Viper) *ModuleHandler {
	return &ModuleHandler{
		config: config,
	}
}

func (h *ModuleHandler) CreateModuleForTeacher(c *fiber.Ctx) error {
	// recuperamos los claims del usuarios
	claims := utils.GetClaims(c)

	// Parseamos el body
	var module types.Module
	if err := c.BodyParser(&module); err != nil {
		log.Println("Error al registrar modulo", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear los datos",
		})
	}

	// Validar datos para registro inicial.
	resp, err := types.Validate(&module)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error en la validación de datos",
			"data":    resp,
		})
	}

	// Crea el modulo en la base de datos y recuperamos los datos del usuario
	// que creo el modulo.
	moduleResponse, err := data.RegisterModuleForTeacher(&module, claims.UserAPI.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al registrar modulo",
			"data":    err,
		})
	}

	// Generamos la url de la imagen del módulo.
	moduleResponse.ImgBackURL = h.config.GetString("APP_HOST") + moduleResponse.ImgBackURL

	return c.Status(fiber.StatusCreated).JSON(moduleResponse)

}

// UpdateModule Actualiza el modulo en la base de datos.
func (h *ModuleHandler) UpdateModule(c *fiber.Ctx) error {

	// claims := utils.GetClaims(c)

	idModule := c.Params("id")
	if idModule == "" {
		log.Println("Error al registrar modulo")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al registrar modulo",
		})
	}

	// convertir a uint el id module
	id, err := strconv.Atoi(idModule)
	if err != nil {
		log.Println("Error al registrar modulo", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al registrar modulo",
		})
	}

	var module types.Module
	// Parseamos el body
	if err := c.BodyParser(&module); err != nil {
		log.Println("Error al registrar modulo", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear los datos",
		})
	}

	// establecemos el ID del módulo
	module.ID = uint(id)

	// Validar datos.
	resp, err := types.Validate(&module)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error en la validación de datos",
			"data":    resp,
		})
	}

	// Actualizamos el módulo en la db
	moduleData, err := data.UpdateModule(&module)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al registrar modulo",
			"data":    err,
		})
	}

	moduleResponse := data.ModuleToApi(*moduleData)

	return c.Status(fiber.StatusOK).JSON(moduleResponse)
}

// GetModulesForTeacher obtiene todos los modules para un teacher
func (h *ModuleHandler) GetModulesForTeacher(c *fiber.Ctx) error {

	claims := utils.GetClaims(c)

	// campos para paginar
	var paginated types.Paginated

	if err := c.QueryParser(&paginated); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear los datos",
			"data":    err,
		})
	}

	// validamos
	_ = paginated.Validate()

	// obtenemos los modules
	modules, details, err := data.GetModulesForTeacher(&paginated, claims.UserAPI.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    err,
		})
	}

	modulesApi := data.ModulesToAPI(modules)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    modulesApi,
		"details": details,
	})
}

func (h *ModuleHandler) GetModules(c *fiber.Ctx) error {

	var paginated types.Paginated

	if err := c.QueryParser(&paginated); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear los datos",
			"data":    err,
		})
	}

	// validamos
	_ = paginated.Validate()

	// obtenemos los modules
	modules, details, err := data.GetModule(&paginated)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    err,
		})
	}

	modulesApi := data.ModulesToAPI(modules)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    modulesApi,
		"details": details,
	})
}

func (h *ModuleHandler) GetModuleWithIsSubscribed(c *fiber.Ctx) error {

	claims := utils.GetClaims(c)

	var paginated types.Paginated

	if err := c.QueryParser(&paginated); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear los datos",
			"data":    err,
		})
	}

	// validamos
	_ = paginated.Validate()

	// obtenemos los modules
	modules, details, err := data.GetModuleWithUserSubscription(&paginated, claims.UserAPI.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    err,
		})
	}

	modulesApi := data.ModuleUserSubToApi(modules)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    modulesApi,
		"details": details,
	})

}

// Subscribe un usuario se podrá suscribir a un modulo
func (h *ModuleHandler) Subscribe(c *fiber.Ctx) error {

	claims := utils.GetClaims(c)

	var req types.ReqSubscription
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear los datos",
			"data":    err,
		})
	}

	// validamos
	err := req.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    err,
		})
	}

	// creamos la subscription
	_, err = data.RegisterSubscription(claims.UserAPI.ID, req.Code)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
	})

}

// Subscriptions recupera todas las subscripciones de un usuario
func (h *ModuleHandler) Subscriptions(c *fiber.Ctx) error {

	var paginated types.Paginated

	if err := c.QueryParser(&paginated); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear los datos",
			"data":    err,
		})
	}

	// validamos
	_ = paginated.Validate()

	claims := utils.GetClaims(c)

	// obtenemos los modules
	modules, details, err := data.GetModuleForStudent(&paginated, claims.UserAPI.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    err,
		})
	}

	modulesApi := data.ModulesToAPI(modules)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    modulesApi,
		"details": details,
	})
}

func (h *ModuleHandler) GetStudents(c *fiber.Ctx) error {
	idModule := c.Params("id")

	// Convertir el idModule a uint
	id, err := strconv.Atoi(idModule)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear el id del modulo",
		})
	}

	// Obtener los estudiantes del módulo
	students, err := data.GetStudentsByModule(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	studentsData := data.UsersToAPI(students)

	return c.Status(fiber.StatusOK).JSON(studentsData)
}

func (h *ModuleHandler) GetModuleByID(c *fiber.Ctx) error {
	idModule := c.Params("id")

	// Convertir el idModule a uint
	id, err := strconv.Atoi(idModule)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear el id del modulo",
		})
	}

	module, err := data.ModuleByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	moduleResponse := data.ModuleToApi(*module)

	return c.JSON(moduleResponse)
}

func (h *ModuleHandler) GenerateTest(c *fiber.Ctx) error {

	claims := utils.GetClaims(c)

	idModule, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear el id del modulo",
		})
	}

	testId, err := data.GenerateTestForStudent(claims.UserAPI.ID, uint(idModule))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"testId": testId,
	})
}

func (h *ModuleHandler) GetTestByID(c *fiber.Ctx) error {

	testId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al parsear el id del test",
		})
	}

	test, err := data.TestByID(uint(testId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(test)

}

func (h *ModuleHandler) GetFeedbackAnswerUser(c *fiber.Ctx) error {

	// La peticion sera en streaming.
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// Recuperar los datos de la pregunta respuesta del usuario y respuesta correcta.
	answerUserID, err := c.ParamsInt("answer_user_id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var answer types.Answer
	if err := c.BodyParser(&answer); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	answerUser, err := data.GetAnswerUserByID(uint(answerUserID))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Pasamos las respuestas del usuario al struct AnswerUser
	answerUser.Answer.TrueOrFalse = answer.TrueOrFalse
	answerUser.Answer.TextOptions = answer.TextOptions
	answerUser.Answer.TextToComplete = answer.TextToComplete

	contentQuestion := ""
	switch answerUser.Question.TypeQuestion {
	case types.QuestionTypeTrueOrFalse:
		contentQuestion = fmt.Sprintf("Necesito que me des retroalimentación para la pregunta, respuesta correcta y la respuesta del estudiante, a continuación te dejo los datos. Pregunta: %s. Respuesta correcta: %t. Respuesta del estudiante: %t", answerUser.Question.TextRoot, answerUser.Question.CorrectAnswer.TrueOrFalse, answerUser.Answer.TrueOrFalse)
	case types.QuestionTypeMultiChoiceText:
		contentQuestion = fmt.Sprintf("Necesito que me des retroalimentación para la pregunta, respuesta correcta y la respuesta del estudiante, a continuación te dejo los datos. Pregunta: %s. Respuesta correcta: %s. Respuesta del estudiante: %s", answerUser.Question.TextRoot, answerUser.Question.CorrectAnswer.TextOptions, answerUser.Answer.TextOptions)
	case types.QuestionTypeCompleteWord:
		contentQuestion = fmt.Sprintf("Necesito que me des retroalimentación para la pregunta de completación, respuesta correcta y la respuesta del estudiante, a continuación te dejo los datos. Pregunta: %s. Respuesta correcta: %s. Respuesta del estudiante: %s", answerUser.Question.TextRoot, answerUser.Question.CorrectAnswer.TextToComplete, answerUser.Answer.TextToComplete)
	case types.QuestionTypeOrderWord:
		contentQuestion = fmt.Sprintf("Necesito que me des retroalimentación para la pregunta de orden de palabras, respuesta correcta y la respuesta del estudiante, a continuación te dejo los datos. Pregunta: %s. Respuesta correcta: %s. Respuesta del estudiante: %s", answerUser.Question.TextRoot, answerUser.Question.CorrectAnswer.TextOptions, answerUser.Answer.TextOptions)
	default:
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// cliente de GPT
	client := gopenai.Setup(h.config.GetString("APP_OPENAI_API_KEY"))
	p := gopenai.ChatCompletionRequestBody{
		Model: "gpt-3.5-turbo",
		Messages: []gopenai.Message{
			{
				Role:    "system",
				Content: "Eres un asistente para estudiante de escuela, donde los estudiantes están aprendiendo ortografía. La respuestas que me debes que dar debe solo tener entre 150 a 250 caracteres.",
			},
			{
				Role:    "user",
				Content: contentQuestion,
			},
		},
		Stream: true,
	}

	// Crea el canal
	resultCh, err := client.GenerateChatCompletion(p)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		respuesta := ""
		// Recuperamos los datos de GPT para enviarlos al cliente.
		for chunk := range resultCh {
			msg := fmt.Sprintf("%s", chunk.Choices[0].Delta.Content)
			fmt.Fprintf(w, "data: %s\n\n", msg)
			respuesta += msg
			err := w.Flush()
			if err != nil {
				fmt.Println(err)
				break
			}

		}

		fmt.Println("respuesta: ", respuesta)

	})

	return nil
}

func (h *ModuleHandler) ValidationAnswerForTestModule(c *fiber.Ctx) error {
	idQuestion, err := c.ParamsInt("answer_user_id")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "error",
			"error":   "Error al recuperar el id de la pregunta",
		})
	}

	var answer types.Answer
	if err := c.BodyParser(&answer); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "error",
			"error":   err.Error(),
		})
	}

	// Evaluar la respuesta del estudiante.
	// Recuperar la answer_user que esta en la base de datos.
	answerUserDB, err := data.GetAnswerUserByID(uint(idQuestion))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "error",
			"error":   err.Error(),
		})
	}

	// establecemos la nueva respuesta del estudiante.
	answerUserDB.Answer.TrueOrFalse = answer.TrueOrFalse
	answerUserDB.Answer.TextOptions = answer.TextOptions
	answerUserDB.Answer.TextToComplete = answer.TextToComplete

	// Evaluación de la pregunta.
	answerUserDB.Responded = true
	switch answerUserDB.Question.TypeQuestion {
	case "true_or_false":
		answerUserDB.IsCorrect = answerUserDB.Question.CorrectAnswer.TrueOrFalse == answerUserDB.Answer.TrueOrFalse
		if answerUserDB.IsCorrect {
			answerUserDB.Score = 10
			answerUserDB.Feedback = "Sigue adelante"
		} else {
			answerUserDB.Score = 0
			answerUserDB.Feedback = "Respuesta incorrecta"
		}
	case "multi_choice_text":
		// caso es multi_chose_text la respuesta viene por TextOpciones.
		if answerUserDB.Question.Options.SelectMode == "single" {
			// si no tiene ninguna opción selection automáticamente es incorrecta
			if len(answerUserDB.Answer.TextOptions) < 1 {
				answerUserDB.IsCorrect = false
				answerUserDB.Score = 0
				answerUserDB.Feedback = "Respuesta incorrecta"
			} else {
				answerUserDB.IsCorrect = utils.ContainsString(answerUserDB.Question.CorrectAnswer.TextOptions, answerUserDB.Answer.TextOptions[0])
				if answerUserDB.IsCorrect {
					answerUserDB.Feedback = "Sigue adelante"
					answerUserDB.Score = 10
				} else {
					answerUserDB.Feedback = "Respuesta incorrecta"
					answerUserDB.Score = 0
				}
			}
		} else {
			// en caso de ser multiple

			points := 0
			// en caso de ser multiple selección se evalúa la respuesta
			for _, correctAnswer := range answerUserDB.Question.CorrectAnswer.TextOptions {
				if utils.ContainsString(answerUserDB.Answer.TextOptions, correctAnswer) {
					points++
				}
			}

			answerUserDB.IsCorrect = points == len(answerUserDB.Question.CorrectAnswer.TextOptions)
			// se calcula el puntaje
			pointsForEachCorrectAnswer := 10 / float32(len(answerUserDB.Question.CorrectAnswer.TextOptions))
			answerUserDB.Score = pointsForEachCorrectAnswer * float32(points)

			if points == 0 {
				answerUserDB.Feedback = "Respuesta incorrecta"
			} else if points < len(answerUserDB.Question.CorrectAnswer.TextOptions) {
				count := len(answerUserDB.Question.CorrectAnswer.TextOptions) - points
				answerUserDB.Feedback = fmt.Sprintf("Te faltó seleccionar %d", count)
			} else {
				answerUserDB.Feedback = "Sigue adelante"
			}
			//if answerUserDB.IsCorrect {
			//	answerUserDB.Feedback = "Sigue adelante"
			//} else {
			//	answerUserDB.Feedback = "Respuesta incorrecta"
			//}
		}

	case "complete_word":

		textToCompleteCorrect := []string(answerUserDB.Question.CorrectAnswer.TextToComplete)

		points := 0

		for _, correctAnswer := range textToCompleteCorrect {
			if utils.ContainsString(answerUserDB.Answer.TextToComplete, correctAnswer) {
				points++
				break
			}
		}

		// Calculamos el puntaje
		answerUserDB.IsCorrect = points == len(textToCompleteCorrect)
		pointsForEachCorrectAnswer := 10 / len(textToCompleteCorrect)
		answerUserDB.Score = float32(pointsForEachCorrectAnswer * points)
		if answerUserDB.IsCorrect {
			answerUserDB.Feedback = "Sigue adelante"
		} else {
			answerUserDB.Feedback = "Respuesta incorrecta"
		}

	case "order_word":
		// analizamos que la respuesta del usuario sea igual que las opciones correctas
		textToCompleteCorrect := []string(answerUserDB.Question.CorrectAnswer.TextOptions)
		textToCompleteUser := []string(answerUserDB.Answer.TextOptions)

		// si el orden está correcto automáticamente es correcta, en caso de que
		// una no sea correcta automáticamente es incorrecta.
		for i, correctAnswer := range textToCompleteCorrect {
			if i >= len(textToCompleteUser) {
				answerUserDB.IsCorrect = false
				answerUserDB.Score = 0
				answerUserDB.Feedback = "Respuesta incorrecta"
				break
			}
			if correctAnswer == textToCompleteUser[i] {
				answerUserDB.IsCorrect = true
				answerUserDB.Score = 10
				answerUserDB.Feedback = "Sigue adelante"
				continue
			} else {
				answerUserDB.IsCorrect = false
				answerUserDB.Score = 0
				answerUserDB.Feedback = "Respuesta incorrecta"
				break
			}
		}

	default:
		answerUserDB.IsCorrect = false
		answerUserDB.Score = 0
		answerUserDB.Feedback = ""
	}

	if !answerUserDB.IsCorrect {
		//err = services.NewGPT(h.config).GenerateFeedbackForQuestion(&answerUserDB)
		//if err != nil {
		//	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		//		"success": "error",
		//		"error":   err.Error(),
		//	})
		//}
		answerUserDB.Feedback = "Respuesta incorrecta"
	} else {
		mensajesMotivadores := []string{
			"¡Buen trabajo!",
			"¡Muy bien!",
			"¡Sigue adelante!",
			"¡Genial!",
		}
		answerUserDB.Feedback = mensajesMotivadores[rand.Intn(len(mensajesMotivadores))]
	}

	// Actualizar cambios en la base de datos.
	err = answerUserDB.UpdateAnswerUser()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "error",
			"error":   err.Error(),
		})
	}

	// retornamos la respuesta del usuario
	answerUserResponse := data.AnswerUserToAPI(answerUserDB)

	return c.Status(fiber.StatusOK).JSON(answerUserResponse)
}

func (h *ModuleHandler) FinishTest(c *fiber.Ctx) error {
	// claims := utils.GetClaims(c)
	testId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "testId not found",
			"error":   err.Error(),
		})
	}

	// Finalizar el test en la base de datos.
	finishTest, err := data.FinishTest(uint(testId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "error",
			"error":   err.Error(),
		})
	}

	return c.JSON(finishTest)
}

// GetMyTestsByModule recupera todos los test de un usuario en un módulo específico.
func (h *ModuleHandler) GetMyTestsByModule(c *fiber.Ctx) error {
	claims := utils.GetClaims(c)
	idModule, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	tests, err := data.GetMyTest(claims.UserAPI.ID, uint(idModule))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	testsAPI := data.TestsModuleToAPI(tests)

	return c.JSON(testsAPI)
}

// GetPointsStudentsForModule recupera todos los puntajes de un usuario con base en todos los modules.
func (h *ModuleHandler) GetPointsStudentsForModule(c *fiber.Ctx) error {

	// Recuperar de los query params las fechas start y end y el límite de elementos.
	startDate := c.Query("start", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.Query("end", time.Now().AddDate(0, 0, 1).Format("2006-01-02")) // un día más para que se incluya la fecha de fin
	limit := c.QueryInt("limit", 10)

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusBadRequest)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Comparar las fechas.
	// La fecha inició debe ser anterior a la fecha fin.
	if start.After(end) {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	lista, err := data.StudentPointsList(start, end, limit)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(lista)
}
