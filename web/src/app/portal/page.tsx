"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { ClientPortalData } from "@/shared/api-client/endpoints";

function PortalContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || (typeof window !== "undefined" ? window.location.pathname.split("/").pop() : "");

  const [portalData, setPortalData] = useState<ClientPortalData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Modal de Aceite Formal
  const [showAcceptModal, setShowAcceptModal] = useState(false);
  const [acceptedByName, setAcceptedByName] = useState("");
  const [accepting, setAccepting] = useState(false);
  const [acceptSuccess, setAcceptSuccess] = useState(false);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  useEffect(() => {
    let isMounted = true;

    async function fetchPortal() {
      if (!token) {
        if (isMounted) {
          setError("Link de acesso não informado ou inválido.");
          setLoading(false);
        }
        return;
      }

      try {
        const res = await fetch(`${apiUrl}/v1/public/portal/${token}`);
        if (!res.ok) {
          throw new Error("Link de acesso inválido ou expirado.");
        }
        const data: ClientPortalData = await res.json();
        if (isMounted) {
          setPortalData(data);
          setError(null);
        }
      } catch (err: unknown) {
        if (isMounted) {
          const msg = err instanceof Error ? err.message : "Falha ao carregar o portal do projeto.";
          setError(msg);
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    }

    fetchPortal();

    // Auto-polling a cada 15 segundos para atualizar a linha do tempo em tempo real
    const interval = setInterval(fetchPortal, 15000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, [token, apiUrl]);

  const handleConfirmAccept = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token || !acceptedByName.trim()) return;

    setAccepting(true);
    try {
      const res = await fetch(`${apiUrl}/v1/public/portal/${token}/accept`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ accepted_by: acceptedByName.trim() }),
      });

      if (!res.ok) {
        const errJson = await res.json().catch(() => ({}));
        throw new Error(errJson.error || "Falha ao registrar aceite formal.");
      }

      const updated: ClientPortalData = await res.json();
      setPortalData(updated);
      setAcceptSuccess(true);
      setShowAcceptModal(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Erro ao registrar o aceite.";
      alert(msg);
    } finally {
      setAccepting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0d1117] text-[#c9d1d9] flex flex-col items-center justify-center p-4">
        <div className="w-10 h-10 border-3 border-[#0e6b73] border-t-transparent rounded-full animate-spin mb-3"></div>
        <div className="text-xs text-[#8b949e]">Carregando portal do seu projeto...</div>
      </div>
    );
  }

  if (error || !portalData) {
    return (
      <div className="min-h-screen bg-[#0d1117] text-[#c9d1d9] flex items-center justify-center p-4 font-sans">
        <div className="bg-[#161b22] border border-[#30363d] rounded-2xl p-6 max-w-sm w-full text-center shadow-xl">
          <div className="w-12 h-12 rounded-2xl bg-[#b42318]/10 text-[#f85149] mx-auto flex items-center justify-center text-xl mb-3">
            🔒
          </div>
          <h2 className="text-base font-bold text-[#f0f6fc] mb-1">Acesso Indisponível</h2>
          <p className="text-xs text-[#8b949e] mb-4 leading-relaxed">
            {error || "Este link de acompanhamento é inválido, expirou ou foi revogado pela administração."}
          </p>
          <div className="text-[11px] text-[#8b949e] bg-[#0d1117] p-3 rounded-xl border border-[#30363d]">
            Caso precise de um novo link, solicite diretamente ao seu gerente de projeto na Geevheit.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#0d1117] text-[#c9d1d9] font-sans antialiased py-6 px-4">
      <div className="max-w-2xl mx-auto space-y-5">
        {/* Cabeçalho do Cliente */}
        <header className="bg-[#161b22] border border-[#30363d] rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between gap-3 mb-3">
            <div className="flex items-center gap-2.5">
              <img
                src="/icon-192.png"
                alt="Geevheit"
                className="w-10 h-10 rounded-xl border border-[#30363d] object-cover"
              />
              <div>
                <span className="text-[10px] font-bold uppercase tracking-wider text-[#4fc3cc]">
                  Portal do Cliente · Geevheit
                </span>
                <h1 className="text-lg font-bold text-[#f0f6fc] m-0 leading-tight">
                  {portalData.project_title}
                </h1>
              </div>
            </div>

            <span className="text-[11px] font-semibold px-2.5 py-1 rounded-full bg-[#0e6b73]/20 text-[#4fc3cc] border border-[#0e6b73]/40">
              {portalData.client_name}
            </span>
          </div>

          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pt-3 border-t border-[#30363d]/60 text-xs text-[#8b949e]">
            <div>
              Status Atual: <b className="text-[#f0f6fc]">{portalData.current_stage_label}</b>
            </div>
            <div className="flex items-center gap-2">
              <span>Progresso Geral:</span>
              <b className="text-[#4fc3cc]">{portalData.progress_percentage}%</b>
            </div>
          </div>

          {/* Barra de Progresso Geral */}
          <div className="h-2 w-full bg-[#0d1117] rounded-full overflow-hidden mt-2.5 flex">
            <div
              style={{ width: `${portalData.progress_percentage}%` }}
              className="bg-gradient-to-r from-[#0e6b73] to-[#4fc3cc] h-full rounded-full transition-all duration-500"
            ></div>
          </div>
        </header>

        {/* Alerta de Aceite Concluído */}
        {(portalData.is_accepted || acceptSuccess) && (
          <div className="bg-[#027a48]/10 border border-[#027a48]/40 rounded-2xl p-4 flex items-start gap-3 text-xs shadow-xs animate-fadeIn">
            <span className="text-xl">✅</span>
            <div>
              <h3 className="font-bold text-[#3fb950] text-sm m-0">Entrega Formalmente Aceita</h3>
              <p className="text-[#8b949e] m-0 mt-0.5 leading-relaxed">
                Este projeto foi homologado e aceito por <b>{portalData.accepted_by || acceptedByName}</b>
                {portalData.accepted_at && (
                  <> em {new Date(portalData.accepted_at).toLocaleDateString("pt-BR", { day: "2-digit", month: "long", year: "numeric", hour: "2-digit", minute: "2-digit" })}</>
                )}
                . O registro de conformidade está gravado com segurança.
              </p>
            </div>
          </div>
        )}

        {/* Linha do Tempo de Etapas (4 Fases) */}
        <section className="bg-[#161b22] border border-[#30363d] rounded-2xl p-5 shadow-sm space-y-4">
          <h2 className="text-sm font-bold text-[#f0f6fc] m-0 flex items-center gap-1.5">
            <span>⏱️</span>
            <span>Linha do Tempo de Produção</span>
          </h2>

          <div className="space-y-3">
            {portalData.stages.map((stage, idx) => {
              const isCompleted = stage.status === "concluido";
              const isCurrent = stage.status === "em_andamento";

              return (
                <div
                  key={stage.key}
                  className={`p-3.5 rounded-xl border transition-all flex items-start gap-3 ${
                    isCompleted
                      ? "border-[#027a48]/40 bg-[#027a48]/5"
                      : isCurrent
                      ? "border-[#0e6b73] bg-[#0e6b73]/10 shadow-xs"
                      : "border-[#30363d] bg-[#0d1117]/40 opacity-70"
                  }`}
                >
                  <div
                    className={`w-7 h-7 rounded-full flex items-center justify-center font-bold text-xs flex-none ${
                      isCompleted
                        ? "bg-[#027a48] text-white"
                        : isCurrent
                        ? "bg-[#0e6b73] text-white ring-2 ring-[#4fc3cc]/40 animate-pulse"
                        : "bg-[#30363d] text-[#8b949e]"
                    }`}
                  >
                    {isCompleted ? "✓" : idx + 1}
                  </div>

                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <h4 className="text-xs sm:text-sm font-bold text-[#f0f6fc] m-0">
                        {stage.label}
                      </h4>
                      <span
                        className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${
                          isCompleted
                            ? "bg-[#027a48]/20 text-[#3fb950]"
                            : isCurrent
                            ? "bg-[#0e6b73]/20 text-[#4fc3cc]"
                            : "bg-[#30363d] text-[#8b949e]"
                        }`}
                      >
                        {isCompleted && "Concluída"}
                        {isCurrent && "Em Andamento"}
                        {!isCompleted && !isCurrent && "Aguardando"}
                      </span>
                    </div>
                    <p className="text-xs text-[#8b949e] m-0 mt-1 leading-relaxed">
                      {stage.description}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        {/* Botão de Aceite Formal da Entrega */}
        {portalData.can_accept && !portalData.is_accepted && !acceptSuccess && (
          <section className="bg-gradient-to-r from-[#0e6b73]/20 via-[#161b22] to-[#161b22] border-2 border-[#0e6b73] rounded-2xl p-5 shadow-lg flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <div className="text-xs font-bold uppercase tracking-wider text-[#4fc3cc] mb-0.5">
                Homologação Concluída
              </div>
              <h3 className="text-base font-bold text-[#f0f6fc] m-0">
                Seu projeto está pronto para entrega final!
              </h3>
              <p className="text-xs text-[#8b949e] m-0 mt-1">
                Revise os entregáveis e confirme o aceite formal para liberação dos acessos de produção.
              </p>
            </div>

            <button
              id="btn-open-accept-modal"
              onClick={() => setShowAcceptModal(true)}
              className="bg-[#027a48] hover:bg-[#2ea043] text-white font-bold text-xs sm:text-sm py-3 px-5 rounded-xl border-0 cursor-pointer shadow-md transition-all flex items-center justify-center gap-2 flex-none"
            >
              <span>✍️</span>
              <span>Aceite Formal da Entrega</span>
            </button>
          </section>
        )}

        {/* Rodapé institucional */}
        <footer className="text-center text-[11px] text-[#8b949e] pt-4 border-t border-[#30363d]/40">
          Geevheit Software House · Sistema Seguro de Orquestração e Gestão de Projetos
        </footer>
      </div>

      {/* Modal de Confirmação em 2 Etapas do Aceite Formal */}
      {showAcceptModal && (
        <div className="fixed inset-0 bg-black/70 z-50 flex items-center justify-center p-4 backdrop-blur-xs">
          <div className="bg-[#161b22] border border-[#30363d] rounded-2xl p-5 max-w-md w-full shadow-2xl text-xs">
            <div className="w-10 h-10 rounded-xl bg-[#027a48]/20 text-[#3fb950] flex items-center justify-center text-lg mb-3">
              ✍️
            </div>
            <h3 className="text-base font-bold text-[#f0f6fc] mb-1">
              Confirmação de Aceite da Entrega
            </h3>
            <p className="text-[#8b949e] mb-4 leading-relaxed">
              Ao confirmar o aceite, você declara formalmente que os testes foram validados e os requisitos do projeto <b>{portalData.project_title}</b> foram cumpridos com sucesso.
            </p>

            <form onSubmit={handleConfirmAccept} className="space-y-3.5">
              <div>
                <label htmlFor="input-accepted-by" className="block font-semibold text-[#8b949e] mb-1">
                  Seu Nome Completo e Cargo *
                </label>
                <input
                  id="input-accepted-by"
                  type="text"
                  required
                  value={acceptedByName}
                  onChange={(e) => setAcceptedByName(e.target.value)}
                  placeholder="Ex: Carlos Mendes — Diretor de Operações"
                  className="w-full px-3 py-2 rounded-xl border border-[#30363d] bg-[#0d1117] text-[#f0f6fc] text-xs focus:outline-none focus:border-[#4fc3cc]"
                />
              </div>

              <div className="p-3 rounded-xl bg-[#0d1117] border border-[#30363d] text-[11px] text-[#8b949e]">
                🔒 <b>Registro Imutável:</b> Este aceite registrará data, hora, IP de origem e hash criptográfico do estado da entrega na trilha de auditoria da Geevheit.
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowAcceptModal(false)}
                  disabled={accepting}
                  className="px-4 py-2 rounded-xl font-semibold border border-[#30363d] bg-transparent text-[#c9d1d9] cursor-pointer hover:bg-[#30363d]/40"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={accepting || !acceptedByName.trim()}
                  className="px-4 py-2 rounded-xl font-bold border-0 bg-[#027a48] hover:bg-[#2ea043] text-white cursor-pointer transition-all disabled:opacity-50"
                >
                  {accepting ? "Registrando Aceite..." : "Confirmar e Assinar Aceite"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default function PortalPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-[#0d1117] text-[#c9d1d9] flex items-center justify-center text-xs">
          Carregando portal...
        </div>
      }
    >
      <PortalContent />
    </Suspense>
  );
}
