"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/shared/api-client/endpoints";

const details = {
  gemini: { purpose: "Gerar páginas, sites e código com contexto do projeto.", url: "https://aistudio.google.com/", limits: "Cota gratuita varia por modelo, região e conta.", key: "GEMINI_API_KEY" },
  groq: { purpose: "Gerar e revisar código usando os modelos disponíveis na Groq.", url: "https://console.groq.com/", limits: "Plano Free limitado por requisições e tokens; confira a cota da sua conta.", key: "GROQ_API_KEY" },
};

export function FactoryToolsPanel({ onShowToast, isDemoMode = false }: { onShowToast: (message: string) => void; isDemoMode?: boolean }) {
  const client = useQueryClient();
  const [confirmed, setConfirmed] = useState(false);
  const [brief, setBrief] = useState("");
  const query = useQuery({ queryKey: ["factory-tools"], queryFn: api.listFactoryTools, retry: false, enabled: !isDemoMode });
  const selection = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) => {
      if (isDemoMode) throw new Error("Entre na sua conta para ativar ferramentas reais.");
      return api.selectFactoryTool(id, active, confirmed);
    },
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ["factory-tools"] });
      await client.invalidateQueries({ queryKey: ["connectors"] });
      onShowToast("Preferência salva. A fábrica verifica a ferramenta antes de cada etapa.");
    },
  });

  async function copyBrief() {
    try {
      await navigator.clipboard.writeText(`Crie um protótipo responsivo para este projeto:\n${brief.trim()}\n\nPriorize celular a partir de 360 px, hierarquia clara, navegação e estados dos botões. Não invente depoimentos, preços ou dados de contato. Entregue uma proposta visual para revisão do dono antes da implementação.`);
      onShowToast("Briefing copiado. Cole no Figma Make e revise o protótipo.");
    } catch { onShowToast("Não foi possível copiar. Selecione e copie o briefing manualmente."); }
  }

  return (
    <section id="factory-tools-panel" className="space-y-4 min-w-0">
      <div className="rounded-2xl border border-[var(--line)] bg-[var(--surf)] p-4 space-y-3">
        <h3 className="m-0 text-base font-bold text-[var(--ink)]">Ferramentas de criação</h3>
        <p className="text-sm text-[var(--mute)] m-0">Escolha quem gera o código da fábrica. Uma API por vez; o Figma fica disponível como apoio ao protótipo.</p>
        <p className="text-xs text-[var(--mute)] m-0">Uma chave configurada não confirma saldo ou disponibilidade. Confira o plano gratuito no provedor. Se a cota acabar, a fábrica informa a falha e aguarda sua decisão.</p>
        <label className="flex items-start gap-2 text-sm text-[var(--ink)]">
          <input id="factory-free-tier-confirmation" type="checkbox" disabled={isDemoMode} checked={confirmed} onChange={e => setConfirmed(e.target.checked)} className="mt-1" />
          Confirmei que a conta e o modelo que vou ativar usam a cota gratuita, sem cobrança habilitada.
        </label>
      </div>

      {isDemoMode && <p role="status" className="text-sm text-[var(--mute)]">Demonstração: entre na sua conta para consultar e ativar as APIs. O briefing do Figma pode ser preparado aqui.</p>}
      {!isDemoMode && query.isPending && <p role="status" className="text-sm text-[var(--mute)]">Consultando ferramentas no servidor…</p>}
      {!isDemoMode && query.isError && <div role="alert" className="rounded-xl border border-red-500/30 p-3 text-sm text-[var(--ink)]">
        Não foi possível consultar a configuração. {query.error.message}
        <button id="factory-tools-retry" onClick={() => query.refetch()} className="block mt-2 underline">Tentar novamente</button>
      </div>}
      {selection.isError && <p role="alert" className="text-sm text-red-500">{selection.error.message}</p>}

      <div className="grid gap-3 sm:grid-cols-2">
        {(isDemoMode ? [{ id: "gemini" as const, name: "Google Gemini", configured: false, active: false }, { id: "groq" as const, name: "Groq", configured: false, active: false }] : query.data ?? []).map(tool => {
          const info = details[tool.id];
          return <article key={tool.id} className="min-w-0 rounded-2xl border border-[var(--line)] bg-[var(--surf)] p-4 space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h4 className="m-0 font-bold text-[var(--ink)]">{tool.name}</h4>
              <span className="text-xs text-[var(--mute)]">{tool.active ? "Selecionado para geração" : "Desligado"}</span>
            </div>
            <p className="text-sm text-[var(--ink)] m-0">{info.purpose}</p>
            <p className="text-xs text-[var(--mute)] m-0">{info.limits}</p>
            <p className="text-xs text-[var(--mute)] break-words m-0">{isDemoMode ? "Configuração real disponível após login." : tool.configured ? "Chave presente no servidor. Conexão ainda não testada por esta tela." : `Para habilitar, configure ${info.key} no servidor.`}</p>
            <div className="flex flex-wrap items-center gap-3">
              <button id={`factory-tool-${tool.id}`} disabled={selection.isPending || (!tool.active && (!tool.configured || !confirmed))} onClick={() => selection.mutate({ id: tool.id, active: !tool.active })} className="rounded-xl bg-[var(--acc)] px-4 py-2 text-sm font-semibold text-white disabled:opacity-40 disabled:cursor-not-allowed">
                {selection.isPending && selection.variables?.id === tool.id ? "Salvando…" : tool.active ? "Desligar geração" : "Usar na fábrica"}
              </button>
              <a href={info.url} target="_blank" rel="noopener noreferrer" className="text-sm text-[var(--acc)] underline">Abrir provedor ↗</a>
            </div>
          </article>;
        })}
      </div>

      <article className="rounded-2xl border border-[var(--line)] bg-[var(--surf)] p-4 space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h4 className="m-0 font-bold text-[var(--ink)]">Figma Make · protótipo</h4>
          <span className="text-xs text-[var(--mute)]">Uso manual</span>
        </div>
        <p className="text-sm text-[var(--mute)] m-0">Prepare a ideia, copie o briefing e crie o protótipo no Figma. O plano Starter tem créditos limitados. Abrir o Figma não conecta a conta nem inicia uma geração na fábrica.</p>
        <label htmlFor="figma-creation-brief" className="block text-sm font-semibold text-[var(--ink)]">Descreva o projeto e o estilo desejado</label>
        <textarea id="figma-creation-brief" value={brief} onChange={e => setBrief(e.target.value)} rows={4} maxLength={12000} placeholder="Ex.: landing page de uma agência de veículos, com catálogo, filtros e contato…" className="w-full rounded-xl border border-[var(--line)] bg-[var(--bg)] p-3 text-sm text-[var(--ink)]" />
        <div className="flex flex-wrap gap-3 items-center">
          <button id="figma-copy-brief" disabled={!brief.trim()} onClick={copyBrief} className="rounded-xl bg-[var(--acc)] px-4 py-2 text-sm font-semibold text-white disabled:opacity-40">Copiar briefing</button>
          <a id="figma-open-make" href="https://www.figma.com/make/" target="_blank" rel="noopener noreferrer" className="text-sm text-[var(--acc)] underline">Abrir Figma Make ↗</a>
          <a href="https://www.figma.com/pricing/" target="_blank" rel="noopener noreferrer" className="text-xs text-[var(--mute)] underline">Consultar créditos</a>
        </div>
      </article>
    </section>
  );
}
