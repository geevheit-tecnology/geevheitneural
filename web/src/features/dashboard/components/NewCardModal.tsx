"use client";

import React, { useState } from "react";
import { api, CardItem, AppItem } from "@/shared/api-client/endpoints";

interface NewCardModalProps {
  isOpen: boolean;
  onClose: () => void;
  apps: AppItem[];
  onCreateCard: (card: Omit<CardItem, "id" | "created_at" | "updated_at">) => void;
  isProcessing?: boolean;
}

export function NewCardModal({
  isOpen,
  onClose,
  apps,
  onCreateCard,
  isProcessing = false,
}: NewCardModalProps) {
  const [title, setTitle] = useState("");
  const [clientName, setClientName] = useState("");
  const [type, setType] = useState<CardItem["type"]>("orcamento");
  const [priceReais, setPriceReais] = useState("");
  const [costReais, setCostReais] = useState("");
  const [appName, setAppName] = useState("Novo");
  const [stage, setStage] = useState("entrada");
  const [productType, setProductType] = useState("saas_b2b");
  const [complexity, setComplexity] = useState("medium");
  const [estimating, setEstimating] = useState(false);
  const [estimateFeedback, setEstimateFeedback] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSuggestEstimate = async () => {
    setEstimating(true);
    setEstimateFeedback(null);
    try {
      const res = await api.estimatePrice({
        title: title || "Novo Projeto",
        description: title,
        product_type: productType,
        complexity,
      });
      setPriceReais((res.suggested_price_cents / 100).toFixed(2).replace(".", ","));
      setCostReais((res.est_cost_cents / 100).toFixed(2).replace(".", ","));
      setEstimateFeedback(
        `Playbook: ${res.playbook} · Prazo sugerido: ${res.estimated_days} dias · Stack: ${res.recommended_stack.join(", ")}`
      );
    } catch {
      const baseMap: Record<string, [number, number, number, string]> = {
        saas_b2b: [8000, 1200, 14, "Go 1.22 + PostgreSQL RLS + Next.js"],
        mobile_app: [5500, 800, 10, "Next.js PWA 360px + Go REST API"],
        landing_page: [1500, 200, 3, "Next.js Static + TailwindCSS"],
      };
      const [basePrice, baseCost, baseDays, stack] = baseMap[productType] || baseMap.saas_b2b;
      const mult = complexity === "high" ? 1.5 : complexity === "low" ? 0.8 : 1.0;
      setPriceReais(Math.round(basePrice * mult).toFixed(2).replace(".", ","));
      setCostReais(Math.round(baseCost * mult).toFixed(2).replace(".", ","));
      setEstimateFeedback(
        `Playbook: ${productType} · Prazo sugerido: ${Math.max(2, Math.round(baseDays * mult))} dias · Stack: ${stack}`
      );
    } finally {
      setEstimating(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !clientName.trim()) return;

    const priceNum = parseFloat(priceReais.replace(",", ".")) || 0;
    const costNum = parseFloat(costReais.replace(",", ".")) || 0;

    onCreateCard({
      title: title.trim(),
      client_name: clientName.trim(),
      client_id: `c_${Date.now()}`,
      type,
      price_cents: Math.round(priceNum * 100),
      est_cost_cents: Math.round(costNum * 100),
      currency: "BRL",
      stage,
      priority: 1,
      app_name: appName,
      due_date: new Date(Date.now() + 7 * 86400000).toISOString(),
    });

    setTitle("");
    setClientName("");
    setPriceReais("");
    setCostReais("");
    setAppName("Novo");
    setStage("entrada");
    onClose();
  };

  return (
    <div
      id="modal-new-card"
      className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4 backdrop-blur-sm"
    >
      <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-md w-full shadow-2xl">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <span className="text-xl">✨</span>
            <h3 className="text-lg font-bold text-[var(--ink)] m-0">Nova Demanda / Cartão</h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Fechar modal"
            className="text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 cursor-pointer text-lg leading-none p-1"
          >
            ✕
          </button>
        </div>

        <p className="text-xs text-[var(--mute)] mb-4 leading-relaxed">
          Cadastre uma nova solicitação de projeto, orçamento ou suporte para iniciar o fluxo.
        </p>

        <form onSubmit={handleSubmit} className="space-y-3.5">
          <div>
            <label
              htmlFor="input-card-title"
              className="block text-xs font-semibold text-[var(--mute)] mb-1"
            >
              Título da Demanda / Projeto *
            </label>
            <input
              id="input-card-title"
              type="text"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Ex: Landing Page de Vendas, Sistema de Estoque"
              className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)] transition-colors"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="input-card-client"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Nome do Cliente *
              </label>
              <input
                id="input-card-client"
                type="text"
                required
                value={clientName}
                onChange={(e) => setClientName(e.target.value)}
                placeholder="Ex: Dr. Roberto, Loja Alfa"
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              />
            </div>

            <div>
              <label
                htmlFor="select-card-type"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Tipo da Demanda
              </label>
              <select
                id="select-card-type"
                value={type}
                onChange={(e) => setType(e.target.value as CardItem["type"])}
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              >
                <option value="orcamento">Orçamento (Novo)</option>
                <option value="bug">Bug / Correção</option>
                <option value="suporte">Suporte / Ajuste</option>
                <option value="financeiro">Financeiro</option>
              </select>
            </div>
          </div>

          <div className="p-2.5 rounded-xl bg-[var(--bg)]/60 border border-[var(--line)] space-y-2">
            <div className="flex items-center justify-between gap-2">
              <select
                id="select-estimate-product-type"
                value={productType}
                onChange={(e) => setProductType(e.target.value)}
                className="flex-1 px-2 py-1.5 text-xs rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)]"
              >
                <option value="saas_b2b">Playbook: SaaS B2B Multi-tenant</option>
                <option value="mobile_app">Playbook: App Mobile / PWA</option>
                <option value="landing_page">Playbook: Landing Page Conversão</option>
              </select>
              <select
                id="select-estimate-complexity"
                value={complexity}
                onChange={(e) => setComplexity(e.target.value)}
                className="px-2 py-1.5 text-xs rounded-lg border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)]"
              >
                <option value="low">Complexidade Baixa</option>
                <option value="medium">Complexidade Média</option>
                <option value="high">Complexidade Alta</option>
              </select>
            </div>
            <button
              type="button"
              id="btn-suggest-price-estimate"
              disabled={estimating}
              onClick={handleSuggestEstimate}
              className="w-full py-2 px-3 rounded-lg font-bold text-xs border border-[var(--acc)]/40 bg-[var(--accs)] text-[var(--acc)] cursor-pointer hover:opacity-90 flex items-center justify-center gap-1.5"
            >
              <span>💡</span>
              <span>
                {estimating
                  ? "Calculando estimativa..."
                  : "Sugerir Valor, Prazo e Margem (Motor de Preço IA - Fase 10)"}
              </span>
            </button>
            {estimateFeedback && (
              <div className="text-[11px] text-[var(--ok)] font-medium leading-snug">
                ✓ {estimateFeedback}
              </div>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="input-card-price"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Valor Cobrado (R$)
              </label>
              <input
                id="input-card-price"
                type="text"
                value={priceReais}
                onChange={(e) => setPriceReais(e.target.value)}
                placeholder="Ex: 3500,00"
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              />
            </div>

            <div>
              <label
                htmlFor="input-card-cost"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Custo Estimado (R$)
              </label>
              <input
                id="input-card-cost"
                type="text"
                value={costReais}
                onChange={(e) => setCostReais(e.target.value)}
                placeholder="Ex: 250,00"
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="select-card-app"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Vincular a App
              </label>
              <select
                id="select-card-app"
                value={appName}
                onChange={(e) => setAppName(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              >
                <option value="Novo">Novo Projeto</option>
                {apps.map((a) => (
                  <option key={a.id} value={a.name}>
                    {a.name} ({a.client_name})
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label
                htmlFor="select-card-stage"
                className="block text-xs font-semibold text-[var(--mute)] mb-1"
              >
                Etapa Inicial
              </label>
              <select
                id="select-card-stage"
                value={stage}
                onChange={(e) => setStage(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] focus:outline-none focus:border-[var(--acc)]"
              >
                <option value="entrada">Entrada</option>
                <option value="proposta_aprovar">Proposta para aprovar 🛡️</option>
                <option value="aguardando_cliente">Aguardando cliente</option>
                <option value="prompt_aprovar">Prompt para aprovar 🛡️</option>
                <option value="em_execucao">Em execução</option>
                <option value="entrega_aprovar">Entrega para aprovar 🛡️</option>
                <option value="entregue">Entregue</option>
              </select>
            </div>
          </div>

          <div className="flex justify-end gap-2.5 pt-3 border-t border-[var(--line)]">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-semibold rounded-xl border border-[var(--line)] bg-transparent text-[var(--ink)] cursor-pointer hover:bg-[var(--line)]/30 transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={isProcessing}
              className="px-4 py-2 text-xs font-semibold rounded-xl bg-[var(--acc)] text-white border-0 cursor-pointer hover:opacity-90 transition-opacity disabled:opacity-50"
            >
              {isProcessing ? "Salvando..." : "Criar Demanda"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
