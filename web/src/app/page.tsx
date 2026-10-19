"use client";

import { useEffect, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { supabase } from "@/shared/auth/supabase";
import { api, CardItem, AppItem, FinancesOverview, ConnectorItem, CardDetails } from "@/shared/api-client/endpoints";
import { useSSE } from "@/shared/hooks/useSSE";
import { LoginForm } from "@/features/auth/components/LoginForm";
import { Header } from "@/features/dashboard/components/Header";
import { PausedBanner } from "@/features/dashboard/components/PausedBanner";
import { BottomNav, TabKey } from "@/features/dashboard/components/BottomNav";
import { CommandLine } from "@/features/dashboard/components/CommandLine";
import { HojeTab } from "@/features/dashboard/components/HojeTab";
import { QuadroTab } from "@/features/dashboard/components/QuadroTab";
import { MeusAppsTab } from "@/features/dashboard/components/MeusAppsTab";
import { FinancasTab } from "@/features/dashboard/components/FinancasTab";
import { ConexoesTab } from "@/features/dashboard/components/ConexoesTab";
import { CardDetailSheet } from "@/features/dashboard/components/CardDetailSheet";

// Dados iniciais de demonstração fiéis ao protótipo aprovado
const INITIAL_CARDS: CardItem[] = [
  { id: "1", type: "orcamento", client_id: "c1", client_name: "Clínica Sorriso", app_name: "Novo", title: "Site para clínica odontológica", price_cents: 320000, est_cost_cents: 26000, currency: "BRL", stage: "entrada", priority: 1, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "2", type: "orcamento", client_id: "c2", client_name: "Casa Verde", app_name: "Novo", title: "Painel de estoque para loja", price_cents: 580000, est_cost_cents: 54000, currency: "BRL", stage: "proposta_aprovar", priority: 2, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "3", type: "bug", client_id: "c3", client_name: "Padaria Central", app_name: "pedidos-app", title: "Erro 500 ao salvar pedido", price_cents: 0, est_cost_cents: 3000, currency: "BRL", stage: "entrada", priority: 3, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "4", type: "orcamento", client_id: "c4", client_name: "Prof. Marina", app_name: "Novo", title: "Landing page de curso", price_cents: 190000, est_cost_cents: 14000, currency: "BRL", stage: "aguardando_cliente", priority: 1, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "5", type: "orcamento", client_id: "c5", client_name: "Lima Imóveis", app_name: "Novo", title: "CRM simples para imobiliária", price_cents: 740000, est_cost_cents: 88000, currency: "BRL", stage: "prompt_aprovar", priority: 2, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "6", type: "suporte", client_id: "c6", client_name: "Oficina Torque", app_name: "os-oficina", title: "Migrar app para o novo banco", price_cents: 60000, est_cost_cents: 5000, currency: "BRL", stage: "em_execucao", priority: 1, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "7", type: "financeiro", client_id: "c7", client_name: "Studio Flor", app_name: "agenda-flor", title: "Cobrança vencida há 12 dias", price_cents: 45000, est_cost_cents: 0, currency: "BRL", stage: "entrada", priority: 2, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "8", type: "orcamento", client_id: "c7", client_name: "Studio Flor", app_name: "agenda-flor", title: "Agenda online para salão", price_cents: 280000, est_cost_cents: 31000, currency: "BRL", stage: "entrega_aprovar", priority: 3, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "9", type: "suporte", client_id: "c3", client_name: "Padaria Central", app_name: "site-padaria", title: "Ajuste de logo no rodapé", price_cents: 0, est_cost_cents: 500, currency: "BRL", stage: "entregue", priority: 0, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
];

const INITIAL_APPS: AppItem[] = [
  { id: "a1", client_name: "Padaria Central", name: "pedidos-app", hosting: "Render", database: "Neon", repository: "", domain: "", status: "ativo", monthly_cost_cents: 3800, monthly_revenue_cents: 18000, net_margin_cents: 14200, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a2", client_name: "Padaria Central", name: "site-padaria", hosting: "Netlify", database: "—", repository: "", domain: "", status: "ativo", monthly_cost_cents: 0, monthly_revenue_cents: 9000, net_margin_cents: 9000, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a3", client_name: "Oficina Torque", name: "os-oficina", hosting: "Vercel", database: "Neon", repository: "", domain: "", status: "ativo", monthly_cost_cents: 4500, monthly_revenue_cents: 30000, net_margin_cents: 25500, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a4", client_name: "Studio Flor", name: "agenda-flor", hosting: "Firebase", database: "Firebase", repository: "", domain: "", status: "ativo", monthly_cost_cents: 5200, monthly_revenue_cents: 45000, net_margin_cents: 39800, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a5", client_name: "Casa Verde", name: "estoque-verde", hosting: "Render", database: "Supabase", repository: "", domain: "", status: "ativo", monthly_cost_cents: 7000, monthly_revenue_cents: 0, net_margin_cents: -7000, is_costly: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a6", client_name: "sem cliente", name: "api-antiga", hosting: "Docker (VPS)", database: "Postgres local", repository: "", domain: "", status: "ativo", monthly_cost_cents: 9600, monthly_revenue_cents: 0, net_margin_cents: -9600, is_costly: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
];

export default function HomePage() {
  const [session, setSession] = useState<unknown>(null);
  const [loadingAuth, setLoadingAuth] = useState(true);
  const [activeTab, setActiveTab] = useState<TabKey>("hoje");
  const [isPaused, setIsPaused] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Bottom Sheet
  const [selectedCardDetails, setSelectedCardDetails] = useState<CardDetails | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);

  // Confirmação de Ações Destrutivas (F3-02)
  const [confirmationData, setConfirmationData] = useState<{
    token: string;
    prompt: string;
  } | null>(null);

  const queryClient = useQueryClient();

  // Inicia SSE para atualização em tempo real
  useSSE((paused) => setIsPaused(paused));

  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(session);
      setLoadingAuth(false);
    });

    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      setSession(session);
    });

    return () => subscription.unsubscribe();
  }, []);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  // 1. Dinheiro Parado
  const { data: moneyWaiting } = useQuery({
    queryKey: ["money-waiting"],
    queryFn: () => api.getMoneyWaiting().catch(() => null),
    enabled: !!session,
  });

  // 2. Cartões
  const { data: cardsData } = useQuery({
    queryKey: ["cards"],
    queryFn: () => api.listCards().catch(() => null),
    enabled: !!session,
  });

  const cards: CardItem[] = cardsData?.cards && cardsData.cards.length > 0 ? cardsData.cards : INITIAL_CARDS;

  // 3. Meus Apps
  const { data: appsData } = useQuery({
    queryKey: ["apps"],
    queryFn: () => api.listApps().catch(() => null),
    enabled: !!session,
  });

  const apps: AppItem[] = appsData?.apps && appsData.apps.length > 0 ? appsData.apps : INITIAL_APPS;

  // 4. Finanças
  const { data: financesData } = useQuery({
    queryKey: ["finances"],
    queryFn: () => api.getFinances().catch(() => null),
    enabled: !!session,
  });

  const finances: FinancesOverview = financesData || {
    total_revenue_cents: apps.reduce((acc, a) => acc + a.monthly_revenue_cents, 0),
    total_cost_cents: apps.reduce((acc, a) => acc + a.monthly_cost_cents, 0),
    net_profit_cents: apps.reduce((acc, a) => acc + (a.monthly_revenue_cents - a.monthly_cost_cents), 0),
    costs_by_service: [
      { service_name: "Docker (VPS)", cost_cents: 9600 },
      { service_name: "Render", cost_cents: 10800 },
      { service_name: "Firebase", cost_cents: 5200 },
      { service_name: "Vercel", cost_cents: 4500 },
    ],
    recent_entries: [],
    suggestion: "São 4 serviços diferentes. Consolidar em 2 ou 3 reduz custo e contas para vigiar.",
  };

  // 5. Conectores
  const { data: connectorsData } = useQuery({
    queryKey: ["connectors"],
    queryFn: () => api.listConnectors().catch(() => null),
    enabled: !!session,
  });

  const connectors: ConnectorItem[] = connectorsData?.connectors || [];

  // Mutações
  const pauseMutation = useMutation({
    mutationFn: () => api.pauseSystem(),
    onSuccess: () => {
      setIsPaused(true);
      showToast("Agentes e conexões pausados. Nenhuma execução externa ocorrerá.");
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
    },
  });

  const resumeMutation = useMutation({
    mutationFn: () => api.resumeSystem(),
    onSuccess: () => {
      setIsPaused(false);
      showToast("Sistema retomado. Os conectores seguem desligados até você ligar.");
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
    },
  });

  const decideMutation = useMutation({
    mutationFn: ({ id, decision, note }: { id: string; decision: "aprovado" | "mais_info" | "recusado"; note?: string }) =>
      api.decideGate(id, decision, note),
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: ["cards"] });
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
      setIsSheetOpen(false);
      if (vars.decision === "aprovado") showToast("Portão aprovado e registrado com sucesso!");
      if (vars.decision === "mais_info") showToast("Pedido de informações registrado no histórico.");
      if (vars.decision === "recusado") showToast("Cartão recusado e arquivado com segurança.");
    },
    onError: (err: Error) => {
      showToast(err.message || "Falha ao registrar decisão");
    },
  });

  const startBuildMutation = useMutation({
    mutationFn: (cardId: string) => api.startBuild(cardId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cards"] });
      showToast("⚡ Pipeline de agentes iniciado! Os agentes autônomos estão construindo.");
      setIsSheetOpen(false);
    },
    onError: (err: Error) => {
      showToast(err.message || "Falha ao iniciar pipeline de construção");
    },
  });

  const createAppMutation = useMutation({
    mutationFn: (data: Parameters<typeof api.createApp>[0]) => api.createApp(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      showToast("App cadastrado com sucesso!");
    },
  });

  const deleteAppMutation = useMutation({
    mutationFn: (id: string) => api.deleteApp(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      showToast("App removido.");
    },
  });

  const createFinanceEntryMutation = useMutation({
    mutationFn: (data: Parameters<typeof api.createFinanceEntry>[0]) => api.createFinanceEntry(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["finances"] });
      showToast("Lançamento financeiro registrado!");
    },
  });

  const processCommandMutation = useMutation({
    mutationFn: (text: string) => api.processCommand(text),
    onSuccess: (res) => {
      if (res.requires_confirmation && res.confirmation_token && res.confirmation_prompt) {
        setConfirmationData({
          token: res.confirmation_token,
          prompt: res.confirmation_prompt,
        });
      } else {
        showToast(res.message);
        queryClient.invalidateQueries({ queryKey: ["cards"] });
        queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
      }
    },
    onError: (err: Error) => {
      showToast(err.message || "Erro ao processar comando");
    },
  });

  const confirmActionMutation = useMutation({
    mutationFn: (token: string) => api.confirmAction(token),
    onSuccess: (res) => {
      setConfirmationData(null);
      showToast(res.message || "Ação confirmada com sucesso!");
      queryClient.invalidateQueries({ queryKey: ["cards"] });
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
    },
    onError: (err: Error) => {
      showToast(err.message || "Falha ao confirmar ação");
      setConfirmationData(null);
    },
  });

  const handleOpenCard = async (cardId: string) => {
    try {
      const details = await api.getCardDetails(cardId);
      setSelectedCardDetails(details);
      setIsSheetOpen(true);
    } catch {
      // Fallback para cartão local se ainda não sincronizado com o banco
      const c = cards.find((x) => x.id === cardId);
      if (c) {
        setSelectedCardDetails({
          card: c,
          events: [{ id: "1", kind: "Pedido recebido", payload: { title: "Pedido inicial recebido" }, actor_type: "owner", at: c.created_at }],
          is_gate: ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage),
          actions: ["Aprovar e avançar"],
          margin_cents: c.price_cents - c.est_cost_cents,
          risk_report: c.stage === "entrega_aprovar" ? {
            id: "r1",
            card_id: c.id,
            tests: [["Testes automáticos", "14 de 14 passaram", "ok"]],
            secrets_scan: [["Segredos no código", "Nenhum encontrado", "ok"]],
            outdated_deps: [["Dependências", "1 aviso de biblioteca desatualizada", "warn"]],
            lgpd_points: [["Dados pessoais (LGPD)", "Guarda telefone: aviso de privacidade conferido", "ok"]],
          } : undefined,
        });
        setIsSheetOpen(true);
      }
    }
  };

  const handleQuickApprove = (cardId: string) => {
    if (isPaused) {
      showToast("Tudo pausado. Retome antes de liberar execuções.");
      return;
    }
    decideMutation.mutate({ id: cardId, decision: "aprovado" });
  };

  const handleAskInfo = (cardId: string) => {
    decideMutation.mutate({ id: cardId, decision: "mais_info", note: "Dono solicitou esclarecimentos adicionais." });
  };

  const handleReject = (cardId: string) => {
    decideMutation.mutate({ id: cardId, decision: "recusado", note: "Recusado pelo dono da software house." });
  };

  // Cálculo da pílula de dinheiro parado
  const pendingCards = cards.filter((c) =>
    ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage)
  );
  const totalWaitingCents = moneyWaiting?.total_cents ?? pendingCards.reduce((acc, c) => acc + c.price_cents, 0);
  const formattedWaiting = (totalWaitingCents / 100).toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
    minimumFractionDigits: 0,
  });
  const decisionsCount = moneyWaiting?.count ?? pendingCards.length;

  const tabTitles: Record<TabKey, string> = {
    hoje: "Hoje",
    quadro: "Quadro",
    apps: "Meus apps",
    fin: "Finanças",
    con: "Conexões",
  };

  if (loadingAuth) {
    return (
      <div className="flex items-center justify-center min-h-screen text-xs text-[var(--mute)]">
        Carregando central...
      </div>
    );
  }

  if (!session) {
    return (
      <main className="flex items-center justify-center min-h-screen bg-[var(--bg)] p-4">
        <LoginForm onSuccess={() => supabase.auth.getSession().then(({ data }) => setSession(data.session))} />
      </main>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-[var(--bg)] text-[var(--ink)] font-sans antialiased">
      {/* Cabeçalho */}
      <Header
        title={tabTitles[activeTab]}
        moneyWaitingFormatted={formattedWaiting}
        pendingDecisionsCount={decisionsCount}
        isPaused={isPaused}
        onTogglePause={() => {
          if (isPaused) resumeMutation.mutate();
          else pauseMutation.mutate();
        }}
        isTogglingPause={pauseMutation.isPending || resumeMutation.isPending}
      />

      {/* Banner de Sistema Pausado */}
      <PausedBanner isPaused={isPaused} onResume={() => resumeMutation.mutate()} />

      {/* Conteúdo Principal com scroll livre */}
      <main className="flex-1 px-4 max-w-[1100px] w-full mx-auto pb-36 pt-1">
        {activeTab === "hoje" && (
          <HojeTab
            cards={cards}
            onOpenCard={handleOpenCard}
            onQuickApprove={handleQuickApprove}
            isProcessing={decideMutation.isPending}
          />
        )}

        {activeTab === "quadro" && (
          <QuadroTab cards={cards} onOpenCard={handleOpenCard} />
        )}

        {activeTab === "apps" && (
          <MeusAppsTab
            apps={apps}
            onCreateApp={(data) => createAppMutation.mutate(data)}
            onDeleteApp={(id) => deleteAppMutation.mutate(id)}
            isProcessing={createAppMutation.isPending}
          />
        )}

        {activeTab === "fin" && (
          <FinancasTab
            finances={finances}
            apps={apps}
            onCreateEntry={(data) => createFinanceEntryMutation.mutate(data)}
            isProcessing={createFinanceEntryMutation.isPending}
          />
        )}

        {activeTab === "con" && (
          <ConexoesTab connectors={connectors} onShowToast={showToast} />
        )}
      </main>

      {/* Barra de Comando de Voz e Texto */}
      <CommandLine
        cards={cards}
        onApprove={handleQuickApprove}
        onReject={handleReject}
        onShowToast={showToast}
        moneySummary={`${formattedWaiting} parados esperando você · ${decisionsCount} decisões`}
        onExecuteCommand={(text) => processCommandMutation.mutate(text)}
      />

      {/* Navegação Inferior */}
      <BottomNav currentTab={activeTab} onTabChange={setActiveTab} />

      {/* Bottom Sheet de Detalhe do Cartão */}
      <CardDetailSheet
        cardDetails={selectedCardDetails}
        isOpen={isSheetOpen}
        onClose={() => setIsSheetOpen(false)}
        onApprove={handleQuickApprove}
        onAskInfo={handleAskInfo}
        onReject={handleReject}
        onBuild={(id) => startBuildMutation.mutate(id)}
        isProcessing={decideMutation.isPending || startBuildMutation.isPending}
      />

      {/* Modal de Confirmação de Ação Destrutiva (F3-02) */}
      {confirmationData && (
        <div className="fixed inset-0 bg-black/70 z-50 flex items-center justify-center p-4">
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-sm w-full shadow-2xl">
            <h3 className="text-base font-bold text-[var(--bad)] mb-2 flex items-center gap-1.5">
              <span>⚠️ Ação Destrutiva</span>
            </h3>
            <p className="text-xs sm:text-sm text-[var(--ink)] mb-4 leading-relaxed">
              {confirmationData.prompt}
            </p>
            <div className="text-[11px] text-[var(--mute)] mb-4 bg-[var(--bg)]/50 p-2 rounded-lg font-mono">
              Token de uso único: {confirmationData.token.substring(0, 12)}...
            </div>
            <div className="flex justify-end gap-2">
              <button
                id="cancel-destructive-btn"
                onClick={() => setConfirmationData(null)}
                className="px-3.5 py-2 rounded-xl text-xs font-semibold border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer"
              >
                Cancelar
              </button>
              <button
                id="confirm-destructive-btn"
                disabled={confirmActionMutation.isPending}
                onClick={() => confirmActionMutation.mutate(confirmationData.token)}
                className="px-3.5 py-2 rounded-xl text-xs font-semibold border-0 bg-[var(--bad)] text-white cursor-pointer hover:opacity-90"
              >
                {confirmActionMutation.isPending ? "Confirmando..." : "Confirmar Ação"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Toast Flutuante */}
      {toastMessage && (
        <div
          id="global-toast"
          className="fixed left-1/2 -translate-x-1/2 bg-[var(--ink)] text-[var(--bg)] py-2.5 px-4 rounded-full text-xs font-medium z-50 shadow-lg pointer-events-none transition-all duration-200"
          style={{ bottom: "calc(110px + env(safe-area-inset-bottom, 0px))" }}
        >
          {toastMessage}
        </div>
      )}
    </div>
  );
}
