"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";

export default function TransportSimPage() {
  const [truckProgress, setTruckProgress] = useState(45);
  const [activeTab, setActiveTab] = useState<"visao" | "financeiro" | "carga">("visao");
  const [selectedPlan, setSelectedPlan] = useState<"completo" | "pro">("completo");

  // Estado do Modal de Pedido / Matrícula
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState("Matrícula Formação Completa 360°");
  const [clientName, setClientName] = useState("");
  const [contact, setContact] = useState("");
  const [company, setCompany] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const openOrderModal = (title: string) => {
    setModalTitle(title);
    setIsModalOpen(true);
    setSubmitted(false);
    setErrorMessage(null);
  };

  const handleOrderSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!clientName.trim() || !contact.trim()) {
      setErrorMessage("Por favor, preencha seu nome e seu WhatsApp ou e-mail.");
      return;
    }

    setSubmitting(true);
    setErrorMessage(null);

    const apiUrl = process.env.NEXT_PUBLIC_API_URL || "https://geevheit-api.onrender.com";
    const payload = {
      client_name: clientName.trim(),
      contact: contact.trim(),
      type: "orcamento",
      description: `Pedido de Matrícula: ${modalTitle}. Empresa/Cidade: ${company.trim() || "Pessoa Física"}. Origem: TransportSim 360 Landing Page.`,
      website: "" // honeypot vazio
    };

    try {
      const res = await fetch(`${apiUrl}/v1/intake`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        setSubmitted(true);
      } else {
        const text = await res.text();
        setErrorMessage(text || "Erro ao registrar pedido. Tente novamente.");
      }
    } catch {
      setErrorMessage("Não foi possível conectar ao servidor. Verifique sua conexão e tente novamente.");
    } finally {
      setSubmitting(false);
    }
  };

  // Animação sutil do caminhão no mapa
  useEffect(() => {
    const interval = setInterval(() => {
      setTruckProgress((prev) => (prev >= 90 ? 15 : prev + 1));
    }, 800);
    return () => clearInterval(interval);
  }, []);

  return (
    <div style={{ backgroundColor: "#F7F9F5", color: "#111827", minHeight: "100vh", fontFamily: "'Inter', system-ui, -apple-system, sans-serif" }}>
      {/* 1. TOP BANNER */}
      <div style={{ backgroundColor: "#0F3822", color: "#E0EED5", padding: "8px 16px", textAlign: "center", fontSize: "13px", fontWeight: 500 }}>
        🚚 Turma com vagas abertas: Ganhe acesso ao simulador de frotas 360° + Certificado Oficial de Logística.
      </div>

      {/* 2. HEADER */}
      <header style={{
        maxWidth: "1280px",
        margin: "0 auto",
        padding: "20px 24px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        borderBottom: "1px solid #E8EDE4"
      }}>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          {/* Logo Icon */}
          <div style={{
            width: "36px",
            height: "36px",
            backgroundColor: "#0F3822",
            borderRadius: "8px",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            color: "#FFF"
          }}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect x="1" y="3" width="15" height="13"></rect>
              <polygon points="16 8 20 8 23 11 23 16 16 16 16 8"></polygon>
              <circle cx="5.5" cy="18.5" r="2.5"></circle>
              <circle cx="18.5" cy="18.5" r="2.5"></circle>
            </svg>
          </div>
          <span style={{ fontSize: "20px", fontWeight: 800, color: "#0F3822", letterSpacing: "-0.5px" }}>
            TransportSim <span style={{ color: "#588135" }}>360</span>
          </span>
        </div>

        <nav style={{ display: "flex", alignItems: "center", gap: "28px" }} className="desktop-nav">
          <a href="#como-funciona" style={{ color: "#374151", textDecoration: "none", fontSize: "15px", fontWeight: 500 }}>Como funciona</a>
          <a href="#o-que-aprende" style={{ color: "#374151", textDecoration: "none", fontSize: "15px", fontWeight: 500 }}>O que você aprende</a>
          <a href="#modulos" style={{ color: "#374151", textDecoration: "none", fontSize: "15px", fontWeight: 500 }}>Grade curricular</a>
          <a href="#preco" style={{ color: "#374151", textDecoration: "none", fontSize: "15px", fontWeight: 500 }}>Planos</a>
        </nav>

        <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <button
            onClick={() => openOrderModal("Acesso à Plataforma / Demonstração")}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: "8px",
              padding: "10px 18px",
              backgroundColor: "#FFF",
              color: "#0F3822",
              border: "1px solid #D5DEC9",
              borderRadius: "10px",
              fontSize: "14px",
              fontWeight: 600,
              cursor: "pointer",
              boxShadow: "0 1px 2px rgba(0,0,0,0.05)"
            }}
          >
            Acessar plataforma ➔
          </button>
        </div>
      </header>

      {/* 3. HERO SECTION (Exatamente como a imagem de referência) */}
      <section style={{ maxWidth: "1280px", margin: "0 auto", padding: "56px 24px 72px" }}>
        <div style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(340px, 1fr))",
          gap: "48px",
          alignItems: "center"
        }}>
          {/* Lado Esquerdo: Textos & CTAs */}
          <div>
            <div style={{
              display: "inline-flex",
              alignItems: "center",
              gap: "6px",
              padding: "4px 12px",
              backgroundColor: "#EBF3E6",
              color: "#3A601E",
              borderRadius: "20px",
              fontSize: "12px",
              fontWeight: 700,
              letterSpacing: "0.8px",
              textTransform: "uppercase",
              marginBottom: "20px"
            }}>
              <span style={{ width: "6px", height: "6px", borderRadius: "50%", backgroundColor: "#588135" }}></span>
              APRENDER. DECIDIR. EVOLUIR.
            </div>

            <h1 style={{
              fontSize: "clamp(36px, 5.5vw, 58px)",
              lineHeight: 1.12,
              fontWeight: 900,
              color: "#0F3822",
              letterSpacing: "-1.5px",
              marginBottom: "24px"
            }}>
              Na logística, cada decisão <span style={{ color: "#588135" }}>move o resultado.</span>
            </h1>

            <p style={{
              fontSize: "18px",
              lineHeight: 1.6,
              color: "#4B5563",
              marginBottom: "36px",
              maxWidth: "520px"
            }}>
              Do primeiro contato comercial ao fechamento financeiro: viva uma operação de transporte completa, antes de assumir uma de verdade.
            </p>

            <div style={{ display: "flex", flexWrap: "wrap", alignItems: "center", gap: "16px" }}>
              <button
                onClick={() => openOrderModal("Matrícula Formação Completa 360°")}
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  gap: "10px",
                  padding: "16px 28px",
                  backgroundColor: "#0F3822",
                  color: "#FFFFFF",
                  borderRadius: "12px",
                  fontSize: "16px",
                  fontWeight: 700,
                  cursor: "pointer",
                  border: "none",
                  boxShadow: "0 4px 14px rgba(15, 56, 34, 0.25)",
                  transition: "transform 0.2s"
                }}
              >
                Começar meu treinamento ➔
              </button>

              <a href="#simulador" style={{
                display: "inline-flex",
                alignItems: "center",
                gap: "6px",
                padding: "16px 20px",
                color: "#0F3822",
                fontSize: "16px",
                fontWeight: 600,
                textDecoration: "none"
              }}>
                Conhecer o simulador ↗
              </a>
            </div>

            {/* Badges de garantia / avaliações */}
            <div style={{ marginTop: "40px", display: "flex", alignItems: "center", gap: "20px" }}>
              <div style={{ display: "flex", gap: "2px", color: "#F59E0B" }}>
                {"★★★★★"}
              </div>
              <span style={{ fontSize: "14px", color: "#6B7280" }}>
                <strong>4.9/5</strong> por mais de 2.800 gestores de tráfego e frota
              </span>
            </div>
          </div>

          {/* Lado Direito: Card da Central de Operações & Mapa */}
          <div style={{
            backgroundColor: "#FFFFFF",
            borderRadius: "24px",
            border: "1px solid #E1E8DC",
            boxShadow: "0 20px 40px -15px rgba(15, 56, 34, 0.08)",
            padding: "24px",
            overflow: "hidden"
          }}>
            {/* Topo do Card */}
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
              <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                <span style={{ width: "8px", height: "8px", borderRadius: "50%", backgroundColor: "#588135" }}></span>
                <span style={{ fontSize: "12px", fontWeight: 700, letterSpacing: "0.5px", color: "#4B5563" }}>
                  CENTRAL DE OPERAÇÕES
                </span>
              </div>
              <span style={{
                fontSize: "11px",
                fontWeight: 600,
                color: "#2D5A1E",
                backgroundColor: "#EBF3E6",
                padding: "3px 10px",
                borderRadius: "12px"
              }}>
                Treinamento
              </span>
            </div>

            {/* Missão em Destaque */}
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "16px" }}>
              <div>
                <span style={{ fontSize: "11px", textTransform: "uppercase", color: "#9CA3AF", fontWeight: 600, letterSpacing: "0.5px" }}>
                  MISSÃO EM DESTAQUE
                </span>
                <h3 style={{ fontSize: "22px", fontWeight: 800, color: "#0F3822", margin: "4px 0" }}>
                  Expresso Sudeste
                </h3>
                <span style={{ fontSize: "14px", color: "#6B7280" }}>
                  São Paulo ➔ Rio de Janeiro
                </span>
              </div>
              <div style={{
                width: "44px",
                height: "44px",
                borderRadius: "50%",
                border: "1px solid #E5E7EB",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "#0F3822"
              }}>
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="1" y="3" width="15" height="13"></rect>
                  <polygon points="16 8 20 8 23 11 23 16 16 16 16 8"></polygon>
                  <circle cx="5.5" cy="18.5" r="2.5"></circle>
                  <circle cx="18.5" cy="18.5" r="2.5"></circle>
                </svg>
              </div>
            </div>

            {/* Visualizador do Mapa com SVG de Rota Real */}
            <div style={{
              backgroundColor: "#EBF3EA",
              borderRadius: "16px",
              border: "1px solid #D5E2D2",
              padding: "16px",
              position: "relative",
              overflow: "hidden"
            }}>
              <div style={{ display: "flex", justifyContent: "space-between", fontSize: "12px", color: "#374151", marginBottom: "12px", fontWeight: 600 }}>
                <span>• MAPA DA OPERAÇÃO</span>
                <span style={{ color: "#6B7280", fontWeight: 400 }}>São Paulo ➔ Rio de Janeiro</span>
              </div>

              {/* Rota Gráfica Estilizada */}
              <div style={{ position: "relative", height: "180px", backgroundColor: "#D4E6D2", borderRadius: "12px", overflow: "hidden", display: "flex", alignItems: "center", justifyContent: "center" }}>
                {/* Relevo e Costas */}
                <svg width="100%" height="100%" style={{ position: "absolute", top: 0, left: 0 }} viewBox="0 0 400 180">
                  {/* Oceano Atlântico */}
                  <path d="M 0,140 Q 150,110 250,135 T 400,110 L 400,180 L 0,180 Z" fill="#B4D5E8" opacity="0.8" />
                  {/* Linha da Rodovia Dutra (BR-116) */}
                  <path d="M 60,110 C 120,95 200,85 340,70" fill="none" stroke="#0F3822" strokeWidth="5" strokeLinecap="round" />
                  <path d="M 60,110 C 120,95 200,85 340,70" fill="none" stroke="#588135" strokeWidth="2" strokeDasharray="6 4" />
                </svg>

                {/* Armazém São Paulo (Origem) */}
                <div style={{
                  position: "absolute",
                  left: "35px",
                  top: "75px",
                  backgroundColor: "#FFFFFF",
                  border: "1px solid #0F3822",
                  borderRadius: "8px",
                  padding: "4px 8px",
                  fontSize: "11px",
                  fontWeight: 700,
                  color: "#0F3822",
                  boxShadow: "0 2px 6px rgba(0,0,0,0.1)",
                  display: "flex",
                  alignItems: "center",
                  gap: "4px"
                }}>
                  🏢 1. São Paulo/SP
                </div>

                {/* Caminhão em Movimento */}
                <div style={{
                  position: "absolute",
                  left: `${truckProgress}%`,
                  top: "65px",
                  transform: "translate(-50%, -50%)",
                  backgroundColor: "#0F3822",
                  color: "#FFFFFF",
                  borderRadius: "10px",
                  padding: "6px 10px",
                  display: "flex",
                  alignItems: "center",
                  gap: "6px",
                  boxShadow: "0 4px 10px rgba(0,0,0,0.25)",
                  transition: "left 0.7s ease",
                  zIndex: 2
                }}>
                  <span style={{ fontSize: "14px" }}>🚚</span>
                  <span style={{ fontSize: "10px", fontWeight: 700 }}>45%</span>
                </div>

                {/* Armazém Rio de Janeiro (Destino) */}
                <div style={{
                  position: "absolute",
                  right: "20px",
                  top: "40px",
                  backgroundColor: "#FFFFFF",
                  border: "1px solid #0F3822",
                  borderRadius: "8px",
                  padding: "4px 8px",
                  fontSize: "11px",
                  fontWeight: 700,
                  color: "#0F3822",
                  boxShadow: "0 2px 6px rgba(0,0,0,0.1)",
                  display: "flex",
                  alignItems: "center",
                  gap: "4px"
                }}>
                  🏢 2. Rio de Janeiro/RJ
                </div>
              </div>

              {/* Rodapé da Telemetria */}
              <div style={{
                marginTop: "12px",
                display: "flex",
                justifyContent: "space-between",
                fontSize: "11px",
                color: "#4B5563",
                fontWeight: 600,
                borderTop: "1px solid #D5E2D2",
                paddingTop: "8px"
              }}>
                <span>--- Percurso rodoviário</span>
                <span>Armazéns de treinamento</span>
                <span style={{ color: "#0F3822" }}>Caminhão • {truckProgress}%</span>
              </div>
            </div>

            {/* Painel de Abas do Simulador no Card */}
            <div style={{ display: "flex", gap: "8px", marginTop: "16px" }}>
              <button
                onClick={() => setActiveTab("visao")}
                style={{
                  flex: 1,
                  padding: "8px",
                  borderRadius: "8px",
                  border: "none",
                  backgroundColor: activeTab === "visao" ? "#EBF3E6" : "#F3F4F6",
                  color: activeTab === "visao" ? "#0F3822" : "#6B7280",
                  fontWeight: 600,
                  fontSize: "12px",
                  cursor: "pointer"
                }}
              >
                📊 Visão Geral
              </button>
              <button
                onClick={() => setActiveTab("financeiro")}
                style={{
                  flex: 1,
                  padding: "8px",
                  borderRadius: "8px",
                  border: "none",
                  backgroundColor: activeTab === "financeiro" ? "#EBF3E6" : "#F3F4F6",
                  color: activeTab === "financeiro" ? "#0F3822" : "#6B7280",
                  fontWeight: 600,
                  fontSize: "12px",
                  cursor: "pointer"
                }}
              >
                💰 DRE do Frete
              </button>
              <button
                onClick={() => setActiveTab("carga")}
                style={{
                  flex: 1,
                  padding: "8px",
                  borderRadius: "8px",
                  border: "none",
                  backgroundColor: activeTab === "carga" ? "#EBF3E6" : "#F3F4F6",
                  color: activeTab === "carga" ? "#0F3822" : "#6B7280",
                  fontWeight: 600,
                  fontSize: "12px",
                  cursor: "pointer"
                }}
              >
                📦 Carga & Manifesto
              </button>
            </div>

            <div style={{ marginTop: "12px", padding: "12px", backgroundColor: "#F9FAF8", borderRadius: "10px", fontSize: "12px", color: "#374151" }}>
              {activeTab === "visao" && (
                <div style={{ display: "flex", justifyContent: "space-between" }}>
                  <div><strong>Tempo estimado:</strong> 6h 30m</div>
                  <div><strong>Pedágios:</strong> 5 praças</div>
                  <div><strong>Status:</strong> Em rota</div>
                </div>
              )}
              {activeTab === "financeiro" && (
                <div style={{ display: "flex", justifyContent: "space-between" }}>
                  <div><strong>Frete Bruto:</strong> R$ 4.850,00</div>
                  <div><strong>Diesel + Pedágios:</strong> R$ 2.120,00</div>
                  <div style={{ color: "#0F3822" }}><strong>Lucro:</strong> R$ 2.730,00</div>
                </div>
              )}
              {activeTab === "carga" && (
                <div style={{ display: "flex", justifyContent: "space-between" }}>
                  <div><strong>Tipo:</strong> Carga Seca Paletizada</div>
                  <div><strong>Peso:</strong> 18.400 kg</div>
                  <div><strong>Seguro:</strong> Averbação 100%</div>
                </div>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* 4. SEÇÃO DE MÉTRICAS / IMPACTO */}
      <section style={{ backgroundColor: "#0F3822", color: "#FFFFFF", padding: "48px 24px" }}>
        <div style={{ maxWidth: "1280px", margin: "0 auto", display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))", gap: "32px", textAlign: "center" }}>
          <div>
            <div style={{ fontSize: "40px", fontWeight: 900, color: "#68A23B" }}>+12.400</div>
            <div style={{ fontSize: "14px", color: "#C6DFC0", marginTop: "4px" }}>Operações simuladas com sucesso</div>
          </div>
          <div>
            <div style={{ fontSize: "40px", fontWeight: 900, color: "#68A23B" }}>-38%</div>
            <div style={{ fontSize: "14px", color: "#C6DFC0", marginTop: "4px" }}>Redução de custos operacionais na prática</div>
          </div>
          <div>
            <div style={{ fontSize: "40px", fontWeight: 900, color: "#68A23B" }}>40+</div>
            <div style={{ fontSize: "14px", color: "#C6DFC0", marginTop: "4px" }}>Cenários de crise e tomada de decisão</div>
          </div>
          <div>
            <div style={{ fontSize: "40px", fontWeight: 900, color: "#68A23B" }}>100%</div>
            <div style={{ fontSize: "14px", color: "#C6DFC0", marginTop: "4px" }}>Online com certificado profissional</div>
          </div>
        </div>
      </section>

      {/* 5. COMO FUNCIONA A SIMULAÇÃO */}
      <section id="como-funciona" style={{ maxWidth: "1280px", margin: "0 auto", padding: "80px 24px" }}>
        <div style={{ textAlign: "center", marginBottom: "56px" }}>
          <span style={{ fontSize: "13px", fontWeight: 700, color: "#588135", letterSpacing: "1px", textTransform: "uppercase" }}>
            Metodologia Prática 360°
          </span>
          <h2 style={{ fontSize: "36px", fontWeight: 900, color: "#0F3822", marginTop: "8px" }}>
            Como você aprende no TransportSim 360
          </h2>
          <p style={{ color: "#6B7280", maxWidth: "600px", margin: "12px auto 0", fontSize: "16px" }}>
            Sem teorias cansativas. Você assume a cadeira de comando de uma transportadora virtual com clientes reais e dinheiro em jogo.
          </p>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))", gap: "24px" }}>
          {[
            {
              step: "01",
              title: "Cotação & Fechamento Comercial",
              desc: "Receba solicitações de frete, avalie peso, cubagem e precifique para ter lucro real sem perder o cliente para concorrentes."
            },
            {
              step: "02",
              title: "Roteirização & Escalação",
              desc: "Defina o veículo correto (Toco, Truck, Carreta), planeje paradas de descanso conforme a Lei do Motorista e calcule pedágios exatos."
            },
            {
              step: "03",
              title: "Gestão de Crise em Rota",
              desc: "Enfrente panes mecânicas, barreiras de fiscalização, quebra de prazo e tome a decisão certa para salvar a margem."
            },
            {
              step: "04",
              title: "Fechamento & DRE da Frota",
              desc: "Analise custos de diesel, manutenção preventiva, consumo por km e emita o relatório final de rentabilidade da operação."
            }
          ].map((item, idx) => (
            <div key={idx} style={{
              backgroundColor: "#FFFFFF",
              padding: "32px 24px",
              borderRadius: "20px",
              border: "1px solid #E5EADF",
              boxShadow: "0 4px 12px rgba(0,0,0,0.03)"
            }}>
              <div style={{ fontSize: "28px", fontWeight: 900, color: "#588135", marginBottom: "16px" }}>
                {item.step}
              </div>
              <h3 style={{ fontSize: "18px", fontWeight: 800, color: "#0F3822", marginBottom: "10px" }}>
                {item.title}
              </h3>
              <p style={{ fontSize: "14px", lineHeight: 1.6, color: "#4B5563" }}>
                {item.desc}
              </p>
            </div>
          ))}
        </div>
      </section>

      {/* 6. TABELA DE PLANOS & INSCRIÇÃO (CHECKOUT DIRETO) */}
      <section id="preco" style={{ backgroundColor: "#F0F5EC", padding: "80px 24px", borderTop: "1px solid #E1E9DD" }}>
        <div style={{ maxWidth: "900px", margin: "0 auto", textAlign: "center" }}>
          <span style={{ fontSize: "13px", fontWeight: 700, color: "#588135", letterSpacing: "1px", textTransform: "uppercase" }}>
            Investimento no seu Futuro
          </span>
          <h2 style={{ fontSize: "36px", fontWeight: 900, color: "#0F3822", marginTop: "8px" }}>
            Escolha seu plano e comece hoje
          </h2>
          <p style={{ color: "#6B7280", margin: "12px auto 40px", fontSize: "16px" }}>
            Acesso imediato ao simulador, todas as missões e emissão de certificado.
          </p>

          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(320px, 1fr))", gap: "24px", textAlign: "left" }}>
            {/* Plano Anual / Completo (Destaque) */}
            <div style={{
              backgroundColor: "#FFFFFF",
              borderRadius: "24px",
              border: "2px solid #0F3822",
              padding: "36px 32px",
              boxShadow: "0 12px 30px rgba(15, 56, 34, 0.12)",
              position: "relative"
            }}>
              <div style={{
                position: "absolute",
                top: "-14px",
                right: "24px",
                backgroundColor: "#588135",
                color: "#FFFFFF",
                fontSize: "11px",
                fontWeight: 800,
                padding: "4px 12px",
                borderRadius: "12px",
                letterSpacing: "0.5px"
              }}>
                MAIS POPULAR
              </div>

              <h3 style={{ fontSize: "22px", fontWeight: 800, color: "#0F3822" }}>Formação Completa 360°</h3>
              <p style={{ fontSize: "14px", color: "#6B7280", marginTop: "4px" }}>Para quem quer dominar operações de ponta a ponta.</p>

              <div style={{ margin: "28px 0" }}>
                <span style={{ fontSize: "14px", color: "#6B7280" }}>12x de</span>
                <div style={{ fontSize: "44px", fontWeight: 900, color: "#0F3822", lineHeight: 1 }}>
                  R$ 49<span style={{ fontSize: "24px" }}>,90</span>
                </div>
                <span style={{ fontSize: "13px", color: "#588135", fontWeight: 600 }}>ou R$ 497,00 à vista</span>
              </div>

              <ul style={{ listStyle: "none", padding: 0, margin: "0 0 32px", fontSize: "14px", color: "#374151", display: "flex", flexDirection: "column", gap: "12px" }}>
                <li>✓ Acesso total ao simulador de frotas por 1 ano</li>
                <li>✓ Todas as 40 missões de transporte rodoviário</li>
                <li>✓ Simulador de cálculo de frete e DRE financeiro</li>
                <li>✓ Certificado oficial com código de validação</li>
                <li>✓ Suporte direto com especialistas de logística</li>
              </ul>

              <button
                onClick={() => openOrderModal("Matrícula Formação Completa 360° (R$ 497,00)")}
                style={{
                  width: "100%",
                  padding: "16px",
                  backgroundColor: "#0F3822",
                  color: "#FFFFFF",
                  border: "none",
                  borderRadius: "12px",
                  fontSize: "16px",
                  fontWeight: 800,
                  cursor: "pointer",
                  boxShadow: "0 4px 12px rgba(15, 56, 34, 0.2)"
                }}
              >
                Garantir minha vaga agora ➔
              </button>
            </div>

            {/* Plano Pro / Corporativo */}
            <div style={{
              backgroundColor: "#FFFFFF",
              borderRadius: "24px",
              border: "1px solid #E5EADF",
              padding: "36px 32px"
            }}>
              <h3 style={{ fontSize: "22px", fontWeight: 800, color: "#0F3822" }}>TransportSim Empresas</h3>
              <p style={{ fontSize: "14px", color: "#6B7280", marginTop: "4px" }}>Para treinar sua equipe de tráfego e expedição.</p>

              <div style={{ margin: "28px 0" }}>
                <span style={{ fontSize: "14px", color: "#6B7280" }}>Sob consulta</span>
                <div style={{ fontSize: "36px", fontWeight: 900, color: "#0F3822", lineHeight: 1 }}>
                  Licenças para Frotas
                </div>
                <span style={{ fontSize: "13px", color: "#6B7280" }}>Treinamento personalizado</span>
              </div>

              <ul style={{ listStyle: "none", padding: 0, margin: "0 0 32px", fontSize: "14px", color: "#374151", display: "flex", flexDirection: "column", gap: "12px" }}>
                <li>✓ Painel do gestor para acompanhar evolução do time</li>
                <li>✓ Relatório de desempenho de cada colaborador</li>
                <li>✓ Cenários customizados com as rotas da sua empresa</li>
                <li>✓ Treinamento de direção preventiva e consumo</li>
              </ul>

              <button
                onClick={() => openOrderModal("Proposta Corporativa para Frotas")}
                style={{
                  width: "100%",
                  padding: "16px",
                  backgroundColor: "#FFFFFF",
                  color: "#0F3822",
                  border: "2px solid #0F3822",
                  borderRadius: "12px",
                  fontSize: "16px",
                  fontWeight: 800,
                  cursor: "pointer"
                }}
              >
                Falar com consultor corporativo ➔
              </button>
            </div>
          </div>

          <div style={{ marginTop: "24px", fontSize: "13px", color: "#6B7280", display: "flex", alignItems: "center", justifyContent: "center", gap: "8px" }}>
            <span>🔒 Pagamento 100% seguro</span> • <span>7 dias de garantia incondicional</span> • <span>Acesso imediato</span>
          </div>
        </div>
      </section>

      {/* 7. FOOTER */}
      <footer style={{ backgroundColor: "#0B2617", color: "#8FA899", padding: "40px 24px", fontSize: "13px", textAlign: "center" }}>
        <p>© 2026 TransportSim 360. Todos os direitos reservados. Plataforma desenvolvida e mantida por Geevheit Neural Center.</p>
        <p style={{ marginTop: "8px" }}>Tecnologia de ponta para operadores, motoristas e transportadoras de todo o Brasil.</p>
      </footer>

      {/* 8. MODAL INTERATIVO DE PEDIDO / MATRÍCULA (LIGADO À API /v1/intake) */}
      {isModalOpen && (
        <div style={{
          position: "fixed",
          inset: 0,
          backgroundColor: "rgba(11, 38, 23, 0.75)",
          backdropFilter: "blur(6px)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: "16px",
          zIndex: 9999
        }}>
          <div style={{
            backgroundColor: "#FFFFFF",
            borderRadius: "24px",
            maxWidth: "480px",
            width: "100%",
            padding: "32px",
            boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.25)",
            position: "relative"
          }}>
            {/* Botão Fechar */}
            <button
              onClick={() => setIsModalOpen(false)}
              style={{
                position: "absolute",
                top: "20px",
                right: "20px",
                background: "none",
                border: "none",
                fontSize: "20px",
                color: "#6B7280",
                cursor: "pointer",
                padding: "4px"
              }}
            >
              ✕
            </button>

            {submitted ? (
              <div style={{ textAlign: "center", padding: "20px 0" }}>
                <div style={{
                  width: "64px",
                  height: "64px",
                  borderRadius: "50%",
                  backgroundColor: "#EBF3E6",
                  color: "#588135",
                  fontSize: "32px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  margin: "0 auto 16px"
                }}>
                  ✓
                </div>
                <h3 style={{ fontSize: "22px", fontWeight: 800, color: "#0F3822", marginBottom: "8px" }}>
                  Pedido Recebido com Sucesso!
                </h3>
                <p style={{ fontSize: "14px", color: "#4B5563", lineHeight: 1.6, marginBottom: "24px" }}>
                  Sua solicitação foi registrada no sistema. Nossa equipe entrará em contato via WhatsApp/E-mail com os detalhes de acesso e pagamento.
                </p>
                <button
                  onClick={() => setIsModalOpen(false)}
                  style={{
                    padding: "12px 28px",
                    backgroundColor: "#0F3822",
                    color: "#FFFFFF",
                    border: "none",
                    borderRadius: "10px",
                    fontWeight: 700,
                    cursor: "pointer"
                  }}
                >
                  Concluir
                </button>
              </div>
            ) : (
              <form onSubmit={handleOrderSubmit}>
                <div style={{
                  display: "inline-flex",
                  alignItems: "center",
                  gap: "6px",
                  backgroundColor: "#EBF3E6",
                  color: "#3A601E",
                  padding: "4px 10px",
                  borderRadius: "20px",
                  fontSize: "11px",
                  fontWeight: 700,
                  marginBottom: "12px"
                }}>
                  <span>🚚 Pedido Oficial</span>
                </div>

                <h3 style={{ fontSize: "20px", fontWeight: 800, color: "#0F3822", marginBottom: "4px" }}>
                  {modalTitle}
                </h3>
                <p style={{ fontSize: "13px", color: "#6B7280", marginBottom: "20px" }}>
                  Preencha os dados abaixo para receber o link de acesso imediato:
                </p>

                {errorMessage && (
                  <div style={{
                    backgroundColor: "#FEF2F2",
                    color: "#991B1B",
                    padding: "10px 14px",
                    borderRadius: "8px",
                    fontSize: "13px",
                    marginBottom: "16px",
                    border: "1px solid #FCA5A5"
                  }}>
                    {errorMessage}
                  </div>
                )}

                <div style={{ marginBottom: "16px" }}>
                  <label style={{ display: "block", fontSize: "12px", fontWeight: 700, color: "#374151", marginBottom: "6px" }}>
                    Seu Nome Completo *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Ex: Carlos Silva"
                    value={clientName}
                    onChange={(e) => setClientName(e.target.value)}
                    style={{
                      width: "100%",
                      padding: "12px 14px",
                      borderRadius: "10px",
                      border: "1px solid #D1D5DB",
                      fontSize: "14px",
                      outline: "none",
                      boxSizing: "border-box"
                    }}
                  />
                </div>

                <div style={{ marginBottom: "16px" }}>
                  <label style={{ display: "block", fontSize: "12px", fontWeight: 700, color: "#374151", marginBottom: "6px" }}>
                    WhatsApp ou E-mail *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Ex: (11) 98765-4321 ou carlos@email.com"
                    value={contact}
                    onChange={(e) => setContact(e.target.value)}
                    style={{
                      width: "100%",
                      padding: "12px 14px",
                      borderRadius: "10px",
                      border: "1px solid #D1D5DB",
                      fontSize: "14px",
                      outline: "none",
                      boxSizing: "border-box"
                    }}
                  />
                </div>

                <div style={{ marginBottom: "24px" }}>
                  <label style={{ display: "block", fontSize: "12px", fontWeight: 700, color: "#374151", marginBottom: "6px" }}>
                    Empresa ou Cidade (opcional)
                  </label>
                  <input
                    type="text"
                    placeholder="Ex: Transportadora Santos / São Paulo"
                    value={company}
                    onChange={(e) => setCompany(e.target.value)}
                    style={{
                      width: "100%",
                      padding: "12px 14px",
                      borderRadius: "10px",
                      border: "1px solid #D1D5DB",
                      fontSize: "14px",
                      outline: "none",
                      boxSizing: "border-box"
                    }}
                  />
                </div>

                <button
                  type="submit"
                  disabled={submitting}
                  style={{
                    width: "100%",
                    padding: "14px",
                    backgroundColor: submitting ? "#9CA3AF" : "#0F3822",
                    color: "#FFFFFF",
                    border: "none",
                    borderRadius: "12px",
                    fontSize: "15px",
                    fontWeight: 800,
                    cursor: submitting ? "not-allowed" : "pointer",
                    boxShadow: "0 4px 12px rgba(15, 56, 34, 0.2)"
                  }}
                >
                  {submitting ? "Registrando pedido..." : "Confirmar Solicitação / Matrícula ➔"}
                </button>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
