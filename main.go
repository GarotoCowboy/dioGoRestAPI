package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Create an object to save user in memory
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Address  string `json:"address"`
}

var (
	users = make(map[string]User)
	mutex = sync.Mutex{}
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/", func(
		rw http.ResponseWriter,
		req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{"message": "Hello World"}`))
		return
	}).Methods("GET", "POST", "PUT", "DELETE")

	//Post handler func
	r.HandleFunc("/api/v1/users", CreateUser).Methods("POST")

	//get Users
	r.HandleFunc("/api/v1/users/{id}", GetUserById).Methods("GET")
	r.HandleFunc("/api/v1/users", ListUsers).Methods("GET")

	//Delete User
	r.HandleFunc("/api/v1/users/{id}", DeleteUser).Methods("DELETE")

	//Update User
	r.HandleFunc("/api/v1/users", EditUser).Methods("PUT")

	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server is running on 8000")
}

// GET USER BY ID IN MEMORY
func GetUserById(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID da URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Converte o ID de string para uint
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido, deve ser um número inteiro positivo", http.StatusBadRequest)
		return
	}

	// Protege o acesso ao mapa
	mutex.Lock()
	user, exists := users[strconv.Itoa(int(uint(id)))]
	mutex.Unlock()

	// Verifica se o usuário existe
	if !exists {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	// Define cabeçalhos e status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Envia a resposta
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Erro ao codificar resposta", http.StatusInternalServerError)
		return
	}
}

// CREATE USERS IN MEMORY
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.ID == 0 {
		http.Error(w, "ID tem que ser maior que 0", http.StatusBadRequest)
		return
	}

	mutex.Lock()

	if _, exists := users[strconv.Itoa(int(user.ID))]; exists {
		mutex.Unlock()
		http.Error(w, "Usuário com este ID já existe", http.StatusConflict)
		return
	}

	users[strconv.Itoa(int(user.ID))] = user
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Erro ao codificar resposta", http.StatusInternalServerError)
		return
	}
}
func EditUser(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID da URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Converte o ID de string para uint
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido, deve ser um número inteiro positivo", http.StatusBadRequest)
		return
	}

	// Decodifica o corpo da requisição
	var user User
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Valida o ID no corpo (deve corresponder ao ID da URL)
	if user.ID != 0 && user.ID != uint(id) {
		http.Error(w, "ID no corpo da requisição não corresponde ao ID na URL", http.StatusBadRequest)
		return
	}

	// Validações dos campos
	if user.Username == "" {
		http.Error(w, "Username é obrigatório", http.StatusBadRequest)
		return
	}
	if user.Password == "" {
		http.Error(w, "Password é obrigatório", http.StatusBadRequest)
		return
	}
	if user.Name == "" {
		http.Error(w, "Name é obrigatório", http.StatusBadRequest)
		return
	}
	// Address é opcional, não requer validação

	// Protege o acesso ao mapa
	mutex.Lock()
	defer mutex.Unlock()

	// Verifica se o usuário existe
	if _, exists := users[strconv.Itoa(int(uint(id)))]; !exists {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	// Atualiza o usuário
	user.ID = uint(id) // Garante que o ID seja o mesmo da URL
	users[strconv.Itoa(int(uint(id)))] = user

	// Define cabeçalhos e status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Envia a resposta
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Erro ao codificar resposta", http.StatusInternalServerError)
		return
	}
}

// LIST USERS IN MEMORY
func ListUsers(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	// Converte o mapa em uma lista para a resposta JSON
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}

	// Define cabeçalhos e status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Envia a resposta
	if err := json.NewEncoder(w).Encode(userList); err != nil {
		http.Error(w, "Erro ao codificar resposta", http.StatusInternalServerError)
		return
	}
}

// DELETE USER IN MEMORY
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID da URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Converte o ID de string para uint
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido, deve ser um número inteiro positivo", http.StatusBadRequest)
		return
	}

	// Protege o acesso ao mapa
	mutex.Lock()
	defer mutex.Unlock()

	// Verifica se o usuário existe
	if _, exists := users[strconv.Itoa(int(uint(id)))]; !exists {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	// Remove o usuário
	delete(users, strconv.Itoa(int(uint(id))))

	// Define status (sem corpo na resposta)
	w.WriteHeader(http.StatusNoContent)
}
