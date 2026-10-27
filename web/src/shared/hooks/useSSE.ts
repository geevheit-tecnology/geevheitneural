"use client";

import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { supabase } from "@/shared/auth/supabase";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "https://geevheit-api.onrender.com";

export function useSSE(onSystemStateChange?: (paused: boolean) => void) {
  const queryClient = useQueryClient();
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    let active = true;

    async function connect() {
      const { data: { session } } = await supabase.auth.getSession();
      if (!active) return;

      // EventSource nativo do browser não envia header Authorization customizado sem polyfill ou query token.
      // Passamos o token como param ou conectamos com cookie/query se disponível.
      const token = session?.access_token;
      const url = token
        ? `${API_BASE_URL}/v1/stream?token=${encodeURIComponent(token)}`
        : `${API_BASE_URL}/v1/stream`;

      try {
        const es = new EventSource(url);
        esRef.current = es;

        es.addEventListener("money_waiting_changed", () => {
          queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
        });

        es.addEventListener("cards_changed", () => {
          queryClient.invalidateQueries({ queryKey: ["cards"] });
          queryClient.invalidateQueries({ queryKey: ["money-waiting"] });
        });

        es.addEventListener("apps_changed", () => {
          queryClient.invalidateQueries({ queryKey: ["apps"] });
          queryClient.invalidateQueries({ queryKey: ["finances"] });
        });

        es.addEventListener("finances_changed", () => {
          queryClient.invalidateQueries({ queryKey: ["finances"] });
        });

        es.addEventListener("system_state_changed", (e) => {
          try {
            const data = JSON.parse(e.data);
            if (typeof data?.paused === "boolean" && onSystemStateChange) {
              onSystemStateChange(data.paused);
            }
          } catch {
            // parsing error ignorado
          }
        });

        es.onerror = () => {
          // Reconexão automática é tratada pelo próprio EventSource
        };
      } catch {
        // Falha silenciosa de conexão SSE
      }
    }

    connect();

    return () => {
      active = false;
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
    };
  }, [queryClient, onSystemStateChange]);
}
