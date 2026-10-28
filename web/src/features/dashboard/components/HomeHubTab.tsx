"use client";

import React, { useState, useMemo } from "react";
import { CardItem, AppItem, FinancesOverview } from "@/shared/api-client/endpoints";
import { AgendaView } from "./AgendaView";
import { HojeTab } from "./HojeTab";
import { AgentsCockpit } from "./AgentsCockpit";

interface HomeHubTabProps {
  cards: CardItem[];
  apps: AppItem[];
  finances: FinancesOverview;
  isPaused: boolean;
  onOpenCard: (cardId: string) => void;
  onQuickApprove: (cardId: string) => void;
  onOpenNewBoardModal: () => void;
  onOpenNewCardModal?: () => void;
  onNavigateToTab: (tab: "quadro" | "apps" | "fin" | "con") => void;
  onClearLegacyData?: () => void;
  isProcessingDecision?: boolean;
}

export function HomeHubTab({
  cards,
  apps,
  finances,
  isPaused,
  onOpenCard,
  onQuickApprove,
  onOpenNewBoardModal,
  onOpenNewCardModal,
  onNavigateToTab,
  onClearLegacyData,
  isProcessingDecision = false,
}: HomeHubTabProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedStageFilter, setSelectedStageFilter] = useState<string | null>(null);
  const [boardViewMode, setBoardViewMode] = useState<"grid" | "list">("grid");
  const [boardSortBy, setBoardSortBy] = useState<"recent" | "name" | "margin">("recent");
  const [hubViewMode, setHubViewMode] = useState<"hub" | "agenda" | "decisoes" | "agentes">("hub");

  // Saudação de acordo com o horário local
  const greeting = useMemo(() => {
    const hour = new Date().getHours();
    if (hour < 12) return "Bom dia";
    if (hour < 18) return "Boa tarde";
    return "Boa noite";
  }, []);

  const todayFormatted = useMemo(() => {
    return new Date().toLocaleDateString("pt-BR", {
      weekday: "long",
      day: "numeric",
      month: "long",
    });
  }, []);

  // Definição das 7 etapas reais de produção (etiquetas)
  const productionStages = [
    { key: "entrada", label: "Entrada", color: "var(--ink)", isGate: false },
    { key: "proposta_aprovar", label: "Proposta para aprovar", color: "var(--warn)", isGate: true },
    { key: "aguardando_cliente", label: "Aguardando cliente", color: "var(--mute)", isGate: false },
    { key: "prompt_aprovar", label: "Prompt para aprovar", color: "var(--warn)", isGate: true },
    { key: "em_execucao", label: "Em execução", color: "var(--acc)", isGate: false },
    { key: "entrega_aprovar", label: "Entrega para aprovar", color: "var(--warn)", isGate: true },
    { key: "entregue", label: "Entregue", color: "var(--ok)", isGate: false },
  ];

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  // Contagem de cartões por etapa de produção
  const stageCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    productionStages.forEach((s) => {
      counts[s.key] = cards.filter((c) => c.stage === s.key).length;
    });
    return counts;
  }, [cards]);

  // Cartões pendentes de portão (decisões do dono)
  const pendingGateCards = useMemo(() => {
    return cards
      .filter((c) => ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage))
      .sort((a, b) => b.price_cents - a.price_cents);
  }, [cards]);

  // Filtro de cartões pela busca global e pela etiqueta selecionada
  const filteredCards = useMemo(() => {
    return cards.filter((card) => {
      const matchSearch =
        searchTerm === "" ||
        card.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
        card.client_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        card.app_name.toLowerCase().includes(searchTerm.toLowerCase());

      const matchStage = selectedStageFilter ? card.stage === selectedStageFilter : true;
      return matchSearch && matchStage;
    });
  }, [cards, searchTerm, selectedStageFilter]);

  // Filtro e ordenação dos quadros/apps
  const sortedAndFilteredApps = useMemo(() => {
    const filtered = apps.filter((app) => {
      return (
        searchTerm === "" ||
        app.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        app.client_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        app.hosting.toLowerCase().includes(searchTerm.toLowerCase())
      );
    });

    return filtered.sort((a, b) => {
      if (boardSortBy === "name") return a.name.localeCompare(b.name);
      if (boardSortBy === "margin") {
        const marginA = a.monthly_revenue_cents - a.monthly_cost_cents;
        const marginB = b.monthly_revenue_cents - b.monthly_cost_cents;
        return marginB - marginA;
      }
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    });
  }, [apps, searchTerm, boardSortBy]);

  // "Continuar de onde parei" (últimos 3 quadros com atividade)
  const recentBoards = useMemo(() => {
    return apps.slice(0, 3);
  }, [apps]);

  // Total de dinheiro parado
  const totalMoneyWaiting = useMemo(() => {
    return pendingGateCards.reduce((acc, c) => acc + c.price_cents, 0);
  }, [pendingGateCards]);

  const [copiedIntake, setCopiedIntake] = useState(false);

  const handleCopyIntakeLink = () => {
    const link = "https://geevheit.vercel.app/intake";
    if (navigator.clipboard) {
      navigator.clipboard.writeText(link);
      setCopiedIntake(true);
      setTimeout(() => setCopiedIntake(false), 3000);
    } else {
      window.open("/intake", "_blank");
    }
  };

  return (
    <div id="home-hub" className="space-y-5 pb-8 animate-fadeIn">
      {/* a) Cabeçalho do Hub de Quadros */}
      <section id="hub-header-section" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 sm:p-5 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-[var(--acc)]">
              Hub de Operação da Software House
            </div>
            <h2 className="text-xl sm:text-2xl font-bold text-[var(--ink)] m-0 leading-tight">
              {greeting}, Dono 👋
            </h2>
            <p className="text-xs text-[var(--mute)] m-0 mt-0.5 capitalize">
              {todayFormatted} · Acompanhe a produção, quadros e decisões
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <button
              id="btn-copy-intake"
              onClick={handleCopyIntakeLink}
              title="Copiar link para clientes solicitarem novos pedidos"
              className="bg-emerald-500/10 border border-emerald-500/30 text-emerald-600 dark:text-emerald-400 font-bold text-xs sm:text-sm py-2 px-3.5 rounded-xl cursor-pointer hover:bg-emerald-500/20 transition-all flex items-center gap-1.5"
            >
              <span>📥</span>
              <span>{copiedIntake ? "✓ Link Copiado!" : "Receber Pedidos"}</span>
            </button>

            <button
              id="btn-hub-new-board"
              onClick={onOpenNewBoardModal}
              className="bg-[var(--acc)] text-white font-bold text-xs sm:text-sm py-2 px-3.5 rounded-xl border-0 cursor-pointer shadow hover:opacity-90 transition-all flex items-center gap-1.5"
            >
              <span>+</span>
              <span>Novo Quadro</span>
            </button>
            {onOpenNewCardModal && (
              <button
                id="btn-hub-new-card"
                onClick={onOpenNewCardModal}
                className="bg-[var(--surf)] border border-[var(--acc)] text-[var(--acc)] font-bold text-xs sm:text-sm py-2 px-3.5 rounded-xl cursor-pointer hover:bg-[var(--accs)] transition-all flex items-center gap-1.5"
              >
                <span>✨</span>
                <span>Nova Demanda</span>
              </button>
            )}
            <button
              id="btn-hub-open-kanban"
              onClick={() => onNavigateToTab("quadro")}
              className="bg-[var(--accs)] text-[var(--ink)] font-semibold text-xs sm:text-sm py-2 px-3 rounded-xl border border-[var(--line)] cursor-pointer hover:bg-[var(--line)]/30 transition-all"
            >
              Ver Kanban
            </button>
            {onClearLegacyData && (
              <button
                id="btn-hub-clear-legacy"
                onClick={onClearLegacyData}
                title="Excluir dados legados e zerar o painel"
                className="bg-red-500/10 border border-red-500/30 text-red-600 dark:text-red-400 font-bold text-xs sm:text-sm py-2 px-3 rounded-xl cursor-pointer hover:bg-red-500/20 transition-all flex items-center gap-1.5"
              >
                <span>🧹</span>
                <span>Limpar Dados Legados</span>
              </button>
            )}
          </div>
        </div>

        {/* Banner dos Links Oficiais da Geevheit */}
        <div className="mt-3.5 pt-3 border-t border-[var(--line)] flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2.5 text-xs">
          <div className="flex flex-wrap items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span className="text-[var(--mute)]">
              <strong>Site Oficial da Software House:</strong>{" "}
              <a
                href="https://geevheith-site.vercel.app"
                target="_blank"
                rel="noreferrer"
                className="text-emerald-600 dark:text-emerald-400 font-bold hover:underline"
              >
                geevheith-site.vercel.app ↗
              </a>
            </span>
            <span className="text-[var(--line)] hidden sm:inline">|</span>
            <span className="text-[var(--mute)]">
              <strong>Captação Direta:</strong>{" "}
              <code className="text-[var(--acc)] font-mono bg-[var(--accs)] px-1.5 py-0.5 rounded">
                geevheit.vercel.app/intake
              </code>
            </span>
          </div>
          <div className="flex items-center gap-2">
            <a
              href="https://geevheith-site.vercel.app/#orcamento"
              target="_blank"
              rel="noreferrer"
              className="text-[var(--acc)] font-semibold hover:underline flex items-center gap-1"
            >
              <span>Abrir formulário no site</span>
              <span>↗</span>
            </a>
          </div>
        </div>
      </section>

      {/* Seletor de Visão do Início: Hub de Quadros | Agenda & Prazos | 3 Decisões */}
      <div className="flex bg-[var(--surf)] p-1 rounded-xl border border-[var(--line)]">
        <button
          id="hub-tab-hub"
          onClick={() => setHubViewMode("hub")}
          className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all cursor-pointer border-0 flex items-center justify-center gap-1.5 ${
            hubViewMode === "hub"
              ? "bg-[var(--acc)] text-white shadow-sm"
              : "bg-transparent text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          <span>🏢</span>
          <span>Hub de Quadros</span>
        </button>

        <button
          id="hub-tab-agenda"
          onClick={() => setHubViewMode("agenda")}
          className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all cursor-pointer border-0 flex items-center justify-center gap-1.5 ${
            hubViewMode === "agenda"
              ? "bg-[var(--acc)] text-white shadow-sm"
              : "bg-transparent text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          <span>📅</span>
          <span>Agenda & Prazos</span>
        </button>

        <button
          id="hub-tab-decisoes"
          onClick={() => setHubViewMode("decisoes")}
          className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all cursor-pointer border-0 flex items-center justify-center gap-1.5 ${
            hubViewMode === "decisoes"
              ? "bg-[var(--acc)] text-white shadow-sm"
              : "bg-transparent text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          <span>⚡</span>
          <span>3 Decisões</span>
        </button>

        <button
          id="hub-tab-agentes"
          onClick={() => setHubViewMode("agentes")}
          className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all cursor-pointer border-0 flex items-center justify-center gap-1.5 ${
            hubViewMode === "agentes"
              ? "bg-[var(--acc)] text-white shadow-sm"
              : "bg-transparent text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          <span>🤖</span>
          <span>Fábrica de IA</span>
        </button>
      </div>

      {hubViewMode === "agenda" ? (
        <AgendaView cards={cards} onOpenCard={onOpenCard} />
      ) : hubViewMode === "decisoes" ? (
        <HojeTab
          cards={cards}
          onOpenCard={onOpenCard}
          onQuickApprove={onQuickApprove}
          isProcessing={isProcessingDecision}
        />
      ) : hubViewMode === "agentes" ? (
        <AgentsCockpit cards={cards} isPaused={isPaused} onOpenCard={onOpenCard} />
      ) : (
        <>
          {/* Busca Global */}
          <section className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
            <div className="relative">
              <input
                id="hub-search-input"
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Buscar por nome de quadro, tarefa, cliente ou tecnologia..."
                className="w-full pl-9 pr-4 py-2.5 text-xs sm:text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)] transition-colors"
              />
              <span className="absolute left-3 top-3 text-[var(--mute)] text-sm pointer-events-none">
                🔍
              </span>
              {searchTerm && (
                <button
                  onClick={() => setSearchTerm("")}
                  className="absolute right-3 top-2.5 text-xs text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 cursor-pointer"
                >
                  Limpar
                </button>
              )}
            </div>
          </section>

      {/* g) Indicadores Gerais da Empresa */}
      <section id="hub-kpis-section" className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 shadow-xs">
          <div className="text-[11px] font-medium text-[var(--mute)]">Quadros Ativos</div>
          <div className="text-lg sm:text-xl font-bold text-[var(--ink)] mt-0.5">{apps.length}</div>
          <div className="text-[10px] text-[var(--mute)] mt-0.5">Sistemas em produção</div>
        </div>

        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 shadow-xs">
          <div className="text-[11px] font-medium text-[var(--mute)]">Dinheiro em Portões</div>
          <div className="text-lg sm:text-xl font-bold text-[var(--warn)] mt-0.5">
            {formatBRL(totalMoneyWaiting)}
          </div>
          <div className="text-[10px] text-[var(--warn)] mt-0.5 font-medium">
            {pendingGateCards.length} decisões aguardando
          </div>
        </div>

        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 shadow-xs">
          <div className="text-[11px] font-medium text-[var(--mute)]">Faturamento Mensal</div>
          <div className="text-lg sm:text-xl font-bold text-[var(--ok)] mt-0.5">
            {formatBRL(finances.total_revenue_cents)}
          </div>
          <div className="text-[10px] text-[var(--ok)] mt-0.5 font-medium">
            Margem: {formatBRL(finances.net_profit_cents)}
          </div>
        </div>

        <div
          id="kpi-agents-btn"
          onClick={() => setHubViewMode("agentes")}
          className="bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 shadow-xs cursor-pointer hover:border-[var(--acc)] hover:bg-[var(--accs)]/10 transition-colors"
        >
          <div className="text-[11px] font-medium text-[var(--mute)] flex items-center justify-between">
            <span>Agentes de IA</span>
            <span className="text-[10px] text-[var(--acc)] font-bold">Cockpit →</span>
          </div>
          <div className="text-lg sm:text-xl font-bold text-[var(--acc)] mt-0.5">5 preparados</div>
          <div className="text-[10px] text-[var(--mute)] mt-0.5">Prontos sob seu comando</div>
        </div>
      </section>

      {/* b) Continuar de Onde Parei */}
      {recentBoards.length > 0 && (
        <section id="hub-resume-section">
          <div className="flex items-center justify-between mb-2 px-1">
            <h3 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-1.5">
              <span>⏱️</span>
              <span>Continuar de onde parei</span>
            </h3>
            <span className="text-xs text-[var(--mute)]">Quadros recentes</span>
          </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {recentBoards.map((app) => {
            const margin = app.monthly_revenue_cents - app.monthly_cost_cents;
            const appCards = cards.filter((c) => c.app_name === app.name);

            return (
              <div
                key={app.id}
                className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3.5 shadow-sm hover:border-[var(--acc)] transition-all flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="text-[11px] px-2 py-0.5 rounded-full font-semibold bg-[var(--accs)] text-[var(--acc)]">
                      {app.hosting} · {app.database}
                    </span>
                    <span className={`text-[11px] font-bold ${margin >= 0 ? "text-[var(--ok)]" : "text-[var(--bad)]"}`}>
                      {formatBRL(margin)}/mês
                    </span>
                  </div>
                  <h4 className="text-sm font-bold text-[var(--ink)] m-0 leading-snug">{app.name}</h4>
                  <div className="text-xs text-[var(--mute)] mt-0.5">{app.client_name}</div>
                </div>

                <div className="mt-3 pt-2.5 border-t border-[var(--line)]/60 flex items-center justify-between text-xs">
                  <span className="text-[var(--mute)]">{appCards.length} cartões no fluxo</span>
                  <button
                    id={`btn-open-recent-board-${app.id}`}
                    onClick={() => onNavigateToTab("quadro")}
                    className="text-[var(--acc)] font-semibold hover:underline bg-transparent border-0 cursor-pointer p-0"
                  >
                    Abrir quadro →
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </section>
      )}

      {/* c) Painel de Produção (Etiquetas Reais) */}
      <section id="hub-production-labels-section" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 sm:p-5 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1 mb-3">
          <div>
            <h3 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-1.5">
              <span>🏷️</span>
              <span>Painel de Produção por Etiquetas</span>
            </h3>
            <p className="text-xs text-[var(--mute)] m-0 mt-0.5">
              Contadores das 7 etapas reais do ciclo de vida dos cartões da empresa
            </p>
          </div>
          {selectedStageFilter && (
            <button
              onClick={() => setSelectedStageFilter(null)}
              className="text-xs text-[var(--acc)] font-semibold hover:underline bg-transparent border-0 cursor-pointer self-start sm:self-auto"
            >
              Limpar filtro de etiqueta
            </button>
          )}
        </div>

        {/* Pílulas de Contagem por Etiqueta */}
        <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-7 gap-2">
          {productionStages.map((stage) => {
            const count = stageCounts[stage.key] || 0;
            const isSelected = selectedStageFilter === stage.key;

            return (
              <button
                key={stage.key}
                id={`stage-filter-${stage.key}`}
                onClick={() => setSelectedStageFilter(isSelected ? null : stage.key)}
                className={`p-2.5 rounded-xl border text-left cursor-pointer transition-all ${
                  isSelected
                    ? "border-[var(--acc)] bg-[var(--accs)] shadow-sm scale-102"
                    : stage.isGate
                    ? "border-[var(--warn)]/40 bg-[var(--warns)]/10 hover:border-[var(--warn)]"
                    : "border-[var(--line)] bg-[var(--bg)]/60 hover:border-[var(--mute)]"
                }`}
              >
                <div className="flex items-center justify-between text-xs mb-1">
                  <span className="font-bold text-base text-[var(--ink)]">{count}</span>
                  {stage.isGate && <span title="Portão exclusivo do dono">🛡️</span>}
                </div>
                <div className="text-[11px] font-medium text-[var(--ink)] leading-tight line-clamp-1">
                  {stage.label}
                </div>
              </button>
            );
          })}
        </div>

        {/* Lista de cartões filtrados pela etiqueta selecionada */}
        {selectedStageFilter && (
          <div className="mt-4 pt-3 border-t border-[var(--line)]/60">
            <div className="text-xs font-semibold text-[var(--mute)] mb-2">
              Cartões em &quot;{productionStages.find((s) => s.key === selectedStageFilter)?.label}&quot; ({filteredCards.length}):
            </div>
            {filteredCards.length === 0 ? (
              <div className="text-xs text-[var(--mute)] text-center py-4 bg-[var(--bg)]/40 rounded-xl">
                Nenhum cartão nesta etapa no momento.
              </div>
            ) : (
              <div className="space-y-2 max-h-60 overflow-y-auto pr-1">
                {filteredCards.map((c) => (
                  <div
                    key={c.id}
                    onClick={() => onOpenCard(c.id)}
                    className="p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] hover:border-[var(--acc)] cursor-pointer flex items-center justify-between gap-2 text-xs transition-colors"
                  >
                    <div>
                      <div className="font-semibold text-[var(--ink)]">{c.title}</div>
                      <div className="text-[11px] text-[var(--mute)]">
                        {c.client_name} · {c.app_name}
                      </div>
                    </div>
                    <div className="text-right flex-none">
                      <div className="font-bold text-[var(--ink)]">
                        {c.price_cents > 0 ? formatBRL(c.price_cents) : "Sem custo"}
                      </div>
                      <span className="text-[10px] text-[var(--acc)] font-medium">Ver detalhes →</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </section>

      {/* f) Alertas e Pendências (Decisões nos Portões) */}
      {pendingGateCards.length > 0 && (
        <section id="hub-alerts-section" className="bg-[var(--surf)] border border-[var(--warn)]/50 rounded-2xl p-4 shadow-sm">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-bold text-[var(--warn)] m-0 flex items-center gap-1.5">
              <span>🛡️</span>
              <span>Portões Aguardando Sua Aprovação ({pendingGateCards.length})</span>
            </h3>
            <span className="text-xs font-semibold text-[var(--warn)]">
              {formatBRL(totalMoneyWaiting)} em jogo
            </span>
          </div>

          <div className="space-y-2.5">
            {pendingGateCards.slice(0, 3).map((card) => (
              <div
                key={card.id}
                className="bg-[var(--bg)] border border-[var(--line)] rounded-xl p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs"
              >
                <div>
                  <div className="flex items-center gap-2 mb-0.5">
                    <span className="font-bold text-[var(--ink)] text-sm">{card.title}</span>
                    <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-[var(--warns)] text-[var(--warn)]">
                      {card.stage === "proposta_aprovar" && "Proposta"}
                      {card.stage === "prompt_aprovar" && "Prompt de Execução"}
                      {card.stage === "entrega_aprovar" && "Entrega Final"}
                    </span>
                  </div>
                  <div className="text-[var(--mute)]">
                    {card.client_name} · {card.app_name} · Valor: <b className="text-[var(--ink)]">{formatBRL(card.price_cents)}</b>
                  </div>
                </div>

                <div className="flex items-center gap-2 self-end sm:self-auto">
                  <button
                    onClick={() => onOpenCard(card.id)}
                    className="px-3 py-1.5 rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] font-semibold cursor-pointer hover:bg-[var(--line)]/30"
                  >
                    Ver detalhes
                  </button>
                  <button
                    onClick={() => onQuickApprove(card.id)}
                    disabled={isProcessingDecision || isPaused}
                    className="px-3 py-1.5 rounded-lg border-0 bg-[var(--acc)] text-white font-bold cursor-pointer hover:opacity-90 disabled:opacity-50"
                  >
                    {isProcessingDecision ? "Aprovando..." : "Aprovar rápido"}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* d) Todos os Quadros (Apps e Projetos) */}
      <section id="hub-all-boards-section" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 sm:p-5 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
          <div>
            <h3 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-1.5">
              <span>📂</span>
              <span>Todos os Quadros e Projetos ({sortedAndFilteredApps.length})</span>
            </h3>
            <p className="text-xs text-[var(--mute)] m-0 mt-0.5">
              Grade e lista dos sistemas operados pelos agentes e equipes
            </p>
          </div>

          <div className="flex items-center gap-2">
            {/* Ordenação */}
            <select
              value={boardSortBy}
              onChange={(e) => setBoardSortBy(e.target.value as "recent" | "name" | "margin")}
              className="text-xs px-2.5 py-1.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none"
            >
              <option value="recent">Mais recentes</option>
              <option value="name">Nome (A-Z)</option>
              <option value="margin">Maior margem</option>
            </select>

            {/* Alternador Grade / Lista */}
            <div className="flex rounded-xl border border-[var(--line)] overflow-hidden">
              <button
                id="btn-view-grid"
                onClick={() => setBoardViewMode("grid")}
                className={`px-2.5 py-1.5 text-xs border-0 cursor-pointer ${
                  boardViewMode === "grid" ? "bg-[var(--acc)] text-white font-bold" : "bg-[var(--bg)] text-[var(--mute)]"
                }`}
                title="Visualização em Grade"
              >
                ⊞ Grade
              </button>
              <button
                id="btn-view-list"
                onClick={() => setBoardViewMode("list")}
                className={`px-2.5 py-1.5 text-xs border-0 cursor-pointer ${
                  boardViewMode === "list" ? "bg-[var(--acc)] text-white font-bold" : "bg-[var(--bg)] text-[var(--mute)]"
                }`}
                title="Visualização em Lista"
              >
                ☰ Lista
              </button>
            </div>
          </div>
        </div>

        {/* Visualização em Grade ou Estado Vazio */}
        {sortedAndFilteredApps.length === 0 ? (
          <div className="text-center py-10 bg-[var(--bg)]/40 rounded-2xl border border-dashed border-[var(--line)] p-6">
            <span className="text-3xl mb-2 block">📋</span>
            <h4 className="text-sm font-bold text-[var(--ink)] m-0 mb-1">Nenhum quadro ou app em produção</h4>
            <p className="text-xs text-[var(--mute)] max-w-sm mx-auto mb-4">
              Comece cadastrando seu primeiro projeto ou aplicativo para organizar o fluxo de trabalho.
            </p>
            <div className="flex flex-wrap items-center justify-center gap-2">
              <button
                onClick={onOpenNewBoardModal}
                className="bg-[var(--acc)] text-white text-xs font-bold py-2 px-4 rounded-xl border-0 cursor-pointer hover:opacity-90 transition-opacity"
              >
                + Criar Primeiro Quadro
              </button>
              {onOpenNewCardModal && (
                <button
                  onClick={onOpenNewCardModal}
                  className="bg-[var(--surf)] border border-[var(--acc)] text-[var(--acc)] text-xs font-bold py-2 px-4 rounded-xl cursor-pointer hover:bg-[var(--accs)] transition-colors"
                >
                  + Nova Demanda
                </button>
              )}
            </div>
          </div>
        ) : boardViewMode === "grid" ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
            {sortedAndFilteredApps.map((app) => {
              const margin = app.monthly_revenue_cents - app.monthly_cost_cents;
              const appCards = cards.filter((c) => c.app_name === app.name);

              return (
                <div
                  key={app.id}
                  className="bg-[var(--bg)] border border-[var(--line)] rounded-2xl p-4 shadow-xs hover:border-[var(--acc)] transition-all flex flex-col justify-between"
                >
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-[11px] px-2 py-0.5 rounded-full font-medium bg-[var(--accs)] text-[var(--acc)]">
                        {app.hosting}
                      </span>
                      <span className={`text-xs font-bold ${margin >= 0 ? "text-[var(--ok)]" : "text-[var(--bad)]"}`}>
                        {formatBRL(margin)}/mês
                      </span>
                    </div>

                    <h4 className="text-base font-bold text-[var(--ink)] m-0 leading-tight">{app.name}</h4>
                    <div className="text-xs text-[var(--mute)] mt-1">{app.client_name}</div>

                    {/* Resumo de etiquetas / progresso */}
                    <div className="mt-3">
                      <div className="flex justify-between text-[10px] text-[var(--mute)] mb-1">
                        <span>Cartões no fluxo</span>
                        <span className="font-semibold">{appCards.length} tarefas</span>
                      </div>
                      <div className="h-1.5 w-full bg-[var(--line)] rounded-full overflow-hidden flex">
                        <div
                          style={{ width: `${Math.min(100, (appCards.length / 5) * 100)}%` }}
                          className="bg-[var(--acc)] h-full rounded-full transition-all"
                        ></div>
                      </div>
                    </div>
                  </div>

                  <div className="mt-4 pt-3 border-t border-[var(--line)]/60 flex items-center justify-between text-xs">
                    <span className="text-[var(--mute)] text-[11px]">Banco: {app.database}</span>
                    <button
                      onClick={() => onNavigateToTab("quadro")}
                      className="text-[var(--acc)] font-bold hover:underline bg-transparent border-0 cursor-pointer p-0"
                    >
                      Abrir Quadro →
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          /* Visualização em Lista */
          <div className="divide-y divide-[var(--line)]/60 bg-[var(--bg)] rounded-xl border border-[var(--line)] overflow-hidden">
            {sortedAndFilteredApps.map((app) => {
              const margin = app.monthly_revenue_cents - app.monthly_cost_cents;
              const appCards = cards.filter((c) => c.app_name === app.name);

              return (
                <div
                  key={app.id}
                  className="p-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-2 hover:bg-[var(--surf)]/50 transition-colors text-xs"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-xl bg-[var(--accs)] text-[var(--acc)] flex items-center justify-center font-bold text-sm">
                      {app.name.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <div className="font-bold text-[var(--ink)] text-sm">{app.name}</div>
                      <div className="text-[11px] text-[var(--mute)]">
                        {app.client_name} · {app.hosting} · {app.database}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center justify-between sm:justify-end gap-4">
                    <div className="text-right">
                      <div className={`font-bold ${margin >= 0 ? "text-[var(--ok)]" : "text-[var(--bad)]"}`}>
                        {formatBRL(margin)}/mês
                      </div>
                      <div className="text-[10px] text-[var(--mute)]">{appCards.length} tarefas no fluxo</div>
                    </div>

                    <button
                      onClick={() => onNavigateToTab("quadro")}
                      className="px-3 py-1.5 rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--acc)] font-bold hover:bg-[var(--accs)] cursor-pointer transition-colors"
                    >
                      Abrir
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>

      {/* h) Saúde do Sistema & i) Ações Rápidas */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {/* Saúde do Sistema */}
        <section id="hub-health-section" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
          <h3 className="text-sm font-bold text-[var(--ink)] mb-3 flex items-center gap-1.5">
            <span>🩺</span>
            <span>Saúde da Infraestrutura</span>
          </h3>

          <div className="space-y-2 text-xs">
            <div className="flex items-center justify-between p-2 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
              <span className="font-medium text-[var(--ink)]">API Go (Clean Arch)</span>
              <span className="text-[var(--ok)] font-bold flex items-center gap-1">
                <span className="w-2 h-2 rounded-full bg-[var(--ok)] inline-block animate-pulse"></span>
                Operacional
              </span>
            </div>

            <div className="flex items-center justify-between p-2 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
              <span className="font-medium text-[var(--ink)]">Supabase PostgreSQL (RLS)</span>
              <span className="text-[var(--ok)] font-bold flex items-center gap-1">
                <span className="w-2 h-2 rounded-full bg-[var(--ok)] inline-block"></span>
                100% Conectado
              </span>
            </div>

            <div className="flex items-center justify-between p-2 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
              <span className="font-medium text-[var(--ink)]">Agentes Autônomos</span>
              <span className="text-[var(--mute)] font-semibold flex items-center gap-1">
                <span>⏸️</span> Em espera (Segurança)
              </span>
            </div>

            <div className="flex items-center justify-between p-2 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)]">
              <span className="font-medium text-[var(--ink)]">Parada de Emergência</span>
              <span className={isPaused ? "text-[var(--bad)] font-bold" : "text-[var(--ok)] font-bold"}>
                {isPaused ? "ATIVADA (Pausado)" : "Normal (Pronto)"}
              </span>
            </div>
          </div>
        </section>

        {/* Ações Rápidas */}
        <section id="hub-quick-actions-section" className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
          <h3 className="text-sm font-bold text-[var(--ink)] mb-3 flex items-center gap-1.5">
            <span>⚡</span>
            <span>Ações Rápidas da Operação</span>
          </h3>

          <div className="grid grid-cols-2 gap-2 text-xs">
            <button
              onClick={onOpenNewBoardModal}
              className="p-3 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-left font-semibold text-[var(--ink)] hover:border-[var(--acc)] hover:bg-[var(--accs)] cursor-pointer transition-colors"
            >
              <div className="text-base mb-1">📋</div>
              <div>Novo Quadro</div>
              <div className="text-[10px] text-[var(--mute)] font-normal">Criar projeto</div>
            </button>

            <button
              onClick={() => onNavigateToTab("quadro")}
              className="p-3 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-left font-semibold text-[var(--ink)] hover:border-[var(--acc)] hover:bg-[var(--accs)] cursor-pointer transition-colors"
            >
              <div className="text-base mb-1">📊</div>
              <div>Quadro Kanban</div>
              <div className="text-[10px] text-[var(--mute)] font-normal">Fluxo de tarefas</div>
            </button>

            <button
              onClick={() => onNavigateToTab("fin")}
              className="p-3 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-left font-semibold text-[var(--ink)] hover:border-[var(--acc)] hover:bg-[var(--accs)] cursor-pointer transition-colors"
            >
              <div className="text-base mb-1">💰</div>
              <div>Finanças</div>
              <div className="text-[10px] text-[var(--mute)] font-normal">DRE e margens</div>
            </button>

            <button
              onClick={() => onNavigateToTab("con")}
              className="p-3 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-left font-semibold text-[var(--ink)] hover:border-[var(--acc)] hover:bg-[var(--accs)] cursor-pointer transition-colors"
            >
              <div className="text-base mb-1">🔌</div>
              <div>Conexões</div>
              <div className="text-[10px] text-[var(--mute)] font-normal">Provedores e IA</div>
            </button>
          </div>
        </section>
      </div>
          </>
        )}
    </div>
  );
}
