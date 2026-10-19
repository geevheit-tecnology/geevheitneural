package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/events"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleSSEStream(pub *events.InMemoryPublisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming não suportado", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		rc := http.NewResponseController(w)
		_ = rc.SetWriteDeadline(time.Time{})

		// Mensagem inicial de conexão estabelecida
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\",\"time\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
		flusher.Flush()

		// Inscrição nos tópicos
		topics := []string{
			app.TopicMoneyWaitingChanged,
			app.TopicSystemStateChanged,
			app.TopicCardsChanged,
			app.TopicAppsChanged,
			app.TopicFinancesChanged,
		}

		channels := make(map[string]<-chan any, len(topics))
		for _, topic := range topics {
			channels[topic] = pub.Subscribe(topic)
		}

		defer func() {
			for topic, ch := range channels {
				pub.Unsubscribe(topic, ch)
			}
		}()

		ticker := time.NewTicker(25 * time.Second) // heartbeat para manter a conexão aberta em proxies como Cloudflare/Render
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return

			case <-ticker.C:
				fmt.Fprint(w, ": heartbeat\n\n")
				flusher.Flush()

			case payload := <-channels[app.TopicMoneyWaitingChanged]:
				sendSSE(w, flusher, "money_waiting_changed", payload)

			case payload := <-channels[app.TopicSystemStateChanged]:
				sendSSE(w, flusher, "system_state_changed", payload)

			case payload := <-channels[app.TopicCardsChanged]:
				sendSSE(w, flusher, "cards_changed", payload)

			case payload := <-channels[app.TopicAppsChanged]:
				sendSSE(w, flusher, "apps_changed", payload)

			case payload := <-channels[app.TopicFinancesChanged]:
				sendSSE(w, flusher, "finances_changed", payload)
			}
		}
	}
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, event string, payload any) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(bytes))
	flusher.Flush()
}
