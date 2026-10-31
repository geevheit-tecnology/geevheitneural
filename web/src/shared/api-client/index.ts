import { supabase } from "@/shared/auth/supabase";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "https://geevheit-api.onrender.com";

export interface ApiError {
  code: string;
  message: string;
}

export class ApiException extends Error {
  constructor(public code: string, message: string, public status: number) {
    super(message);
    this.name = "ApiException";
  }
}

export async function apiClient<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const { data: { session } } = await supabase.auth.getSession();

  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");

  if (session?.access_token) {
    headers.set("Authorization", `Bearer ${session.access_token}`);
  }

  const url = `${API_BASE_URL}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`;

  // Timeout resiliente de 3.5s para não travar o mobile em cold sleep do Render
  const controller = new AbortController();
  const timeoutMs = 3500;
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const res = await fetch(url, {
      ...options,
      headers,
      signal: options.signal || controller.signal,
    });
    clearTimeout(timeoutId);

    if (!res.ok) {
      let errorCode = "unknown_error";
      let errorMessage = `HTTP Error ${res.status}`;

      try {
        const errBody = await res.json();
        if (errBody?.error?.code) {
          errorCode = errBody.error.code;
        }
        if (errBody?.error?.message) {
          errorMessage = errBody.error.message;
        }
      } catch {
        // Falha ao parsear JSON de erro
      }

      throw new ApiException(errorCode, errorMessage, res.status);
    }

    return res.json();
  } catch (err: unknown) {
    clearTimeout(timeoutId);
    if ((err as Error)?.name === "AbortError") {
      throw new ApiException("timeout", "A API demorou para responder. A execução não foi confirmada; confira o quadro antes de reenviar.", 408);
    }
    throw err;
  }
}
