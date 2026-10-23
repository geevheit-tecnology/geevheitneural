"use client";

import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  FinancesOverview,
  AppItem,
  CreateFinanceEntryInput,
  api,
  SubscriptionItem,
  CreateSubscriptionInput,
} from "@/shared/api-client/endpoints";

interface FinancasTabProps {
  finances: FinancesOverview | null;
  apps: AppItem[];
  onCreateEntry: (data: CreateFinanceEntryInput) => void;
  isProcessing: boolean;
}

export function FinancasTab({ finances, apps, onCreateEntry, isProcessing }: FinancasTabProps) {
  const queryClient = useQueryClient();
  const [subTab, setSubTab] = useState<"lancamentos" | "recorrencia">("recorrencia");

  // Estados do modal de lançamento manual
  const [showModal, setShowModal] = useState(false);
  const [kind, setKind] = useState<"receita" | "custo">("custo");
  const [amountReais, setAmountReais] = useState("");
  const [service, setService] = useState("Render");
  const [description, setDescription] = useState("");

  // Estados do modal de nova assinatura
  const [showSubModal, setShowSubModal] = useState(false);
  const [subPlan, setSubPlan] = useState("");
  const [subAmountReais, setSubAmountReais] = useState("");
  const [subInterval, setSubInterval] = useState<"mensal" | "trimestral" | "semestral" | "anual">("mensal");
  const [subDueDay, setSubDueDay] = useState("10");
  const [subAppId, setSubAppId] = useState("");
  const [subDesc, setSubDesc] = useState("");

  // Queries para MRR e Assinaturas (Fase 14)
  const { data: mrrMetrics, isLoading: loadingMRR } = useQuery({
    queryKey: ["mrr-metrics"],
    queryFn: () => api.getMRRMetrics(),
    enabled: subTab === "recorrencia",
  });

  const { data: subscriptions = [], isLoading: loadingSubs } = useQuery({
    queryKey: ["subscriptions"],
    queryFn: () => api.getSubscriptions(),
    enabled: subTab === "recorrencia",
  });

  const { data: upcomingCharges = [] } = useQuery({
    queryKey: ["subscription-charges"],
    queryFn: () => api.getSubscriptionCharges(),
    enabled: subTab === "recorrencia",
  });

  // Mutations para Assinaturas
  const createSubMutation = useMutation({
    mutationFn: (data: CreateSubscriptionInput) => api.createSubscription(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      queryClient.invalidateQueries({ queryKey: ["mrr-metrics"] });
      queryClient.invalidateQueries({ queryKey: ["subscription-charges"] });
      setShowSubModal(false);
      setSubPlan("");
      setSubAmountReais("");
      setSubDesc("");
    },
  });

  const updateSubStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      api.updateSubscriptionStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      queryClient.invalidateQueries({ queryKey: ["mrr-metrics"] });
    },
  });

  const payChargeMutation = useMutation({
    mutationFn: (id: string) => api.paySubscriptionCharge(id, "pix"),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscription-charges"] });
      queryClient.invalidateQueries({ queryKey: ["mrr-metrics"] });
      queryClient.invalidateQueries({ queryKey: ["finances"] });
    },
  });

  const runDailyMutation = useMutation({
    mutationFn: () => api.processDailySubscriptions(),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      queryClient.invalidateQueries({ queryKey: ["subscription-charges"] });
      queryClient.invalidateQueries({ queryKey: ["mrr-metrics"] });
      alert(`Rotina diária executada com sucesso!\nCobranças vencidas: ${res.overdue_count}\nNovas cobranças geradas: ${res.generated_count}`);
    },
  });

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

  const handleCreateSubscription = (e: React.FormEvent) => {
    e.preventDefault();
    const val = parseFloat(subAmountReais.replace(",", "."));
    if (isNaN(val) || val <= 0) return;

    const selectedApp = apps.find((a) => a.id === subAppId);
    const clientId = selectedApp?.client_id || "a0000000-0000-0000-0000-000000000001";

    createSubMutation.mutate({
      client_id: clientId,
      app_id: subAppId || undefined,
      plan: subPlan,
      description: subDesc,
      amount_cents: Math.round(val * 100),
      currency: "BRL",
      interval: subInterval,
      due_day: parseInt(subDueDay, 10) || 10,
    });
  };

  return (
    <div id="tab-financas" className="space-y-4">
      {/* Seletor de Sub-Abas: MRR e Recorrência x Lançamentos */}
      <div className="flex bg-[var(--surf)] p-1 rounded-2xl border border-[var(--line)]">
        <button
          type="button"
          onClick={() => setSubTab("recorrencia")}
          className={`flex-1 py-2 text-xs font-bold rounded-xl transition-all cursor-pointer ${
            subTab === "recorrencia"
              ? "bg-[var(--acc)] text-white shadow"
              : "text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          Mensalidades & MRR
        </button>
        <button
          type="button"
          onClick={() => setSubTab("lancamentos")}
          className={`flex-1 py-2 text-xs font-bold rounded-xl transition-all cursor-pointer ${
            subTab === "lancamentos"
              ? "bg-[var(--acc)] text-white shadow"
              : "text-[var(--mute)] hover:text-[var(--ink)]"
          }`}
        >
          Visão Geral & Lançamentos
        </button>
      </div>

      {subTab === "recorrencia" ? (
        <div className="space-y-4">
          {/* Métricas de Negócio MRR */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3">
              <small className="text-xs text-[var(--mute)] block">MRR Atual</small>
              <div className="text-lg font-bold text-[var(--ok)] mt-0.5">
                {loadingMRR ? "..." : formatBRL(mrrMetrics?.current_mrr_cents || 0)}
              </div>
              <span className="text-[10px] text-[var(--mute)]">Faturamento recorrente</span>
            </div>

            <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3">
              <small className="text-xs text-[var(--mute)] block">MRR Previsto</small>
              <div className="text-lg font-bold text-[var(--ink)] mt-0.5">
                {loadingMRR ? "..." : formatBRL(mrrMetrics?.projected_mrr_cents || 0)}
              </div>
              <span className="text-[10px] text-[var(--mute)]">Potencial com pausas</span>
            </div>

            <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-3">
              <small className="text-xs text-[var(--mute)] block">Assinaturas Ativas</small>
              <div className="text-lg font-bold text-[var(--ink)] mt-0.5">
                {loadingMRR ? "..." : mrrMetrics?.active_count || 0}
              </div>
              <span className="text-[10px] text-[var(--mute)]">
                {mrrMetrics?.paused_count || 0} pausadas · {mrrMetrics?.canceled_count || 0} churn
              </span>
            </div>

            <div className={`border rounded-2xl p-3 ${
              (mrrMetrics?.delinquent_count || 0) > 0
                ? "bg-red-500/10 border-red-500/30"
                : "bg-[var(--surf)] border-[var(--line)]"
            }`}>
              <small className="text-xs text-[var(--mute)] block">Inadimplência</small>
              <div className={`text-lg font-bold mt-0.5 ${
                (mrrMetrics?.delinquent_count || 0) > 0 ? "text-[var(--bad)]" : "text-[var(--ink)]"
              }`}>
                {loadingMRR ? "..." : formatBRL(mrrMetrics?.delinquent_amount_cents || 0)}
              </div>
              <span className="text-[10px] text-[var(--mute)]">
                {mrrMetrics?.delinquent_count || 0} faturas atrasadas
              </span>
            </div>
          </div>

          {/* Barra de Ações de Recorrência */}
          <div className="flex items-center justify-between gap-2">
            <button
              onClick={() => runDailyMutation.mutate()}
              disabled={runDailyMutation.isPending}
              className="text-xs font-semibold py-1.5 px-3 rounded-xl border border-[var(--line)] bg-[var(--surf)] text-[var(--mute)] hover:text-[var(--ink)] cursor-pointer"
            >
              {runDailyMutation.isPending ? "Processando..." : "↻ Rodar Verificação Diária"}
            </button>
            <button
              id="open-sub-modal-btn"
              onClick={() => setShowSubModal(true)}
              className="text-xs font-semibold py-1.5 px-3 rounded-xl bg-[var(--acc)] text-white hover:opacity-90 cursor-pointer"
            >
              + Nova Assinatura
            </button>
          </div>

          {/* Painel de Cobranças Atrasadas e Próximas */}
          {upcomingCharges.length > 0 && (
            <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 space-y-3">
              <h3 className="text-xs font-bold text-[var(--ink)] uppercase tracking-wider">
                Faturas Pendentes & A Vencer
              </h3>
              <div className="space-y-2">
                {upcomingCharges.map((ch) => {
                  const isOverdue = ch.status === "atrasada";
                  return (
                    <div
                      key={ch.id}
                      className={`flex items-center justify-between p-3 rounded-xl border text-xs ${
                        isOverdue
                          ? "bg-red-500/5 border-red-500/30"
                          : "bg-[var(--bg)] border-[var(--line)]"
                      }`}
                    >
                      <div>
                        <div className="font-bold text-[var(--ink)]">
                          {formatBRL(ch.amount_cents)} via PIX
                        </div>
                        <div className="text-[11px] text-[var(--mute)]">
                          Vencimento: {new Date(ch.due_date).toLocaleDateString("pt-BR")} ·{" "}
                          <span className={isOverdue ? "text-red-500 font-semibold" : "text-amber-500"}>
                            {isOverdue ? "Atrasada / Vencida" : "Pendente"}
                          </span>
                        </div>
                      </div>
                      <button
                        onClick={() => payChargeMutation.mutate(ch.id)}
                        disabled={payChargeMutation.isPending}
                        className="py-1 px-2.5 rounded-lg text-xs font-semibold bg-[var(--ok)] text-white hover:opacity-90 cursor-pointer"
                      >
                        {payChargeMutation.isPending ? "Gravando..." : "Marcar Pago"}
                      </button>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Lista de Assinaturas Cadastradas */}
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
            <h3 className="text-sm font-bold text-[var(--ink)] mb-3">Assinaturas Recorrentes</h3>
            {loadingSubs ? (
              <p className="text-xs text-[var(--mute)]">Carregando assinaturas...</p>
            ) : subscriptions.length === 0 ? (
              <div className="text-center py-6 text-xs text-[var(--mute)]">
                Nenhuma assinatura cadastrada ainda. Clique em &quot;+ Nova Assinatura&quot; para começar.
              </div>
            ) : (
              <div className="space-y-3">
                {subscriptions.map((sub: SubscriptionItem) => {
                  const statusColors: Record<string, string> = {
                    ativa: "bg-emerald-500/10 text-emerald-500 border-emerald-500/30",
                    pausada: "bg-amber-500/10 text-amber-500 border-amber-500/30",
                    inadimplente: "bg-red-500/10 text-red-500 border-red-500/30",
                    cancelada: "bg-gray-500/10 text-gray-400 border-gray-500/30",
                  };

                  return (
                    <div
                      key={sub.id}
                      className="p-3 rounded-xl border border-[var(--line)] bg-[var(--bg)] flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <b className="text-[var(--ink)] text-sm">{sub.plan}</b>
                          <span
                            className={`px-2 py-0.5 rounded-md border text-[10px] uppercase font-bold ${
                              statusColors[sub.status] || "bg-[var(--surf)]"
                            }`}
                          >
                            {sub.status}
                          </span>
                        </div>
                        <div className="text-[11px] text-[var(--mute)] mt-1">
                          {formatBRL(sub.amount_cents)} / {sub.interval} (Dia {sub.due_day}) · Próximo:{" "}
                          {new Date(sub.next_due_date).toLocaleDateString("pt-BR")}
                        </div>
                        {sub.description && (
                          <div className="text-[11px] text-[var(--mute)] mt-0.5">{sub.description}</div>
                        )}
                      </div>

                      <div className="flex items-center gap-2">
                        {sub.status === "ativa" && (
                          <button
                            onClick={() =>
                              updateSubStatusMutation.mutate({ id: sub.id, status: "pausada" })
                            }
                            className="py-1 px-2.5 rounded-lg border border-[var(--line)] text-[var(--ink)] hover:bg-[var(--surf)] cursor-pointer"
                          >
                            Pausar
                          </button>
                        )}
                        {sub.status === "pausada" && (
                          <button
                            onClick={() =>
                              updateSubStatusMutation.mutate({ id: sub.id, status: "ativa" })
                            }
                            className="py-1 px-2.5 rounded-lg bg-[var(--acc)] text-white hover:opacity-90 cursor-pointer"
                          >
                            Reativar
                          </button>
                        )}
                        {sub.status !== "cancelada" && (
                          <button
                            onClick={() => {
                              if (confirm(`Deseja cancelar a assinatura ${sub.plan}?`)) {
                                updateSubStatusMutation.mutate({ id: sub.id, status: "cancelada" });
                              }
                            }}
                            className="py-1 px-2 rounded-lg border border-red-500/30 text-red-500 hover:bg-red-500/10 cursor-pointer"
                          >
                            Cancelar
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      ) : (
        /* Aba Tradicional de Visão Geral & Lançamentos */
        <div className="space-y-3">
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
                      {formatBRL(app.monthly_revenue_cents)} / {formatBRL(app.monthly_cost_cents)}
                    </span>
                  </div>
                  <div className="h-2 w-full bg-[var(--line)] rounded-full overflow-hidden flex">
                    <div
                      className="bg-[var(--ok)] h-full"
                      style={{
                        width: `${(app.monthly_revenue_cents / maxVal) * 100}%`,
                      }}
                    />
                    <div
                      className="bg-[var(--bad)] h-full"
                      style={{
                        width: `${(app.monthly_cost_cents / maxVal) * 100}%`,
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Visão de Custos por Serviço */}
          <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm">
            <h3 className="text-sm font-bold text-[var(--ink)] m-0 mb-3">Custos por serviço</h3>
            <div className="space-y-2">
              {finances?.costs_by_service?.map((s) => (
                <div
                  key={s.service_name}
                  className="flex justify-between items-center text-xs p-2 rounded-xl border border-[var(--line)] bg-[var(--bg)]"
                >
                  <span className="text-[var(--ink)] font-semibold">{s.service_name}</span>
                  <span className="text-[var(--mute)]">{formatBRL(s.cost_cents)}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Sugestão de Otimização */}
          {finances?.suggestion && (
            <div className="bg-amber-500/10 border border-amber-500/20 rounded-2xl p-3 text-xs text-amber-500">
              💡 <b>Sugestão da Fábrica:</b> {finances.suggestion}
            </div>
          )}
        </div>
      )}

      {/* Modal: Nova Assinatura Recorrente (Fase 14) */}
      {showSubModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4">
          <form
            onSubmit={handleCreateSubscription}
            className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-sm w-full shadow-2xl space-y-3"
          >
            <h3 className="text-base font-bold text-[var(--ink)] m-0">Nova Assinatura Recorrente</h3>

            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Nome do Plano / Serviço</label>
              <input
                type="text"
                required
                placeholder="Ex.: Manutenção Pro, Hospedagem VIP..."
                value={subPlan}
                onChange={(e) => setSubPlan(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              />
            </div>

            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">App / Projeto Vinculado</label>
              <select
                value={subAppId}
                onChange={(e) => setSubAppId(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              >
                <option value="">Nenhum app específico (Geral)</option>
                {apps.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.name} ({a.client_name})
                  </option>
                ))}
              </select>
            </div>

            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="text-xs text-[var(--mute)] block mb-1">Valor em R$</label>
                <input
                  type="text"
                  required
                  placeholder="Ex.: 350,00"
                  value={subAmountReais}
                  onChange={(e) => setSubAmountReais(e.target.value)}
                  className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
                />
              </div>
              <div>
                <label className="text-xs text-[var(--mute)] block mb-1">Dia Vencimento</label>
                <input
                  type="number"
                  min="1"
                  max="31"
                  required
                  value={subDueDay}
                  onChange={(e) => setSubDueDay(e.target.value)}
                  className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
                />
              </div>
            </div>

            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Periodicidade</label>
              <select
                value={subInterval}
                onChange={(e) =>
                  setSubInterval(e.target.value as "mensal" | "trimestral" | "semestral" | "anual")
                }
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              >
                <option value="mensal">Mensal (100% no MRR)</option>
                <option value="trimestral">Trimestral (dividido por 3 no MRR)</option>
                <option value="semestral">Semestral (dividido por 6 no MRR)</option>
                <option value="anual">Anual (dividido por 12 no MRR)</option>
              </select>
            </div>

            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Descrição</label>
              <input
                type="text"
                placeholder="Ex.: Manutenção corretiva e backups"
                value={subDesc}
                onChange={(e) => setSubDesc(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setShowSubModal(false)}
                className="px-3 py-2 text-xs font-semibold rounded-xl border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer"
              >
                Cancelar
              </button>
              <button
                type="submit"
                disabled={createSubMutation.isPending}
                className="px-3 py-2 text-xs font-semibold rounded-xl bg-[var(--acc)] text-white border-0 cursor-pointer"
              >
                {createSubMutation.isPending ? "Cadastrando..." : "Cadastrar Assinatura"}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Modal: Lançamento Financeiro Manual */}
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
