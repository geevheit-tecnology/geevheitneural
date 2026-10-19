"use client";

import { useState } from "react";

export default function IntakePage() {
  const [clientName, setClientName] = useState("");
  const [contact, setContact] = useState("");
  const [type, setType] = useState("feature");
  const [description, setDescription] = useState("");
  const [website, setWebsite] = useState(""); // Honeypot oculto

  const [submitting, setSubmitting] = useState(false);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8085";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!clientName.trim() || !description.trim()) {
      setError("Por favor, preencha o seu nome e a descrição do projeto.");
      return;
    }

    setSubmitting(true);
    setError(null);
    setStatusMessage("Enviando solicitação...");

    const payload = {
      client_name: clientName.trim(),
      contact: contact.trim(),
      type,
      description: description.trim(),
      website: website.trim(), // Se preenchido por bot, a API descarta silenciosamente
    };

    // Auto-retry com backoff para tolerar cold-start da camada gratuita do Render (até 45s)
    const maxRetries = 3;
    let attempt = 0;
    let success = false;

    while (attempt < maxRetries && !success) {
      attempt++;
      try {
        if (attempt > 1) {
          setStatusMessage(`Acordando servidor de produção seguro (tentativa ${attempt}/${maxRetries})...`);
        }

        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 20000); // 20s timeout por tentativa

        const res = await fetch(`${apiUrl}/v1/intake`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (res.ok) {
          success = true;
          setSubmitted(true);
          break;
        } else if (res.status >= 500 && attempt < maxRetries) {
          // Servidor acordando ou erro transitório, espera 3s antes do retry
          await new Promise((r) => setTimeout(r, 3000));
          continue;
        } else {
          const errText = await res.text();
          throw new Error(errText || "Erro ao enviar solicitação.");
        }
      } catch {
        if (attempt >= maxRetries) {
          setError(
            "O servidor está demorando mais do que o esperado para responder. Por favor, tente novamente em alguns segundos ou fale diretamente pelo WhatsApp."
          );
        } else {
          await new Promise((r) => setTimeout(r, 4000));
        }
      }
    }

    setSubmitting(false);
    setStatusMessage(null);
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-lg bg-slate-900 border border-slate-800 rounded-3xl p-6 sm:p-8 shadow-2xl relative overflow-hidden">
        {/* Glow de fundo */}
        <div className="absolute top-0 right-1/2 translate-x-1/2 w-80 h-32 bg-indigo-500/10 blur-3xl pointer-events-none rounded-full" />

        <div className="text-center mb-6">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-xs font-semibold mb-3">
            <span>🚀 Nova Demanda</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">Conte-nos sobre sua ideia</h1>
          <p className="text-slate-400 text-sm mt-1">
            Receba um orçamento e cronograma personalizado em até 24 horas.
          </p>
        </div>

        {submitted ? (
          <div className="p-6 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-center space-y-4">
            <div className="w-14 h-14 bg-emerald-500/20 text-emerald-400 rounded-full flex items-center justify-center mx-auto text-2xl">
              ✓
            </div>
            <h3 className="text-xl font-bold text-emerald-400">Demanda Recebida com Sucesso!</h3>
            <p className="text-sm text-slate-300 leading-relaxed">
              Obrigado pelo contato! Nossa equipe técnica já recebeu a sua solicitação e entrará em contato com a proposta comercial.
            </p>
            <button
              onClick={() => {
                setSubmitted(false);
                setClientName("");
                setContact("");
                setDescription("");
              }}
              className="px-4 py-2 rounded-xl bg-slate-800 text-slate-200 text-xs font-medium hover:bg-slate-700 transition-colors"
            >
              Enviar outra solicitação
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs text-center">
                {error}
              </div>
            )}

            {statusMessage && (
              <div className="p-3 rounded-xl bg-indigo-500/10 border border-indigo-500/30 text-indigo-300 text-xs text-center flex items-center justify-center gap-2">
                <div className="w-3.5 h-3.5 border-2 border-indigo-400 border-t-transparent rounded-full animate-spin" />
                <span>{statusMessage}</span>
              </div>
            )}

            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1.5">
                Seu Nome ou Empresa *
              </label>
              <input
                id="input-intake-name"
                type="text"
                required
                value={clientName}
                onChange={(e) => setClientName(e.target.value)}
                placeholder="Ex: João da Silva / Padaria Central"
                className="w-full px-4 py-3 rounded-xl bg-slate-800/80 border border-slate-700 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 text-slate-100 text-sm outline-none transition"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1.5">
                Contato (WhatsApp ou E-mail) *
              </label>
              <input
                id="input-intake-contact"
                type="text"
                required
                value={contact}
                onChange={(e) => setContact(e.target.value)}
                placeholder="Ex: (11) 98765-4321 ou joao@empresa.com"
                className="w-full px-4 py-3 rounded-xl bg-slate-800/80 border border-slate-700 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 text-slate-100 text-sm outline-none transition"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1.5">
                Tipo de Projeto
              </label>
              <select
                id="select-intake-type"
                value={type}
                onChange={(e) => setType(e.target.value)}
                className="w-full px-4 py-3 rounded-xl bg-slate-800/80 border border-slate-700 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 text-slate-100 text-sm outline-none transition"
              >
                <option value="feature">Novo Aplicativo / Sistema Web</option>
                <option value="mobile">Aplicativo Mobile (PWA)</option>
                <option value="suporte">Manutenção / Evolução de Software</option>
                <option value="bug">Correção de Problema Crítico</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1.5">
                O que você precisa que o sistema faça? *
              </label>
              <textarea
                id="textarea-intake-desc"
                rows={4}
                required
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Descreva o objetivo do projeto, funcionalidades principais e prazo esperado..."
                className="w-full px-4 py-3 rounded-xl bg-slate-800/80 border border-slate-700 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 text-slate-100 text-sm outline-none transition resize-none"
              />
            </div>

            {/* Honeypot field (escondido de humanos, preenchido por robôs) */}
            <div className="hidden" aria-hidden="true">
              <label>Website</label>
              <input
                type="text"
                name="website"
                tabIndex={-1}
                autoComplete="off"
                value={website}
                onChange={(e) => setWebsite(e.target.value)}
              />
            </div>

            <button
              id="btn-submit-intake"
              type="submit"
              disabled={submitting}
              className="w-full py-3.5 px-6 rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-bold text-sm transition-all shadow-lg shadow-indigo-600/20 active:scale-[0.99] flex items-center justify-center gap-2 mt-2"
            >
              {submitting ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Enviando solicitação...</span>
                </>
              ) : (
                <span>Enviar Solicitação de Orçamento</span>
              )}
            </button>
            <p className="text-[11px] text-center text-slate-500 mt-2">
              Seus dados serão tratados em conformidade estrita com a LGPD.
            </p>
          </form>
        )}
      </div>
    </div>
  );
}
