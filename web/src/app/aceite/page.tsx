"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";

interface ProposalView {
  card_title: string;
  client_name: string;
  amount_cents: number;
  currency: string;
  body: string;
  status: string;
}

function AceiteContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || (typeof window !== "undefined" ? window.location.pathname.split("/").pop() : "");

  const [proposal, setProposal] = useState<ProposalView | null>(null);
  const [loading, setLoading] = useState(true);
  const [accepting, setAccepting] = useState(false);
  const [accepted, setAccepted] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8085";

  useEffect(() => {
    let isMounted = true;

    async function fetchProposal() {
      if (!token) {
        if (isMounted) {
          setError("Token de proposta não fornecido.");
          setLoading(false);
        }
        return;
      }

      try {
        const res = await fetch(`${apiUrl}/v1/public/proposals/${token}`);
        if (!res.ok) {
          throw new Error("Proposta não encontrada ou link expirado.");
        }
        const data = await res.json();
        if (isMounted) {
          setProposal(data);
          if (data.status === "aceita") {
            setAccepted(true);
          }
        }
      } catch (err: unknown) {
        if (isMounted) {
          const msg = err instanceof Error ? err.message : "Falha ao carregar proposta.";
          setError(msg);
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    }

    fetchProposal();

    return () => {
      isMounted = false;
    };
  }, [token, apiUrl]);

  const handleAccept = async () => {
    if (!token) return;
    setAccepting(true);
    setError(null);

    try {
      const res = await fetch(`${apiUrl}/v1/public/proposals/${token}/accept`, {
        method: "POST",
      });
      if (!res.ok) {
        const errData = await res.text();
        throw new Error(errData || "Falha ao aceitar proposta.");
      }
      setAccepted(true);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Erro ao processar o aceite.";
      setError(msg);
    } finally {
      setAccepting(false);
    }
  };

  const formatBRL = (cents: number) => {
    return new Intl.NumberFormat("pt-BR", {
      style: "currency",
      currency: "BRL",
    }).format(cents / 100);
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-lg bg-slate-900 border border-slate-800 rounded-3xl p-6 sm:p-8 shadow-2xl relative overflow-hidden">
        {/* Glow de fundo */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-72 h-32 bg-emerald-500/10 blur-3xl pointer-events-none rounded-full" />

        <div className="text-center mb-6">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold mb-3">
            <span>⚡ Geevheit Software House</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">Proposta Comercial</h1>
          <p className="text-slate-400 text-sm mt-1">Revisão e confirmação de escopo</p>
        </div>

        {loading && (
          <div className="py-12 flex flex-col items-center justify-center gap-3">
            <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
            <span className="text-sm text-slate-400">Carregando dados da proposta...</span>
          </div>
        )}

        {error && (
          <div className="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm mb-6 text-center">
            {error}
          </div>
        )}

        {proposal && !loading && (
          <div className="space-y-6">
            <div className="p-4 rounded-2xl bg-slate-800/60 border border-slate-700/60 space-y-3">
              <div>
                <span className="text-xs text-slate-400 block font-medium">Cliente</span>
                <span className="text-base font-semibold text-slate-100">{proposal.client_name}</span>
              </div>
              <div>
                <span className="text-xs text-slate-400 block font-medium">Projeto / Demanda</span>
                <span className="text-base font-semibold text-slate-200">{proposal.card_title}</span>
              </div>
              <div>
                <span className="text-xs text-slate-400 block font-medium">Investimento</span>
                <span className="text-2xl font-black text-emerald-400">
                  {formatBRL(proposal.amount_cents)}
                </span>
              </div>
            </div>

            {proposal.body && (
              <div className="p-4 rounded-2xl bg-slate-800/40 border border-slate-800">
                <span className="text-xs text-slate-400 block font-medium mb-1">Escopo e Condições</span>
                <p className="text-sm text-slate-300 whitespace-pre-wrap leading-relaxed">{proposal.body}</p>
              </div>
            )}

            {accepted ? (
              <div className="p-5 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-center space-y-2">
                <div className="w-12 h-12 bg-emerald-500/20 text-emerald-400 rounded-full flex items-center justify-center mx-auto text-xl">
                  ✓
                </div>
                <h3 className="text-lg font-bold text-emerald-400">Proposta Aceita!</h3>
                <p className="text-xs text-slate-300 leading-relaxed">
                  Obrigado pela confiança! O projeto foi enviado diretamente para a nossa equipe técnica de engenharia. Em breve você receberá atualizações sobre o início da execução.
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                <button
                  id="btn-accept-proposal"
                  onClick={handleAccept}
                  disabled={accepting}
                  className="w-full py-4 px-6 rounded-2xl bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 text-slate-950 font-bold text-base transition-all shadow-lg shadow-emerald-500/20 active:scale-[0.99] flex items-center justify-center gap-2"
                >
                  {accepting ? (
                    <>
                      <div className="w-5 h-5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
                      <span>Confirmando aceite...</span>
                    </>
                  ) : (
                    <span>Aprovar e Iniciar Projeto</span>
                  )}
                </button>
                <p className="text-[11px] text-center text-slate-500">
                  Ao clicar em aprovar, você concorda com o escopo e os valores apresentados nesta proposta.
                </p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default function AceitePage() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-slate-950 text-slate-400 flex items-center justify-center">Carregando...</div>}>
      <AceiteContent />
    </Suspense>
  );
}
