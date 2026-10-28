package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/geevheit/neuralcenter/api/config"
	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	"github.com/geevheit/neuralcenter/api/internal/adapters/auth"
	"github.com/geevheit/neuralcenter/api/internal/adapters/events"
	"github.com/geevheit/neuralcenter/api/internal/adapters/http/handlers"
	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/adapters/notifications"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	localpub "github.com/geevheit/neuralcenter/api/internal/adapters/publisher"
	"github.com/geevheit/neuralcenter/api/internal/adapters/sandbox"
	"github.com/geevheit/neuralcenter/api/internal/adapters/webhooks"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/command"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Configuração e Logger estruturado (slog JSON sem PII)
	cfg := config.Load()

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting_geevheit_api",
		slog.String("port", cfg.Port),
		slog.String("log_level", cfg.LogLevel),
	)

	// 2. Conexão ao PostgreSQL via pgxpool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed_to_parse_database_url", slog.Any("error", err))
		os.Exit(1)
	}

	// Se configurado para Simple Protocol (Render / pooler Supabase porta 6543)
	if cfg.DBSimpleProtocol {
		poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("failed_to_connect_to_postgres", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Warn("database_ping_warning", slog.Any("error", err))
	} else {
		logger.Info("connected_to_postgres_successfully")
	}

	// 3. Composition Root: Injeção de dependência manual (SOLID)
	clk := clock.SystemClock{}
	uow := postgres.NewUnitOfWork(pool)
	publisher := events.NewInMemoryPublisher()
	tokenValidator := auth.NewTokenValidator(cfg.OwnerUserID, cfg.SupabaseJWTSecret)

	q := sqlcgen.New(pool)
	moneyReader := postgres.NewMoneyWaitingReader(q)
	cardRepo := postgres.NewCardRepository(q)
	appRepo := postgres.NewAppRepository(q)
	financeRepo := postgres.NewFinanceRepository(q)
	connectorRepo := postgres.NewConnectorRepository(q)
	agentRepo := postgres.NewAgentRepository(q)

	// Casos de Uso
	decideGateUC := app.NewDecideGateUseCase(uow, publisher, clk)
	pauseSystemUC := app.NewPauseSystemUseCase(uow, publisher, clk)
	resumeSystemUC := app.NewResumeSystemUseCase(uow, publisher, clk)
	getMoneyWaitingUC := app.NewGetMoneyWaitingUseCase(moneyReader)

	listCardsUC := app.NewListCardsUseCase(cardRepo)
	getCardDetailsUC := app.NewGetCardDetailsUseCase(cardRepo)
	deleteCardUC := app.NewDeleteCardUseCase(cardRepo, publisher)

	listAppsUC := app.NewListAppsUseCase(appRepo)
	createAppUC := app.NewCreateAppUseCase(appRepo, publisher)
	updateAppUC := app.NewUpdateAppUseCase(appRepo, publisher)
	deleteAppUC := app.NewDeleteAppUseCase(appRepo, publisher)
	appHandlers := handlers.NewAppHandlers(listAppsUC, createAppUC, updateAppUC, deleteAppUC)

	overviewFinancesUC := app.NewGetFinancesOverviewUseCase(financeRepo)
	createLedgerEntryUC := app.NewCreateLedgerEntryUseCase(financeRepo, publisher)
	financeHandlers := handlers.NewFinanceHandlers(overviewFinancesUC, createLedgerEntryUC)

	listConnectorsUC := app.NewListConnectorsUseCase(connectorRepo)

	// Fase 3: Comandos com confirmação de ações destrutivas e Ficha de Risco
	confirmSvc := command.NewConfirmationService(5 * time.Minute)
	processCmdUC := app.NewProcessCommandUseCase(cardRepo, moneyReader, financeRepo, decideGateUC, pauseSystemUC, resumeSystemUC, confirmSvc)
	confirmActionUC := app.NewConfirmActionUseCase(confirmSvc, decideGateUC, pauseSystemUC, deleteCardUC)
	commandHandlers := handlers.NewCommandHandlers(processCmdUC, confirmActionUC)

	getRiskUC := app.NewGetRiskReportUseCase(cardRepo)
	saveRiskUC := app.NewSaveRiskReportUseCase(cardRepo, publisher)
	riskHandlers := handlers.NewRiskHandlers(getRiskUC, saveRiskUC)

	// Fase 4: Comercial, Intake Público, Propostas com Aceite e LGPD
	intakeUC := app.NewIntakeUseCase(uow, publisher)
	intakeHandler := handlers.NewIntakeHandler(intakeUC)

	proposalUC := app.NewProposalUseCases(uow, publisher, cfg.BaseURL)
	proposalHandler := handlers.NewProposalHandler(proposalUC)

	clientLGPD_UC := app.NewClientLGPDUseCases(uow)
	clientLGPDHandler := handlers.NewClientLGPDHandler(clientLGPD_UC)

	// Fase 5: Contratos dos Agentes de IA com ExecutionGuard
	agentUC := app.NewAgentUseCases(uow, publisher)
	agentHandler := handlers.NewAgentHandler(agentUC)

	// Sprint 02: Fases 6 a 10 (Fábrica Autônoma, Pipelines, Playbooks e Métricas)
	if err := app.SeedPlaybooksAndModules(context.Background(), uow); err != nil {
		logger.Error("seed_playbooks_error", slog.Any("error", err))
	}
	// Sprint 02 e Fase 16 (Fábrica Autônoma, Sandboxes e Provedor Gemini Free Tier)
	localSandbox, err := sandbox.NewLocalSandboxManager("")
	if err != nil {
		logger.Error("sandbox_init_failed", slog.Any("error", err))
	}
	localPublisher := localpub.NewLocalPublisher(cfg.BaseURL, localSandbox)

	var agentRunner pipeline.AgentRunner = agents.NewMockAgentRunner()
	geminiKey := os.Getenv("GEMINI_API_KEY")
	groqKey := os.Getenv("GROQ_API_KEY")
	runnerType := strings.ToLower(os.Getenv("AGENT_RUNNER_TYPE"))

	if (runnerType == "groq" || runnerType == "") && groqKey != "" {
		logger.Info("autonomous_factory_runner_enabled", slog.String("provider", "groq_ultra_fast_free_tier"), slog.String("model", os.Getenv("GROQ_MODEL")))
		agentRunner = agents.NewGroqAgentRunner(agents.GroqConfig{
			APIKey: groqKey,
			Model:  os.Getenv("GROQ_MODEL"),
		})
	} else if (runnerType == "gemini" || runnerType == "") && geminiKey != "" {
		logger.Info("autonomous_factory_runner_enabled", slog.String("provider", "google_gemini_free_tier"))
		agentRunner = agents.NewGeminiAgentRunner(agents.GeminiConfig{
			APIKey: geminiKey,
			Model:  os.Getenv("GEMINI_MODEL"),
		})
	}
	guard := system.NewExecutionGuard()

	startBuildUC := app.NewStartBuildUseCase(uow, publisher, guard, agentRunner, localSandbox, localPublisher)
	getPipelineUC := app.NewGetCardPipelineUseCase(uow)

	getPublicScopeUC := app.NewGetPublicScopeUseCase(uow)
	approvePublicScopeUC := app.NewApprovePublicScopeUseCase(uow, publisher)

	listPlaybooksUC := app.NewListPlaybooksUseCase(uow)
	listModulesUC := app.NewListModulesUseCase(uow)
	getScoreboardUC := app.NewGetFactoryScoreboardUseCase(uow)
	estimatePriceUC := app.NewEstimatePriceUseCase(uow)

	// Sprint 03: Fase 11 - Motor de Pagamentos PIX e Webhooks
	generatePixUC := app.NewGeneratePixPaymentUseCase(uow, publisher, clk)
	getPixUC := app.NewGetPixPaymentUseCase(uow)
	processPixWebhookUC := app.NewProcessPixWebhookUseCase(uow, publisher, clk)

	// Sprint 03: Fase 13 - Portal Exclusivo do Cliente
	portalUC := app.NewPortalUseCases(uow, publisher, cfg.BaseURL)
	portalHandler := handlers.NewPortalHandler(portalUC)

	// Sprint 03: Fase 14 - Gestão de Mensalidades (MRR) e Recorrência
	subscriptionUC := app.NewSubscriptionsUseCases(uow, publisher)
	subscriptionHandlers := handlers.NewSubscriptionHandlers(subscriptionUC)

	// Sprint 03: Fase 15 - Notificações de Negócio e Webhooks
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")
	whatsAppURL := os.Getenv("WHATSAPP_API_URL")
	whatsAppToken := os.Getenv("WHATSAPP_API_TOKEN")
	multiNotifier := notifications.NewMultiNotifier(logger, telegramToken, telegramChatID, whatsAppURL, whatsAppToken)
	webhookDispatcher := webhooks.NewDispatcher(logger, postgres.NewWebhookRepository(sqlcgen.New(pool)), false)
	notificationWebhooksUC := app.NewNotificationWebhooksUseCases(uow, multiNotifier, webhookDispatcher, false)
	notificationWebhooksHandlers := handlers.NewNotificationWebhookHandlers(notificationWebhooksUC)

	// Listener em background para despachar eventos de negócio a notificações e webhooks
	go func() {
		cardCh := publisher.Subscribe(app.TopicCardsChanged)
		financeCh := publisher.Subscribe(app.TopicFinancesChanged)
		for {
			select {
			case payload, ok := <-cardCh:
				if !ok {
					return
				}
				if m, ok := payload.(map[string]any); ok {
					if evt, ok := m["event"].(string); ok && evt == "portal_delivery_accepted" {
						notificationWebhooksUC.DispatchBusinessEvent(context.Background(), "portal.delivery_accepted", "Aceite Formal Recebido", "Cliente aceitou formalmente a entrega do projeto", m)
					}
				}
			case payload, ok := <-financeCh:
				if !ok {
					return
				}
				if m, ok := payload.(map[string]any); ok {
					if evt, ok := m["event"].(string); ok && evt == "pix_payment_confirmed" {
						notificationWebhooksUC.DispatchBusinessEvent(context.Background(), "payment.pix_confirmed", "Pagamento PIX Confirmado", "Pagamento de cliente liquidado via PIX", m)
					}
				}
			}
		}
	}()

	// 4. Roteamento (net/http padrão do Go 1.22+)
	mux := http.NewServeMux()

	// Endpoints públicos de infraestrutura, intake/propostas/escopo, portal do cliente e webhook PIX
	mux.HandleFunc("GET /healthz", handlers.HandleHealthz)
	mux.HandleFunc("GET /readyz", handlers.HandleReadyz(pool))
	mux.HandleFunc("POST /v1/intake", intakeHandler.SubmitIntake)
	mux.HandleFunc("GET /v1/public/proposals/{token}", proposalHandler.GetPublicProposal)
	mux.HandleFunc("POST /v1/public/proposals/{token}/accept", proposalHandler.AcceptPublicProposal)
	mux.HandleFunc("GET /v1/public/scope/{token}", handlers.HandleGetPublicScope(getPublicScopeUC))
	mux.HandleFunc("POST /v1/public/scope/{token}/approve", handlers.HandleApprovePublicScope(approvePublicScopeUC))
	mux.HandleFunc("GET /v1/public/portal/{token}", portalHandler.GetPublicPortal)
	mux.HandleFunc("POST /v1/public/portal/{token}/accept", portalHandler.AcceptPublicDelivery)
	mux.HandleFunc("POST /v1/payments/pix/webhook", handlers.HandlePixWebhook(processPixWebhookUC, ""))

	// Endpoints restritos para Agentes de IA (validação de token via hash + ExecutionGuard)
	requireAgentAuth := middleware.RequireAgentAuth(agentRepo)
	mux.Handle("GET /v1/agent/cards", requireAgentAuth(http.HandlerFunc(agentHandler.ListCards)))
	mux.Handle("GET /v1/agent/cards/{id}", requireAgentAuth(http.HandlerFunc(agentHandler.GetCard)))
	mux.Handle("POST /v1/agent/cards/{id}/progress", requireAgentAuth(http.HandlerFunc(agentHandler.RecordProgress)))
	mux.Handle("POST /v1/agent/cards/{id}/submit-delivery", requireAgentAuth(http.HandlerFunc(agentHandler.SubmitDelivery)))

	// Endpoints protegidos pelo dono com 2FA (AAL2)
	requireAuth := middleware.RequireOwnerAuth(tokenValidator)

	// Decisões e Sistema
	mux.Handle("POST /v1/cards/{id}/decision", requireAuth(handlers.HandleDecideGate(decideGateUC)))
	mux.Handle("POST /v1/system/pause", requireAuth(handlers.HandlePauseSystem(pauseSystemUC)))
	mux.Handle("POST /v1/system/resume", requireAuth(handlers.HandleResumeSystem(resumeSystemUC)))
	mux.Handle("GET /v1/money-waiting", requireAuth(handlers.HandleGetMoneyWaiting(getMoneyWaitingUC)))

	// Comandos e Ações Destrutivas (F3-01, F3-02, F3-03)
	mux.Handle("POST /v1/commands", requireAuth(commandHandlers.HandleProcessCommand()))
	mux.Handle("POST /v1/commands/confirm", requireAuth(commandHandlers.HandleConfirmAction()))

	// Ficha de Risco (F3-04)
	mux.Handle("GET /v1/cards/{id}/risk", requireAuth(riskHandlers.Get()))
	mux.Handle("PUT /v1/cards/{id}/risk", requireAuth(riskHandlers.Save()))

	// Propostas e Clientes / LGPD (F4-02, F4-04, F12)
	mux.Handle("POST /v1/proposals", requireAuth(http.HandlerFunc(proposalHandler.CreateProposal)))
	mux.Handle("GET /v1/cards/{id}/proposal-document", requireAuth(http.HandlerFunc(proposalHandler.GetProposalDocument)))
	mux.Handle("POST /v1/cards/{id}/proposal-document", requireAuth(http.HandlerFunc(proposalHandler.GetProposalDocument)))
	mux.Handle("GET /v1/clients/{id}/export", requireAuth(http.HandlerFunc(clientLGPDHandler.ExportClient)))
	mux.Handle("DELETE /v1/clients/{id}/purge", requireAuth(http.HandlerFunc(clientLGPDHandler.PurgeClient)))

	// Cartões (Hoje e Quadro Kanban)
	mux.Handle("GET /v1/cards", requireAuth(handlers.HandleListCards(listCardsUC)))
	mux.Handle("GET /v1/cards/{id}", requireAuth(handlers.HandleGetCardDetails(getCardDetailsUC)))
	mux.Handle("DELETE /v1/cards/{id}", requireAuth(handlers.HandleDeleteCard(deleteCardUC)))

	// Portal do Cliente (Fase 13)
	mux.Handle("POST /v1/cards/{id}/portal-link", requireAuth(http.HandlerFunc(portalHandler.GeneratePortalLink)))
	mux.Handle("DELETE /v1/cards/{id}/portal-link", requireAuth(http.HandlerFunc(portalHandler.RevokePortalLink)))
	mux.Handle("GET /v1/cards/{id}/portal-link", requireAuth(http.HandlerFunc(portalHandler.GetPortalStatus)))

	// Meus Apps
	mux.Handle("GET /v1/apps", requireAuth(appHandlers.List()))
	mux.Handle("POST /v1/apps", requireAuth(appHandlers.Create()))
	mux.Handle("PUT /v1/apps/{id}", requireAuth(appHandlers.Update()))
	mux.Handle("DELETE /v1/apps/{id}", requireAuth(appHandlers.Delete()))

	// Finanças
	mux.Handle("GET /v1/finances", requireAuth(financeHandlers.Overview()))
	mux.Handle("POST /v1/finances/entries", requireAuth(financeHandlers.CreateEntry()))
	mux.Handle("GET /v1/finances/mrr", requireAuth(subscriptionHandlers.GetMRRMetrics()))
	mux.Handle("GET /v1/finances/charges", requireAuth(subscriptionHandlers.ListCharges()))
	mux.Handle("POST /v1/finances/charges/{id}/pay", requireAuth(subscriptionHandlers.PayCharge()))

	// Gestão de Mensalidades (MRR) e Recorrência (Fase 14)
	mux.Handle("POST /v1/subscriptions", requireAuth(subscriptionHandlers.CreateSubscription()))
	mux.Handle("GET /v1/subscriptions", requireAuth(subscriptionHandlers.ListSubscriptions()))
	mux.Handle("PATCH /v1/subscriptions/{id}/status", requireAuth(subscriptionHandlers.UpdateStatus()))
	mux.Handle("POST /v1/subscriptions/process-daily", requireAuth(subscriptionHandlers.ProcessDaily()))

	// Webhooks de Saída e Notificações de Negócio (Fase 15)
	mux.Handle("GET /v1/webhooks", requireAuth(notificationWebhooksHandlers.ListEndpoints()))
	mux.Handle("POST /v1/webhooks", requireAuth(notificationWebhooksHandlers.CreateEndpoint()))
	mux.Handle("DELETE /v1/webhooks/{id}", requireAuth(notificationWebhooksHandlers.DeleteEndpoint()))
	mux.Handle("GET /v1/webhooks/deliveries", requireAuth(notificationWebhooksHandlers.ListDeliveries()))
	mux.Handle("POST /v1/webhooks/{id}/redeliver", requireAuth(notificationWebhooksHandlers.Redeliver()))
	mux.Handle("GET /v1/notifications/preferences", requireAuth(notificationWebhooksHandlers.GetPreferences()))
	mux.Handle("PUT /v1/notifications/preferences", requireAuth(notificationWebhooksHandlers.SavePreferences()))

	// Conexões
	mux.Handle("GET /v1/connectors", requireAuth(handlers.HandleListConnectors(listConnectorsUC)))

	// Fábrica Autônoma, Pipelines, Playbooks e Placar (Sprint 02: Fases 6 a 10)
	mux.Handle("POST /v1/cards/{id}/pipeline/start", requireAuth(handlers.HandleStartBuild(startBuildUC)))
	mux.Handle("GET /v1/cards/{id}/pipeline", requireAuth(handlers.HandleGetCardPipeline(getPipelineUC)))
	mux.Handle("GET /v1/playbooks", requireAuth(handlers.HandleListPlaybooks(listPlaybooksUC)))
	mux.Handle("GET /v1/modules", requireAuth(handlers.HandleListModules(listModulesUC)))
	mux.Handle("GET /v1/factory/scoreboard", requireAuth(handlers.HandleGetFactoryScoreboard(getScoreboardUC)))
	mux.Handle("POST /v1/factory/estimate", requireAuth(handlers.HandleEstimatePrice(estimatePriceUC)))
	mux.HandleFunc("GET /v1/cards/{id}/preview", handlers.HandleGetCardPreview(localSandbox))
	mux.Handle("GET /v1/cards/{id}/workspace", requireAuth(handlers.HandleGetCardWorkspace(localSandbox)))


	// Cobrança PIX (Fase 11)
	mux.Handle("POST /v1/cards/{id}/pix", requireAuth(handlers.HandleGeneratePix(generatePixUC, cfg.OwnerUserID)))
	mux.Handle("GET /v1/cards/{id}/pix", requireAuth(handlers.HandleGetPix(getPixUC)))

	// SSE Stream em tempo real
	mux.Handle("GET /v1/stream", requireAuth(handlers.HandleSSEStream(publisher)))

	// Middlewares globais aplicados na ordem correta
	var handler http.Handler = mux
	handler = middleware.CORS(cfg.CORSAllowedOrigins)(handler)
	handler = middleware.Logging(logger)(handler)
	handler = middleware.Recover(logger)(handler)
	handler = middleware.RequestID(handler)

	// 5. Servidor HTTP com timeouts e Graceful Shutdown
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server_listening", slog.String("addr", server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	// Captura sinais de interrupção do sistema operacional
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server_fatal_error", slog.Any("error", err))
		}

	case sig := <-shutdown:
		logger.Info("shutdown_signal_received", slog.String("signal", sig.String()))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful_shutdown_failed", slog.Any("error", err))
			_ = server.Close()
		}
		logger.Info("server_stopped_gracefully")
	}
}
