"use client";

import React, { useState } from "react";
import { AppItem, CreateAppInput } from "@/shared/api-client/endpoints";

interface MeusAppsTabProps {
  apps: AppItem[];
  onCreateApp: (data: CreateAppInput) => void;
  onDeleteApp: (id: string) => void;
  isProcessing: boolean;
}

export function MeusAppsTab({ apps, onCreateApp, onDeleteApp, isProcessing }: MeusAppsTabProps) {
  const [showModal, setShowModal] = useState(false);
  const [name, setName] = useState("");
  const [hosting, setHosting] = useState("Render");
  const [database, setDatabase] = useState("Supabase");

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    onCreateApp({
      name: name.trim(),
      hosting,
      database,
    });
    setName("");
    setShowModal(false);
  };

  return (
    <div id="tab-apps" className="space-y-3">
      <div className="bg-[var(--accs)] text-[var(--ink)] rounded-xl p-3 text-xs sm:text-sm font-medium flex items-center justify-between gap-2">
        <span>Todos os seus apps num só lugar: onde rodam, qual banco usam e se pagam a própria conta.</span>
        <button
          id="open-new-app-modal-btn"
          onClick={() => setShowModal(true)}
          className="bg-[var(--acc)] text-white font-semibold text-xs py-1.5 px-3 rounded-lg border-0 cursor-pointer flex-none hover:opacity-90"
        >
          + Novo App
        </button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {apps.length === 0 ? (
          <div className="col-span-full text-center py-12 bg-[var(--surf)] border border-dashed border-[var(--line)] rounded-2xl p-6">
            <span className="text-3xl mb-2 block">📱</span>
            <h4 className="text-sm font-bold text-[var(--ink)] m-0 mb-1">Nenhum aplicativo cadastrado</h4>
            <p className="text-xs text-[var(--mute)] max-w-sm mx-auto mb-4">
              Cadastre seus apps para monitorar onde rodam, custos e margem líquida.
            </p>
            <button
              onClick={() => setShowModal(true)}
              className="bg-[var(--acc)] text-white text-xs font-bold py-2 px-4 rounded-xl border-0 cursor-pointer hover:opacity-90"
            >
              + Cadastrar Primeiro App
            </button>
          </div>
        ) : (
          apps.map((app) => {
          const margin = app.monthly_revenue_cents - app.monthly_cost_cents;
          const isNegative = margin < 0;

          return (
            <div
              key={app.id}
              className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-4 shadow-sm"
            >
              <div className="flex items-center justify-between mb-1">
                <h3 className="text-base font-bold text-[var(--ink)] m-0">{app.name}</h3>
                <button
                  onClick={() => {
                    if (confirm(`Excluir app ${app.name}?`)) onDeleteApp(app.id);
                  }}
                  className="text-[var(--mute)] hover:text-[var(--bad)] text-xs bg-transparent border-0 cursor-pointer p-1"
                >
                  Excluir
                </button>
              </div>
              <div className="text-xs text-[var(--mute)] mb-3">{app.client_name}</div>

              <div className="space-y-1.5 text-xs bg-[var(--bg)]/50 border border-[var(--line)] rounded-xl p-2.5 mb-3">
                <div className="flex justify-between">
                  <span className="text-[var(--mute)]">Hospedagem</span>
                  <b className="text-[var(--ink)]">{app.hosting || "—"}</b>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--mute)]">Banco</span>
                  <b className="text-[var(--ink)]">{app.database || "—"}</b>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--mute)]">Custo por mês</span>
                  <b className="text-[var(--ink)]">{formatBRL(app.monthly_cost_cents)}</b>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--mute)]">Receita por mês</span>
                  <b className="text-[var(--ink)]">{formatBRL(app.monthly_revenue_cents)}</b>
                </div>
                <div className="flex justify-between border-t border-[var(--line)] pt-1.5 font-medium">
                  <span className="text-[var(--mute)]">Resultado</span>
                  <b className={margin >= 0 ? "text-[var(--ok)]" : "text-[var(--bad)]"}>
                    {formatBRL(margin)}
                  </b>
                </div>
              </div>

              {isNegative && (
                <div className="bg-[var(--warns)] text-[var(--warn)] rounded-xl p-2.5 text-xs font-semibold leading-relaxed">
                  ⚠️ Custa mais do que rende. Candidato a migrar ou encerrar.
                </div>
              )}
            </div>
          );
        })
      )}
      </div>

      {/* Modal de Cadastro de App */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4">
          <form
            onSubmit={handleSubmit}
            className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 max-w-sm w-full shadow-2xl space-y-3"
          >
            <h3 className="text-base font-bold text-[var(--ink)] m-0">Cadastrar Novo App</h3>
            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Nome do app</label>
              <input
                id="new-app-name-input"
                type="text"
                required
                placeholder="Ex.: delivery-bot"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              />
            </div>
            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Hospedagem</label>
              <select
                value={hosting}
                onChange={(e) => setHosting(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              >
                <option value="Render">Render</option>
                <option value="Vercel">Vercel</option>
                <option value="Netlify">Netlify</option>
                <option value="Docker (VPS)">Docker (VPS)</option>
                <option value="Firebase">Firebase</option>
              </select>
            </div>
            <div>
              <label className="text-xs text-[var(--mute)] block mb-1">Banco de Dados</label>
              <select
                value={database}
                onChange={(e) => setDatabase(e.target.value)}
                className="w-full text-xs p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
              >
                <option value="Supabase">Supabase</option>
                <option value="Neon">Neon</option>
                <option value="Postgres local">Postgres local</option>
                <option value="Firebase">Firebase</option>
                <option value="—">Sem banco</option>
              </select>
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
                Salvar App
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
