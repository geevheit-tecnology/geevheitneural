"use client";

import React from "react";
import { CardDetails } from "@/shared/api-client/endpoints";

interface CardDetailSheetProps {
  cardDetails: CardDetails | null;
  isOpen: boolean;
  onClose: () => void;
  onApprove: (cardId: string) => void;
  onAskInfo: (cardId: string) => void;
  onReject: (cardId: string) => void;
  isProcessing: boolean;
}

export function CardDetailSheet({
  cardDetails,
  isOpen,
  onClose,
  onApprove,
  onAskInfo,
  onReject,
  isProcessing,
}: CardDetailSheetProps) {
  const [riskAccepted, setRiskAccepted] = React.useState(false);

  if (!isOpen || !cardDetails) return null;

  const { card, events, risk_report, actions, margin_cents } = cardDetails;
  const isDeliveryGate = card.stage === "entrega_aprovar";

  const typeConfig: Record<string, { label: string; color: string }> = {
    orcamento: { label: "Orçamento", color: "#0e6b73" },
    bug: { label: "Bug", color: "#b42318" },
    suporte: { label: "Suporte", color: "#b45309" },
    financeiro: { label: "Financeiro", color: "#6d4bb5" },
  };

  const currentType = typeConfig[card.type] || { label: card.type, color: "var(--mute)" };

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  return (
    <div
      id="card-detail-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      className="fixed inset-0 bg-black/70 z-50 flex items-end justify-center sm:items-center sm:p-4"
    >
      <div
        id="card-detail-sheet"
        className="bg-[var(--surf)] border-t sm:border border-[var(--line)] w-full max-w-[560px] rounded-t-3xl sm:rounded-2xl p-5 max-h-[85vh] overflow-y-auto animate-in slide-in-from-bottom-5 duration-200"
        style={{ paddingBottom: "calc(20px + env(safe-area-inset-bottom, 0px))" }}
      >
        {/* Header do Cartão */}
        <div className="flex items-center justify-between mb-2">
          <span
            className="text-xs px-2.5 py-0.5 rounded-full font-semibold"
            style={{ backgroundColor: "var(--accs)", color: currentType.color }}
          >
            {currentType.label}
          </span>
          <button
            onClick={onClose}
            className="text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 text-xl cursor-pointer p-1"
          >
            ✕
          </button>
        </div>

        <h3 className="text-lg font-bold text-[var(--ink)] m-0 mb-1 leading-snug">
          {card.title}
        </h3>
        <div className="text-xs text-[var(--mute)] mb-4">
          {card.client_name} · {card.app_name !== "Novo" ? card.app_name : "Novo Projeto"}
        </div>

        {/* Linhas de Valores Financeiros */}
        <div className="bg-[var(--bg)]/50 border border-[var(--line)] rounded-xl p-3 space-y-2 mb-4 text-sm">
          <div className="flex justify-between">
            <span className="text-[var(--mute)]">Valor cobrado</span>
            <b className="text-[var(--ink)]">
              {card.price_cents > 0 ? formatBRL(card.price_cents) : "Sem cobrança"}
            </b>
          </div>
          <div className="flex justify-between">
            <span className="text-[var(--mute)]">Custo estimado</span>
            <b className="text-[var(--ink)]">{formatBRL(card.est_cost_cents)}</b>
          </div>
          <div className="flex justify-between border-t border-[var(--line)] pt-2 font-medium">
            <span className="text-[var(--mute)]">Margem prevista</span>
            <b className={margin_cents >= 0 ? "text-[var(--ok)]" : "text-[var(--bad)]"}>
              {formatBRL(margin_cents)}
            </b>
          </div>
        </div>

        {/* Ficha de Risco (quando presente na entrega) */}
        {risk_report && (
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-xl p-3 mb-4 text-xs">
            <b className="text-[var(--ink)] block mb-2 text-sm">Ficha de Risco de Entrega</b>
            <div className="space-y-1.5">
              {risk_report.tests?.map((t, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{t[0]}</span>
                  <b className={t[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {t[1]}
                  </b>
                </div>
              ))}
              {risk_report.secrets_scan?.map((s, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{s[0]}</span>
                  <b className={s[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {s[1]}
                  </b>
                </div>
              ))}
              {risk_report.outdated_deps?.map((d, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{d[0]}</span>
                  <b className={d[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {d[1]}
                  </b>
                </div>
              ))}
              {risk_report.lgpd_points?.map((l, idx) => (
                <div key={idx} className="flex justify-between items-center">
                  <span className="text-[var(--mute)]">{l[0]}</span>
                  <b className={l[2] === "ok" ? "text-[var(--ok)]" : "text-[var(--warn)]"}>
                    {l[1]}
                  </b>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Histórico Cronológico */}
        <div className="mb-5">
          <span className="text-xs font-semibold text-[var(--mute)] uppercase tracking-wider block mb-2">
            Histórico
          </span>
          <ul className="m-0 p-0 pl-3.5 border-l-2 border-[var(--line)] space-y-2 list-none text-xs text-[var(--mute)]">
            {events.length > 0 ? (
              events.map((ev, i) => (
                <li key={i} className="leading-relaxed">
                  <span className="text-[var(--ink)] font-medium">
                    {String(ev.payload?.title || ev.kind)}
                  </span>
                  {Boolean(ev.payload?.note) && <span className="block italic text-[11px]">{String(ev.payload?.note)}</span>}
                </li>
              ))
            ) : (
              <li>Pedido registrado no sistema</li>
            )}
          </ul>
        </div>

        {/* Botões de Ação */}
        <div className="grid gap-2">
          {isDeliveryGate && (
            <label className="flex items-start gap-2 bg-[var(--warns)]/30 border border-[var(--warn)]/40 p-2.5 rounded-xl text-xs cursor-pointer mb-1 text-[var(--ink)]">
              <input
                id="risk-accepted-checkbox"
                type="checkbox"
                checked={riskAccepted}
                onChange={(e) => setRiskAccepted(e.target.checked)}
                className="mt-0.5 accent-[var(--acc)] cursor-pointer"
              />
              <span className="leading-tight">
                Declaro que avaliei a <b>Ficha de Risco</b> acima (testes, segredos, dependências e LGPD) antes de autorizar a entrega ao cliente.
              </span>
            </label>
          )}

          {actions.length > 0 && (
            <button
              id="sheet-approve-btn"
              disabled={isProcessing || (isDeliveryGate && !riskAccepted)}
              onClick={() => onApprove(card.id)}
              className={`w-full py-3 px-4 rounded-xl font-semibold text-sm border-0 transition-opacity ${
                isDeliveryGate && !riskAccepted
                  ? "bg-[var(--line)] text-[var(--mute)] cursor-not-allowed"
                  : "bg-[var(--acc)] text-white cursor-pointer hover:opacity-90"
              }`}
            >
              {actions[0]}
            </button>
          )}

          {card.stage !== "entregue" && card.stage !== "recusado" && (
            <div className="grid grid-cols-2 gap-2">
              <button
                id="sheet-ask-info-btn"
                disabled={isProcessing}
                onClick={() => onAskInfo(card.id)}
                className="py-2.5 px-3 rounded-xl font-semibold text-xs border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] cursor-pointer hover:border-[var(--acc)]"
              >
                Pedir mais informações
              </button>
              <button
                id="sheet-reject-btn"
                disabled={isProcessing}
                onClick={() => onReject(card.id)}
                className="py-2.5 px-3 rounded-xl font-semibold text-xs border border-[var(--line)] bg-[var(--surf)] text-[var(--bad)] cursor-pointer hover:border-[var(--bad)]"
              >
                Recusar
              </button>
            </div>
          )}

          <button
            id="sheet-close-btn"
            onClick={onClose}
            className="w-full py-2.5 px-4 rounded-xl font-semibold text-xs border border-[var(--line)] bg-transparent text-[var(--mute)] cursor-pointer hover:text-[var(--ink)]"
          >
            Fechar
          </button>
        </div>
      </div>
    </div>
  );
}
