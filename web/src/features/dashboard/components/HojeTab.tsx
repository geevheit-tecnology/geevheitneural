"use client";

import React, { useState } from "react";
import { CardItem } from "@/shared/api-client/endpoints";
import { AgendaView } from "./AgendaView";

interface HojeTabProps {
  cards: CardItem[];
  onOpenCard: (cardId: string) => void;
  onQuickApprove: (cardId: string) => void;
  isProcessing: boolean;
}

export function HojeTab({ cards, onOpenCard, onQuickApprove, isProcessing }: HojeTabProps) {
  const [viewMode, setViewMode] = useState<"decisoes" | "agenda">("decisoes");

  // Filtra cartões que exigem decisão do dono ou entrada
  const pendingStages = ["entrada", "proposta_aprovar", "prompt_aprovar", "entrega_aprovar"];
  const priorityCards = cards
    .filter((c) => pendingStages.includes(c.stage))
    .sort((a, b) => b.price_cents - a.price_cents)
    .slice(0, 3);

  const typeConfig: Record<string, { label: string; color: string }> = {
    orcamento: { label: "Orçamento", color: "#0e6b73" },
    bug: { label: "Bug", color: "#b42318" },
    suporte: { label: "Suporte", color: "#b45309" },
    financeiro: { label: "Financeiro", color: "#6d4bb5" },
  };

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  const getRecommendation = (card: CardItem) => {
    switch (card.stage) {
      case "entrada":
        return "Responda hoje para não deixar o cliente esperando.";
      case "proposta_aprovar":
        const margin = card.price_cents - card.est_cost_cents;
        return `Aprovar hoje evita o cliente esfriar. Margem prevista: ${formatBRL(margin)}.`;
      case "prompt_aprovar":
        return "O sinal já foi pago. Aprovar libera a execução.";
      case "entrega_aprovar":
        return "Os testes passaram. Aprovar libera a entrega e o recebimento.";
      default:
        return "Decisão aguardando sua revisão.";
    }
  };

  const getActionLabel = (stage: string) => {
    switch (stage) {
      case "entrada":
        return "Mover para proposta";
      case "proposta_aprovar":
        return "Aprovar e enviar ao cliente";
      case "prompt_aprovar":
        return "Aprovar prompt e liberar";
      case "entrega_aprovar":
        return "Aprovar e entregar ao cliente";
      default:
        return "Avançar";
    }
  };

  return (
    <div id="tab-hoje" className="space-y-3">
      {/* Alternador de Modo: 3 Decisões vs Agenda Completa */}
      <div className="flex bg-[var(--surf)] p-1 rounded-xl border border-[var(--line)]">
        <button
          id="toggle-view-decisoes"
          onClick={() => setViewMode("decisoes")}
          className={`flex-1 py-1.5 text-xs font-bold rounded-lg transition-all cursor-pointer border-0 ${
            viewMode === "decisoes"
              ? "bg-[var(--acc)] text-white shadow-sm"
              : "bg-transparent text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          ⚡ 3 Decisões do Dia
        </button>
        <button
          id="toggle-view-agenda"
          onClick={() => setViewMode("agenda")}
          className={`flex-1 py-1.5 text-xs font-bold rounded-lg transition-all cursor-pointer border-0 flex items-center justify-center gap-1.5 ${
            viewMode === "agenda"
              ? "bg-[var(--acc)] text-white shadow-sm"
              : "bg-transparent text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          <span>📅</span>
          <span>Agenda & Prazos</span>
        </button>
      </div>

      {viewMode === "agenda" ? (
        <AgendaView cards={cards} onOpenCard={onOpenCard} />
      ) : (
        <>
          <div className="bg-[var(--accs)] text-[var(--ink)] rounded-xl p-3 text-xs sm:text-sm font-medium">
            Suas 3 decisões do dia, ordenadas por dinheiro em jogo.
          </div>

          {priorityCards.length === 0 ? (
            <div className="text-center py-12 text-[var(--mute)] text-sm">
              Nada esperando por você hoje. Todas as decisões estão em dia!
            </div>
          ) : (

        priorityCards.map((card) => {
          const tConfig = typeConfig[card.type] || { label: card.type, color: "var(--mute)" };
          const actionLabel = getActionLabel(card.stage);

          return (
            <div
              key={card.id}
              className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm"
            >
              <div className="flex items-center justify-between mb-2">
                <span
                  className="text-xs px-2.5 py-0.5 rounded-full font-semibold"
                  style={{ backgroundColor: "var(--accs)", color: tConfig.color }}
                >
                  {tConfig.label}
                </span>
                <span className="text-xs font-semibold text-[var(--ink)]">
                  {card.price_cents > 0 ? `${formatBRL(card.price_cents)} em jogo` : "Sem valor"}
                </span>
              </div>

              <h3 className="text-base font-bold text-[var(--ink)] m-0 mb-1 leading-snug">
                {card.title}
              </h3>
              <div className="text-xs text-[var(--mute)] mb-3">{card.client_name}</div>

              <p className="text-xs sm:text-sm text-[var(--ink)] bg-[var(--bg)]/50 border border-[var(--line)] rounded-xl p-2.5 m-0 mb-3.5 leading-relaxed">
                💡 {getRecommendation(card)}
              </p>

              <div className="grid grid-cols-2 gap-2">
                <button
                  id={`hoje-approve-${card.id}`}
                  disabled={isProcessing}
                  onClick={() => onQuickApprove(card.id)}
                  className="py-2.5 px-3 rounded-xl font-semibold text-xs bg-[var(--acc)] text-white border-0 cursor-pointer hover:opacity-90 transition-opacity"
                >
                  {actionLabel}
                </button>
                <button
                  id={`hoje-details-${card.id}`}
                  onClick={() => onOpenCard(card.id)}
                  className="py-2.5 px-3 rounded-xl font-semibold text-xs border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] cursor-pointer hover:border-[var(--acc)] transition-colors"
                >
                  Ver detalhes
                </button>
              </div>
            </div>
          );
        })
      )}
        </>
      )}
    </div>
  );
}
