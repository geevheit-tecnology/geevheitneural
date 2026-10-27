import { createClient } from "@supabase/supabase-js";

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || "https://tggzeapudgzswdojtndj.supabase.co";
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InRnZ3plYXB1ZGd6c3dkb2p0bmRqIiwicm9sZSI6ImFub24iLCJpYXQiOjE3OTI1Mzk3OTEsImV4cCI6MjEwODExNTc5MX0.PmxAW-rEJ2qrXwYlkmV8mXzPb0TqdBzjNtXd0JChhVo";

export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
  auth: {
    persistSession: true,
    autoRefreshToken: true,
  },
});
