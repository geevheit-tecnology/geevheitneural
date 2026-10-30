"use client";

import React, { useState, useRef, useEffect } from "react";
import { CardItem } from "@/shared/api-client/endpoints";

interface CommandLineProps {
  cards: CardItem[];
  onApprove: (cardId: string) => void;
  onReject: (cardId: string) => void;
  onShowToast: (msg: string) => void;
  moneySummary: string;
  onExecuteCommand?: (text: string) => void;
  onOpenChat?: () => void;
}

export function CommandLine({
  cards,
  onApprove,
  onReject,
  onShowToast,
  moneySummary,
  onExecuteCommand,
  onOpenChat,
}: CommandLineProps) {
  const [inputVal, setInputVal] = useState("");
  const [isListening, setIsListening] = useState(false);
  const activeRecognitionRef = useRef<any>(null);

  useEffect(() => {
    return () => {
      if (activeRecognitionRef.current) {
        try {
          activeRecognitionRef.current.abort();
        } catch {}
      }
    };
  }, []);

  const normalize = (s: string) =>
    s
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "");

  const handleCommand = (rawText: string) => {
    if (onExecuteCommand) {
      onExecuteCommand(rawText);
      return;
    }

    const q = normalize(rawText.trim());
    if (!q) return;

    if (/status|resumo|decis|pendent/.test(q)) {
      onShowToast(moneySummary);
      return;
    }

    const isApprove = /aprov|aceit|libera|entreg/.test(q);
    const isReject = /recus|rejeit/.test(q);

    // Procura por cliente ou título do cartão
    const matches = cards.filter((c) => {
      if (c.stage === "entregue" || c.stage === "recusado") return false;
      const combined = normalize(`${c.client_name} ${c.title}`);
      return combined.split(/\s+/).some((w) => w.length >= 4 && q.includes(w));
    });

    if (matches.length === 0) {
      onShowToast("Não achei esse cartão. Tente o nome do cliente.");
      return;
    }

    if (matches.length > 1) {
      onShowToast(`Achei ${matches.length} cartões. Diga o nome completo do cliente.`);
      return;
    }

    const targetCard = matches[0];

    if (isReject) {
      onReject(targetCard.id);
      return;
    }

    if (isApprove) {
      onApprove(targetCard.id);
      return;
    }

    onShowToast(`Cartão encontrado: ${targetCard.title} (${targetCard.client_name})`);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputVal) return;
    handleCommand(inputVal);
    setInputVal("");
  };

  const handleMicClick = () => {
    interface SpeechEvent {
      results: { [key: number]: { [key: number]: { transcript: string } } };
    }
    const win = window as unknown as {
      SpeechRecognition?: new () => any;
      webkitSpeechRecognition?: new () => any;
    };

    const SpeechRecognition = win.SpeechRecognition || win.webkitSpeechRecognition;

    if (!SpeechRecognition) {
      onShowToast("Seu navegador não tem ditado de voz nativo. Use o teclado do celular.");
      return;
    }

    if (activeRecognitionRef.current) {
      try {
        activeRecognitionRef.current.abort();
      } catch {}
      activeRecognitionRef.current = null;
      setIsListening(false);
      return;
    }

    try {
      const recognition = new SpeechRecognition();
      recognition.lang = "pt-BR";
      recognition.interimResults = false;
      recognition.maxAlternatives = 1;
      activeRecognitionRef.current = recognition;

      let hasHandledResult = false;

      recognition.onstart = () => {
        setIsListening(true);
        onShowToast("Ouvindo sua ordem...");
      };

      recognition.onresult = (event: SpeechEvent) => {
        if (hasHandledResult) return;
        hasHandledResult = true;

        const transcript = event.results?.[0]?.[0]?.transcript;
        try {
          recognition.stop();
        } catch {}
        activeRecognitionRef.current = null;
        setIsListening(false);

        if (transcript && transcript.trim()) {
          setInputVal(transcript.trim());
          handleCommand(transcript.trim());
          setInputVal("");
        }
      };

      recognition.onerror = () => {
        activeRecognitionRef.current = null;
        setIsListening(false);
        onShowToast("Não entendi a fala. Tente novamente.");
      };

      recognition.onend = () => {
        activeRecognitionRef.current = null;
        setIsListening(false);
      };

      recognition.start();
    } catch {
      activeRecognitionRef.current = null;
      setIsListening(false);
      onShowToast("Microfone inacessível no momento.");
    }
  };

  return (
    <form
      id="command-line-form"
      onSubmit={handleSubmit}
      className="fixed left-0 right-0 z-40 bg-[var(--bg)] border-t border-[var(--line)] p-2 px-3 flex gap-2"
      style={{ bottom: "calc(54px + env(safe-area-inset-bottom, 0px))" }}
    >
      {onOpenChat && (
        <button
          type="button"
          id="btn-open-jarvis-drawer"
          onClick={onOpenChat}
          title="Abrir Chat com o Jarvis"
          className="w-9 h-9 rounded-full bg-[var(--accs)] text-[var(--acc)] text-base flex items-center justify-center border border-[var(--line)] cursor-pointer hover:scale-105 transition-transform flex-none shadow-xs"
        >
          🧠
        </button>
      )}
      <input
        id="cmd-input"
        type="text"
        placeholder="Dite uma ordem. Ex.: aprovar proposta ou construir projeto"
        value={inputVal}
        onChange={(e) => setInputVal(e.target.value)}
        className="flex-1 min-w-0 text-xs sm:text-sm py-2 px-3.5 rounded-full border border-[var(--line)] bg-[var(--surf)] text-[var(--ink)] outline-none focus:border-[var(--acc)]"
      />
      <button
        type="button"
        id="cmd-mic-btn"
        onClick={handleMicClick}
        aria-label="Falar comando de voz"
        className={`px-3 py-2 rounded-full font-semibold text-xs border-0 cursor-pointer flex-none transition-colors ${
          isListening
            ? "bg-[var(--bad)] text-white animate-pulse"
            : "bg-[var(--accs)] text-[var(--acc)] hover:opacity-90"
        }`}
      >
        {isListening ? "Gravando" : "Falar"}
      </button>
      <button
        type="submit"
        id="cmd-send-btn"
        className="px-4 py-2 rounded-full font-semibold text-xs border-0 bg-[var(--acc)] text-white cursor-pointer flex-none hover:opacity-90 transition-opacity"
      >
        Enviar
      </button>
    </form>
  );
}
