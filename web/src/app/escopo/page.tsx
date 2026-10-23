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

  const [loading, setLoading] = useState(true);
  const [data, setData] = useState<ScopeData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [approving, setApproving] = useState(false);
  const [approvedSuccess, setApprovedSuccess] = useState(false);

  const apiURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8085";

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
          throw new Error("Especificação de escopo não encontrada ou link expirado.");
        }
        return res.json();
      })
      .then((scope: ScopeData) => {
        setData(scope);
        if (scope.client_approved) {
          setApprovedSuccess(true);
        }
      })
      .catch((err) => {
        setError(err.message);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [token, apiURL]);

  const handleApprove = async () => {
    if (!token) return;
    setApproving(true);
    try {
      const res = await fetch(`${apiURL}/v1/public/scope/${token}/approve`, {
        method: "POST",
      });
      if (!res.ok) {
        throw new Error("Falha ao registrar aprovação do escopo.");
      }
      setApprovedSuccess(true);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Erro ao aprovar escopo.";
      alert(message);
    } finally {
      setApproving(false);
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
                {approving ? "Registrando aprovação..." : "Aprovar Escopo e Iniciar Produção"}
              </button>
            </div>
          ) : (
            <div className="text-center pt-2">
              <span className="text-xs text-emerald-600 font-medium">
                Obrigado! Sua aprovação foi registrada de forma imutável nos servidores da Geevheit.
              </span>
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
