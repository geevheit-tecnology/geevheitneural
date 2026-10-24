package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

// HandleGetCardPreview entrega a interface visual compilada no workspace do cartão
func HandleGetCardPreview(sandbox app.SandboxManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão é obrigatório")
			return
		}

		if sandbox == nil {
			middleware.WriteError(w, http.StatusServiceUnavailable, "sandbox_unavailable", "gerenciador de sandbox não configurado")
			return
		}

		files, err := sandbox.ReadFiles(cardID)
		if err != nil || len(files) == 0 {
			// Resposta visual amigável enquanto o build está sendo gerado
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>GeevheitNeural — Aguardando Build</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen flex items-center justify-center p-6 font-sans">
  <div class="text-center max-w-md bg-slate-900 border border-slate-800 rounded-2xl p-8">
    <div class="w-12 h-12 border-4 border-emerald-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
    <h2 class="text-xl font-bold text-white mb-2">Build em Processamento</h2>
    <p class="text-slate-400 text-sm">O pipeline da Fábrica Autônoma está gerando os arquivos para o cartão <strong>%s</strong>.</p>
  </div>
</body>
</html>`, cardID)))
			return
		}

		// Procura por index.html no workspace
		for _, f := range files {
			if strings.EqualFold(f.Path, "index.html") {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(f.Content))
				return
			}
		}

		// Se não tiver index.html, entrega o primeiro arquivo com visualizador
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(files[0].Content))
	}
}

// HandleGetCardWorkspace retorna os arquivos e árvore de código em JSON
func HandleGetCardWorkspace(sandbox app.SandboxManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão é obrigatório")
			return
		}

		if sandbox == nil {
			middleware.WriteError(w, http.StatusServiceUnavailable, "sandbox_unavailable", "gerenciador de sandbox não configurado")
			return
		}

		files, err := sandbox.ReadFiles(cardID)
		if err != nil {
			middleware.WriteJSON(w, http.StatusOK, map[string]any{
				"card_id": cardID,
				"files":   []app.ProjectFile{},
				"count":   0,
			})
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"card_id": cardID,
			"files":   files,
			"count":   len(files),
		})
	}
}
