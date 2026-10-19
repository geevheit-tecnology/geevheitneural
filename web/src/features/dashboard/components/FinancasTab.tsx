"use client";

import React, { useState } from "react";
import { FinancesOverview, AppItem, CreateFinanceEntryInput } from "@/shared/api-client/endpoints";

interface FinancasTabProps {
  finances: FinancesOverview | null;
  apps: AppItem[];
  onCreateEntry: (data: CreateFinanceEntryInput) => void;
  isProcessing: boolean;
}

export function FinancasTab({ finances, apps, onCreateEntry, isProcessing }: FinancasTabProps) {
  const [showModal, setShowModal] = useState(false);
  const [kind, setKind] = useState<"receita" | "custo">("custo");
  const [amountReais, setAmountReais] = useState("");
  const [service, setService] = useState("Render");
  const [description, setDescription] = useState("");

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  const totalRev = finances?.total_revenue_cents || 0;
  const totalCost = finances?.total_cost_cents || 0;
  const netProfit = totalRev - totalCost;

  const maxVal = Math.max(
    1,
    ...apps.map((a) => Math.max(a.monthly_revenue_cents, a.monthly_cost_cents))
  );

  const handleEntrySubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const val = parseFloat(amountReais.replace(",", "."));
    if (isNaN(val) || val <= 0) return;

    onCreateEntry({
      kind,
      amount_cents: Math.round(val * 100),
      service,
      description,
    });
    setAmountReais("");
    setDescription("");
    setShowModal(false);
  };

  return (
    <div id="tab-financas" className="space-y-3">
      {/* 3 KPIs Principais */}
      <div className="grid grid-cols-3 gap-2">
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3">
          <small className="text-xs text-[var(--mute)] block">Receita</small>
          <div className="text-base sm:text-lg font-bold text-[var(--ink)] mt-0.5">
            {formatBRL(totalRev)}
          </div>
        </div>
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3">
          <small className="text-xs text-[var(--mute)] block">Custo</small>
          <div className="text-base sm:text-lg font-bold text-[var(--ink)] mt-0.5">
            {formatBRL(totalCost)}
          </div>
        </div>
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3">
          <small className="text-xs text-[var(--mute)] block">Lucro</small>
          <div
            className={`text-base sm:text-lg font-bold mt-0.5 ${
              netProfit >= 0 ? "text-[var(--ok)]" : "text-[var(--bad)]"
            }`}
          >
            {formatBRL(netProfit)}
          </div>
        </div>
      </div>

      {/* Ação de Novo Lançamento */}
      <div className="flex justify-end">
        <button
          id="open-finance-modal-btn"
          onClick={() => setShowModal(true)}
          className="text-xs font-semibold py-1.5 px-3 rounded-xl border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] hover:border-[var(--acc)] cursor-pointer"
        >
          + Novo Lançamento
        </button>
      </div>

      {/* Detalhamento Por App com Barra de Proporção */}
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
        <h3 className="text-sm font-bold text-[var(--ink)] m-0 mb-3">Por app</h3>
        <div className="space-y-3">
          {apps.map((app) => (
            <div key={app.id}>
              <div className="flex justify-between items-center text-xs mb-1">
                <b className="text-[var(--ink)]">{app.name}</b>
                <span className="text-[var(--mute)]">
                  {formatBRL(app.monthly_revenue_cents)} entra · {formatBRL(app.monthly_cost_cents)} sai
                </span>
              </div>
              <div className="h-2 rounded-full bg-[var(--line)] overflow-hidden flex">
                <i
                  style={{
                    width: `${(app.monthly_revenue_cents / maxVal) * 50}%`,
                    backgroundColor: "var(--ok)",
                  }}
                  className="block h-full"
                />
                <i
                  style={{
                    width: `${(app.monthly_cost_cents / maxVal) * 50}%`,
                    backgroundColor: "var(--bad)",
                  }}
                  className="block h-full"
                />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Custos por Serviço e Sugestão de Consolidação */}
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
        <h3 className="text-sm font-bold text-[var(--ink)] m-0 mb-3">Custo por serviço</h3>
        <div className="space-y-2 text-xs">
          {finances?.costs_by_service?.map((s, idx) => (
            <div key={idx} className="flex justify-between items-center py-0.5">
              <span className="text-[var(--mute)]">{s.service_name}</span>
              <b className="text-[var(--ink)]">{formatBRL(s.cost_cents)}</b>
            </div>
          ))}
        </div>

        {finances?.suggestion && (
          <div className="bg-[var(--accs)] text-[var(--ink)] rounded-xl p-3 text-xs mt-3 leading-relaxed">
            💡 {finances.suggestion}
          </div>
        )}
      </div>

      {/* Fase 10: Placar da Fábrica Autônoma */}
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 space-y-3">
        <div className="flex items-center justify-between">
          <b className="text-sm font-bold text-[var(--ink)] flex items-center gap-1.5">
            <span>🏭</span> Placar da Fábrica de Software
          </b>
          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600">
            Fábrica Ativa
          </span>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs">
          <div className="bg-[var(--bg)]/60 border border-[var(--line)] p-2.5 rounded-xl">
            <span className="text-[var(--mute)] block text-[11px]">Lead Time Médio</span>
            <b className="text-base text-[var(--ink)]">48.5h</b>
          </div>
          <div className="bg-[var(--bg)]/60 border border-[var(--line)] p-2.5 rounded-xl">
            <span className="text-[var(--mute)] block text-[11px]">Aprovação de 1ª</span>
            <b className="text-base text-emerald-600">92.5%</b>
          </div>
          <div className="bg-[var(--bg)]/60 border border-[var(--line)] p-2.5 rounded-xl">
            <span className="text-[var(--mute)] block text-[11px]">Decisão do Dono</span>
            <b className="text-base text-[var(--acc)]">1.4h</b>
          </div>
          <div className="bg-[var(--bg)]/60 border border-[var(--line)] p-2.5 rounded-xl">
            <span className="text-[var(--mute)] block text-[11px]">Defeitos Pós-Entrega</span>
            <b className="text-base text-emerald-600">0</b>
          </div>
        </div>

        {/* Fase 8: Playbooks Catalogados */}
        <div className="pt-2 border-t border-[var(--line)] space-y-2">
          <span className="text-xs font-semibold text-[var(--mute)] block">
            Biblioteca de Playbooks & Autonomia:
          </span>
          <div className="grid gap-2 sm:grid-cols-3 text-xs">
            <div className="bg-[var(--bg)]/40 border border-[var(--line)] p-2.5 rounded-xl">
              <b className="text-[var(--ink)] block">SaaS B2B Multi-tenant</b>
              <span className="text-[11px] text-[var(--mute)] block mt-0.5">Go + Postgres RLS + Next.js</span>
              <span className="text-[10px] font-semibold text-[var(--acc)] mt-1 block">Autonomia Nível 2</span>
            </div>
            <div className="bg-[var(--bg)]/40 border border-[var(--line)] p-2.5 rounded-xl">
              <b className="text-[var(--ink)] block">App Mobile / PWA</b>
              <span className="text-[11px] text-[var(--mute)] block mt-0.5">Next.js PWA + REST + SSE</span>
              <span className="text-[10px] font-semibold text-[var(--acc)] mt-1 block">Autonomia Nível 2</span>
            </div>
            <div className="bg-[var(--bg)]/40 border border-[var(--line)] p-2.5 rounded-xl">
              <b className="text-[var(--ink)] block">Landing Page Conversão</b>
              <span className="text-[11px] text-[var(--mute)] block mt-0.5">Next.js Static + Antispam</span>
              <span className="text-[10px] font-semibold text-emerald-600 mt-1 block">Autonomia Nível 3</span>
            </div>
          </div>
        </div>
      </div>

      {/* Modal de Lançamento */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4">
          <form
            onSubmit={handleEntrySubmit}
            className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-sm w-full shadow-2xl space-y-3"
          >
            <h3 className="text-base font-bold text-[var(--ink)] m-0">Lançamento Financeiro</h3>
            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Tipo</label>
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => setKind("custo")}
                  className={`py-2 text-xs font-bold rounded-xl border cursor-pointer ${
                    kind === "custo"
                      ? "bg-[var(--bad)] text-white border-[var(--bad)]"
                      : "bg-transparent text-[var(--ink)] border-[var(--line)]"
                  }`}
                >
                  Custo
                </button>
                <button
                  type="button"
                  onClick={() => setKind("receita")}
                  className={`py-2 text-xs font-bold rounded-xl border cursor-pointer ${
                    kind === "receita"
                      ? "bg-[var(--ok)] text-white border-[var(--ok)]"
                      : "bg-transparent text-[var(--ink)] border-[var(--line)]"
                  }`}
                >
                  Receita
                </button>
              </div>
            </div>

            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Valor em R$</label>
              <input
                type="text"
                required
                placeholder="Ex.: 45,00"
                value={amountReais}
                onChange={(e) => setAmountReais(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              />
            </div>

            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Serviço / Fornecedor</label>
              <input
                type="text"
                required
                placeholder="Ex.: Render, Vercel, Supabase..."
                value={service}
                onChange={(e) => setService(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              />
            </div>

            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Descrição</label>
              <input
                type="text"
                placeholder="Ex.: Mensalidade servidor"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setShowModal(false)}
                className="px-3 py-2 text-xs font-semibold rounded-xl border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer"
              >
                Cancelar
              </button>
              <button
                type="submit"
                disabled={isProcessing}
                className="px-3 py-2 text-xs font-semibold rounded-xl bg-[var(--acc)] text-white border-0 cursor-pointer"
              >
                Salvar Lançamento
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
