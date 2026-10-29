"use client";

import React from "react";
import { CardItem } from "@/shared/api-client/endpoints";

interface QuadroTabProps {
  cards: CardItem[];
  onOpenCard: (cardId: string) => void;
}

export function QuadroTab({ cards, onOpenCard }: QuadroTabProps) {
  const stages = [
    { key: "entrada", label: "Entrada", isGate: false },
    { key: "proposta_aprovar", label: "Proposta para aprovar", isGate: true },
    { key: "aguardando_cliente", label: "Aguardando cliente", isGate: false },
    { key: "prompt_aprovar", label: "Prompt para aprovar", isGate: true },
    { key: "em_execucao", label: "Em execução", isGate: false },
    { key: "entrega_aprovar", label: "Entrega para aprovar", isGate: true },
    { key: "entregue", label: "Entregue", isGate: false },
  ];

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

  const getStageAgentMeta = (card: CardItem) => {
    switch (card.stage) {
      case "entrada":
        return { role: "requisitos", action: "Aguardando triagem BDD", aiCostCents: 0 };
      case "proposta_aprovar":
        return { role: "requisitos", action: "Escopo BDD + Orçamento pronto", aiCostCents: 8 };
      case "aguardando_cliente":
        return { role: "requisitos", action: "Link /escopo enviado ao cliente", aiCostCents: 8 };
      case "prompt_aprovar":
        return { role: "arquiteto", action: "Playbook + Arquitetura definidos", aiCostCents: 20 };
      case "em_execucao":
        return { role: "dev + qa", action: "Construindo & testando em sandbox", aiCostCents: 42 };
      case "entrega_aprovar":
        return { role: "publicador", action: "Prévia ao vivo + Ficha de Risco pronta", aiCostCents: 68 };
      case "entregue":
        return { role: "suporte", action: "Monitoramento pós-entrega ativo", aiCostCents: 75 };
      default:
        return null;
    }
  };

  const isSensitiveCard = (card: CardItem) => {
    if (card.type === "financeiro") return true;
    const lower = (card.title || "").toLowerCase();
    return ["pagamento", "pix", "financeir", "autentica", "login", "2fa", "lgpd", "cobrança", "cobranca"].some(
      (kw) => lower.includes(kw)
    );
  };

  return (
    <div id="tab-quadro">
      <div
        id="kanban-scroll-container"
        className="flex gap-3 overflow-x-auto snap-x snap-mandatory py-1 px-4 -mx-4"
        style={{ scrollbarWidth: "thin" }}
      >
        {stages.map((stage) => {
          const stageCards = cards.filter((c) => c.stage === stage.key);

          return (
            <div
              key={stage.key}
              className={`flex-none w-[82vw] sm:w-[290px] snap-start rounded-2xl p-3 border transition-colors self-start ${
                stage.isGate
                  ? "border-[var(--warn)] bg-[var(--warns)]/10"
                  : "border-[var(--line)] bg-[var(--surf)]/60"
              }`}
            >
              <div className="flex items-center justify-between mb-3 px-1">
                <h2 className="text-xs font-bold text-[var(--ink)] m-0 flex items-center gap-1.5">
                  {stage.isGate && <span className="text-[var(--warn)]">🛡️</span>}
                  <span>{stage.label}</span>
                </h2>
                <span className="text-xs font-medium text-[var(--mute)] bg-[var(--bg)] px-2 py-0.5 rounded-full border border-[var(--line)]">
                  {stageCards.length}
                </span>
              </div>

              <div className="space-y-2.5 min-h-[80px]">
                {stageCards.length === 0 ? (
                  <div className="text-xs text-[var(--mute)] p-3 text-center rounded-xl border border-dashed border-[var(--line)]">
                    Nada por aqui
                  </div>
                ) : (
                  stageCards.map((card) => {
                    const tConfig = typeConfig[card.type] || {
                      label: card.type,
                      color: "var(--mute)",
                    };

                    return (
                      <button
                        key={card.id}
                        id={`card-item-${card.id}`}
                        onClick={() => onOpenCard(card.id)}
                        className="w-full text-left bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 cursor-pointer transition-all hover:border-[var(--acc)] hover:shadow-sm"
                        style={{ borderLeftWidth: "4px", borderLeftColor: tConfig.color }}
                      >
                        <b className="block text-xs sm:text-sm font-semibold text-[var(--ink)] mb-0.5 leading-snug">
                          {card.title}
                        </b>
                        <small className="block text-xs text-[var(--mute)] mb-2">
                          {card.client_name}
                          {card.app_name !== "Novo" ? ` · ${card.app_name}` : ""}
                        </small>

                        {/* Fase 6 (sprit02): Estado visível do agente atuante, o que fez e custo de IA */}
                        {(() => {
                          const agentMeta = getStageAgentMeta(card);
                          if (!agentMeta) return null;
                          return (
                            <div className="mb-2 p-1.5 rounded-lg bg-[var(--bg)]/70 border border-[var(--line)] text-[10px]">
                              <div className="flex items-center justify-between font-bold text-[var(--ink)]">
                                <span className="uppercase text-[var(--acc)]">🤖 {agentMeta.role}</span>
                                <span className="text-[var(--mute)] font-mono">
                                  IA: R$ {(agentMeta.aiCostCents / 100).toFixed(2).replace(".", ",")}
                                </span>
                              </div>
                              <div className="text-[var(--mute)] truncate mt-0.5">{agentMeta.action}</div>
                            </div>
                          );
                        })()}

                        {isSensitiveCard(card) && (
                          <div className="mb-2 text-[10px] font-bold px-2 py-0.5 rounded bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30 inline-block">
                            🛡️ revisão_humana_obrigatória
                          </div>
                        )}

                        <div className="flex items-center justify-between text-xs">
                          <span
                            className="text-[11px] px-2 py-0.5 rounded-full font-semibold"
                            style={{
                              backgroundColor: "var(--accs)",
                              color: tConfig.color,
                            }}
                          >
                            {tConfig.label}
                          </span>
                          <span className="font-semibold text-[var(--ink)]">
                            {card.price_cents > 0 ? formatBRL(card.price_cents) : "Sem custo"}
                          </span>
                        </div>
                      </button>
                    );
                  })
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
