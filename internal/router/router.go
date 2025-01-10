package router

import (
	"github.com/ZnNr/user-reward-controler/internal/handlers"
	"github.com/ZnNr/user-reward-controler/internal/logging"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewRouter создает новый маршрутизатор и регистрирует маршруты.
func NewRouter(
	taskHandler *handlers.TaskHandler,
	userHandler *handlers.UserHandler,
	referralHandler *handlers.ReferralHandler,
	logger *zap.Logger,
) *mux.Router {
	r := mux.NewRouter()

	// Миддлвары для логирования
	r.Use(logging.LoggingMiddleware(logger))

	// Регистрируем маршруты для задач (Tasks)
	r.HandleFunc("/tasks", taskHandler.GetTasks).Methods("GET")                        // Получить все задачи
	r.HandleFunc("/tasks/{id}", taskHandler.GetTaskByID).Methods("GET")                // Получить задачу по ID
	r.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")                     // Создать новую задачу
	r.HandleFunc("/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")                 // Обновить задачу
	r.HandleFunc("/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")              // Удалить задачу
	r.HandleFunc("/tasks/{id}/status", taskHandler.UpdateTaskStatus).Methods("PATCH")  // Обновить статус задачи
	r.HandleFunc("/tasks/{id}/description", taskHandler.GetDescription).Methods("GET") // Получить описание задачи с возможностью пагинации

	// Регистрируем маршруты для пользователей (Users)
	r.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	r.HandleFunc("/users/{user_id}", userHandler.GetUserByID).Methods("GET")
	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/{user_id}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{user_id}", userHandler.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/email", userHandler.GetUserByEmail).Methods("GET")
	r.HandleFunc("/users/{user_id}/balance", userHandler.UpdateBalance).Methods("PUT")
	r.HandleFunc("/users/{user_id}/full-info", userHandler.GetUserFullInfo).Methods("GET")
	r.HandleFunc("/users/{user_id}/summary", userHandler.GetUserSummary).Methods("GET")
	r.HandleFunc("/users/invite", userHandler.InviteUser).Methods("POST")
	r.HandleFunc("/users/leader", userHandler.GetLeaderByBalance).Methods("GET")
	r.HandleFunc("/users/top", userHandler.GetTopUsers).Methods("GET")

	// Регистрируем маршруты для рефералов
	r.HandleFunc("/referrals", referralHandler.GetReferralsByUserID).Methods("GET")      // Изменено на GetReferralsByUserID
	r.HandleFunc("/referrals/{referral_id}", referralHandler.GetReferral).Methods("GET") // Изменено на GetReferral
	r.HandleFunc("/referrals", referralHandler.CreateReferral).Methods("POST")
	r.HandleFunc("/referrals/{referral_id}", referralHandler.UpdateReferral).Methods("PUT")    // Изменено на UpdateReferral
	r.HandleFunc("/referrals/{referral_id}", referralHandler.DeleteReferral).Methods("DELETE") // Изменено на DeleteReferral

	return r
}
