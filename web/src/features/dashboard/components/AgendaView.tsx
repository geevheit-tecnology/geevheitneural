"use client";

import React, { useState } from "react";
import { CardItem } from "@/shared/api-client/endpoints";

interface AgendaViewProps {
  cards: CardItem[];
  onOpenCard: (cardId: string) => void;
}

export function AgendaView({ cards, onOpenCard }: AgendaViewProps) {
  const [selectedFilter, setSelectedFilter] = useState<"todos" | "hoje" | "atrasados" | "semana">("todos");
  const [selectedDayOffset, setSelectedDayOffset] = useState<number | null>(null);

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  // Gera os próximos 7 dias para o calendário rápido no topo
  const weekDays = Array.from({ length: 7 }, (_, i) => {
    const d = new Date(today);
    d.setDate(today.getDate() + i);
    return d;
  });

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  const getCardDueDate = (card: CardItem): Date => {
    if (card.due_date) {
      return new Date(card.due_date);
    }
    // Fallback simulado para cards sem data explícita com base no ID
    const dummyOffset = (parseInt(card.id.replace(/\D/g, "") || "1", 10) % 5) - 1;
    const d = new Date(today);
    d.setDate(today.getDate() + dummyOffset);
    return d;
  };

  // Classifica e enriquece os cartões com status de prazo
  const enrichedCards = cards.map((card) => {
    const dueDate = getCardDueDate(card);
    const dueDateZero = new Date(dueDate);
    dueDateZero.setHours(0, 0, 0, 0);

    const diffTime = dueDateZero.getTime() - today.getTime();
    const diffDays = Math.round(diffTime / (1000 * 60 * 60 * 24));

    let statusType: "atrasado" | "hoje" | "semana" | "futuro" = "futuro";
    let statusLabel = "";
    let statusBadgeColor = "";

    if (diffDays < 0) {
      statusType = "atrasado";
      statusLabel = `Atrasado há ${Math.abs(diffDays)} dia(s)`;
      statusBadgeColor = "bg-rose-500/15 text-rose-500 border-rose-500/20";
    } else if (diffDays === 0) {
      statusType = "hoje";
      statusLabel = "Vence Hoje!";
      statusBadgeColor = "bg-amber-500/15 text-amber-500 border-amber-500/20";
    } else if (diffDays <= 7) {
      statusType = "semana";
      statusLabel = diffDays === 1 ? "Vence amanhã" : `Vence em ${diffDays} dias`;
      statusBadgeColor = "bg-emerald-500/15 text-emerald-500 border-emerald-500/20";
    } else {
      statusType = "futuro";
      statusLabel = `Previsto para ${dueDate.toLocaleDateString("pt-BR", { day: "2-digit", month: "2-digit" })}`;
      statusBadgeColor = "bg-slate-500/15 text-slate-400 border-slate-500/20";
    }

    return {
      ...card,
      calculatedDueDate: dueDate,
      diffDays,
      statusType,
      statusLabel,
      statusBadgeColor,
    };
  });

  // Filtra de acordo com o filtro selecionado ou dia do calendário
  const filteredCards = enrichedCards.filter((c) => {
    if (selectedDayOffset !== null) {
      return c.diffDays === selectedDayOffset;
    }
    if (selectedFilter === "hoje") return c.statusType === "hoje";
    if (selectedFilter === "atrasados") return c.statusType === "atrasado";
    if (selectedFilter === "semana") return c.diffDays >= 0 && c.diffDays <= 7;
    return true;
  }).sort((a, b) => a.diffDays - b.diffDays);

  const overdueCount = enrichedCards.filter((c) => c.statusType === "atrasado").length;
  const todayCount = enrichedCards.filter((c) => c.statusType === "hoje").length;
  const weekCount = enrichedCards.filter((c) => c.diffDays >= 0 && c.diffDays <= 7).length;

  return (
    <div id="agenda-view-container" className="space-y-4">
      {/* 1. Calendário Rápido Semanal */}
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3 shadow-sm">
        <div className="flex items-center justify-between mb-2 px-1">
          <span className="text-xs font-bold text-[var(--ink)] uppercase tracking-wider flex items-center gap-1.5">
            <span>📅</span>
            <span>Agenda Semanal</span>
          </span>
          {selectedDayOffset !== null && (
            <button
              onClick={() => setSelectedDayOffset(null)}
              className="text-[11px] font-semibold text-[var(--acc)] hover:underline bg-transparent border-0 cursor-pointer p-0"
            >
              Ver todos os dias
            </button>
          )}
        </div>

        <div className="grid grid-cols-7 gap-1.5 text-center">
          {weekDays.map((date, idx) => {
            const isToday = idx === 0;
            const isSelected = selectedDayOffset === idx;
            const dayCards = enrichedCards.filter((c) => c.diffDays === idx);
            const count = dayCards.length;

            return (
              <button
                key={idx}
                id={`calendar-day-${idx}`}
                onClick={() => setSelectedDayOffset(isSelected ? null : idx)}
                className={`flex flex-col items-center justify-center py-2 px-1 rounded-xl transition-all border cursor-pointer ${
                  isSelected
                    ? "bg-[var(--acc)] text-white border-[var(--acc)] shadow-md shadow-[var(--acc)]/20 scale-[1.03]"
                    : isToday
                    ? "bg-[var(--accs)] border-[var(--acc)]/40 text-[var(--ink)]"
                    : "bg-[var(--bg)]/50 border-[var(--line)] text-[var(--mute)] hover:border-[var(--line-strong)]"
                }`}
              >
                <span className="text-[10px] font-medium uppercase">
                  {date.toLocaleDateString("pt-BR", { weekday: "narrow" })}
                </span>
                <span className={`text-sm font-bold my-0.5 ${isSelected ? "text-white" : "text-[var(--ink)]"}`}>
                  {date.getDate()}
                </span>
                {count > 0 ? (
                  <span
                    className={`text-[10px] px-1.5 py-0.2 rounded-full font-bold ${
                      isSelected
                        ? "bg-white/20 text-white"
                        : isToday
                        ? "bg-amber-500/20 text-amber-500"
                        : "bg-[var(--accs)] text-[var(--acc)]"
                    }`}
                  >
                    {count}
                  </span>
                ) : (
                  <span className="text-[10px] opacity-0">•</span>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* 2. Filtros de Atalho */}
      <div className="flex gap-2 overflow-x-auto pb-1 -mx-1 px-1" style={{ scrollbarWidth: "none" }}>
        <button
          onClick={() => { setSelectedFilter("todos"); setSelectedDayOffset(null); }}
          className={`flex-none text-xs font-semibold px-3 py-1.5 rounded-full transition-colors border cursor-pointer ${
            selectedFilter === "todos" && selectedDayOffset === null
              ? "bg-[var(--acc)] text-white border-[var(--acc)]"
              : "bg-[var(--surf)] text-[var(--mute)] border-[var(--line)]"
          }`}
        >
          Todas as Entregas ({enrichedCards.length})
        </button>

        <button
          onClick={() => { setSelectedFilter("hoje"); setSelectedDayOffset(null); }}
          className={`flex-none text-xs font-semibold px-3 py-1.5 rounded-full transition-colors border cursor-pointer flex items-center gap-1.5 ${
            selectedFilter === "hoje" && selectedDayOffset === null
              ? "bg-amber-500 text-slate-950 border-amber-500 font-bold"
              : "bg-[var(--surf)] text-amber-500 border-amber-500/30"
          }`}
        >
          <span className="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
          Hoje ({todayCount})
        </button>

        {overdueCount > 0 && (
          <button
            onClick={() => { setSelectedFilter("atrasados"); setSelectedDayOffset(null); }}
            className={`flex-none text-xs font-semibold px-3 py-1.5 rounded-full transition-colors border cursor-pointer flex items-center gap-1.5 ${
              selectedFilter === "atrasados" && selectedDayOffset === null
                ? "bg-rose-600 text-white border-rose-600 font-bold"
                : "bg-[var(--surf)] text-rose-500 border-rose-500/30"
            }`}
          >
            <span className="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse"></span>
            Atrasados ({overdueCount})
          </button>
        )}

        <button
          onClick={() => { setSelectedFilter("semana"); setSelectedDayOffset(null); }}
          className={`flex-none text-xs font-semibold px-3 py-1.5 rounded-full transition-colors border cursor-pointer ${
            selectedFilter === "semana" && selectedDayOffset === null
              ? "bg-[var(--acc)] text-white border-[var(--acc)]"
              : "bg-[var(--surf)] text-[var(--mute)] border-[var(--line)]"
          }`}
        >
          Próximos 7 Dias ({weekCount})
        </button>
      </div>

      {/* 3. Lista de Demandas & Prazos da Agenda */}
      <div className="space-y-2.5">
        {filteredCards.length === 0 ? (
          <div className="text-center py-12 bg-[var(--surf)]/50 border border-[var(--line)] rounded-2xl text-[var(--mute)] text-sm">
            Nenhuma entrega ou compromisso agendado para este filtro.
          </div>
        ) : (
          filteredCards.map((card) => (
            <div
              key={card.id}
              onClick={() => onOpenCard(card.id)}
              className="bg-[var(--surf)] border border-[var(--line)] hover:border-[var(--acc)] rounded-2xl p-4 shadow-sm transition-all cursor-pointer hover:shadow-md"
            >
              <div className="flex items-center justify-between mb-2">
                <span
                  className={`text-[11px] px-2 py-0.5 rounded-full font-bold border ${card.statusBadgeColor}`}
                >
                  {card.statusLabel}
                </span>

                <span className="text-xs font-extrabold text-[var(--ink)]">
                  {card.price_cents > 0 ? formatBRL(card.price_cents) : "Suporte / Garantia"}
                </span>
              </div>

              <h3 className="text-base font-bold text-[var(--ink)] m-0 mb-1 leading-snug">
                {card.title}
              </h3>

              <div className="flex items-center justify-between text-xs text-[var(--mute)] mt-3 pt-2.5 border-t border-[var(--line)]/60">
                <span className="flex items-center gap-1.5">
                  <span className="font-semibold text-[var(--ink)]">{card.client_name}</span>
                  <span>•</span>
                  <span>{card.app_name}</span>
                </span>

                <span className="text-[11px] font-medium text-[var(--acc)] flex items-center gap-1">
                  Ver detalhes →
                </span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
