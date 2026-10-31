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
  onExecuteCommand?: (text: string) => Promise<string | void>;
}

export type AssistantPersona = "jarvis" | "jake";

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
  onExecuteCommand,
}: JarvisChatDrawerProps) {
  // 1. Configuração de Persona (Jarvis = Masculino | Jake = Feminina)
  const [persona, setPersona] = useState<AssistantPersona>(() => {
    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("geev_assistant_persona");
      if (stored === "jake" || stored === "jarvis") return stored;
    }
    return "jarvis";
  });

  // Tom e Velocidade da voz
  const [pitch, setPitch] = useState<number>(() => {
    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("geev_assistant_pitch");
      if (stored) return parseFloat(stored);
    }
    return 0.85; // Tom padrão ligeiramente mais grave
  });

  const [rate, setRate] = useState<number>(() => {
    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("geev_assistant_rate");
      if (stored) return parseFloat(stored);
    }
    return 1.05;
  });

  const [selectedVoiceURI, setSelectedVoiceURI] = useState<string>(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("geev_assistant_voice_uri") || "";
    }
    return "";
  });

  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [availableVoices, setAvailableVoices] = useState<SpeechSynthesisVoice[]>([]);

  // Carrega vozes do navegador
  useEffect(() => {
    if (typeof window === "undefined" || !("speechSynthesis" in window)) return;

    const loadVoices = () => {
      const all = window.speechSynthesis.getVoices();
      const pt = all.filter((v) => v.lang.toLowerCase().startsWith("pt"));
      setAvailableVoices(pt.length > 0 ? pt : all);
    };

    loadVoices();
    window.speechSynthesis.onvoiceschanged = loadVoices;
    return () => {
      if (window.speechSynthesis) {
        window.speechSynthesis.onvoiceschanged = null;
      }
    };
  }, []);

  const [messages, setMessages] = useState<ChatMessage[]>(() => {
    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("geev_jarvis_chat_history");
      if (stored) {
        try {
          const list: ChatMessage[] = JSON.parse(stored);
          return list.map((m) => ({
            ...m,
            text: m.text.replace(/Comandante/g, "Senhor").replace(/comandante/g, "senhor"),
          }));
        } catch {}
      }
    }
    return [
      {
        id: "msg_welcome",
        sender: "jarvis",
        text:
          persona === "jake"
            ? "Olá, Senhor. Sou a Jake da GeevheitNeural. Estou monitorando os 5 agentes, faturamento e portões. Como posso te ajudar agora?"
            : "Olá, Senhor. Sou o Jarvis da GeevheitNeural. Estou monitorando os 5 agentes, faturamento e portões. Como posso te ajudar agora?",
        timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
      },
    ];
  });

  const [inputVal, setInputVal] = useState("");
  const [isListening, setIsListening] = useState(false);
  const [isThinking, setIsThinking] = useState(false);
  const [voiceEnabled, setVoiceEnabled] = useState<boolean>(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("geev_voice_feedback_enabled") === "true";
    }
    return false; // Silencioso por padrão, usuário liga quando quiser
  });

  const clearChatHistory = () => {
    if (typeof window !== "undefined") {
      localStorage.removeItem("geev_jarvis_chat_history");
      if ("speechSynthesis" in window) {
        window.speechSynthesis.cancel();
      }
    }
    setMessages([
      {
        id: "msg_welcome_cleared",
        sender: "jarvis",
        text:
          persona === "jake"
            ? "Histórico limpo, Senhor! Sou a Jake. Pronta para receber suas ordens de projetos, aprovações ou finanças."
            : "Histórico limpo, Senhor! Sou o Jarvis. Pronto para receber suas ordens de projetos, aprovações ou finanças.",
        timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
      },
    ]);
  };

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const activeRecognitionRef = useRef<any>(null);
  const isSpeakingRef = useRef<boolean>(false);
  const lastHandledTextRef = useRef<{ text: string; time: number }>({ text: "", time: 0 });

  // Cleanup na desmontagem
  useEffect(() => {
    return () => {
      if (activeRecognitionRef.current) {
        try {
          activeRecognitionRef.current.abort();
        } catch {}
      }
      if (typeof window !== "undefined" && "speechSynthesis" in window) {
        try {
          window.speechSynthesis.cancel();
        } catch {}
      }
    };
  }, []);

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

  // Salva configurações de Persona e Voz
  const savePersonaConfig = (
    newPersona: AssistantPersona,
    newPitch?: number,
    newRate?: number,
    newVoiceURI?: string
  ) => {
    setPersona(newPersona);
    localStorage.setItem("geev_assistant_persona", newPersona);

    const actualPitch =
      newPitch !== undefined ? newPitch : newPersona === "jake" ? 1.15 : 0.85;
    setPitch(actualPitch);
    localStorage.setItem("geev_assistant_pitch", String(actualPitch));

    const actualRate = newRate !== undefined ? newRate : 1.05;
    setRate(actualRate);
    localStorage.setItem("geev_assistant_rate", String(actualRate));

    if (newVoiceURI !== undefined) {
      setSelectedVoiceURI(newVoiceURI);
      localStorage.setItem("geev_assistant_voice_uri", newVoiceURI);
    }
  };

  // Função para falar (Text-to-Speech) com proteção contra eco no microfone
  const speakText = (text: string, customPitch?: number, customRate?: number) => {
    if (!voiceEnabled || typeof window === "undefined" || !("speechSynthesis" in window)) {
      return;
    }
    try {
      // Aborta imediatamente qualquer reconhecimento de voz ativo para impedir que o microfone ouça o próprio alto-falante
      if (activeRecognitionRef.current) {
        try {
          activeRecognitionRef.current.abort();
        } catch {}
        activeRecognitionRef.current = null;
        setIsListening(false);
      }

      window.speechSynthesis.cancel(); // Para falas anteriores
      isSpeakingRef.current = true;

      const utterance = new SpeechSynthesisUtterance(text);
      utterance.lang = "pt-BR";
      utterance.rate = customRate !== undefined ? customRate : rate;
      utterance.pitch = customPitch !== undefined ? customPitch : pitch;

      utterance.onend = () => {
        isSpeakingRef.current = false;
      };
      utterance.onerror = () => {
        isSpeakingRef.current = false;
      };

      // Seleção de voz instalada
      if (selectedVoiceURI) {
        const found = availableVoices.find((v) => v.voiceURI === selectedVoiceURI);
        if (found) utterance.voice = found;
      } else {
        // Seleção inteligente baseada na persona
        if (persona === "jake") {
          const fem = availableVoices.find((v) =>
            /luciana|leticia|francisca|maria|victoria|vitoria|female|feminina/i.test(v.name)
          );
          if (fem) utterance.voice = fem;
        } else {
          const masc = availableVoices.find((v) =>
            /daniel|felipe|ricardo|male|masculin|jorge/i.test(v.name)
          );
          if (masc) utterance.voice = masc;
        }
      }

      window.speechSynthesis.speak(utterance);
    } catch (err) {
      isSpeakingRef.current = false;
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
      const intro =
        persona === "jake"
          ? "Voz ativada, Senhor. Sou a Jake e responderei em áudio."
          : "Voz ativada, Senhor. Sou o Jarvis e responderei em áudio.";
      speakText(intro);
    } else {
      if (typeof window !== "undefined" && "speechSynthesis" in window) {
        window.speechSynthesis.cancel();
      }
    }
  };

  const testVoiceSample = () => {
    const sample =
      persona === "jake"
        ? "Olá, Senhor! Sou a Jake. Estou pronta para monitorar a fábrica de software e aprovar portões com o senhor."
        : "Olá, Senhor! Sou o Jarvis. Todos os sistemas e agentes estão operacionais sob suas diretrizes.";
    speakText(sample);
  };

  const formatBRL = (cents: number) => {
    return (cents / 100).toLocaleString("pt-BR", {
      style: "currency",
      currency: "BRL",
      minimumFractionDigits: 0,
    });
  };

  // Motor de Inteligência e Resposta
  const processAssistantResponse = (userPrompt: string): string => {
    const q = userPrompt
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "");

    const pendingGateCards = cards.filter((c) =>
      ["proposta_aprovar", "prompt_aprovar", "entrega_aprovar"].includes(c.stage)
    );
    const totalMoney = pendingGateCards.reduce((acc, c) => acc + c.price_cents, 0);
    const assistantName = persona === "jake" ? "Jake" : "Jarvis";

    // 0. Dúvidas sobre Especificações Extras / Como detalhar demandas
    if (/especific|detalhe|requisito|mais campo|regras|complex/.test(q)) {
      return `Para adicionar especificações técnicas extras a qualquer projeto: clique no cartão da demanda no quadro. Na aba de detalhes, você pode escrever o escopo completo, regras de negócio e critérios BDD. No Portão 2 (Prompt), o Arquiteto lerá essas regras para gerar o código sob medida.`;
    }

    // 1. Pergunta sobre Status / Resumo da Software House
    if (/status|resumo|visao|como esta|painel/.test(q)) {
      return `Relatório da Software House: Temos ${apps.length} aplicativo(s) ativos e ${cards.length} demanda(s) no fluxo. Há ${formatBRL(totalMoney)} parados em ${pendingGateCards.length} decisão(ões) de portão. O faturamento mensal previsto é de ${formatBRL(finances.total_revenue_cents)} com margem líquida de ${formatBRL(finances.net_profit_cents)}. Os agentes estão ${isPaused ? "pausados" : "operando normalmente sob o motor Groq/Gemini"}.`;
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
      return `Os 5 agentes estão preparados com os motores Groq e Google Gemini. Temos ${activeDev} demanda(s) em execução no sandbox isolado.`;
    }

    // 4. Pausar ou Retomar
    if (/pausar|parar|congelar/.test(q)) {
      if (isPaused) {
        return "O sistema já está pausado pelo ExecutionGuard.";
      }
      onTogglePause();
      return `Ordem recebida pelo ${assistantName}: Sistema pausado imediatamente. Todos os agentes foram congelados.`;
    }

    if (/retomar|continuar|despausar|liberar/.test(q)) {
      if (!isPaused) {
        return "O sistema já está em operação ativa.";
      }
      onTogglePause();
      return `Ordem recebida pelo ${assistantName}: Sistema retomado. Os agentes estão prontos para prosseguir.`;
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
    if (/criar|novo projeto|nova demanda|novo app|cadastrar|fazer/.test(q)) {
      onOpenNewCardModal();
      return "Abrindo o modal de cadastro de nova demanda na tela para você preencher os dados.";
    }

    // 7. Identidade / Nome
    if (/quem e voce|seu nome|jarvis|jake/.test(q)) {
      return `Sou ${assistantName}, a inteligência operacional da GeevheitNeural. Você pode alternar entre Jarvis (voz masculina) e Jake (voz feminina) clicando no ícone de engrenagem no cabeçalho.`;
    }

    // 8. Resposta Geral
    return "Compreendido, Senhor! Estou pronto para agir. Você pode me pedir para criar uma demanda (ex: 'Criar app de entregas'), construir uma solução com a fábrica autônoma, aprovar um portão ou excluir um projeto antigo.";
  };

  const handleSend = async (text: string) => {
    const trimmed = text.trim();
    if (!trimmed) return;

    // Proteção contra duplicação/eco: se o mesmo texto foi enviado a menos de 2 segundos, ignora
    const now = Date.now();
    const cleanLower = trimmed.toLowerCase();
    if (
      lastHandledTextRef.current.text === cleanLower &&
      now - lastHandledTextRef.current.time < 2000
    ) {
      return;
    }
    lastHandledTextRef.current = { text: cleanLower, time: now };

    // Se estiver falando, cala imediatamente
    if (typeof window !== "undefined" && "speechSynthesis" in window) {
      try {
        window.speechSynthesis.cancel();
      } catch {}
      isSpeakingRef.current = false;
    }

    const userMsg: ChatMessage = {
      id: `usr_${Date.now()}`,
      sender: "user",
      text: trimmed,
      timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
    };

    setMessages((prev) => [...prev, userMsg]);
    setInputVal("");
    setIsThinking(true);

    const q = trimmed.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");

    // 1. Respostas conversacionais imediatas de identidade
    if (/quem e voce|seu nome/.test(q)) {
      setIsThinking(false);
      const assistantName = persona === "jake" ? "Jake" : "Jarvis";
      const reply = `Sou ${assistantName}, a inteligência operacional da GeevheitNeural. Você pode alternar entre Jarvis (voz masculina) e Jake (voz feminina) clicando no ícone de engrenagem no cabeçalho.`;
      const assistantMsg: ChatMessage = {
        id: `ast_${Date.now()}`,
        sender: "jarvis",
        text: reply,
        timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
      };
      setMessages((prev) => [...prev, assistantMsg]);
      speakText(reply);
      return;
    }

    // 2. Execução de comandos reais através da central
    if (onExecuteCommand) {
      try {
        const cmdResult = await onExecuteCommand(trimmed);
        setIsThinking(false);
        if (cmdResult && typeof cmdResult === "string") {
          const assistantMsg: ChatMessage = {
            id: `ast_${Date.now()}`,
            sender: "jarvis",
            text: cmdResult,
            timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
          };
          setMessages((prev) => [...prev, assistantMsg]);
          speakText(cmdResult);
          return;
        }
      } catch (err: unknown) {
        setIsThinking(false);
        const errorMsg = (err as Error)?.message || "Não foi possível executar a ordem solicitada.";
        const assistantMsg: ChatMessage = {
          id: `ast_${Date.now()}`,
          sender: "jarvis",
          text: `Aviso operacional: ${errorMsg}`,
          timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
        };
        setMessages((prev) => [...prev, assistantMsg]);
        speakText(`Aviso operacional: ${errorMsg}`);
        return;
      }
    }

    // 3. Fallback inteligente local
    setTimeout(() => {
      setIsThinking(false);
      const reply = processAssistantResponse(trimmed);
      const assistantMsg: ChatMessage = {
        id: `ast_${Date.now()}`,
        sender: "jarvis",
        text: reply,
        timestamp: new Date().toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" }),
      };
      setMessages((prev) => [...prev, assistantMsg]);
      speakText(reply);
    }, 400);
  };

  // Reconhecimento de Voz nativo com cancelamento de eco
  const handleMicClick = () => {
    interface SpeechEvent {
      results: { [key: number]: { [key: number]: { transcript: string } } };
    }
    const win = window as unknown as {
      SpeechRecognition?: new () => any;
      webkitSpeechRecognition?: new () => any;
    };

    const SpeechRec = win.SpeechRecognition || win.webkitSpeechRecognition;
    if (!SpeechRec) {
      alert("Reconhecimento de voz não suportado neste navegador. Use o teclado.");
      return;
    }

    // Se já estiver ouvindo, clicar novamente desliga o microfone
    if (activeRecognitionRef.current) {
      try {
        activeRecognitionRef.current.abort();
      } catch {}
      activeRecognitionRef.current = null;
      setIsListening(false);
      return;
    }

    // Se o sintetizador estiver falando, interrompe a fala imediatamente ao ligar o microfone
    if (typeof window !== "undefined" && "speechSynthesis" in window) {
      try {
        window.speechSynthesis.cancel();
      } catch {}
      isSpeakingRef.current = false;
    }

    try {
      const recognition = new SpeechRec();
      recognition.lang = "pt-BR";
      recognition.interimResults = false;
      recognition.maxAlternatives = 1;
      activeRecognitionRef.current = recognition;

      let hasHandledResult = false;

      recognition.onstart = () => {
        setIsListening(true);
      };

      recognition.onresult = (e: SpeechEvent) => {
        if (hasHandledResult || activeRecognitionRef.current !== recognition || isSpeakingRef.current) return;
        hasHandledResult = true;

        const transcript = e.results?.[0]?.[0]?.transcript;

        // Fecha o microfone imediatamente no primeiro resultado final para evitar captação de eco
        try {
          recognition.stop();
        } catch {}
        activeRecognitionRef.current = null;
        setIsListening(false);

        if (transcript && transcript.trim()) {
          handleSend(transcript.trim());
        }
      };

      recognition.onerror = () => {
        if (activeRecognitionRef.current !== recognition) return;
        activeRecognitionRef.current = null;
        setIsListening(false);
      };

      recognition.onend = () => {
        if (activeRecognitionRef.current !== recognition) return;
        activeRecognitionRef.current = null;
        setIsListening(false);
      };

      recognition.start();
    } catch {
      activeRecognitionRef.current = null;
      setIsListening(false);
    }
  };

  if (!isOpen) return null;

  const isJake = persona === "jake";

  return (
    <div
      id="modal-jarvis-chat"
      className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex justify-end"
    >
      <div className="w-full max-w-md bg-[var(--bg)] border-l border-[var(--line)] h-full flex flex-col shadow-2xl animate-slideLeft">
        {/* Cabeçalho do Chat */}
        <div className="p-3.5 bg-[var(--surf)] border-b border-[var(--line)] flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div
              className={`w-9 h-9 rounded-xl flex items-center justify-center font-bold text-base shadow-sm ${
                isJake
                  ? "bg-gradient-to-tr from-purple-500 to-pink-500 text-white"
                  : "bg-gradient-to-tr from-[var(--acc)] to-teal-400 text-white"
              }`}
            >
              {isJake ? "✨" : "🧠"}
            </div>
            <div>
              <div className="flex items-center gap-1.5">
                <h3 className="text-sm font-bold text-[var(--ink)] m-0 leading-tight">
                  {isJake ? "Jake Neural" : "Jarvis Neural"}
                </h3>
                <span className="w-2 h-2 rounded-full bg-[var(--ok)] inline-block animate-pulse"></span>
              </div>
              <div className="text-[11px] text-[var(--mute)]">
                {isJake ? "Assistente Feminina" : "Assistente Masculino"} • Fábrica de Software
              </div>
            </div>
          </div>

          <div className="flex items-center gap-1.5">
            {/* Botão de Configurações de Voz / Persona */}
            <button
              id="btn-assistant-settings"
              onClick={() => setIsSettingsOpen(!isSettingsOpen)}
              title="Configurar Voz e Persona (Jarvis ou Jake)"
              className={`p-1.5 px-2 rounded-lg text-xs font-semibold border cursor-pointer transition-all flex items-center gap-1 ${
                isSettingsOpen
                  ? "bg-[var(--acc)] text-white border-[var(--acc)]"
                  : "bg-[var(--bg)] border-[var(--line)] text-[var(--ink)] hover:border-[var(--acc)]"
              }`}
            >
              <span>⚙️</span>
              <span className="text-[10px] hidden sm:inline">Voz</span>
            </button>

            {/* Botão de Alternar Som On/Off */}
            <button
              id="btn-toggle-voice"
              onClick={toggleVoice}
              title={voiceEnabled ? "Desativar áudio falado" : "Ativar áudio falado"}
              className={`p-1.5 px-2 rounded-lg text-xs font-semibold border cursor-pointer transition-all flex items-center gap-1 ${
                voiceEnabled
                  ? "bg-[var(--accs)] border-[var(--acc)] text-[var(--acc)] shadow-xs"
                  : "bg-[var(--bg)] border-[var(--line)] text-[var(--mute)] hover:text-[var(--ink)]"
              }`}
            >
              <span>{voiceEnabled ? "🔊" : "🔇"}</span>
            </button>

            {/* Botão Limpar Histórico do Chat */}
            <button
              id="btn-clear-chat"
              type="button"
              onClick={clearChatHistory}
              title="Limpar mensagens e zerar conversa do assistente"
              className="p-1.5 px-2 rounded-lg text-xs font-semibold border border-[var(--line)] bg-[var(--bg)] text-[var(--mute)] hover:text-red-500 hover:border-red-500/40 cursor-pointer transition-all flex items-center gap-1"
            >
              <span>🧹</span>
              <span className="text-[10px] hidden sm:inline">Limpar</span>
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

        {/* Painel Expansível de Configurações do Assistente (Jarvis / Jake) */}
        {isSettingsOpen && (
          <div className="p-4 bg-[var(--surf)] border-b border-[var(--line)] space-y-3.5 animate-fadeIn">
            <div className="flex items-center justify-between">
              <h4 className="text-xs font-bold text-[var(--ink)] m-0 flex items-center gap-1.5">
                <span>⚙️</span>
                <span>Configurar Assistente de Voz</span>
              </h4>
              <button
                onClick={() => setIsSettingsOpen(false)}
                className="text-[11px] text-[var(--mute)] hover:text-[var(--ink)] bg-transparent border-0 cursor-pointer"
              >
                Concluir
              </button>
            </div>

            {/* Seletor de Persona: Jarvis (Masculino) vs Jake (Feminina) */}
            <div>
              <label className="text-[11px] font-semibold text-[var(--ink)] block mb-1.5">
                Identidade e Persona:
              </label>
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => savePersonaConfig("jarvis", 0.85, 1.05)}
                  className={`p-2 rounded-xl border text-xs font-bold flex items-center justify-center gap-2 cursor-pointer transition-all ${
                    !isJake
                      ? "bg-[var(--acc)] text-white border-[var(--acc)] shadow-sm"
                      : "bg-[var(--bg)] border-[var(--line)] text-[var(--mute)] hover:text-[var(--ink)]"
                  }`}
                >
                  <span>🧠</span>
                  <span>Jarvis (Masculino)</span>
                </button>

                <button
                  type="button"
                  onClick={() => savePersonaConfig("jake", 1.15, 1.05)}
                  className={`p-2 rounded-xl border text-xs font-bold flex items-center justify-center gap-2 cursor-pointer transition-all ${
                    isJake
                      ? "bg-gradient-to-r from-purple-500 to-pink-500 text-white border-transparent shadow-sm"
                      : "bg-[var(--bg)] border-[var(--line)] text-[var(--mute)] hover:text-[var(--ink)]"
                  }`}
                >
                  <span>✨</span>
                  <span>Jake (Feminina)</span>
                </button>
              </div>
            </div>

            {/* Seletor de Vozes instaladas no SO */}
            {availableVoices.length > 0 && (
              <div>
                <label className="text-[11px] font-semibold text-[var(--ink)] block mb-1">
                  Voz do Dispositivo:
                </label>
                <select
                  value={selectedVoiceURI}
                  onChange={(e) => savePersonaConfig(persona, pitch, rate, e.target.value)}
                  className="w-full text-xs p-2 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none"
                >
                  <option value="">Automático da Persona ({isJake ? "Feminina" : "Masculina"})</option>
                  {availableVoices.map((v) => (
                    <option key={v.voiceURI} value={v.voiceURI}>
                      {v.name} ({v.lang})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Controles de Tom (Pitch) e Velocidade (Rate) */}
            <div className="grid grid-cols-2 gap-3 text-xs">
              <div>
                <div className="flex justify-between text-[10px] text-[var(--mute)] mb-1">
                  <span>Tom da Voz</span>
                  <span>{pitch < 1 ? "Grave" : pitch > 1 ? "Agudo" : "Médio"}</span>
                </div>
                <input
                  type="range"
                  min="0.5"
                  max="1.5"
                  step="0.05"
                  value={pitch}
                  onChange={(e) => savePersonaConfig(persona, parseFloat(e.target.value), rate)}
                  className="w-full accent-[var(--acc)] cursor-pointer"
                />
              </div>

              <div>
                <div className="flex justify-between text-[10px] text-[var(--mute)] mb-1">
                  <span>Velocidade</span>
                  <span>{rate.toFixed(2)}x</span>
                </div>
                <input
                  type="range"
                  min="0.8"
                  max="1.4"
                  step="0.05"
                  value={rate}
                  onChange={(e) => savePersonaConfig(persona, pitch, parseFloat(e.target.value))}
                  className="w-full accent-[var(--acc)] cursor-pointer"
                />
              </div>
            </div>

            {/* Botão de Teste de Voz */}
            <div className="pt-1 flex items-center justify-between">
              <span className="text-[10px] text-[var(--mute)]">
                {isJake ? "Jake usará tom feminino" : "Jarvis usará tom masculino"}
              </span>
              <button
                type="button"
                onClick={testVoiceSample}
                className="text-xs font-semibold px-3 py-1.5 rounded-lg bg-[var(--bg)] border border-[var(--line)] text-[var(--ink)] hover:border-[var(--acc)] cursor-pointer flex items-center gap-1.5"
              >
                <span>🔊</span>
                <span>Ouvir Voz de Teste</span>
              </button>
            </div>
          </div>
        )}

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
          {isThinking && (
            <div className="flex flex-col items-start animate-pulse">
              <div className="bg-[var(--surf)] border border-[var(--acc)]/40 text-[var(--ink)] rounded-2xl p-3 text-xs flex items-center gap-2 shadow-xs">
                <span className="w-2 h-2 rounded-full bg-[var(--acc)] animate-ping"></span>
                <span>{isJake ? "Jake" : "Jarvis"} executando sua ordem na central...</span>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Sugestões Rápidas de Ordens */}
        <div className="p-2 px-3 border-t border-[var(--line)] bg-[var(--surf)]/50 overflow-x-auto flex gap-1.5 no-scrollbar">
          <button
            onClick={() => handleSend("Abrir laboratório de IA")}
            className="text-[11px] whitespace-nowrap bg-purple-500/10 border border-purple-500/30 hover:bg-purple-500/20 rounded-full px-2.5 py-1 text-purple-600 dark:text-purple-400 font-semibold cursor-pointer"
          >
            🧪 Laboratório IA
          </button>
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
            onClick={() => handleSend("Como adiciono especificações extras nos cartões?")}
            className="text-[11px] whitespace-nowrap bg-[var(--bg)] border border-[var(--line)] hover:border-[var(--acc)] rounded-full px-2.5 py-1 text-[var(--ink)] cursor-pointer"
          >
            📝 Especificações
          </button>
          <button
            onClick={() => handleSend("Excluir dados legados e zerar painel")}
            className="text-[11px] whitespace-nowrap bg-[var(--bg)] border border-red-500/40 hover:bg-red-500/10 rounded-full px-2.5 py-1 text-red-600 dark:text-red-400 cursor-pointer"
          >
            🧹 Zerar Painel
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
            disabled={isThinking}
            onChange={(e) => setInputVal(e.target.value)}
            placeholder={
              isThinking
                ? "Executando ordem..."
                : `Fale com ${isJake ? "a Jake" : "o Jarvis"} ou dite uma ordem...`
            }
            className="flex-1 text-xs sm:text-sm py-2 px-3.5 rounded-xl border border-[var(--line)] bg-[var(--bg)] text-[var(--ink)] outline-none focus:border-[var(--acc)] disabled:opacity-60"
          />

          <button
            type="button"
            onClick={handleMicClick}
            disabled={isThinking}
            title="Ditar por voz"
            className={`p-2 rounded-xl text-xs font-bold border-0 cursor-pointer flex-none transition-all ${
              isListening
                ? "bg-[var(--bad)] text-white animate-pulse"
                : "bg-[var(--accs)] text-[var(--acc)] hover:opacity-90 disabled:opacity-50"
            }`}
          >
            {isListening ? "Gravando..." : "🎙️"}
          </button>

          <button
            type="submit"
            disabled={isThinking || !inputVal.trim()}
            className="py-2 px-3.5 rounded-xl text-xs font-bold bg-[var(--acc)] text-white border-0 cursor-pointer hover:opacity-90 transition-opacity flex-none disabled:opacity-50"
          >
            {isThinking ? "..." : "Enviar"}
          </button>
        </form>
      </div>
    </div>
  );
}
