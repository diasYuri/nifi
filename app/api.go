package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

type Item struct {
	ID    int     `json:"id"`
	Nome  string  `json:"nome"`
	Preco float64 `json:"preco"`
}

func main() {
	http.HandleFunc("/itens-do-pedido", itensDoPedidoHandler)
	log.Fatal(http.ListenAndServe(":8091", nil))
}

func itensDoPedidoHandler(w http.ResponseWriter, r *http.Request) {
	// Extrair o id do pedido da query parameter
	idPedido := r.URL.Query().Get("id")
	if idPedido == "" {
		http.Error(w, "É necessário fornecer o ID do pedido", http.StatusBadRequest)
		return
	}

	// Gerar um número aleatório de itens entre 1 e o tamanho máximo da lista de itens
	numItens := rand.Intn(10)
	fmt.Printf("Itens do pedido %s", idPedido)

	// Selecionar aleatoriamente os itens da lista de itens do pedido
	itensSelecionados := make([]Item, numItens)
	for i := 0; i < numItens; i++ {
		itensSelecionados[i] = Item{
			ID:    rand.Intn(1000),
			Nome:  fmt.Sprintf("Item %d", rand.Intn(230)),
			Preco: rand.Float64(),
		}
	}

	// Serializar a lista de itens selecionados para JSON e escrever na resposta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(itensSelecionados); err != nil {
		http.Error(w, "Erro ao serializar resposta", http.StatusInternalServerError)
		return
	}
}
