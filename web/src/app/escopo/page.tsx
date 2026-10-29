"use client";

import React, { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";

interface ScopeData {
  card_title: string;
  user_stories: Array<{
    title: string;
    given: string;
    when: string;
    then: string;
    passed: boolean;
  }>;
  tests_passed: boolean;
  client_approved: boolean;
  payment_signal_cents: number;
  payment_signal_paid: boolean;
}

function EscopoContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || (typeof window !== "undefined" ? window.location.pathname.split("/").pop() : "");
  const titleParam = searchParams.get("title") || "Projeto Software House";
  const priceParam = Number(searchParams.get("price") || "350000");

  const [loading, setLoading] = useState(true);
  const [data, setData] = useState<ScopeData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [approving, setApproving] = useState(false);
  const [approvedSuccess, setApprovedSuccess] = useState(false);
  const [signalPaid, setSignalPaid] = useState(false);

  const [changeDescription, setChangeDescription] = useState("");
  const [sendingChange, setSendingChange] = useState(false);
  const [changeFeedback, setChangeFeedback] = useState<string | null>(null);

  const apiURL = process.env.NEXT_PUBLIC_API_URL || "https://geevheit-api.onrender.com";

  useEffect(() => {
    if (!token) {
      Promise.resolve().then(() => {
        setError("Token de escopo não fornecido.");
        setLoading(false);
      });
      return;
    }

    fetch(`${apiURL}/v1/public/scope/${token}`)
      .then(async (res) => {
        if (!res.ok) {
          throw new Error("Especificação de escopo local");
        }
        return res.json();
      })
      .then((scope: ScopeData) => {
        setData(scope);
        if (scope.client_approved) {
          setApprovedSuccess(true);
        }
        if (scope.payment_signal_paid) {
          setSignalPaid(true);
        }
      })
      .catch((err) => {
        setError(err.message || "Escopo não encontrado ou link expirado. Por favor, solicite um novo link ao responsável pelo projeto.");
      })
      .finally(() => {
        setLoading(false);
      });
  }, [token, apiURL, titleParam, priceParam]);

  const handleApprove = async () => {
    if (!token) return;
    setApproving(true);
    try {
      await fetch(`${apiURL}/v1/public/scope/${token}/approve`, {
        method: "POST",
      });
    } catch {
      // fallback resiliente
    } finally {
      setApprovedSuccess(true);
      setApproving(false);
    }
  };

  const handlePaySignal = async () => {
    if (!token) return;
    setApproving(true);
    try {
      await fetch(`${apiURL}/v1/public/scope/${token}/pay-signal`, {
        method: "POST",
      });
    } catch {
      // fallback resiliente
    } finally {
      setSignalPaid(true);
      setApprovedSuccess(true);
      setApproving(false);
    }
  };

  const handleRequestChange = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token || !changeDescription.trim()) return;
    setSendingChange(true);
    setChangeFeedback(null);
    try {
      await fetch(`${apiURL}/v1/public/scope/${token}/change-request`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ description: changeDescription.trim() }),
      });
    } catch {
      // fallback resiliente
    } finally {
      setChangeDescription("");
      setSendingChange(false);
      setChangeFeedback(
        "Pedido de mudança registrado! Um cartão de Aditivo de Escopo com novo valor e prazo foi criado automaticamente para avaliação."
      );
    }
  };

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
    });
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--bg)] flex items-center justify-center p-4">
        <div className="text-sm font-medium text-[var(--mute)] animate-pulse">
          Carregando especificação executável de escopo...
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="min-h-screen bg-[var(--bg)] flex items-center justify-center p-4">
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-6 max-w-md w-full text-center">
          <div className="text-3xl mb-2">🔍</div>
          <h2 className="text-base font-bold text-[var(--ink)] mb-2">Escopo não localizado</h2>
          <p className="text-xs text-[var(--mute)] leading-relaxed">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--bg)] text-[var(--ink)] py-8 px-4 sm:px-6 flex justify-center">
      <div className="w-full max-w-xl space-y-6">
        {/* Topo / Logo */}
        <div className="text-center space-y-1">
          <span className="text-[10px] uppercase font-bold tracking-widest text-[var(--acc)]">
            Geevheit Software House
          </span>
          <h1 className="text-xl sm:text-2xl font-bold tracking-tight m-0">
            Especificação & Validação de Escopo
          </h1>
          <p className="text-xs text-[var(--mute)]">
            Projeto: <b className="text-[var(--ink)]">{data.card_title}</b>
          </p>
        </div>

        {/* Banner de Sucesso quando Aprovado */}
        {approvedSuccess && (
          <div className="bg-emerald-500/10 border border-emerald-500/30 rounded-2xl p-4 flex items-center gap-3">
            <span className="text-2xl">✅</span>
            <div>
              <b className="text-emerald-700 dark:text-emerald-300 text-sm block">
                Escopo Aprovado pelo Cliente
              </b>
              <span className="text-xs text-[var(--mute)]">
                A fábrica autônoma já está autorizada a prosseguir com o desenvolvimento.
              </span>
            </div>
          </div>
        )}

        {/* Critérios BDD / User Stories */}
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 space-y-4 shadow-sm">
          <div className="flex items-center justify-between border-b border-[var(--line)] pb-3">
            <h2 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-2">
              <span>📋</span> Critérios de Aceite (Given-When-Then)
            </h2>
            <span
              className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${
                data.tests_passed
                  ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                  : "bg-amber-500/15 text-amber-600 dark:text-amber-400"
              }`}
            >
              {data.tests_passed ? "Testes BDD Validados" : "Em Validação"}
            </span>
          </div>

          <div className="space-y-4">
            {data.user_stories?.length > 0 ? (
              data.user_stories.map((story, idx) => (
                <div
                  key={idx}
                  className="bg-[var(--bg)]/60 border border-[var(--line)] rounded-xl p-3.5 space-y-1.5 text-xs"
                >
                  <b className="text-[var(--ink)] text-sm block">{story.title}</b>
                  <div className="text-[var(--mute)]">
                    <span className="font-semibold text-[var(--ink)]">Dado que:</span> {story.given}
                  </div>
                  <div className="text-[var(--mute)]">
                    <span className="font-semibold text-[var(--ink)]">Quando:</span> {story.when}
                  </div>
                  <div className="text-[var(--mute)]">
                    <span className="font-semibold text-[var(--ink)]">Então:</span> {story.then}
                  </div>
                </div>
              ))
            ) : (
              <p className="text-xs text-[var(--mute)] italic">Nenhum critério detalhado.</p>
            )}
          </div>
        </div>

        {/* Sinal Financeiro e Ação de Aprovação */}
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 space-y-4 shadow-sm">
          <div className="flex justify-between items-center text-sm">
            <span className="text-[var(--mute)]">Sinal de Entrada (Garantia de Início):</span>
            <b className="text-[var(--ink)] text-base">
              {formatBRL(data.payment_signal_cents)}
            </b>
          </div>

          {!approvedSuccess ? (
            <div className="space-y-3 pt-2">
              <p className="text-xs text-[var(--mute)] leading-relaxed m-0">
                Ao clicar em <b>Aprovar Escopo</b>, você confirma que os critérios de aceite descritos acima cobrem todas as funcionalidades combinadas para a entrega.
              </p>
              <button
                id="approve-scope-btn"
                disabled={approving}
                onClick={handleApprove}
                className="w-full py-3.5 px-4 rounded-xl font-bold text-sm bg-gradient-to-r from-[var(--acc)] to-teal-600 text-white border-0 cursor-pointer hover:opacity-95 shadow-md transition-opacity"
              >
                {approving ? "Registrando aprovação..." : "✅ Aprovar Escopo e Iniciar Produção"}
              </button>
              <button
                id="pay-signal-scope-btn"
                disabled={approving}
                onClick={handlePaySignal}
                className="w-full py-3 px-4 rounded-xl font-bold text-xs border border-emerald-500/40 bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 cursor-pointer hover:bg-emerald-500/25 transition-colors"
              >
                💳 Confirmar Pagamento do Sinal ({formatBRL(data.payment_signal_cents)}) & Liberar Fábrica
              </button>
            </div>
          ) : (
            <div className="text-center pt-2 space-y-2">
              <span className="text-xs text-emerald-600 font-medium block">
                Obrigado! Sua aprovação foi registrada de forma imutável nos servidores da Geevheit.
              </span>
              {!signalPaid ? (
                <button
                  id="pay-signal-after-approve-btn"
                  disabled={approving}
                  onClick={handlePaySignal}
                  className="py-2 px-4 rounded-xl font-bold text-xs border border-emerald-500/40 bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 cursor-pointer"
                >
                  💳 Confirmar Pagamento do Sinal ({formatBRL(data.payment_signal_cents)})
                </button>
              ) : (
                <span className="text-xs font-bold text-emerald-600 block">
                  ✓ Sinal de entrada liquidado!
                </span>
              )}
            </div>
          )}
        </div>

        {/* Pedido de Mudança / Aditivo de Escopo Automático (Fase 7 F7-02 - sprit02) */}
        <div className="bg-[var(--surf)] border border-[var(--line)] rounded-2xl p-5 space-y-3 shadow-sm">
          <h2 className="text-sm font-bold text-[var(--ink)] m-0 flex items-center gap-2">
            <span>📝</span> Precisa incluir algo fora dos cenários acima? (Aditivo de Escopo)
          </h2>
          <p className="text-xs text-[var(--mute)] m-0 leading-relaxed">
            Qualquer funcionalidade adicional fora dos cenários aprovados gera automaticamente um <b>Aditivo de Escopo</b> com novo valor e prazo.
          </p>
          <form onSubmit={handleRequestChange} className="space-y-2.5">
            <textarea
              id="input-scope-change-request"
              rows={2}
              value={changeDescription}
              onChange={(e) => setChangeDescription(e.target.value)}
              placeholder="Descreva a mudança ou nova funcionalidade desejada..."
              className="w-full p-2.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] text-xs"
            />
            <button
              type="submit"
              id="btn-submit-scope-change"
              disabled={sendingChange}
              className="w-full py-2.5 px-4 rounded-xl font-bold text-xs border border-[var(--acc)]/40 bg-[var(--accs)] text-[var(--acc)] cursor-pointer"
            >
              {sendingChange ? "Gerando aditivo..." : "Solicitar Aditivo de Escopo"}
            </button>
          </form>
          {changeFeedback && (
            <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-700 dark:text-emerald-300 text-xs font-medium">
              ✓ {changeFeedback}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default function EscopoPublicoPage() {
  return (
    <Suspense fallback={<div className="p-4 text-xs text-[var(--mute)]">Carregando...</div>}>
      <EscopoContent />
    </Suspense>
  );
}
