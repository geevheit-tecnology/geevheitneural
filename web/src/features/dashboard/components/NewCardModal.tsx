"use client";

import React, { useState } from "react";
import { CardItem, AppItem } from "@/shared/api-client/endpoints";

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

  if (!isOpen) return null;

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
