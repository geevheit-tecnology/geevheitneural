"use client";

import React, { useState, useEffect, useRef } from "react";
import { CardItem, AppItem, FinancesOverview } from "@/shared/api-client/endpoints";

export interface ChatMessage {
  id: string;
  sender: "user" | "jarvis";
  text: string;
  timestamp: string;
}

interface JarvisChatDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  cards: CardItem[];
  apps: AppItem[];
  finances: FinancesOverview;
  isPaused: boolean;
  onApproveCard: (cardId: string) => void;
  onTogglePause: () => void;
  onOpenNewCardModal: () => void;
}

export function JarvisChatDrawer({
  isOpen,
  onClose,
  cards,
  apps,
  finances,
  isPaused,
  onApproveCard,
  onTogglePause,
  onOpenNewCardModal,
}: JarvisChatDrawerProps) {
  const [messages, setMessages] = useState<ChatMessage[]>(() => {
    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("geev_jarvis_chat_history");
      if (stored) {
        try {
          return JSON.parse(stored);
        } catch {}
      }
    }
    return [
      {
        id: "msg_welcome",
        sender: "jarvis",
        text: "Olá, Comandante. Sou o Jarvis da GeevheitNeural. Estou monitorando os 5 agentes, faturamento e portões. Como posso te ajudar agora?",
        timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
      },
    ];
  });

  const [inputVal, setInputVal] = useState("");
  const [isListening, setIsListening] = useState(false);
  const [voiceEnabled, setVoiceEnabled] = useState<boolean>(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("geev_voice_feedback_enabled") === "true";
    }
    return false; // Silencioso por padrão, usuário liga quando quiser
  });

  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll para última mensagem
  useEffect(() => {
    if (isOpen) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, isOpen]);

  // Salva histórico no localStorage
  useEffect(() => {
    if (typeof window !== "undefined") {
      try {
        localStorage.setItem("geev_jarvis_chat_history", JSON.stringify(messages));
      } catch {}
    }
  }, [messages]);

  // Função para o Jarvis falar (Text-to-Speech)
  const speakText = (text: string) => {
    if (!voiceEnabled || typeof window === "undefined" || !("speechSynthesis" in window)) {
      return;
    }
    try {
      window.speechSynthesis.cancel(); // Para falas anteriores
      const utterance = new SpeechSynthesisUtterance(text);
      utterance.lang = "pt-BR";
      utterance.rate = 1.05;
      utterance.pitch = 0.95;
      window.speechSynthesis.speak(utterance);
    } catch (err) {
      console.warn("Falha na síntese de voz:", err);
    }
  };

  const toggleVoice = () => {
    const nextVal = !voiceEnabled;
    setVoiceEnabled(nextVal);
    if (typeof window !== "undefined") {
      localStorage.setItem("geev_voice_feedback_enabled", String(nextVal));
    }
    if (nextVal) {
      speakText("Voz ativada, Comandante. Responderei em áudio e texto.");
    } else {
      if (typeof window !== "undefined" && "speechSynthesis" in window) {
        window.speechSynthesis.cancel();
      }
    }
  };

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  // Motor de Inteligência e Resposta do Jarvis
  const processJarvisResponse = (userPrompt: string): string => {
    const q = userPrompt
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "");

    const pendingGateCards = cards.filter((c) =>
      ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage)
    );
    const totalMoney = pendingGateCards.reduce((acc, c) => acc + c.price_cents, 0);

    // 1. Pergunta sobre Status / Resumo da Software House
    if (/status|resumo|visao|como esta|painel/.test(q)) {
      return `Relatório da Software House: Temos ${apps.length} aplicativo(s) ativos e ${cards.length} demanda(s) no fluxo. Há ${formatBRL(totalMoney)} parados em ${pendingGateCards.length} decisão(ões) de portão. O faturamento mensal previsto é de ${formatBRL(finances.total_revenue_cents)} com margem líquida de ${formatBRL(finances.net_profit_cents)}. Os 5 agentes estão ${isPaused ? "pausados" : "operando normalmente"}.`;
    }

    // 2. Dinheiro Parado
    if (/dinheiro|quanto|valor|parado|esperando/.test(q)) {
      if (pendingGateCards.length === 0) {
        return "Não há nenhum centavo parado em portões no momento. Todas as propostas e entregas estão em dia!";
      }
      return `Temos exatamente ${formatBRL(totalMoney)} aguardando sua decisão em ${pendingGateCards.length} cartão(ões). Se você aprovar, o sinal é cobrado ou a entrega é liberada.`;
    }

    // 3. Status dos Agentes de IA
    if (/agente|ia|robo|fabrica|dev|qa|arquiteto/.test(q)) {
      if (isPaused) {
        return "Os 5 agentes (Comercial, Arquiteto, Dev, QA, DevOps) estão PAUSADOS pelo botão de emergência. Nenhuma execução externa ocorrerá até você retomar o sistema.";
      }
      const activeDev = cards.filter((c) => c.stage === "em_execucao").length;
      return `Os 5 agentes estão preparados no Google AI Studio (Gemini 1.5 Flash Free Tier com taxa de 15 RPM). Temos ${activeDev} demanda(s) em execução autônoma no sandbox isolado.`;
    }

    // 4. Pausar ou Retomar
    if (/pausar|parar|congelar/.test(q)) {
      if (isPaused) {
        return "O sistema já está pausado pelo ExecutionGuard.";
      }
      onTogglePause();
      return "Ordem recebida: Sistema pausado imediatamente. Todos os agentes foram congelados.";
    }

    if (/retomar|continuar|despausar|liberar/.test(q)) {
      if (!isPaused) {
        return "O sistema já está em operação ativa.";
      }
      onTogglePause();
      return "Ordem recebida: Sistema retomado. Os agentes estão prontos para prosseguir sob suas diretrizes.";
    }

    // 5. Ordem de Aprovação
    if (/aprovar|aprova|libera/.test(q)) {
      const match = cards.find(
        (c) =>
          ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage) &&
          (q.includes(c.client_name.toLowerCase()) || q.includes(c.title.toLowerCase()))
      );

      if (match) {
        onApproveCard(match.id);
        return `Ordem executada com sucesso: Aprovei o portão de ${match.title} do cliente ${match.client_name}!`;
      }

      if (pendingGateCards.length === 1) {
        onApproveCard(pendingGateCards[0].id);
        return `Aprovei a única decisão pendente: ${pendingGateCards[0].title} de ${pendingGateCards[0].client_name}!`;
      }

      if (pendingGateCards.length > 1) {
        return `Encontrei ${pendingGateCards.length} portões aguardando aprovação. Diga o nome do cliente ou do projeto para eu aprovar com segurança.`;
      }

      return "Não há nenhum portão pendente de aprovação neste momento.";
    }

    // 6. Cadastrar Novo Projeto / Demanda
    if (/criar|novo projeto|nova demanda|novo app|cadastrar/.test(q)) {
      onOpenNewCardModal();
      return "Abrindo o modal de cadastro de nova demanda na tela para você preencher os dados.";
    }

    // 7. Resposta Geral Inteligente do Jarvis
    return `Compreendido. Sua ordem "${userPrompt}" foi registrada no histórico. Como comandante da central, você pode me pedir resumos de faturamento, consultar agentes, pausar o sistema ou aprovar portões pelo nome.`;
  };

  const handleSend = (text: string) => {
    if (!text.trim()) return;

    const userMsg: ChatMessage = {
      id: `usr_${Date.now()}`,
      sender: "user",
      text: text.trim(),
      timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
    };

    setMessages((prev) => [...prev, userMsg]);
    setInputVal("");

    // Resposta do Jarvis com leve atraso realista
    setTimeout(() => {
      const reply = processJarvisResponse(text);
      const jarvisMsg: ChatMessage = {
        id: `jrv_${Date.now()}`,
        sender: "jarvis",
        text: reply,
        timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
      };
      setMessages((prev) => [...prev, jarvisMsg]);
      speakText(reply);
    }, 400);
  };

  // Reconhecimento de Voz nativo
  const handleMicClick = () => {
    interface SpeechEvent {
      results: { [key: number]: { [key: number]: { transcript: string } } };
    }
    const win = window as unknown as {
      SpeechRecognition?: new () => {
        lang: string;
        interimResults: boolean;
        onstart: () => void;
        onresult: (e: SpeechEvent) => void;
        onerror: () => void;
        onend: () => void;
        start: () => void;
      };
      webkitSpeechRecognition?: new () => {
        lang: string;
        interimResults: boolean;
        onstart: () => void;
        onresult: (e: SpeechEvent) => void;
        onerror: () => void;
        onend: () => void;
        start: () => void;
      };
    };

    const SpeechRecognition = win.SpeechRecognition || win.webkitSpeechRecognition;
    if (!SpeechRecognition) {
      alert("Seu navegador não suporta reconhecimento de voz direto. Use o teclado.");
      return;
    }

    try {
      const recognition = new SpeechRecognition();
      recognition.lang = "pt-BR";
      recognition.interimResults = false;

      recognition.onstart = () => {
        setIsListening(true);
      };

      recognition.onresult = (e: SpeechEvent) => {
        const transcript = e.results[0][0].transcript;
        setIsListening(false);
        handleSend(transcript);
      };

      recognition.onerror = () => setIsListening(false);
      recognition.onend = () => setIsListening(false);

      recognition.start();
    } catch {
      setIsListening(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div
      id="modal-jarvis-chat"
      className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex justify-end"
    >
      <div className="w-full max-w-md bg-[var(--bg)] border-l border-[var(--line)] h-full flex flex-col shadow-2xl animate-slideLeft">
        {/* Cabeçalho do Chat com Controle de Voz */}
        <div className="p-3.5 bg-[var(--surf)] border-b border-[var(--line)] flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-[var(--acc)] to-teal-400 text-white flex items-center justify-center font-bold text-base shadow-sm">
              🧠
            </div>
            <div>
              <div className="flex items-center gap-1.5">
                <h3 className="text-sm font-bold text-[var(--ink)] m-0 leading-tight">
                  Jarvis Neural
                </h3>
                <span className="w-2 h-2 rounded-full bg-[var(--ok)] inline-block animate-pulse"></span>
              </div>
              <div className="text-[11px] text-[var(--mute)]">Fábrica de Software Autônoma</div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Botão de Alternar Voz (Audio On/Off) */}
            <button
              id="btn-toggle-voice"
              onClick={toggleVoice}
              title={voiceEnabled ? "Desativar voz (ficar silencioso)" : "Ativar voz do Jarvis"}
              className={`p-1.5 px-2 rounded-lg text-xs font-semibold border cursor-pointer transition-all flex items-center gap-1 ${
                voiceEnabled
                  ? "bg-[var(--accs)] border-[var(--acc)] text-[var(--acc)] shadow-xs"
                  : "bg-[var(--bg)] border-[var(--line)] text-[var(--mute)] hover:text-[var(--ink)]"
              }`}
            >
              <span>{voiceEnabled ? "🔊" : "🔇"}</span>
              <span className="text-[10px]">{voiceEnabled ? "Voz Ligada" : "Silencioso"}</span>
            </button>

            {/* Fechar Gaveta */}
            <button
              onClick={onClose}
              className="p-1 text-base text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 cursor-pointer"
            >
              ✕
            </button>
          </div>
        </div>

        {/* Mensagens de Chat */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {messages.map((m) => (
            <div
              key={m.id}
              className={`flex flex-col ${m.sender === "user" ? "items-end" : "items-start"}`}
            >
              <div
                className={`max-w-[85%] rounded-2xl p-3 text-xs sm:text-sm leading-relaxed shadow-xs ${
                  m.sender === "user"
                    ? "bg-[var(--acc)] text-white rounded-br-xs"
                    : "bg-[var(--surf)] border border-[var(--line)] text-[var(--ink)] rounded-bl-xs"
                }`}
              >
                {m.text}
              </div>
              <span className="text-[10px] text-[var(--mute)] mt-1 px-1">{m.timestamp}</span>
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        {/* Sugestões Rápidas de Ordens */}
        <div className="p-2 px-3 border-t border-[var(--line)] bg-[var(--surf)]/50 overflow-x-auto flex gap-1.5 no-scrollbar">
          <button
            onClick={() => handleSend("Status da software house")}
            className="text-[11px] whitespace-nowrap bg-[var(--bg)] border border-[var(--line)] hover:border-[var(--acc)] rounded-full px-2.5 py-1 text-[var(--ink)] cursor-pointer"
          >
            📊 Resumo Geral
          </button>
          <button
            onClick={() => handleSend("Quanto dinheiro parado esperando aprovação?")}
            className="text-[11px] whitespace-nowrap bg-[var(--bg)] border border-[var(--line)] hover:border-[var(--acc)] rounded-full px-2.5 py-1 text-[var(--ink)] cursor-pointer"
          >
            💰 Dinheiro Parado
          </button>
          <button
            onClick={() => handleSend("Qual o status dos 5 agentes?")}
            className="text-[11px] whitespace-nowrap bg-[var(--bg)] border border-[var(--line)] hover:border-[var(--acc)] rounded-full px-2.5 py-1 text-[var(--ink)] cursor-pointer"
          >
            🤖 Status dos Agentes
          </button>
          <button
            onClick={() => handleSend(isPaused ? "Retomar sistema" : "Pausar tudo agora")}
            className="text-[11px] whitespace-nowrap bg-[var(--bg)] border border-[var(--line)] hover:border-[var(--bad)] rounded-full px-2.5 py-1 text-[var(--bad)] cursor-pointer"
          >
            {isPaused ? "▶️ Retomar" : "🛑 Parar Tudo"}
          </button>
        </div>

        {/* Input de Envio com Microfone */}
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSend(inputVal);
          }}
          className="p-3 bg-[var(--surf)] border-t border-[var(--line)] flex items-center gap-2"
        >
          <input
            type="text"
            value={inputVal}
            onChange={(e) => setInputVal(e.target.value)}
            placeholder="Pergunte ao Jarvis ou dite uma ordem..."
            className="flex-1 text-xs sm:text-sm py-2 px-3.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none focus:border-[var(--acc)]"
          />

          <button
            type="button"
            onClick={handleMicClick}
            title="Ditar por voz"
            className={`p-2 rounded-xl text-xs font-bold border-0 cursor-pointer flex-none transition-all ${
              isListening
                ? "bg-[var(--bad)] text-white animate-pulse"
                : "bg-[var(--accs)] text-[var(--acc)] hover:opacity-90"
            }`}
          >
            {isListening ? "Gravando..." : "🎙️"}
          </button>

          <button
            type="submit"
            className="py-2 px-3.5 rounded-xl text-xs font-bold bg-[var(--acc)] text-white border-0 cursor-pointer hover:opacity-90 transition-opacity flex-none"
          >
            Enviar
          </button>
        </form>
      </div>
    </div>
  );
}
