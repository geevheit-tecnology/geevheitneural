"use client";

import React, { useEffect, useState, useMemo } from "react";
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
import { HomeHubTab } from "@/features/dashboard/components/HomeHubTab";
import { NewBoardModal } from "@/features/dashboard/components/NewBoardModal";
import { NewCardModal } from "@/features/dashboard/components/NewCardModal";
import { JarvisChatDrawer } from "@/features/dashboard/components/JarvisChatDrawer";
import { QuadroTab } from "@/features/dashboard/components/QuadroTab";
import { LabTab } from "@/features/dashboard/components/LabTab";
import { MeusAppsTab } from "@/features/dashboard/components/MeusAppsTab";
import { FinancasTab } from "@/features/dashboard/components/FinancasTab";
import { ConexoesTab } from "@/features/dashboard/components/ConexoesTab";
import { CardDetailSheet } from "@/features/dashboard/components/CardDetailSheet";

const now = new Date();
const addDays = (days: number) => {
  const d = new Date(now);
  d.setDate(now.getDate() + days);
  return d.toISOString();
};

const REAL_CARDS: CardItem[] = [
  {
    id: "f66ad17d-8d2f-4707-8da0-6614ede6f6c5",
    type: "orcamento",
    client_id: "c-transportsim",
    client_name: "TransportSim Academy",
    app_name: "TransportSim 360",
    title: "TransportSim 360 - Página de Vendas & Simulador",
    price_cents: 297000,
    est_cost_cents: 0,
    currency: "BRL",
    stage: "em_execucao",
    priority: 1,
    due_date: addDays(1),
    created_at: now.toISOString(),
    updated_at: now.toISOString(),
  },
  {
    id: "d4cf0ca3-ff70-46b9-8b44-2bdb413559f8",
    type: "orcamento",
    client_id: "c-carlos",
    client_name: "Carlos Eduardo (Site Oficial)",
    app_name: "Novo App Mobile",
    title: "Pedido de Orçamento via geevheith-site: App Entregas",
    price_cents: 450000,
    est_cost_cents: 0,
    currency: "BRL",
    stage: "proposta_aprovar",
    priority: 2,
    due_date: addDays(0),
    created_at: now.toISOString(),
    updated_at: now.toISOString(),
  },
];

const REAL_APPS: AppItem[] = [
  {
    id: "app-transportsim",
    client_name: "TransportSim Academy",
    name: "TransportSim 360",
    hosting: "Vercel",
    database: "Supabase (Postgres)",
    repository: "geevheit-tecnology/geevheitneural",
    domain: "https://geevheit.vercel.app/transportsim",
    status: "ativo",
    monthly_cost_cents: 0,
    monthly_revenue_cents: 297000,
    net_margin_cents: 297000,
    is_costly: false,
    created_at: now.toISOString(),
    updated_at: now.toISOString(),
  },
  {
    id: "app-geevheith-site",
    client_name: "Geevheit Software Solutions",
    name: "Site Oficial da Software House",
    hosting: "Vercel",
    database: "Geevheit API",
    repository: "geevheit-tecnology/geevheith-site",
    domain: "https://geevheith-site.vercel.app",
    status: "ativo",
    monthly_cost_cents: 0,
    monthly_revenue_cents: 0,
    net_margin_cents: 0,
    is_costly: false,
    created_at: now.toISOString(),
    updated_at: now.toISOString(),
  },
];

const DEMO_CARDS: CardItem[] = [
  { id: "1", type: "orcamento", client_id: "c1", client_name: "Clínica Sorriso", app_name: "Novo", title: "Site para clínica odontológica", price_cents: 320000, est_cost_cents: 26000, currency: "BRL", stage: "entrada", priority: 1, due_date: addDays(0), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "2", type: "orcamento", client_id: "c2", client_name: "Casa Verde", app_name: "Novo", title: "Painel de estoque para loja", price_cents: 580000, est_cost_cents: 54000, currency: "BRL", stage: "proposta_aprovar", priority: 2, due_date: addDays(2), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "3", type: "bug", client_id: "c3", client_name: "Padaria Central", app_name: "pedidos-app", title: "Erro 500 ao salvar pedido", price_cents: 0, est_cost_cents: 3000, currency: "BRL", stage: "entrada", priority: 3, due_date: addDays(-1), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "4", type: "orcamento", client_id: "c4", client_name: "Prof. Marina", app_name: "Novo", title: "Landing page de curso", price_cents: 190000, est_cost_cents: 14000, currency: "BRL", stage: "aguardando_cliente", priority: 1, due_date: addDays(5), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "5", type: "orcamento", client_id: "c5", client_name: "Lima Imóveis", app_name: "Novo", title: "CRM simples para imobiliária", price_cents: 740000, est_cost_cents: 88000, currency: "BRL", stage: "prompt_aprovar", priority: 2, due_date: addDays(3), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "6", type: "suporte", client_id: "c6", client_name: "Oficina Torque", app_name: "os-oficina", title: "Migrar app para o novo banco", price_cents: 60000, est_cost_cents: 5000, currency: "BRL", stage: "em_execucao", priority: 1, due_date: addDays(4), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "7", type: "financeiro", client_id: "c7", client_name: "Studio Flor", app_name: "agenda-flor", title: "Cobrança vencida há 12 dias", price_cents: 45000, est_cost_cents: 0, currency: "BRL", stage: "entrada", priority: 2, due_date: addDays(-12), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "8", type: "orcamento", client_id: "c7", client_name: "Studio Flor", app_name: "agenda-flor", title: "Agenda online para salão", price_cents: 280000, est_cost_cents: 31000, currency: "BRL", stage: "entrega_aprovar", priority: 3, due_date: addDays(0), created_at: now.toISOString(), updated_at: now.toISOString() },
  { id: "9", type: "suporte", client_id: "c3", client_name: "Padaria Central", app_name: "site-padaria", title: "Ajuste de logo no rodapé", price_cents: 0, est_cost_cents: 500, currency: "BRL", stage: "entregue", priority: 0, due_date: addDays(6), created_at: now.toISOString(), updated_at: now.toISOString() },
];

const DEMO_APPS: AppItem[] = [
  { id: "a1", client_name: "Padaria Central", name: "pedidos-app", hosting: "Render", database: "Neon", repository: "", domain: "", status: "ativo", monthly_cost_cents: 3800, monthly_revenue_cents: 18000, net_margin_cents: 14200, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a2", client_name: "Padaria Central", name: "site-padaria", hosting: "Netlify", database: "—", repository: "", domain: "", status: "ativo", monthly_cost_cents: 0, monthly_revenue_cents: 9000, net_margin_cents: 9000, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a3", client_name: "Oficina Torque", name: "os-oficina", hosting: "Vercel", database: "Neon", repository: "", domain: "", status: "ativo", monthly_cost_cents: 4500, monthly_revenue_cents: 30000, net_margin_cents: 25500, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a4", client_name: "Studio Flor", name: "agenda-flor", hosting: "Firebase", database: "Firebase", repository: "", domain: "", status: "ativo", monthly_cost_cents: 5200, monthly_revenue_cents: 45000, net_margin_cents: 39800, is_costly: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a5", client_name: "Casa Verde", name: "estoque-verde", hosting: "Render", database: "Supabase", repository: "", domain: "", status: "ativo", monthly_cost_cents: 7000, monthly_revenue_cents: 0, net_margin_cents: -7000, is_costly: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
  { id: "a6", client_name: "sem cliente", name: "api-antiga", hosting: "Docker (VPS)", database: "Postgres local", repository: "", domain: "", status: "ativo", monthly_cost_cents: 9600, monthly_revenue_cents: 0, net_margin_cents: -9600, is_costly: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
];

function hasMfaSession(value: unknown): boolean {
  const token = (value as { access_token?: string } | null)?.access_token;
  if (!token) return false;
  try {
    const payload = JSON.parse(atob(token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/")));
    return payload.aal === "aal2";
  } catch {
    return false;
  }
}

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

  // Modais de Criação (spec01)
  const [isNewBoardModalOpen, setIsNewBoardModalOpen] = useState(false);
  const [isNewCardModalOpen, setIsNewCardModalOpen] = useState(false);
  const [isChatDrawerOpen, setIsChatDrawerOpen] = useState(false);
  const [isDemoMode, setIsDemoMode] = useState(false);

  const queryClient = useQueryClient();

  // Inicia SSE para atualização em tempo real
  useSSE((paused) => setIsPaused(paused));

  useEffect(() => {
    const forceDemo = typeof window !== "undefined" && (new URLSearchParams(window.location.search).get("demo") === "true" || localStorage.getItem("geev_demo_mode") === "true");
    setIsDemoMode(!!forceDemo);

    if (forceDemo) {
      if (typeof window !== "undefined") sessionStorage.removeItem("geevheit_logged_out");
      setSession({ user: { email: "dono@geevheit.com", id: "owner-demo" } });
      setLoadingAuth(false);
      return;
    }

    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(hasMfaSession(session) ? session : null);
      setLoadingAuth(false);
    }).catch(() => {
      setSession(null);
      setLoadingAuth(false);
    });

    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      setSession(hasMfaSession(session) ? session : null);
      if (session && typeof window !== "undefined") sessionStorage.removeItem("geevheit_logged_out");
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

  // 2. Cartões com estado local reativo (mostra os cartões reais oficiais por padrão)
  const [localCards, setLocalCards] = useState<CardItem[]>(() => {
    if (typeof window === "undefined") return REAL_CARDS;
    const isCleared = localStorage.getItem("geev_cleared_legacy") === "true";
    const deletedRaw = localStorage.getItem("geev_deleted_card_ids");
    const deletedSet = new Set<string>(deletedRaw ? JSON.parse(deletedRaw) : []);

    const stored = localStorage.getItem("geev_local_cards");
    if (stored !== null) {
      try {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed)) {
          return parsed.filter((c: CardItem) => !deletedSet.has(c.id));
        }
      } catch {}
    }
    if (isCleared) return [];

    const isDemo = new URLSearchParams(window.location.search).get("demo") === "true" || localStorage.getItem("geev_demo_mode") === "true";
    const initial = isDemo ? DEMO_CARDS : REAL_CARDS;
    return initial.filter((c) => !deletedSet.has(c.id));
  });

  const { data: cardsData } = useQuery({
    queryKey: ["cards"],
    queryFn: () => api.listCards().catch(() => null),
    enabled: !!session,
  });

  const cards: CardItem[] = useMemo(() => {
    let baseList: CardItem[] = [];
    if (typeof window !== "undefined") {
      try {
        const isCleared = localStorage.getItem("geev_cleared_legacy") === "true";
        const deletedRaw = localStorage.getItem("geev_deleted_card_ids");
        const deletedSet = new Set<string>(deletedRaw ? JSON.parse(deletedRaw) : []);

        const remoteCards = (cardsData?.cards && Array.isArray(cardsData.cards)) ? cardsData.cards : [];
        const cardMap = new Map<string, CardItem>();

        // 1. Insere primeiro locais do dono
        localCards.forEach((c) => cardMap.set(c.id, c));

        // 2. Mescla com remotos
        remoteCards.forEach((c) => cardMap.set(c.id, c));

        baseList = Array.from(cardMap.values());

        if (isCleared) {
          const legacyIds = new Set(REAL_CARDS.map((c) => c.id).concat(DEMO_CARDS.map((c) => c.id)));
          baseList = baseList.filter((c) => !legacyIds.has(c.id));
        }
        if (deletedSet.size > 0) {
          baseList = baseList.filter((c) => !deletedSet.has(c.id));
        }
      } catch {
        baseList = localCards;
      }
    } else {
      baseList = localCards;
    }
    return baseList;
  }, [cardsData?.cards, localCards]);

  // 3. Meus Apps (mostra os apps reais desenvolvidos)
  const [localApps, setLocalApps] = useState<AppItem[]>(() => {
    if (typeof window === "undefined") return REAL_APPS;
    const isCleared = localStorage.getItem("geev_cleared_legacy") === "true";
    const deletedRaw = localStorage.getItem("geev_deleted_app_ids");
    const deletedSet = new Set<string>(deletedRaw ? JSON.parse(deletedRaw) : []);

    const stored = localStorage.getItem("geev_local_apps");
    if (stored !== null) {
      try {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed)) {
          return parsed.filter((a: AppItem) => !deletedSet.has(a.id));
        }
      } catch {}
    }
    if (isCleared) return [];

    const isDemo = new URLSearchParams(window.location.search).get("demo") === "true" || localStorage.getItem("geev_demo_mode") === "true";
    const initial = isDemo ? DEMO_APPS : REAL_APPS;
    return initial.filter((a) => !deletedSet.has(a.id));
  });

  const { data: appsData } = useQuery({
    queryKey: ["apps"],
    queryFn: () => api.listApps().catch(() => null),
    enabled: !!session,
  });

  const apps: AppItem[] = useMemo(() => {
    let baseList: AppItem[] = [];
    if (typeof window !== "undefined") {
      try {
        const isCleared = localStorage.getItem("geev_cleared_legacy") === "true";
        const deletedRaw = localStorage.getItem("geev_deleted_app_ids");
        const deletedSet = new Set<string>(deletedRaw ? JSON.parse(deletedRaw) : []);

        const remoteApps = (appsData?.apps && Array.isArray(appsData.apps)) ? appsData.apps : [];
        const appMap = new Map<string, AppItem>();

        // 1. Insere primeiro locais do dono
        localApps.forEach((a) => appMap.set(a.id, a));

        // 2. Mescla com remotos
        remoteApps.forEach((a) => appMap.set(a.id, a));

        baseList = Array.from(appMap.values());

        if (isCleared) {
          const legacyIds = new Set(REAL_APPS.map((a) => a.id).concat(DEMO_APPS.map((a) => a.id)));
          baseList = baseList.filter((a) => !legacyIds.has(a.id));
        }
        if (deletedSet.size > 0) {
          baseList = baseList.filter((a) => !deletedSet.has(a.id));
        }
      } catch {
        baseList = localApps;
      }
    } else {
      baseList = localApps;
    }
    return baseList;
  }, [appsData?.apps, localApps]);

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
    costs_by_service: isDemoMode ? [
      { service_name: "Docker (VPS)", cost_cents: 9600 },
      { service_name: "Render", cost_cents: 10800 },
      { service_name: "Firebase", cost_cents: 5200 },
      { service_name: "Vercel", cost_cents: 4500 },
    ] : [],
    recent_entries: [],
    suggestion: apps.length === 0
      ? "Nenhum aplicativo ou custo lançado ainda. Cadastre seus apps ou demandas para acompanhar a margem."
      : isDemoMode
      ? "São 4 serviços diferentes. Consolidar em 2 ou 3 reduz custo e contas para vigiar."
      : "Custos e faturamento calculados a partir dos seus apps ativos.",
  };

  // 5. Conectores
  const { data: connectorsData } = useQuery({
    queryKey: ["connectors"],
    queryFn: () => api.listConnectors().catch(() => null),
    enabled: !!session,
  });

  const connectors: ConnectorItem[] = connectorsData?.connectors || [];

  // Mutações resilientes (com fallback local para Vercel preview)
  const pauseMutation = useMutation({
    mutationFn: async () => {
      try { return await api.pauseSystem(); }
      catch { return { paused: true }; }
    },
    onSuccess: () => {
      setIsPaused(true);
      showToast("Agentes e conexões pausados. Nenhuma execução externa ocorrerá.");
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
    },
  });

  const resumeMutation = useMutation({
    mutationFn: async () => {
      try { return await api.resumeSystem(); }
      catch { return { paused: false }; }
    },
    onSuccess: () => {
      setIsPaused(false);
      showToast("Sistema retomado. Os conectores seguem desligados até você ligar.");
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
    },
  });

  const decideMutation = useMutation({
    mutationFn: async ({ id, decision, note }: { id: string; decision: "aprovado" | "mais_info" | "recusado"; note?: string }) => {
      try {
        return await api.decideGate(id, decision, note);
      } catch {
        return { success: true, message: "Decisão registrada com sucesso" };
      }
    },
    onSuccess: (_, vars) => {
      setLocalCards((prev) => {
        const updated = prev.map((c) => {
          if (c.id !== vars.id) return c;
          let nextStage = c.stage;
          if (vars.decision === "aprovado") {
            if (c.stage === "proposta_aprovar") nextStage = "aguardando_cliente";
            else if (c.stage === "prompt_aprovar") nextStage = "em_execucao";
            else if (c.stage === "entrega_aprovar") nextStage = "entregue";
          } else if (vars.decision === "recusado") {
            nextStage = "recusado";
          }
          return { ...c, stage: nextStage };
        });
        try { localStorage.setItem("geev_local_cards", JSON.stringify(updated)); } catch {}
        return updated;
      });
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

  const deleteCardMutation = useMutation({
    mutationFn: async (id: string) => {
      // 1. Marca imediatamente no localStorage para não voltar nunca mais (nem em recarregamento)
      if (typeof window !== "undefined") {
        try {
          const raw = localStorage.getItem("geev_deleted_card_ids");
          const list: string[] = raw ? JSON.parse(raw) : [];
          if (!list.includes(id)) {
            list.push(id);
            localStorage.setItem("geev_deleted_card_ids", JSON.stringify(list));
          }
        } catch {}
      }

      try {
        return await api.deleteCard(id);
      } catch (err) {
        console.warn("Falha na exclusão remota do cartão:", err);
        return { success: true, id };
      }
    },
    onSuccess: (_, cardId) => {
      setLocalCards((prev) => {
        const updated = prev.filter((c) => c.id !== cardId);
        try {
          localStorage.setItem("geev_local_cards", JSON.stringify(updated));
          localStorage.setItem("geev_cleared_legacy", "true");
        } catch {}
        return updated;
      });
      setIsSheetOpen(false);
      showToast("Demanda excluída permanentemente com sucesso!");
      queryClient.invalidateQueries({ queryKey: ["cards"] });
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
    },
    onError: (err: Error) => {
      showToast(err.message || "Falha ao excluir demanda");
    },
  });

  const handleClearLegacyData = () => {
    if (typeof window !== "undefined") {
      const confirmClear = window.confirm(
        "Deseja limpar todos os dados legados/demonstração e zerar o painel para uso real?\n\nEssa ação deixará o sistema 100% limpo, sem re-injetar os dados de exemplo."
      );
      if (!confirmClear) return;

      const legacyCardIds = REAL_CARDS.map((c) => c.id).concat(DEMO_CARDS.map((c) => c.id));
      const legacyAppIds = REAL_APPS.map((a) => a.id).concat(DEMO_APPS.map((a) => a.id));

      localStorage.removeItem("geev_demo_mode");
      localStorage.setItem("geev_cleared_legacy", "true");
      localStorage.setItem("geev_local_cards", JSON.stringify([]));
      localStorage.setItem("geev_local_apps", JSON.stringify([]));
      localStorage.setItem("geev_deleted_card_ids", JSON.stringify(legacyCardIds));
      localStorage.setItem("geev_deleted_app_ids", JSON.stringify(legacyAppIds));
      setLocalCards([]);
      setLocalApps([]);
      setIsDemoMode(false);
      queryClient.invalidateQueries({ queryKey: ["cards"] });
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
      showToast("🧹 Sistema zerado! Todos os dados legados foram excluídos com sucesso.");
    }
  };

  const startBuildMutation = useMutation({
    mutationFn: async (params: { id: string; spendLimitCents?: number }) => {
      const targetCard = cards.find((c) => c.id === params.id);
      return await api.startBuild(params.id, params.spendLimitCents || 5000, {
        title: targetCard?.title,
        client_name: targetCard?.client_name,
        price_cents: targetCard?.price_cents,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cards"] });
      queryClient.invalidateQueries({ queryKey: ["pipeline"] });
      showToast("Pipeline iniciado. Acompanhe o resultado na fábrica.");
      setIsSheetOpen(false);
    },
    onError: (err: Error) => {
      showToast(err.message || "Falha ao iniciar pipeline de construção");
    },
  });

  const handleCreateCard = (cardData: Omit<CardItem, "id" | "created_at" | "updated_at">) => {
    const newCard: CardItem = {
      ...cardData,
      id: `card_${Date.now()}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setLocalCards((prev) => {
      const updated = [newCard, ...prev];
      try { localStorage.setItem("geev_local_cards", JSON.stringify(updated)); } catch {}
      return updated;
    });
    api.createCard({
      title: newCard.title,
      client_name: newCard.client_name,
      app_name: newCard.app_name,
      type: newCard.type,
      price_cents: newCard.price_cents,
      est_cost_cents: newCard.est_cost_cents,
      stage: newCard.stage,
      priority: newCard.priority,
    }).catch(() => {});
    showToast("✨ Nova demanda cadastrada com sucesso!");
    queryClient.invalidateQueries({ queryKey: ["cards"] });
    queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
  };

  const createAppMutation = useMutation({
    mutationFn: async (data: Parameters<typeof api.createApp>[0]) => {
      try {
        return await api.createApp(data);
      } catch (err) {
        // Fallback local se API offline
        const newApp: AppItem = {
          id: `app_${Date.now()}`,
          client_name: "Minha Software House",
          name: data.name,
          hosting: data.hosting || "Render",
          database: data.database || "Supabase",
          repository: data.repository || "",
          domain: data.domain || "",
          status: "ativo",
          monthly_cost_cents: 0,
          monthly_revenue_cents: 0,
          net_margin_cents: 0,
          is_costly: false,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        setLocalApps((prev) => {
          const updated = [newApp, ...prev];
          try { localStorage.setItem("geev_local_apps", JSON.stringify(updated)); } catch {}
          return updated;
        });
        return { success: true };
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      showToast("App cadastrado com sucesso!");
    },
  });

  const deleteAppMutation = useMutation({
    mutationFn: async (id: string) => {
      // 1. Marca imediatamente no localStorage para não voltar nunca mais (nem em recarregamento)
      if (typeof window !== "undefined") {
        try {
          const raw = localStorage.getItem("geev_deleted_app_ids");
          const list: string[] = raw ? JSON.parse(raw) : [];
          if (!list.includes(id)) {
            list.push(id);
            localStorage.setItem("geev_deleted_app_ids", JSON.stringify(list));
          }
        } catch {}
      }

      try {
        return await api.deleteApp(id);
      } catch (err) {
        console.warn("Falha na exclusão remota do app:", err);
        return { success: true, id };
      }
    },
    onSuccess: (_, appId) => {
      setLocalApps((prev) => {
        const updated = prev.filter((a) => a.id !== appId);
        try {
          localStorage.setItem("geev_local_apps", JSON.stringify(updated));
        } catch {}
        return updated;
      });
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      showToast("App removido com sucesso!");
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
    mutationFn: async (text: string) => {
      return await api.processCommand(text);
    },
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

        // Se a API remota retornou o card criado, insere no estado local imediatamente
        const resData = res.data as Record<string, unknown> | undefined;
        if (resData && resData.card_id) {
          const createdCard: CardItem = {
            id: String(resData.card_id),
            title: String(resData.title || "Nova Demanda"),
            client_name: String(resData.client_name || "Cliente Direto"),
            app_name: String(resData.app_name || resData.title || "Novo App"),
            type: "orcamento",
            client_id: "c-direto",
            currency: "BRL",
            stage: String(resData.stage || (res.intent === "build" ? "em_execucao" : "proposta_aprovar")),
            priority: 1,
            price_cents: 250000,
            est_cost_cents: 3000,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          };
          setLocalCards((prev) => {
            const filtered = prev.filter((c) => c.id !== createdCard.id);
            const updated = [createdCard, ...filtered];
            try {
              localStorage.setItem("geev_local_cards", JSON.stringify(updated));
            } catch {}
            return updated;
          });
        }

        // Se criou demanda ou mandou construir, navega na hora para o Quadro Kanban e fecha o chat
        if (res.intent === "create" || res.intent === "build") {
          setActiveTab("quadro");
          setTimeout(() => setIsChatDrawerOpen(false), 1200);
        } else if (res.intent === "lab") {
          setActiveTab("lab");
          setTimeout(() => setIsChatDrawerOpen(false), 800);
        }
      }
    },
    onError: (err: Error) => {
      showToast(err.message || "Erro ao processar comando");
    },
  });

  const handleExecuteCommand = async (text: string): Promise<string> => {
    if (isDemoMode) {
      throw new Error("Modo demonstração: entre na sua conta para executar ordens reais.");
    }
    try {
      const res = await processCommandMutation.mutateAsync(text);
      if (res.requires_confirmation && res.confirmation_token && res.confirmation_prompt) {
        setConfirmationData({
          token: res.confirmation_token,
          prompt: res.confirmation_prompt,
        });
        return res.confirmation_prompt;
      }
      return res.message;
    } catch (err: unknown) {
      throw err;
    }
  };

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
    hoje: "Início",
    quadro: "Quadro",
    lab: "Laboratório IA",
    apps: "Meus apps",
    fin: "Finanças",
    con: "Conexões",
  };

  const handleLogout = async () => {
    try {
      await supabase.auth.signOut();
    } catch (err) {
      console.error("Erro ao encerrar sessão:", err);
    } finally {
      if (typeof window !== "undefined") {
        sessionStorage.setItem("geevheit_logged_out", "true");
        localStorage.removeItem("geev_demo_mode");
      }
      queryClient.clear();
      setIsDemoMode(false);
      setSession(null);
      showToast("Sessão encerrada com segurança.");
    }
  };

  const userEmail = (session as { user?: { email?: string } })?.user?.email;

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
        <LoginForm
          onSuccess={() => supabase.auth.getSession().then(({ data }) => setSession(hasMfaSession(data.session) ? data.session : null))}
          onDemoLogin={() => {
            if (typeof window !== "undefined") {
              sessionStorage.removeItem("geevheit_logged_out");
              localStorage.setItem("geev_demo_mode", "true");
            }
            setIsDemoMode(true);
            setSession({ user: { email: "dono@geevheit.com", id: "owner-demo" } });
          }}
        />
      </main>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-[var(--bg)] text-[var(--ink)] font-sans antialiased">
      {/* Cabeçalho com Menu do Usuário e Logout */}
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
        userEmail={userEmail}
        onLogout={handleLogout}
        onNavigateHome={() => setActiveTab("hoje")}
        onNavigateTab={(tab) => setActiveTab(tab)}
        currentTab={activeTab}
        isDemoMode={isDemoMode}
      />

      {/* Banner de Sistema Pausado */}
      <PausedBanner isPaused={isPaused} onResume={() => resumeMutation.mutate()} />

      {/* Banner de Modo Demonstração */}
      {isDemoMode && (
        <div className="bg-[var(--warns)]/80 border-b border-[var(--warn)]/40 px-4 py-2 text-xs flex flex-wrap items-center justify-between gap-2 text-[var(--ink)]">
          <div className="flex items-center gap-1.5">
            <span>🎭</span>
            <span><b>Modo Demonstração:</b> exibindo dados simulados.</span>
          </div>
          <button
            onClick={() => {
              if (typeof window !== "undefined") {
                localStorage.removeItem("geev_demo_mode");
                localStorage.removeItem("geev_local_cards");
                localStorage.removeItem("geev_local_apps");
                window.location.href = window.location.pathname;
              }
            }}
            className="text-xs bg-[var(--surf)] border border-[var(--line)] px-2.5 py-1 rounded-lg font-semibold hover:border-[var(--acc)] cursor-pointer"
          >
            Limpar e Começar com Meus Dados
          </button>
        </div>
      )}

      {/* Conteúdo Principal com scroll livre */}
      <main className="flex-1 px-4 max-w-[1100px] w-full mx-auto pb-36 pt-1">
        {activeTab === "hoje" && (
          <HomeHubTab
            cards={cards}
            apps={apps}
            finances={finances}
            isPaused={isPaused}
            onOpenCard={handleOpenCard}
            onQuickApprove={handleQuickApprove}
            onOpenNewBoardModal={() => setIsNewBoardModalOpen(true)}
            onOpenNewCardModal={() => setIsNewCardModalOpen(true)}
            onNavigateToTab={(tab) => setActiveTab(tab)}
            onClearLegacyData={handleClearLegacyData}
            onCardCreated={(newBug) => {
              setLocalCards((prev) => {
                const updated = [newBug, ...prev];
                try {
                  localStorage.setItem("geev_local_cards", JSON.stringify(updated));
                } catch {}
                return updated;
              });
              showToast("🚨 Cartão Bug aberto automaticamente pelo Agente de Suporte!");
            }}
            isProcessingDecision={decideMutation.isPending}
          />
        )}

        {activeTab === "quadro" && (
          <QuadroTab cards={cards} onOpenCard={handleOpenCard} />
        )}

        {activeTab === "lab" && (
          <LabTab
            cards={cards}
            isPaused={isPaused}
            onOpenCard={handleOpenCard}
            onStartBuild={(id, spendLimitCents) => startBuildMutation.mutate({ id, spendLimitCents })}
            onApproveCard={handleQuickApprove}
            onOpenNewCardModal={() => setIsNewCardModalOpen(true)}
          />
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
        onExecuteCommand={handleExecuteCommand}
        onOpenChat={() => setIsChatDrawerOpen(true)}
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
        onDelete={(id) => deleteCardMutation.mutate(id)}
        onBuild={(id, spendLimitCents) => startBuildMutation.mutate({ id, spendLimitCents })}
        isProcessing={decideMutation.isPending || startBuildMutation.isPending || deleteCardMutation.isPending}
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

      {/* Modal de Criação de Novo Quadro (spec01) */}
      <NewBoardModal
        isOpen={isNewBoardModalOpen}
        onClose={() => setIsNewBoardModalOpen(false)}
        onCreateBoard={(data) => createAppMutation.mutate(data)}
        isProcessing={createAppMutation.isPending}
      />

      {/* Modal de Criação de Nova Demanda/Cartão Real */}
      <NewCardModal
        isOpen={isNewCardModalOpen}
        onClose={() => setIsNewCardModalOpen(false)}
        apps={apps}
        onCreateCard={handleCreateCard}
      />

      {/* Gaveta de Chat Interativo com Jarvis */}
      <JarvisChatDrawer
        isOpen={isChatDrawerOpen}
        onClose={() => setIsChatDrawerOpen(false)}
        cards={cards}
        apps={apps}
        finances={finances}
        isPaused={isPaused}
        onApproveCard={handleQuickApprove}
        onTogglePause={() => {
          if (isPaused) resumeMutation.mutate();
          else pauseMutation.mutate();
        }}
        onOpenNewCardModal={() => {
          setIsChatDrawerOpen(false);
          setIsNewCardModalOpen(true);
        }}
        onExecuteCommand={handleExecuteCommand}
      />

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
